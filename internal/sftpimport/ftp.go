package sftpimport

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ftpEntry struct {
	name    string
	isDir   bool
	size    int64
	modTime time.Time
}

type ftpConn struct {
	conn     net.Conn
	tp       *textproto.Conn
	host     string
	pinnedIP string
	ftps     bool
	tlsCfg   *tls.Config
	stopCtx  func()
}

func dialFTP(ctx context.Context, o Options, allowed []string) (*ftpConn, error) {
	address, err := pinnedAddress(ctx, o.Host, o.Port, allowed)
	if err != nil {
		return nil, err
	}
	pinnedHost, _, _ := net.SplitHostPort(address)
	raw, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, problem("FTP connection failed")
	}
	stop := context.AfterFunc(ctx, func() { _ = raw.Close() })
	fc := &ftpConn{
		conn:     raw,
		tp:       textproto.NewConn(raw),
		host:     o.Host,
		pinnedIP: pinnedHost,
		ftps:     o.Protocol == "ftps",
		tlsCfg:   &tls.Config{ServerName: o.Host, MinVersion: tls.VersionTLS12},
		stopCtx:  func() { stop() },
	}
	ok := false
	defer func() {
		if !ok {
			fc.close()
		}
	}()

	_ = raw.SetDeadline(time.Now().Add(15 * time.Second))
	if _, _, err := fc.tp.ReadResponse(220); err != nil {
		return nil, problem("FTP greeting failed")
	}

	if fc.ftps {
		if code, _, err := fc.cmd(234, "AUTH TLS"); err != nil {
			if code == 534 || code == 500 || code == 502 {
				return nil, problem("FTPS AUTH TLS not supported by server")
			}
			return nil, problem("FTPS TLS negotiation failed")
		}
		tlsConn := tls.Client(raw, fc.tlsCfg)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return nil, problem("FTPS TLS handshake failed")
		}
		fc.conn = tlsConn
		fc.tp = textproto.NewConn(tlsConn)
		if _, _, err := fc.cmd(200, "PBSZ 0"); err != nil {
			return nil, problem("FTPS PBSZ failed")
		}
		if _, _, err := fc.cmd(200, "PROT P"); err != nil {
			return nil, problem("FTPS PROT P failed")
		}
	}

	code, _, err := fc.cmd(0, "USER %s", o.Username)
	if err != nil {
		return nil, problem("FTP USER failed")
	}
	if code == 331 {
		if _, _, err := fc.cmd(230, "PASS %s", o.Password); err != nil {
			return nil, problem("FTP authentication failed")
		}
	} else if code != 230 {
		return nil, problem("FTP authentication failed")
	}

	if _, _, err := fc.cmd(200, "TYPE I"); err != nil {
		return nil, problem("FTP TYPE I failed")
	}
	_ = fc.conn.SetDeadline(time.Time{})
	ok = true
	return fc, nil
}

func (fc *ftpConn) close() {
	if fc.stopCtx != nil {
		fc.stopCtx()
	}
	if fc.tp != nil {
		_, _, _ = fc.cmd(0, "QUIT")
		_ = fc.tp.Close()
	}
	if fc.conn != nil {
		_ = fc.conn.Close()
	}
}

func (fc *ftpConn) cmd(expect int, format string, args ...any) (int, string, error) {
	_ = fc.conn.SetDeadline(time.Now().Add(30 * time.Second))
	id, err := fc.tp.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	fc.tp.StartResponse(id)
	defer fc.tp.EndResponse(id)
	code, msg, err := fc.tp.ReadResponse(expect)
	_ = fc.conn.SetDeadline(time.Time{})
	return code, msg, err
}

func (fc *ftpConn) resolveRoot(remotePath string) (string, error) {
	if remotePath == "" || remotePath == "." {
		_, msg, err := fc.cmd(257, "PWD")
		if err != nil {
			return "/", nil
		}
		if start := strings.IndexByte(msg, '"'); start >= 0 {
			if end := strings.IndexByte(msg[start+1:], '"'); end >= 0 {
				return path.Clean("/" + msg[start+1:start+1+end]), nil
			}
		}
		return "/", nil
	}
	if _, _, err := fc.cmd(250, "CWD %s", remotePath); err != nil {
		return "", problem("remote root unavailable")
	}
	_, msg, err := fc.cmd(257, "PWD")
	if err == nil {
		if start := strings.IndexByte(msg, '"'); start >= 0 {
			if end := strings.IndexByte(msg[start+1:], '"'); end >= 0 {
				return path.Clean("/" + msg[start+1:start+1+end]), nil
			}
		}
	}
	return path.Clean("/" + remotePath), nil
}

