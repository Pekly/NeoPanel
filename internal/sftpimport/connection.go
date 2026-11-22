package sftpimport

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"net"
	"net/netip"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Injectable only inside this package; production always uses the system resolver.
var lookupIP = net.DefaultResolver.LookupNetIP

var forbidden = prefixes("0.0.0.0/8", "100.64.0.0/10", "169.254.0.0/16", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/3", "168.63.129.16/32", "100::/64", "2001::/23", "2001:db8::/32", "2002::/16", "3fff::/20", "fd00:ec2::254/128", "fd20:ce::254/128")

func prefixes(ss ...string) []netip.Prefix {
	var ps []netip.Prefix
	for _, s := range ss {
		ps = append(ps, netip.MustParsePrefix(s))
	}
	return ps
}
func permitted(ip netip.Addr, allow bool) bool {
	if !ip.IsValid() || ip.Zone() != "" {
		return false
	}
	ip = ip.Unmap()
	if ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsLinkLocalMulticast() {
		return false
	}
	for _, p := range forbidden {
		if p.Contains(ip) {
			return false
		}
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return allow
	}
	// Restrict IPv6 to currently allocated global unicast, excluding transition,
	// documentation and special-purpose ranges (including NAT64).
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	return ip.IsGlobalUnicast()
}
func pinnedAddress(ctx context.Context, host string, port int, allowed []string) (string, error) {
	allow := false
	for _, entry := range allowed {
		h, p, e := net.SplitHostPort(entry)
		if e != nil {
			continue
		}
		n, e := strconv.Atoi(p)
		if e != nil || n < 1 || n > 65535 {
			continue
		}
		h, n, e = hostPort(h, n)
		if e == nil && h == host && n == port {
			allow = true
		}
	}
	var ips []netip.Addr
	if ip, e := netip.ParseAddr(host); e == nil {
		ips = []netip.Addr{ip}
	} else {
		var err error
		ips, err = lookupIP(ctx, "ip", host)
		if err != nil {
			return "", problem("host resolution failed")
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(ips) == 0 || len(ips) > 64 {
		return "", problem("invalid host addresses")
	}
	for _, ip := range ips {
		if !permitted(ip, allow) {
			return "", problem("outbound address denied")
		}
	}
	return net.JoinHostPort(ips[0].Unmap().String(), strconv.Itoa(port)), nil
}

type idleConn struct{ net.Conn }

func (c idleConn) Read(p []byte) (int, error) {
	_ = c.SetReadDeadline(time.Now().Add(30 * time.Second))
	return c.Conn.Read(p)
}
func (c idleConn) Write(p []byte) (int, error) {
	_ = c.SetWriteDeadline(time.Now().Add(30 * time.Second))
	return c.Conn.Write(p)
}

// dialSSH resolves once, validates every answer, then dials only a literal IP.
// Context cancellation closes the transport, interrupting all SSH/SFTP calls.
func dialSSH(ctx context.Context, host string, port int, allowed []string, cfg *ssh.ClientConfig) (*ssh.Client, func(), error) {
	address, err := pinnedAddress(ctx, host, port, allowed)
	if err != nil {
		return nil, nil, err
	}
	raw, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, nil, problem("connection failed")
	}
	stop := context.AfterFunc(ctx, func() { _ = raw.Close() })
	timer := time.AfterFunc(10*time.Second, func() { _ = raw.Close() })
	conn, chans, reqs, err := ssh.NewClientConn(idleConn{raw}, net.JoinHostPort(host, strconv.Itoa(port)), cfg)
	timer.Stop()
	if err != nil {
		stop()
		_ = raw.Close()
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		return nil, nil, problem("SSH handshake or authentication failed")
	}
	client := ssh.NewClient(conn, chans, reqs)
	return client, func() { stop(); _ = raw.Close(); _ = client.Close() }, nil
}

// HostKey discovers an untrusted key for administrator verification. No SSH
// authentication request is sent: the host-key callback aborts the handshake.
func HostKey(ctx context.Context, host string, port int, allowedHosts []string) (HostKeyInfo, error) {
	var info HostKeyInfo
	host, port, err := hostPort(host, port)
	if err != nil {
		return info, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cfg := &ssh.ClientConfig{User: "host-key-probe", HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
		info = HostKeyInfo{ssh.FingerprintSHA256(key), key.Type()}
		return errors.New("host key collected")
	}}
	client, closeFn, err := dialSSH(ctx, host, port, allowedHosts, cfg)
	if client != nil {
		closeFn()
	}
	if ctx.Err() != nil {
		return HostKeyInfo{}, ctx.Err()
	}
	if info.Fingerprint != "" {
		return info, nil
	}
	return info, err
}

type remote struct {
	client *sftp.Client
	frames *frameReader
	close  func()
	root   string
}

func connect(ctx context.Context, o Options, allowed []string) (*remote, error) {
	var auth ssh.AuthMethod
	if o.PrivateKey != "" {
		s, err := signer(o)
		if err != nil {
			return nil, problem("invalid private key")
		}
		auth = ssh.PublicKeys(s)
	} else {
		auth = ssh.Password(o.Password)
	}
	var mismatch atomic.Bool
	cfg := &ssh.ClientConfig{User: o.Username, Auth: []ssh.AuthMethod{auth}, HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
		if subtle.ConstantTimeCompare([]byte(ssh.FingerprintSHA256(key)), []byte(o.Fingerprint)) != 1 {
			mismatch.Store(true)
			return problem("host key mismatch")
		}
		return nil
	}}
	client, closeFn, err := dialSSH(ctx, o.Host, o.Port, allowed, cfg)
	if err != nil {
		if mismatch.Load() {
			return nil, problem("host key mismatch")
		}
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			closeFn()
		}
	}()
	session, err := client.NewSession()
	if err != nil {
		return nil, problem("SFTP session failed")
	}
	rd, err := session.StdoutPipe()
	if err != nil {
		return nil, problem("SFTP session failed")
	}
	wr, err := session.StdinPipe()
	if err != nil {
		return nil, problem("SFTP session failed")
	}
	// Drain only bounded diagnostics; never expose server-controlled text.
	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, problem("SFTP session failed")
	}
	go func() { _, _ = io.Copy(io.Discard, io.LimitReader(stderr, 1<<20)); closeFn() }()
	if err = session.RequestSubsystem("sftp"); err != nil {
		return nil, problem("SFTP subsystem failed")
	}
	frames := &frameReader{source: rd, maxData: o.MaxBytes + 1}
	c, err := sftp.NewClientPipe(frames, wr, sftp.UseConcurrentReads(false), sftp.MaxPacket(32768))
	if err != nil {
		return nil, problem("SFTP protocol failed")
	}
	r := &remote{client: c, frames: frames, close: func() { closeFn(); _ = c.Close(); _ = session.Close() }}
	root, err := c.RealPath(o.RemotePath)
	if err != nil || !safePath(root, true) || len(root) == 0 || root[0] != '/' {
		return nil, problem("invalid remote root")
	}
	r.root = path.Clean(root)
	// Reject symlinks in every resolved root component. RealPath is called once;
	// requested path components are checked too so a configured symlink is refused.
	if err = r.checkDirs(root); err != nil {
		return nil, err
	}
	if err = r.checkRequested(o.RemotePath); err != nil {
		return nil, err
	}
	success = true
	return r, nil
}
func (r *remote) checkRequested(p string) error {
	// A relative path is interpreted by the server relative to its SFTP home.
	for p != "." && p != "/" {
		st, e := r.client.Lstat(p)
		if e != nil || st.Mode().Type() != os.ModeDir {
			return problem("remote root is not a real directory")
		}
		next := path.Dir(p)
		if next == p {
			break
		}
		p = next
	}
	return nil
}
