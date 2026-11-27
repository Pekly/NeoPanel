package sftpimport

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type testInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (i testInfo) Name() string       { return i.name }
func (i testInfo) Size() int64        { return i.size }
func (i testInfo) Mode() os.FileMode  { return i.mode }
func (i testInfo) ModTime() time.Time { return time.Unix(1700000000, 0) }
func (i testInfo) IsDir() bool        { return i.mode.IsDir() }
func (i testInfo) Sys() any           { return nil }

type infoList []os.FileInfo

func (l infoList) ListAt(dst []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF
	}
	n := copy(dst, l[offset:])
	if n < len(dst) {
		return n, io.EOF
	}
	return n, nil
}

type testFS struct {
	nodes     map[string]testInfo
	data      map[string]string
	lists     atomic.Int32
	writes    atomic.Int32
	realpaths atomic.Int32
	stall     <-chan struct{}
	reading   chan struct{}
	once      sync.Once
	changed   atomic.Bool
	listError bool
}

func newFS() *testFS {
	return &testFS{nodes: map[string]testInfo{"/": {name: "/", mode: os.ModeDir | 0755}}, data: map[string]string{}}
}
func (f *testFS) add(p, data string, mode os.FileMode) {
	f.nodes[p] = testInfo{path.Base(p), int64(len(data)), mode}
	f.data[p] = data
}
func (f *testFS) RealPath(p string) (string, error) {
	f.realpaths.Add(1)
	return path.Join("/", p), nil
}
func (f *testFS) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	if r.Method == "List" {
		f.lists.Add(1)
		if f.listError {
			return nil, errors.New("SECRET remote uncontrolled diagnostic")
		}
		var l infoList
		for p, st := range f.nodes {
			if p != "/" && path.Dir(p) == r.Filepath {
				l = append(l, st)
			}
		}
		sort.Slice(l, func(i, j int) bool { return l[i].Name() < l[j].Name() })
		return l, nil
	}
	st, ok := f.nodes[r.Filepath]
	if !ok {
		return nil, os.ErrNotExist
	}
	if f.changed.Load() && !st.IsDir() {
		st.mode = os.ModeSymlink | 0777
	}
	return infoList{st}, nil
}
func (f *testFS) Lstat(r *sftp.Request) (sftp.ListerAt, error) { return f.Filelist(r) }

type testReader struct {
	*strings.Reader
	fs *testFS
}

func (r *testReader) ReadAt(p []byte, off int64) (int, error) {
	if r.fs.reading != nil {
		r.fs.once.Do(func() { close(r.fs.reading) })
	}
	if r.fs.stall != nil {
		<-r.fs.stall
	}
	return r.Reader.ReadAt(p, off)
}
func (f *testFS) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	return &testReader{strings.NewReader(f.data[r.Filepath]), f}, nil
}
func (f *testFS) Filewrite(*sftp.Request) (io.WriterAt, error) {
	f.writes.Add(1)
	return nil, os.ErrPermission
}
func (f *testFS) Filecmd(*sftp.Request) error { f.writes.Add(1); return os.ErrPermission }

type testServer struct {
	options      Options
	allowed      []string
	auth         atomic.Int32
	privateKey   string
	encryptedKey string
}