func (fc *ftpConn) openData(ctx context.Context, command string) (net.Conn, error) {
	var port int
	if code, msg, err := fc.cmd(0, "EPSV"); err == nil && code == 229 {
		if s := strings.Index(msg, "(|||"); s >= 0 {
			if e := strings.Index(msg[s+4:], "|)"); e > 0 {
				port, _ = strconv.Atoi(msg[s+4 : s+4+e])
			}
		}
	}
	if port <= 0 {
		code, msg, err := fc.cmd(227, "PASV")
		if err != nil || code != 227 {
			return nil, problem("FTP passive mode failed")
		}
		s, e := strings.IndexByte(msg, '('), strings.LastIndexByte(msg, ')')
		if s < 0 || e <= s {
			return nil, problem("invalid PASV response")
		}
		parts := strings.Split(msg[s+1:e], ",")
		if len(parts) != 6 {
			return nil, problem("invalid PASV address")
		}
		p1, e1 := strconv.Atoi(strings.TrimSpace(parts[4]))
		p2, e2 := strconv.Atoi(strings.TrimSpace(parts[5]))
		if e1 != nil || e2 != nil {
			return nil, problem("invalid PASV port")
		}
		port = p1*256 + p2
	}
	if port < 1 || port > 65535 {
		return nil, problem("invalid FTP data port")
	}

	// Pin data connection to the validated control IP to prevent PASV SSRF redirects.
	dataAddr := net.JoinHostPort(fc.pinnedIP, strconv.Itoa(port))
	dc, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", dataAddr)
	if err != nil {
		return nil, problem("FTP data connection failed")
	}

	_ = fc.conn.SetDeadline(time.Now().Add(30 * time.Second))
	id, err := fc.tp.Cmd("%s", command)
	if err != nil {
		_ = dc.Close()
		return nil, err
	}
	fc.tp.StartResponse(id)
	code, _, err := fc.tp.ReadResponse(0)
	fc.tp.EndResponse(id)
	_ = fc.conn.SetDeadline(time.Time{})
	if err != nil || (code != 125 && code != 150) {
		_ = dc.Close()
		return nil, problem(fmt.Sprintf("FTP command %s rejected (code %d)", strings.SplitN(command, " ", 2)[0], code))
	}

	if fc.ftps {
		tlsData := tls.Client(dc, fc.tlsCfg)
		if err := tlsData.HandshakeContext(ctx); err != nil {
			_ = dc.Close()
			return nil, problem("FTPS data TLS handshake failed")
		}
		return tlsData, nil
	}
	return dc, nil
}

func (fc *ftpConn) finishData() error {
	_ = fc.conn.SetDeadline(time.Now().Add(30 * time.Second))
	code, _, err := fc.tp.ReadResponse(0)
	_ = fc.conn.SetDeadline(time.Time{})
	if err != nil || (code != 226 && code != 250) {
		return problem("FTP transfer completion failed")
	}
	return nil
}

func (fc *ftpConn) listDir(ctx context.Context, dir string) ([]ftpEntry, error) {
	// Try MLSD first, fallback to LIST
	dc, err := fc.openData(ctx, "MLSD "+dir)
	mlsd := err == nil
	if !mlsd {
		dc, err = fc.openData(ctx, "LIST "+dir)
		if err != nil {
			return nil, problem("FTP directory listing failed")
		}
	}
	var lines []string
	sc := bufio.NewScanner(io.LimitReader(dc, 16<<20))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	_ = dc.Close()
	if err := fc.finishData(); err != nil {
		return nil, err
	}
	if err := sc.Err(); err != nil {
		return nil, problem("FTP listing read failed")
	}

	var entries []ftpEntry
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ent ftpEntry
		var ok bool
		if mlsd {
			ent, ok = parseMLSDLine(line)
		} else {
			ent, ok = parseLISTLine(line)
		}
		if !ok || ent.name == "." || ent.name == ".." {
			continue
		}
		entries = append(entries, ent)
	}
	return entries, nil
}

