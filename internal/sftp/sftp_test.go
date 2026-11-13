package sftp

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	sftpclient "github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

const testPassword = "local-integration-password"

type fixture struct {
	db         *database.DB
	server     *SFTPServer
	root, addr string
	signer     ssh.Signer
	users      map[string]*database.User
	members    map[string]*database.ServerSubuser
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	must(t, err)
	db, err := database.InitDB(filepath.Join(root, "test.db"))
	must(t, err)
	t.Cleanup(func() { db.Close() })
	f := &fixture{db: db, root: filepath.Join(root, "servers"), users: map[string]*database.User{}, members: map[string]*database.ServerSubuser{}}
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	must(t, err)
	_, private, err := ed25519.GenerateKey(rand.Reader)
	must(t, err)
	f.signer, err = ssh.NewSignerFromKey(private)
	must(t, err)
	for _, name := range []string{"owner", "otherowner", "admin", "reader", "editor", "member", "outsider", "accessonly", "dotted.user"} {
		role := "user"
		if name == "admin" {
			role = "admin"
		}
		u, err := db.CreateUser(name, string(hash), role)
		must(t, err)
		f.users[name] = u
		_, err = db.AddSSHKey(u.ID, "integration", ssh.FingerprintSHA256(f.signer.PublicKey()), string(ssh.MarshalAuthorizedKey(f.signer.PublicKey())))
		must(t, err)
	}
	for i, id := range []string{"srv-alpha", "srv-beta"} {
		owner := f.users["owner"].ID
		if i == 1 {
			owner = f.users["otherowner"].ID
		}
		must(t, db.CreateServer(&database.Server{ID: id, Name: "duplicate-alias", OwnerID: owner, Port: 25565 + i}))
		must(t, os.MkdirAll(filepath.Join(f.root, id, "world"), 0755))
		must(t, os.WriteFile(filepath.Join(f.root, id, "file.txt"), []byte(id+" contents"), 0644))
		must(t, os.WriteFile(filepath.Join(f.root, id, "world", "level.dat"), []byte("keep-world"), 0644))
	}
	for name, perms := range map[string][]string{
		"reader":      {"sftp.access", "files.view"},
		"editor":      {"sftp.access", "files.view", "files.edit", "files.delete"},
		"member":      {"files.view", "files.edit", "files.delete"},
		"accessonly":  {"sftp.access"},
		"dotted.user": {"sftp.access", "files.view"},
	} {
		m, err := db.AddServerSubuser("srv-alpha", f.users[name].ID, perms)
		must(t, err)
		f.members[name] = m
	}
	cfg := &config.Config{}
	cfg.SFTP.Enabled = true
	cfg.SFTP.KeyPath = filepath.Join(root, "host-key")
	cfg.Paths.ServersDir = f.root
	f.server, err = NewSFTPServer(cfg, db)
	must(t, err)
	// Exercise the production SSH handshake, channel negotiation, and handlers,
	// but bind only loopback and avoid Start's host firewall configuration.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	must(t, err)
	f.addr = listener.Addr().String()
	f.server.listener = listener
	f.server.wg.Add(1)
	go f.server.acceptLoop()
	t.Cleanup(f.server.Stop)
	return f
}

func (f *fixture) auth(kind string) ssh.AuthMethod {
	switch kind {
	case "key":
		return ssh.PublicKeys(f.signer)
	case "keyboard":
		return ssh.KeyboardInteractive(func(_ string, _ string, questions []string, _ []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = testPassword
			}
			return answers, nil
		})
	default:
		return ssh.Password(testPassword)
	}
}

func (f *fixture) dial(user, kind string) (*ssh.Client, error) {
	return ssh.Dial("tcp", f.addr, &ssh.ClientConfig{User: user, Auth: []ssh.AuthMethod{f.auth(kind)}, HostKeyAlgorithms: []string{ssh.KeyAlgoED25519}, HostKeyCallback: ssh.FixedHostKey(f.server.sshConfHostKeyForTest()), Timeout: 5 * time.Second})
}