func serve(t *testing.T, f *testFS, raw func(ssh.Channel)) *testServer {
	t.Helper()
	_, hostPrivate, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	hostSigner, e := ssh.NewSignerFromKey(hostPrivate)
	if e != nil {
		t.Fatal(e)
	}
	_, userPrivate, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	userSigner, e := ssh.NewSignerFromKey(userPrivate)
	if e != nil {
		t.Fatal(e)
	}
	der, e := x509.MarshalPKCS8PrivateKey(userPrivate)
	if e != nil {
		t.Fatal(e)
	}
	encrypted, e := ssh.MarshalPrivateKeyWithPassphrase(userPrivate, "test", []byte("phrase"))
	if e != nil {
		t.Fatal(e)
	}
	s := &testServer{privateKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), encryptedKey: string(pem.EncodeToMemory(encrypted))}
	cfg := &ssh.ServerConfig{PasswordCallback: func(_ ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
		s.auth.Add(1)
		if string(p) != "SECRET" {
			return nil, errors.New("bad password")
		}
		return nil, nil
	}, PublicKeyCallback: func(_ ssh.ConnMetadata, k ssh.PublicKey) (*ssh.Permissions, error) {
		s.auth.Add(1)
		if !bytes.Equal(k.Marshal(), userSigner.PublicKey().Marshal()) {
			return nil, errors.New("bad key")
		}
		return nil, nil
	}}
	cfg.AddHostKey(hostSigner)
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	_, p, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(p)
	s.options = Options{Host: "127.0.0.1", Port: port, Username: "import", Password: "SECRET", Fingerprint: ssh.FingerprintSHA256(hostSigner.PublicKey()), RemotePath: "/"}
	s.allowed = []string{listener.Addr().String()}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var conns []net.Conn
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, e := listener.Accept()
			if e != nil {
				return
			}
			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer c.Close()
				conn, channels, reqs, e := ssh.NewServerConn(c, cfg)
				if e != nil {
					return
				}
				defer conn.Close()
				go ssh.DiscardRequests(reqs)
				for incoming := range channels {
					if incoming.ChannelType() != "session" {
						_ = incoming.Reject(ssh.UnknownChannelType, "denied")
						continue
					}
					ch, requests, e := incoming.Accept()
					if e != nil {
						return
					}
					for req := range requests {
						var sub struct{ Name string }
						if req.Type != "subsystem" || ssh.Unmarshal(req.Payload, &sub) != nil || sub.Name != "sftp" {
							_ = req.Reply(false, nil)
							continue
						}
						_ = req.Reply(true, nil)
						if raw != nil {
							raw(ch)
						} else {
							server := sftp.NewRequestServer(ch, sftp.Handlers{FileGet: f, FilePut: f, FileCmd: f, FileList: f})
							_ = server.Serve()
							_ = server.Close()
						}
						_ = ch.Close()
						break
					}
				}
			}()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		<-done
		mu.Lock()
		for _, c := range conns {
			_ = c.Close()
		}
		mu.Unlock()
		wg.Wait()
	})
	return s
}
func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestPasswordTransferFiltersAndProgress(t *testing.T) {
	f := newFS()
	f.add("/dir", "", os.ModeDir|0777)
	f.add("/dir/keep", "content", 0777)
	f.add("/dir/drop", "omit", 0644)
	f.add("/other", "not selected", 0644)
	f.add("/dir/.hidden", "secret", 0644)
	s := serve(t, f, nil)
	o := s.options
	o.IncludePaths = []string{"dir"}
	o.ExcludePaths = []string{"dir/drop"}
	o.SkipHidden = true
	o.PreserveTimes = true
	dest := t.TempDir()
	var final Progress
	scanning := false
	err := Transfer(testContext(t), o, s.allowed, dest, func(p Progress) {
		if p.State == "scanning" {
			scanning = true
		}
		final = p
	})
	if err != nil {
		t.Fatal(err)
	}
	b, e := os.ReadFile(filepath.Join(dest, "dir", "keep"))
	if e != nil || string(b) != "content" {
		t.Fatalf("content %q %v", b, e)
	}
	for _, p := range []string{"dir/drop", "other", "dir/.hidden"} {
		if _, e = os.Stat(filepath.Join(dest, filepath.FromSlash(p))); !os.IsNotExist(e) {
			t.Fatalf("unexpected %s", p)
		}
	}
	if !scanning || final.State != "transferring" || final.FilesTotal != 1 || final.FilesDone != 1 || final.BytesDone != 7 || final.BytesTotal != 7 {
		t.Fatalf("progress %+v", final)
	}
	st, e := os.Stat(filepath.Join(dest, "dir", "keep"))
	if e != nil {
		t.Fatal(e)
	}
	if st.ModTime().Unix() != 1700000000 {
		t.Fatal("mtime not preserved")
	}
	if runtime.GOOS != "windows" && st.Mode().Perm() != 0644 {
		t.Fatal("wrong file mode")
	}
	if f.writes.Load() != 0 || f.realpaths.Load() != 1 {
		t.Fatalf("writes=%d realpaths=%d", f.writes.Load(), f.realpaths.Load())
	}
}
func TestSelectedFileTraversesAncestors(t *testing.T) {
	f := newFS()
	f.add("/parent", "", os.ModeDir|0755)
	f.add("/parent/child", "yes", 0644)
	f.add("/parent/no", "no", 0644)
	s := serve(t, f, nil)
	o := s.options
	o.IncludePaths = []string{"parent/child"}
	dest := t.TempDir()
	if e := Transfer(testContext(t), o, s.allowed, dest, nil); e != nil {
		t.Fatal(e)
	}
	entries, e := os.ReadDir(filepath.Join(dest, "parent"))
	if e != nil || len(entries) != 1 || entries[0].Name() != "child" {
		t.Fatalf("selection %v %v", entries, e)
	}
}
func TestHostKeyAndFingerprintBeforeAuth(t *testing.T) {
	s := serve(t, newFS(), nil)
	info, e := HostKey(testContext(t), s.options.Host, s.options.Port, s.allowed)
	if e != nil || info.Fingerprint != s.options.Fingerprint || info.Algorithm != "ssh-ed25519" {
		t.Fatalf("host key %+v %v", info, e)
	}
	if s.auth.Load() != 0 {
		t.Fatal("discovery authenticated")
	}
	o := s.options
	o.Fingerprint = "SHA256:" + strings.Repeat("A", 43)
	_, e = Probe(testContext(t), o, s.allowed)
	if e == nil || !strings.Contains(e.Error(), "mismatch") {
		t.Fatalf("mismatch %v", e)
	}
	if s.auth.Load() != 0 {
		t.Fatal("credentials sent before pin verification")
	}
}
func TestPrivateKeyAuthentication(t *testing.T) {
	s := serve(t, newFS(), nil)
	for _, encrypted := range []bool{false, true} {
		o := s.options
		o.Password = ""
		o.PrivateKey = s.privateKey
		if encrypted {
			o.PrivateKey = s.encryptedKey
			o.Passphrase = "phrase"
		}
		if _, e := Probe(testContext(t), o, s.allowed); e != nil {
			t.Fatal(e)
		}
	}
	if s.auth.Load() == 0 {
		t.Fatal("key was not authenticated")
	}
}
func TestProbeCappedAndNonRecursive(t *testing.T) {
	f := newFS()
	for i := 0; i < 205; i++ {
		f.add(fmt.Sprintf("/file%03d", i), "x", 0644)
	}
	f.add("/dir", "", os.ModeDir|0755)
	f.add("/dir/inner", "x", 0644)
	s := serve(t, f, nil)
	p, e := Probe(testContext(t), s.options, s.allowed)
	if e != nil {
		t.Fatal(e)
	}
	if !p.Truncated || len(p.Entries) != 200 || p.RemotePath != "/" || f.lists.Load() != 1 {
		t.Fatalf("preview %+v lists %d", p, f.lists.Load())
	}
}
func TestUnsafeRemoteEntries(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode os.FileMode
	}{{"../escape", 0644}, {"nested/escape", 0644}, {"back\\slash", 0644}, {"NUL.txt", 0644}, {"trail.", 0644}, {"colon:stream", 0644}, {"link", os.ModeSymlink | 0777}, {"pipe", os.ModeNamedPipe | 0600}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFS()
			f.add("/entry", "bad", tc.mode)
			st := f.nodes["/entry"]
			st.name = tc.name
			f.nodes["/entry"] = st
			s := serve(t, f, nil)
			dest := t.TempDir()
			if e := Transfer(testContext(t), s.options, s.allowed, dest, nil); e == nil {
				t.Fatal("unsafe entry accepted")
			}
			entries, _ := os.ReadDir(dest)
			if len(entries) != 0 {
				t.Fatal("unsafe entry written")
			}
		})
	}
}
func TestCaseFoldCollision(t *testing.T) {
	f := newFS()
	f.add("/a", "a", 0644)
	f.add("/A", "b", 0644)
	s := serve(t, f, nil)
	err := Transfer(testContext(t), s.options, s.allowed, t.TempDir(), nil)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		if err == nil || !strings.Contains(err.Error(), "collision") {
			t.Fatalf("collision %v", err)
		}
	} else if err != nil {
		t.Fatalf("case-sensitive OS should allow distinct-case files: %v", err)
	}
	if foldKey("s") != foldKey("ſ") {
		t.Fatal("incomplete Unicode folding")
	}
}
func TestMetadataAndActualByteLimits(t *testing.T) {
	for _, lie := range []bool{false, true} {
		t.Run(fmt.Sprint(lie), func(t *testing.T) {
			f := newFS()
			f.add("/file", "too much", 0644)
			if lie {
				st := f.nodes["/file"]
				st.size = 2
				f.nodes["/file"] = st
			}
			s := serve(t, f, nil)
			o := s.options
			o.MaxBytes = 2
			dest := t.TempDir()
			e := Transfer(testContext(t), o, s.allowed, dest, nil)
			if e == nil {
				t.Fatal("byte limit accepted")
			}
			st, _ := os.Stat(filepath.Join(dest, "file"))
			if st != nil && st.Size() > 2 {
				t.Fatal("over-budget bytes written")
			}
		})
	}
}
func TestSymlinkRecheck(t *testing.T) {
	f := newFS()
	f.add("/file", "x", 0644)
	s := serve(t, f, nil)
	e := Transfer(testContext(t), s.options, s.allowed, t.TempDir(), func(p Progress) {
		if p.State == "transferring" {
			f.changed.Store(true)
		}
	})
	if e == nil {
		t.Fatal("post-scan symlink change accepted")
	}
}
func TestCancellationStalledRead(t *testing.T) {
	release := make(chan struct{})
	f := newFS()
	f.add("/file", "content", 0644)
	f.stall = release
	f.reading = make(chan struct{})
	s := serve(t, f, nil)
	defer close(release)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	dest := t.TempDir()
	go func() { done <- Transfer(ctx, s.options, s.allowed, dest, nil) }()
	select {
	case <-f.reading:
	case <-time.After(3 * time.Second):
		t.Fatal("read did not start")
	}
	cancel()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatalf("cancel %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("stalled read ignored cancellation")
	}
}
func TestCancellationBandwidthSleep(t *testing.T) {
	f := newFS()
	f.add("/file", strings.Repeat("x", 32768), 0644)
	s := serve(t, f, nil)
	o := s.options
	o.BandwidthKiB = 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	start := time.Now()
	e := Transfer(ctx, o, s.allowed, t.TempDir(), func(p Progress) {
		if p.BytesDone > 0 {
			cancel()
		}
	})
	if !errors.Is(e, context.Canceled) || time.Since(start) > 2*time.Second {
		t.Fatalf("sleep cancellation %v", e)
	}
}
func TestDestinationPolicy(t *testing.T) {
	s := serve(t, newFS(), nil)
	dir := t.TempDir()
	file := filepath.Join(dir, "exists")
	if e := os.WriteFile(file, []byte("original"), 0600); e != nil {
		t.Fatal(e)
	}
	for _, dest := range []string{dir, file, filepath.Join(dir, "missing")} {
		if e := Transfer(testContext(t), s.options, s.allowed, dest, nil); e == nil {
			t.Fatalf("destination accepted %s", dest)
		}
	}
	b, _ := os.ReadFile(file)
	if string(b) != "original" || s.auth.Load() != 0 {
		t.Fatal("destination validation was not before auth or overwrote")
	}
	if runtime.GOOS != "windows" {
		public := t.TempDir()
		if e := os.Chmod(public, 0755); e != nil {
			t.Fatal(e)
		}
		if _, e := destinationRoot(public); e == nil {
			t.Fatal("nonprivate destination accepted")
		}
	}
}
func TestOutboundPolicyAndDNSPinning(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "::1", "fc00::1", "::ffff:10.0.0.1"} {
		a := netip.MustParseAddr(ip)
		if permitted(a, false) || !permitted(a, true) {
			t.Fatalf("private policy %s", ip)
		}
	}
	for _, ip := range []string{"0.0.0.0", "169.254.169.254", "100.100.100.200", "224.0.0.1", "192.0.2.1", "198.51.100.2", "203.0.113.1", "168.63.129.16", "fe80::1", "ff02::1", "::", "::ffff:169.254.169.254", "2001:db8::1", "64:ff9b::a00:1"} {
		if permitted(netip.MustParseAddr(ip), true) || permitted(netip.MustParseAddr(ip), false) {
			t.Fatalf("special policy %s", ip)
		}
	}
	if !permitted(netip.MustParseAddr("8.8.8.8"), false) || !permitted(netip.MustParseAddr("2606:4700:4700::1111"), false) {
		t.Fatal("public address denied")
	}
	old := lookupIP
	defer func() { lookupIP = old }()
	lookupIP = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")}, nil
	}
	if _, e := pinnedAddress(testContext(t), "mixed.test", 22, nil); e == nil {
		t.Fatal("not all DNS results validated")
	}
	s := serve(t, newFS(), nil)
	var calls atomic.Int32
	lookupIP = func(context.Context, string, string) ([]netip.Addr, error) {
		if calls.Add(1) > 1 {
			return []netip.Addr{netip.MustParseAddr("169.254.169.254")}, nil
		}
		return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
	}
	o := s.options
	o.Host = "rebind.test"
	allowed := []string{net.JoinHostPort(o.Host, strconv.Itoa(o.Port))}
	if _, e := Probe(testContext(t), o, allowed); e != nil {
		t.Fatal(e)
	}
	if calls.Load() != 1 {
		t.Fatal("DNS resolved again during dial")
	}
}
func TestDNSCancellation(t *testing.T) {
	old := lookupIP
	defer func() { lookupIP = old }()
	lookupIP = func(ctx context.Context, _, _ string) ([]netip.Addr, error) { <-ctx.Done(); return nil, ctx.Err() }
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, e := HostKey(ctx, "stall.test", 22, nil)
	if !errors.Is(e, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatalf("DNS cancellation %v", e)
	}
}
func TestMalformedPacketBeforeAllocation(t *testing.T) {
	s := serve(t, newFS(), func(ch ssh.Channel) {
		var init [9]byte
		_, _ = io.ReadFull(ch, init[:])
		_, _ = ch.Write([]byte{0x7f, 0xff, 0xff, 0xff})
		_, _ = io.Copy(io.Discard, ch)
	})
	start := time.Now()
	if _, e := Probe(testContext(t), s.options, s.allowed); e == nil {
		t.Fatal("oversize packet accepted")
	}
	if time.Since(start) > time.Second {
		t.Fatal("waited for oversized payload")
	}
}
func TestFrameGrammarAndBudgets(t *testing.T) {
	bodies := [][]byte{{104, 0, 0, 0, 1, 0xff, 0xff, 0xff, 0xff}, {105, 0, 0, 0, 1, 0x80, 0, 0, 0, 0xff, 0xff, 0xff, 0xff}, {103, 0, 0, 0, 1, 0xff, 0xff, 0xff, 0xff}, {102, 0, 0, 0, 1, 0, 0, 0, 8, 'x'}}
	for _, b := range bodies {
		if _, _, ok := validResponse(b, true); ok {
			t.Fatalf("malformed accepted %x", b)
		}
	}
	body := []byte{2, 0, 0, 0, 3}
	packet := make([]byte, 4)
	binary.BigEndian.PutUint32(packet, uint32(len(body)))
	packet = append(packet, body...)
	r := frameReader{source: bytes.NewReader(packet), metadata: maxMetadata, maxData: 10}
	if _, e := r.Read(make([]byte, 100)); e == nil {
		t.Fatal("metadata limit accepted")
	}
	r = frameReader{source: bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff}), maxData: 10}
	if _, e := r.Read(make([]byte, 1)); e == nil {
		t.Fatal("length limit accepted")
	}
}
func TestErrorsAreSanitized(t *testing.T) {
	f := newFS()
	f.listError = true
	s := serve(t, f, nil)
	_, e := Probe(testContext(t), s.options, s.allowed)
	if e == nil || strings.Contains(e.Error(), "SECRET") || strings.Contains(e.Error(), "uncontrolled") {
		t.Fatalf("unsanitized %v", e)
	}
}
func TestValidate(t *testing.T) {
	base := Options{Host: "Example.COM", Username: "user", Password: "password", Fingerprint: "SHA256:" + strings.Repeat("A", 43)}
	o, e := Validate(base)
	if e != nil || o.Host != "example.com" || o.Port != 22 || o.RemotePath != "." || o.MaxBytes != 10<<30 || o.TimeoutMinutes != 60 {
		t.Fatalf("defaults %+v %v", o, e)
	}
	changes := []func(*Options){func(o *Options) { o.Host = "https://example.com" }, func(o *Options) { o.Host = "fe80::1%eth0" }, func(o *Options) { o.Port = 65536 }, func(o *Options) { o.Username = "" }, func(o *Options) { o.Password = "" }, func(o *Options) { o.PrivateKey = "key" }, func(o *Options) { o.Passphrase = "phrase" }, func(o *Options) { o.Fingerprint = "SHA256:bad" }, func(o *Options) { o.MaxBytes = -1 }, func(o *Options) { o.MaxBytes = 101 << 30 }, func(o *Options) { o.TimeoutMinutes = 241 }, func(o *Options) { o.BandwidthKiB = 1048577 }, func(o *Options) { o.PrivateKey = strings.Repeat("x", 65537); o.Password = "" }, func(o *Options) { o.IncludePaths = make([]string, 129) }}
	for i, change := range changes {
		v := base
		change(&v)
		if _, e := Validate(v); e == nil {
			t.Fatalf("invalid %d accepted", i)
		}
	}
	for _, p := range []string{"../bad", "a/../b", "a\\b", "C:/a", "/ok/../bad", "foo.", "NUL.txt", "a/COM1", "a//b", "a/ ", strings.Repeat("a", 4097)} {
		v := base
		v.RemotePath = p
		if _, e := Validate(v); e == nil {
			t.Fatalf("path accepted %q", p)
		}
	}
	for _, p := range []string{"/", ".", "/srv/world", "world/region"} {
		v := base
		v.RemotePath = p
		if _, e := Validate(v); e != nil {
			t.Fatalf("valid path %q: %v", p, e)
		}
	}
}