func parseMLSDLine(line string) (ftpEntry, bool) {
	idx := strings.IndexByte(line, ' ')
	if idx <= 0 {
		return ftpEntry{}, false
	}
	factsStr, name := line[:idx], strings.TrimRight(line[idx+1:], "\r\n")
	if name == "" {
		return ftpEntry{}, false
	}
	var ent ftpEntry
	ent.name = name
	for _, kv := range strings.Split(factsStr, ";") {
		if kv == "" {
			continue
		}
		p := strings.SplitN(kv, "=", 2)
		if len(p) != 2 {
			continue
		}
		k, v := strings.ToLower(p[0]), p[1]
		switch k {
		case "type":
			tl := strings.ToLower(v)
			if tl == "cdir" || tl == "pdir" || strings.Contains(tl, "symlink") || strings.Contains(tl, "link") {
				return ftpEntry{}, false
			}
			ent.isDir = tl == "dir"
		case "size":
			ent.size, _ = strconv.ParseInt(v, 10, 64)
		case "modify":
			if len(v) >= 14 {
				if t, err := time.Parse("20060102150405", v[:14]); err == nil {
					ent.modTime = t
				}
			}
		}
	}
	return ent, true
}

func parseLISTLine(line string) (ftpEntry, bool) {
	line = strings.TrimRight(line, "\r\n")
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return ftpEntry{}, false
	}
	// Unix-style: drwxr-xr-x 2 user group 4096 Nov 25 12:00 filename
	perm := fields[0]
	if len(perm) >= 10 && (perm[0] == 'd' || perm[0] == '-' || perm[0] == 'l') {
		if perm[0] == 'l' {
			return ftpEntry{}, false // reject symlinks
		}
		if len(fields) < 9 {
			return ftpEntry{}, false
		}
		size, _ := strconv.ParseInt(fields[4], 10, 64)
		// Find the start of the filename after field 7 (time/year)
		pos := 0
		for i := 0; i < 8; i++ {
			idx := strings.Index(line[pos:], fields[i])
			if idx < 0 {
				return ftpEntry{}, false
			}
			pos += idx + len(fields[i])
		}
		name := strings.TrimSpace(line[pos:])
		if name == "" {
			return ftpEntry{}, false
		}
		return ftpEntry{
			name:  name,
			isDir: perm[0] == 'd',
			size:  size,
		}, true
	}
	// Windows IIS style: 11-25-26 12:00PM <DIR> foldername
	if len(fields) >= 4 {
		isDir := strings.EqualFold(fields[2], "<DIR>")
		var size int64
		if !isDir {
			size, _ = strconv.ParseInt(fields[2], 10, 64)
		}
		idx := strings.Index(line, fields[2])
		if idx < 0 {
			return ftpEntry{}, false
		}
		name := strings.TrimSpace(line[idx+len(fields[2]):])
		if name == "" {
			return ftpEntry{}, false
		}
		return ftpEntry{name: name, isDir: isDir, size: size}, true
	}
	return ftpEntry{}, false
}

func probeFTP(ctx context.Context, o Options, allowedHosts []string) (preview Preview, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.TimeoutMinutes)*time.Minute)
	defer cancel()
	fc, err := dialFTP(ctx, o, allowedHosts)
	if err != nil {
		return preview, err
	}
	defer fc.close()

	root, err := fc.resolveRoot(o.RemotePath)
	if err != nil {
		return preview, err
	}
	preview.RemotePath = root
	preview.Entries = []Entry{}

	entries, err := fc.listDir(ctx, root)
	if err != nil {
		return preview, err
	}
	seen := map[string]bool{}
	for _, ent := range entries {
		if !safeComponent(ent.name) {
			return Preview{}, problem("unsafe remote entry")
		}
		key := collisionKey(ent.name)
		if seen[key] {
			return Preview{}, problem("case-fold path collision")
		}
		seen[key] = true
		if !selected(o, ent.name, ent.isDir) {
			continue
		}
		if len(preview.Entries) == 200 {
			preview.Truncated = true
			continue
		}
		preview.Entries = append(preview.Entries, Entry{ent.name, ent.name, ent.isDir, ent.size})
	}
	return preview, nil
}

type ftpPlanItem struct {
	rel   string
	entry ftpEntry
}