// Read the generated host key rather than accepting arbitrary SSH host keys.
func (s *SFTPServer) sshConfHostKeyForTest() ssh.PublicKey {
	signer, err := getOrGenerateEd25519HostKey(s.cfg.SFTP.KeyPath + ".ed25519")
	if err != nil {
		panic(err)
	}
	return signer.PublicKey()
}

func (f *fixture) connect(t *testing.T, user, kind string) (*ssh.Client, *sftpclient.Client) {
	t.Helper()
	sshClient, err := f.dial(user, kind)
	must(t, err)
	t.Cleanup(func() { sshClient.Close() })
	client, err := sftpclient.NewClient(sshClient)
	must(t, err)
	t.Cleanup(func() { client.Close() })
	return sshClient, client
}

func denied(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("operation unexpectedly authorized")
	}
}

func TestSSHAuthenticationAndRouting(t *testing.T) {
	f := newFixture(t)
	for _, kind := range []string{"password", "key", "keyboard"} {
		t.Run(kind, func(t *testing.T) {
			for _, user := range []string{"outsider", "outsider#srv-alpha", "member", "member#srv-alpha", "reader#srv-beta", "admin", "owner#duplicate-alias", "owner#srv-", "owner#", "owner#unknown"} {
				t.Run("denied/"+user, func(t *testing.T) {
					c, err := f.dial(user, kind)
					if c != nil {
						c.Close()
					}
					denied(t, err)
				})
			}
			for _, user := range []string{"reader", "reader#srv-alpha", "reader/srv-alpha", "reader.srv-alpha", "reader#alpha", "reader#srv-al", "dotted.user", "owner", "admin#srv-alpha", "admin#srv-beta"} {
				t.Run("allowed/"+user, func(t *testing.T) {
					_, c := f.connect(t, user, kind)
					entries, err := c.ReadDir("/")
					must(t, err)
					for _, entry := range entries {
						if strings.HasPrefix(entry.Name(), "srv-") {
							t.Fatal("shared servers root exposed")
						}
					}
					file, err := c.Open("/file.txt")
					must(t, err)
					defer file.Close()
					contents, err := io.ReadAll(file)
					must(t, err)
					expected := "srv-alpha"
					if strings.HasSuffix(user, "srv-beta") {
						expected = "srv-beta"
					}
					if string(contents) != expected+" contents" {
						t.Fatalf("wrong jail: %q", contents)
					}
				})
			}
		})
	}
	// A non-SFTP membership must not influence automatic routing.
	_, err := f.db.AddServerSubuser("srv-beta", f.users["reader"].ID, []string{"files.view"})
	must(t, err)
	c, err := f.dial("reader", "password")
	must(t, err)
	c.Close()
	members, err := f.db.ListServerSubusers("srv-beta")
	must(t, err)
	must(t, f.db.UpdateServerSubuser("srv-beta", members[0].ID, []string{"sftp.access", "files.view"}))
	c, err = f.dial("reader", "password")
	if c != nil {
		c.Close()
	}
	denied(t, err)
	// Exact IDs remain unambiguous even when another server has the same alias.
	if f.server.findServer("srv-alpha") == nil || f.server.findServer("duplicate-alias") != nil {
		t.Fatal("alias resolution failed")
	}
}