func TestMultiChunkTransferAndShortFile(t *testing.T) {
	for _, short := range []bool{false, true} {
		t.Run(fmt.Sprint(short), func(t *testing.T) {
			f := newFS()
			content := strings.Repeat("payload", 20000)
			f.add("/file", content, 0644)
			if short {
				f.data["/file"] = "short"
			}
			s := serve(t, f, nil)
			o := s.options
			o.MaxBytes = int64(len(content))
			dest := t.TempDir()
			e := Transfer(testContext(t), o, s.allowed, dest, nil)
			if short {
				if e == nil {
					t.Fatal("shrunk source accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			b, e := os.ReadFile(filepath.Join(dest, "file"))
			if e != nil || string(b) != content {
				t.Fatal("multi-chunk content mismatch")
			}
		})
	}
}
func TestExcludedSubtreeNotListed(t *testing.T) {
	f := newFS()
	f.add("/skip", "", os.ModeDir|0755)
	f.add("/skip/link", "", os.ModeSymlink|0777)
	f.add("/keep", "x", 0644)
	s := serve(t, f, nil)
	o := s.options
	o.ExcludePaths = []string{"skip"}
	if e := Transfer(testContext(t), o, s.allowed, t.TempDir(), nil); e != nil {
		t.Fatal(e)
	}
	if f.lists.Load() != 1 {
		t.Fatal("excluded subtree traversed")
	}
}
func TestDepthLimit(t *testing.T) {
	f := newFS()
	p := ""
	for i := 0; i < 65; i++ {
		p += "/d"
		f.add(p, "", os.ModeDir|0755)
	}
	s := serve(t, f, nil)
	e := Transfer(testContext(t), s.options, s.allowed, t.TempDir(), nil)
	if e == nil || !strings.Contains(e.Error(), "depth") {
		t.Fatalf("depth limit: %v", e)
	}
}
func TestRootSymlinkAndDestinationSymlink(t *testing.T) {
	f := newFS()
	f.add("/link", "", os.ModeSymlink|0777)
	s := serve(t, f, nil)
	o := s.options
	o.RemotePath = "link"
	if _, e := Probe(testContext(t), o, s.allowed); e == nil {
		t.Fatal("symlink root accepted")
	}
	dir := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if e := os.Symlink(dir, link); e == nil {
		if _, e := destinationRoot(link); e == nil {
			t.Fatal("destination symlink accepted")
		}
	} else {
		t.Log("OS does not grant local symlink creation; remote symlink checks still tested")
	}
}
func TestHandshakeCancellation(t *testing.T) {
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		c, e := listener.Accept()
		if e == nil {
			defer c.Close()
			_, _ = io.Copy(io.Discard, c)
		}
	}()
	_, p, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(p)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, e = HostKey(ctx, "127.0.0.1", port, []string{listener.Addr().String()})
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatalf("handshake cancellation %v", e)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handshake socket not closed")
	}
}
func framed(b []byte) []byte {
	p := make([]byte, 4)
	binary.BigEndian.PutUint32(p, uint32(len(b)))
	return append(p, b...)
}
func TestFrameEntryAndDataBudgets(t *testing.T) {
	// A single minimal NAME entry pushed beyond the cumulative count cap.
	body := []byte{104, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 'x', 0, 0, 0, 0, 0, 0, 0, 0}
	r := &frameReader{source: bytes.NewReader(framed(body)), names: maxEntries + 1, maxData: 10}
	r.listing.Store(true)
	if _, e := r.Read(make([]byte, 100)); e == nil {
		t.Fatal("entry budget accepted")
	}
	body = []byte{103, 0, 0, 0, 1, 0, 0, 0, 2, 'a', 'b'}
	r = &frameReader{source: bytes.NewReader(framed(body)), maxData: 1}
	r.readRemaining.Store(2)
	if _, e := r.Read(make([]byte, 100)); e == nil {
		t.Fatal("wire byte budget accepted")
	}
	r = &frameReader{source: bytes.NewReader(framed(body)), maxData: 100}
	r.readRemaining.Store(1)
	if _, e := r.Read(make([]byte, 100)); e == nil {
		t.Fatal("overlong DATA accepted")
	}
	for _, ip := range []string{"fd00:ec2::254", "fd20:ce::254"} {
		if permitted(netip.MustParseAddr(ip), true) {
			t.Fatal("IPv6 cloud metadata accepted")
		}
	}
}
func FuzzResponseGuard(f *testing.F) {
	f.Add([]byte{2, 0, 0, 0, 3})
	f.Add([]byte{105, 0, 0, 0, 1, 0x80, 0, 0, 0, 255, 255, 255, 255})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) <= maxPacket {
			_, _, _ = validResponse(b, true)
		}
	})
}

func TestWindowsNamesAndExactAllowlist(t *testing.T) {
	for _, name := range []string{"CON .txt", "LPT1 .log", "COM¹.log"} {
		if safeComponent(name) {
			t.Fatalf("Windows device accepted: %s", name)
		}
	}
	if foldKey("I") != foldKey("ı") {
		t.Fatal("Windows upcase collision missed")
	}
	for _, allow := range [][]string{nil, {"127.0.0.1:0"}, {"127.0.0.1:23"}, {"other.test:22"}} {
		if _, e := pinnedAddress(context.Background(), "127.0.0.1", 22, allow); e == nil {
			t.Fatal("nonmatching allowlist permitted loopback")
		}
	}
}