func transferFTP(ctx context.Context, o Options, allowedHosts []string, destination string, progress func(Progress)) (err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.TimeoutMinutes)*time.Minute)
	defer cancel()
	local, err := destinationRoot(destination)
	if err != nil {
		return err
	}
	defer local.Close()

	fc, err := dialFTP(ctx, o, allowedHosts)
	if err != nil {
		return err
	}
	defer fc.close()

	root, err := fc.resolveRoot(o.RemotePath)
	if err != nil {
		return err
	}

	p := Progress{State: "scanning"}
	emit := func() {
		if progress != nil {
			progress(p)
		}
	}
	emit()

	var plan []ftpPlanItem
	seen := map[string]bool{}
	count := 0
	var scan func(string, int) error
	scan = func(rel string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > 64 {
			return problem("remote depth limit exceeded")
		}
		full := path.Join(root, rel)
		entries, e := fc.listDir(ctx, full)
		if e != nil {
			return e
		}
		for _, ent := range entries {
			count++
			if count > maxEntries {
				return problem("remote entry limit exceeded")
			}
			child := path.Join(rel, ent.name)
			if !safeComponent(ent.name) || !safePath(child, false) || len(path.Join(root, child)) > 4096 {
				return problem("unsafe remote path")
			}
			key := collisionKey(child)
			if seen[key] {
				return problem("case-fold path collision")
			}
			seen[key] = true
			if !selected(o, child, ent.isDir) {
				continue
			}
			plan = append(plan, ftpPlanItem{child, ent})
			if !ent.isDir {
				if ent.size > o.MaxBytes-p.BytesTotal {
					return problem("remote byte limit exceeded")
				}
				p.FilesTotal++
				p.BytesTotal += ent.size
			}
			p.CurrentPath = child
			emit()
			if ent.isDir {
				if e = scan(child, depth+1); e != nil {
					return e
				}
			}
		}
		return nil
	}

	if err = scan("", 0); err != nil {
		return err
	}

	p.State = "transferring"
	p.CurrentPath = ""
	emit()

	start := time.Now()
	var lastSpeedBytes int64
	lastSpeedTick := start
	updateSpeed := func() {
		now := time.Now()
		elapsed := now.Sub(lastSpeedTick)
		if elapsed >= 250*time.Millisecond {
			delta := p.BytesDone - lastSpeedBytes
			p.SpeedBPS = int64(float64(delta) / elapsed.Seconds())
			lastSpeedBytes = p.BytesDone
			lastSpeedTick = now
		} else if p.SpeedBPS == 0 {
			tot := now.Sub(start).Seconds()
			if tot > 0.05 {
				p.SpeedBPS = int64(float64(p.BytesDone) / tot)
			}
		}
	}

	buf := make([]byte, 32768)
	for _, it := range plan {
		if err = ctx.Err(); err != nil {
			return err
		}
		p.CurrentPath = it.rel
		emit()

		if it.entry.isDir {
			if e := local.Mkdir(filepath.FromSlash(it.rel), 0755); e != nil {
				return problem("cannot create staging directory")
			}
			continue
		}

		full := path.Join(root, it.rel)
		dc, err := fc.openData(ctx, "RETR "+full)
		if err != nil {
			return err
		}
		target, err := local.OpenFile(filepath.FromSlash(it.rel), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			_ = dc.Close()
			_ = fc.finishData()
			return problem("cannot create staging file")
		}

		for {
			if err = ctx.Err(); err != nil {
				_ = target.Close()
				_ = dc.Close()
				return err
			}
			n, readErr := dc.Read(buf)
			if n > 0 {
				if int64(n) > o.MaxBytes-p.BytesDone {
					_ = target.Close()
					_ = dc.Close()
					return problem("remote byte limit exceeded")
				}
				if _, wErr := target.Write(buf[:n]); wErr != nil {
					_ = target.Close()
					_ = dc.Close()
					return problem("cannot write staging file")
				}
				p.BytesDone += int64(n)
				updateSpeed()
				emit()
				if o.BandwidthKiB > 0 {
					delay := time.Duration(float64(p.BytesDone)/float64(int64(o.BandwidthKiB)*1024)*float64(time.Second)) - time.Since(start)
					if delay > 0 {
						timer := time.NewTimer(delay)
						select {
						case <-ctx.Done():
							timer.Stop()
							_ = target.Close()
							_ = dc.Close()
							return ctx.Err()
						case <-timer.C:
						}
					}
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				_ = target.Close()
				_ = dc.Close()
				return problem("FTP file read failed")
			}
		}
		_ = target.Close()
		_ = dc.Close()
		if err := fc.finishData(); err != nil {
			return err
		}
		if o.PreserveTimes && !it.entry.modTime.IsZero() {
			_ = local.Chtimes(filepath.FromSlash(it.rel), it.entry.modTime, it.entry.modTime)
		}
		p.FilesDone++
		emit()
	}
	return nil
}