func TestSFTPPermissions(t *testing.T) {
	f := newFixture(t)
	for _, kind := range []string{"password", "key"} {
		t.Run(kind, func(t *testing.T) {
			_, c := f.connect(t, "reader", kind)
			file, err := c.Create("/denied.txt")
			if file != nil {
				file.Close()
			}
			denied(t, err)
			denied(t, c.Mkdir("/denied-dir"))
			denied(t, c.Rename("/file.txt", "/renamed.txt"))
			denied(t, c.PosixRename("/file.txt", "/renamed.txt"))
			denied(t, c.Truncate("/file.txt", 0))
			denied(t, c.Chmod("/file.txt", 0600))
			denied(t, c.Remove("/file.txt"))
			denied(t, c.RemoveDirectory("/world"))
			denied(t, c.Symlink("/file.txt", "/link"))
			denied(t, c.Link("/file.txt", "/hardlink"))
		})
	}
	_, c := f.connect(t, "accessonly", "password")
	_, err := c.ReadDir("/")
	denied(t, err)
	_, err = c.Stat("/file.txt")
	denied(t, err)
	_, err = c.RealPath("/")
	denied(t, err)
	file, err := c.Open("/file.txt")
	if file != nil {
		file.Close()
	}
	denied(t, err)
	file, err = c.Create("/denied.txt")
	if file != nil {
		file.Close()
	}
	denied(t, err)
	// An edit-only grant allows uploads, not reads or deletes.
	must(t, f.db.UpdateServerSubuser("srv-alpha", f.members["accessonly"].ID, []string{"sftp.access", "files.edit"}))
	file, err = c.OpenFile("/upload.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	must(t, err)
	_, err = file.WriteAt([]byte("upload"), 0)
	must(t, err)
	must(t, file.Close())
	file, err = c.OpenFile("/upload.txt", os.O_RDWR)
	if file != nil {
		file.Close()
	}
	denied(t, err)
	denied(t, c.Remove("/upload.txt"))
}

func TestSFTPRevocationOnEstablishedHandles(t *testing.T) {
	for _, kind := range []string{"password", "key", "keyboard"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			_, c := f.connect(t, "editor", kind)
			read, err := c.Open("/file.txt")
			must(t, err)
			defer read.Close()
			write, err := c.OpenFile("/file.txt", os.O_WRONLY)
			must(t, err)
			defer write.Close()
			both, err := c.OpenFile("/file.txt", os.O_RDWR)
			must(t, err)
			defer both.Close()
			buf := make([]byte, 1)
			_, err = read.ReadAt(buf, 0)
			must(t, err)
			_, err = write.WriteAt([]byte("s"), 0)
			must(t, err)
			update := func(perms ...string) { must(t, f.db.UpdateServerSubuser("srv-alpha", f.members["editor"].ID, perms)) }
			update("sftp.access", "files.edit", "files.delete")
			_, err = read.ReadAt(buf, 0)
			denied(t, err)
			_, err = both.ReadAt(buf, 0)
			denied(t, err)
			_, err = read.Stat()
			denied(t, err)
			_, err = write.WriteAt([]byte("s"), 0)
			must(t, err)
			update("sftp.access", "files.view", "files.delete")
			_, err = write.WriteAt([]byte("x"), 0)
			denied(t, err)
			_, err = both.WriteAt([]byte("x"), 0)
			denied(t, err)
			denied(t, write.Truncate(0))
			_, err = read.ReadAt(buf, 0)
			must(t, err)
			update("files.view", "files.edit", "files.delete")
			_, err = read.ReadAt(buf, 0)
			denied(t, err)
			_, err = write.WriteAt([]byte("x"), 0)
			denied(t, err)
			_, err = c.ReadDir("/")
			denied(t, err)
			denied(t, c.Remove("/file.txt"))
			update("sftp.access", "files.view", "files.edit", "files.delete")
			_, err = read.ReadAt(buf, 0)
			must(t, err)
			must(t, f.db.DeleteServerSubuser("srv-alpha", f.members["editor"].ID))
			_, err = read.ReadAt(buf, 0)
			denied(t, err)
			_, err = write.WriteAt([]byte("x"), 0)
			denied(t, err)
		})
	}
}

func TestOwnerAdminAndDynamicIdentity(t *testing.T) {
	for _, user := range []string{"owner", "admin"} {
		t.Run(user, func(t *testing.T) {
			f := newFixture(t)
			_, c := f.connect(t, user+"#srv-alpha", "key")
			must(t, c.Mkdir("/new-dir"))
			file, err := c.Create("/new-dir/new.txt")
			must(t, err)
			_, err = file.WriteAt([]byte("allowed"), 0)
			must(t, err)
			must(t, file.Close())
			must(t, c.Rename("/new-dir/new.txt", "/new-dir/renamed.txt"))
			must(t, c.Remove("/new-dir/renamed.txt"))
			must(t, c.RemoveDirectory("/new-dir"))
			file, err = c.Open("/file.txt")
			must(t, err)
			defer file.Close()
			if user == "admin" {
				must(t, f.db.UpdateUserRole(f.users[user].ID, "user"))
			} else {
				must(t, f.db.UpdateServerOwner("srv-alpha", f.users["otherowner"].ID))
			}
			_, err = file.ReadAt(make([]byte, 1), 0)
			denied(t, err)
			_, err = c.ReadDir("/")
			denied(t, err)
		})
	}
	t.Run("deleted-user", func(t *testing.T) {
		f := newFixture(t)
		_, c := f.connect(t, "editor", "password")
		file, err := c.Open("/file.txt")
		must(t, err)
		defer file.Close()
		must(t, f.db.DeleteUser(f.users["editor"].ID))
		_, err = file.ReadAt(make([]byte, 1), 0)
		denied(t, err)
	})
	t.Run("deleted-server", func(t *testing.T) {
		f := newFixture(t)
		_, c := f.connect(t, "owner", "password")
		file, err := c.Open("/file.txt")
		must(t, err)
		defer file.Close()
		must(t, f.db.DeleteServer("srv-alpha"))
		_, err = file.ReadAt(make([]byte, 1), 0)
		denied(t, err)
	})
}

func TestSFTPJailAndNonRecursiveRmdir(t *testing.T) {
	f := newFixture(t)
	_, c := f.connect(t, "owner", "password")
	for _, p := range []string{"/../srv-beta/file.txt", "/../../srv-beta/file.txt", `\..\srv-beta\file.txt`, "/srv-beta/file.txt", "C:/file.txt", "/file.txt:stream"} {
		file, err := c.Open(p)
		if file != nil {
			file.Close()
		}
		denied(t, err)
	}
	denied(t, c.Rename("/file.txt", "/../srv-beta/stolen.txt"))
	denied(t, c.RemoveDirectory("/world"))
	denied(t, c.RemoveDirectory("/"))
	denied(t, c.Rename("/", "/renamed-root"))
	denied(t, c.Symlink("/../srv-beta", "/escape"))
	denied(t, c.Link("/file.txt", "/hardlink"))
	data, err := os.ReadFile(filepath.Join(f.root, "srv-alpha", "world", "level.dat"))
	must(t, err)
	if string(data) != "keep-world" {
		t.Fatal("world was changed")
	}
	data, err = os.ReadFile(filepath.Join(f.root, "srv-beta", "file.txt"))
	must(t, err)
	if string(data) != "srv-beta contents" {
		t.Fatal("other server was changed")
	}
	t.Run("preexisting-symlink", func(t *testing.T) {
		link := filepath.Join(f.root, "srv-alpha", "escape")
		if err := os.Symlink(filepath.Join(f.root, "srv-beta"), link); err != nil {
			t.Skipf("host cannot create symlinks: %v", err)
		}
		file, err := c.Open("/escape/file.txt")
		if file != nil {
			file.Close()
		}
		denied(t, err)
		file, err = c.Create("/escape/file.txt")
		if file != nil {
			file.Close()
		}
		denied(t, err)
		_, err = c.ReadDir("/escape")
		denied(t, err)
		denied(t, c.Remove("/escape/file.txt"))
		denied(t, c.Rename("/file.txt", "/escape/replaced.txt"))
	})
}

func TestSSHExecAndForwardingDenied(t *testing.T) {
	f := newFixture(t)
	client, _ := f.connect(t, "owner", "password")
	for _, command := range []string{"id", "sftp-server", "echo sftp"} {
		session, err := client.NewSession()
		must(t, err)
		denied(t, session.Start(command))
		session.Close()
	}
	session, err := client.NewSession()
	must(t, err)
	denied(t, session.Shell())
	session.Close()
	conn, err := client.Dial("tcp", "127.0.0.1:22")
	if conn != nil {
		conn.Close()
	}
	denied(t, err)
	listener, err := client.Listen("tcp", "127.0.0.1:0")
	if listener != nil {
		listener.Close()
	}
	denied(t, err)
}

// Raw SFTP packets retain an OPENDIR handle across DB revocation. The high-level
// client closes directory handles at the end of ReadDir, so it cannot test this.
func TestSFTPDirectoryHandleRevocation(t *testing.T) {
	f := newFixture(t)
	client, err := f.dial("reader", "password")
	must(t, err)
	defer client.Close()
	session, err := client.NewSession()
	must(t, err)
	defer session.Close()
	in, err := session.StdinPipe()
	must(t, err)
	out, err := session.StdoutPipe()
	must(t, err)
	must(t, session.RequestSubsystem("sftp"))
	send := func(body []byte) {
		must(t, binary.Write(in, binary.BigEndian, uint32(len(body))))
		_, err := in.Write(body)
		must(t, err)
	}
	receive := func() []byte {
		var size uint32
		must(t, binary.Read(out, binary.BigEndian, &size))
		if size > 65536 {
			t.Fatal("unexpected packet size")
		}
		body := make([]byte, size)
		_, err := io.ReadFull(out, body)
		must(t, err)
		return body
	}
	send([]byte{1, 0, 0, 0, 3}) // INIT v3
	if reply := receive(); reply[0] != 2 {
		t.Fatalf("not VERSION: %v", reply)
	}
	send([]byte{11, 0, 0, 0, 1, 0, 0, 0, 1, '/'}) // OPENDIR id=1, /
	reply := receive()
	if len(reply) < 9 || reply[0] != 102 {
		t.Fatalf("not HANDLE: %v", reply)
	}
	handle := append([]byte(nil), reply[5:]...)
	must(t, f.db.UpdateServerSubuser("srv-alpha", f.members["reader"].ID, []string{"files.view"}))
	send(append([]byte{12, 0, 0, 0, 2}, handle...)) // READDIR id=2
	reply = receive()
	if len(reply) < 9 || reply[0] != 101 || binary.BigEndian.Uint32(reply[5:9]) != 3 {
		t.Fatalf("expected permission denied STATUS: %v", reply)
	}
}

func TestSSHInvalidCredentials(t *testing.T) {
	f := newFixture(t)
	_, private, err := ed25519.GenerateKey(rand.Reader)
	must(t, err)
	wrongKey, err := ssh.NewSignerFromKey(private)
	must(t, err)
	for _, auth := range []ssh.AuthMethod{
		ssh.Password("wrong"), ssh.PublicKeys(wrongKey),
		ssh.KeyboardInteractive(func(_, _ string, _ []string, _ []bool) ([]string, error) { return []string{"wrong"}, nil }),
	} {
		c, err := ssh.Dial("tcp", f.addr, &ssh.ClientConfig{
			User: "owner", Auth: []ssh.AuthMethod{auth}, Timeout: 5 * time.Second,
			HostKeyAlgorithms: []string{ssh.KeyAlgoED25519}, HostKeyCallback: ssh.FixedHostKey(f.server.sshConfHostKeyForTest()),
		})
		if c != nil {
			c.Close()
		}
		denied(t, err)
		if !strings.Contains(err.Error(), "unable to authenticate") {
			t.Fatalf("failure was not authentication: %v", err)
		}
	}
}

func TestSFTPReplacedJailDenied(t *testing.T) {
	f := newFixture(t)
	_, c := f.connect(t, "owner", "key")
	// Establish the SSH session before replacing the jail with a cross-server link.
	root := filepath.Join(f.root, "srv-alpha")
	must(t, os.Rename(root, root+"-saved"))
	if err := os.Symlink(filepath.Join(f.root, "srv-beta"), root); err != nil {
		t.Skipf("host cannot create symlinks: %v", err)
	}
	_, err := c.ReadDir("/")
	denied(t, err)
	file, err := c.Open("/file.txt")
	if file != nil {
		file.Close()
	}
	denied(t, err)
	file, err = c.Create("/file.txt")
	if file != nil {
		file.Close()
	}
	denied(t, err)
	sshClient, err := f.dial("owner", "password")
	if sshClient != nil {
		sshClient.Close()
	}
	denied(t, err)
}

func TestServerIDValidation(t *testing.T) {
	for _, id := range []string{"", ".", "..", "../srv-beta", `..\srv-beta`, "C:escape", "srv-alpha.", "srv-alpha ", "srv\x00alpha"} {
		if validServerID(id) {
			t.Fatalf("accepted invalid server ID %q", id)
		}
	}
	for _, id := range []string{"srv-alpha", "1234", "abc-def-123", "srv.name"} {
		if !validServerID(id) {
			t.Fatalf("rejected valid server ID %q", id)
		}
	}
}
