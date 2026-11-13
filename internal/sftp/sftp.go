package sftp

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPServer struct {
	cfg      *config.Config
	db       *database.DB
	listener net.Listener
	sshConf  *ssh.ServerConfig
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewSFTPServer(cfg *config.Config, db *database.DB) (*SFTPServer, error) {
	if !cfg.SFTP.Enabled {
		return nil, nil
	}

	rsaSigner, err := getOrGenerateHostKey(cfg.SFTP.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("sftp rsa host key setup failed: %w", err)
	}

	edSigner, err := getOrGenerateEd25519HostKey(cfg.SFTP.KeyPath + ".ed25519")
	if err != nil {
		log.Printf("[WARN] SFTP Ed25519 host key setup skipped: %v", err)
	}

	s := &SFTPServer{
		cfg:      cfg,
		db:       db,
		stopChan: make(chan struct{}),
	}

	sshConf := &ssh.ServerConfig{
		PasswordCallback:            s.authenticate,
		KeyboardInteractiveCallback: s.authenticateKeyboardInteractive,
		PublicKeyCallback:           s.authenticatePublicKey,
		MaxAuthTries:                6,
		ServerVersion:               "SSH-2.0-NeoPanel_SFTP",
	}
	if edSigner != nil {
		sshConf.AddHostKey(edSigner)
	}

	// Support modern RSA SHA-2 signature algorithms (rsa-sha2-256, rsa-sha2-512) for OpenSSH 8.8+ & modern SFTP clients
	if algSigner, ok := rsaSigner.(ssh.AlgorithmSigner); ok {
		if sha2Signer, err := ssh.NewSignerWithAlgorithms(algSigner, []string{
			ssh.KeyAlgoRSASHA512,
			ssh.KeyAlgoRSASHA256,
			ssh.KeyAlgoRSA,
		}); err == nil {
			sshConf.AddHostKey(sha2Signer)
		} else {
			sshConf.AddHostKey(rsaSigner)
		}
	} else {
		sshConf.AddHostKey(rsaSigner)
	}

	s.sshConf = sshConf
	return s, nil
}

func (s *SFTPServer) IsListening() bool {
	return s != nil && s.listener != nil
}

func (s *SFTPServer) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.cfg.SFTP.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind SFTP port %d: %w", s.cfg.SFTP.Port, err)
	}
	s.listener = listener
	log.Printf("[SFTP] Built-in SFTP server listening on %s", addr)

	// Ensure firewall allows SFTP port on Linux host
	if _, err := exec.LookPath("ufw"); err == nil {
		_ = exec.Command("ufw", "allow", fmt.Sprintf("%d/tcp", s.cfg.SFTP.Port)).Run()
	}

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *SFTPServer) Stop() {
	if s.listener != nil {
		close(s.stopChan)
		_ = s.listener.Close()
		s.wg.Wait()
		log.Printf("[SFTP] SFTP server stopped")
	}
}

func (s *SFTPServer) findServer(targetServer string) *database.Server {
	targetServer = strings.TrimSpace(targetServer)
	if targetServer == "" {
		return nil
	}
	if srv, err := s.db.GetServer(targetServer); err == nil && srv != nil {
		return srv
	}
	servers, _ := s.db.ListServers()
	for _, sv := range servers {
		if strings.EqualFold(sv.ID, targetServer) ||
			strings.EqualFold(sv.Name, targetServer) ||
			strings.EqualFold(sv.ID, "srv-"+targetServer) ||
			strings.HasPrefix(strings.ToLower(sv.ID), strings.ToLower(targetServer)) {
			return sv
		}
	}
	return nil
}

func (s *SFTPServer) resolveUsernameAndServer(rawUser string) (string, *database.Server, error) {
	rawUser = strings.TrimSpace(rawUser)
	if rawUser == "" {
		return "", nil, fmt.Errorf("empty username")
	}

	var candUser, candSrv string

	// 1. Support '#' or '/' delimiter (e.g. user#srv-1 or user/srv-1)
	if idx := strings.IndexAny(rawUser, "#/"); idx > 0 && idx < len(rawUser)-1 {
		candUser = rawUser[:idx]
		candSrv = rawUser[idx+1:]
	} else if idx := strings.LastIndex(rawUser, "."); idx > 0 && idx < len(rawUser)-1 {
		// 2. Support '.' delimiter (e.g. user.srv-1)
		part1 := rawUser[:idx]
		part2 := rawUser[idx+1:]
		if srv := s.findServer(part2); srv != nil {
			candUser = part1
			candSrv = part2
		} else if u, _ := s.db.GetUserByEmailOrUsername(rawUser); u != nil {
			candUser = u.Username
			candSrv = ""
		} else {
			candUser = part1
			candSrv = part2
		}
	} else {
		candUser = rawUser
		candSrv = ""
	}

	u, err := s.db.GetUserByEmailOrUsername(candUser)
	if err != nil || u == nil {
		if u2, err2 := s.db.GetUserByEmailOrUsername(rawUser); err2 == nil && u2 != nil {
			u = u2
			candSrv = ""
		} else {
			return "", nil, fmt.Errorf("user %q not found", candUser)
		}
	}

	actualUsername := u.Username

	if candSrv != "" {
		srv := s.findServer(candSrv)
		if srv == nil {
			return actualUsername, nil, fmt.Errorf("server %q not found", candSrv)
		}
		if u.Role != "admin" && srv.OwnerID != u.ID {
			perms, _ := s.db.GetUserServerPermissions(srv.ID, u.ID)
			if len(perms) == 0 {
				return actualUsername, nil, fmt.Errorf("access denied to server %q", candSrv)
			}
		}
		return actualUsername, srv, nil
	}

	// Auto-route single server accounts directly
	srvs, err := s.db.ListServersForUser(u.ID, u.Role)
	if err == nil && len(srvs) == 1 {
		return actualUsername, srvs[0], nil
	}

	if u.Role == "admin" {
		return actualUsername, nil, nil
	}

	if len(srvs) > 0 {
		return actualUsername, srvs[0], nil
	}

	return actualUsername, nil, nil
}

func (s *SFTPServer) authenticate(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	rawUser := c.User()
	username, srv, srvErr := s.resolveUsernameAndServer(rawUser)
	if srvErr != nil {
		return nil, srvErr
	}

	// Always authenticate via NeoPanel user credentials.
	// Node token auth is never valid for SFTP connections – that prevents
	// clients from connecting as "node-daemon" using the node secret.
	user, err := s.db.GetUserByEmailOrUsername(username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("authentication failed for user %q", rawUser)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), pass); err != nil {
		return nil, fmt.Errorf("invalid password for user %q", rawUser)
	}

	jailPath := s.cfg.Paths.ServersDir
	if srv != nil {
		jailPath = filepath.Join(s.cfg.Paths.ServersDir, srv.ID)
	}

	_ = os.MkdirAll(jailPath, 0755)

	return &ssh.Permissions{
		Extensions: map[string]string{
			"jail_path": jailPath,
			"username":  username,
		},
	}, nil
}

func (s *SFTPServer) authenticateKeyboardInteractive(c ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	answers, err := challenge("NeoPanel SFTP", "Enter your NeoPanel panel password.", []string{"Password: "}, []bool{false})
	if err != nil || len(answers) == 0 {
		return nil, fmt.Errorf("keyboard-interactive authentication failed for user %q", c.User())
	}
	return s.authenticate(c, []byte(answers[0]))
}

func (s *SFTPServer) authenticatePublicKey(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	rawUser := c.User()
	username, srv, srvErr := s.resolveUsernameAndServer(rawUser)
	if srvErr != nil {
		return nil, srvErr
	}

	keys, err := s.db.GetAllSSHKeysForUser(username)
	if err != nil || len(keys) == 0 {
		return nil, fmt.Errorf("no authorized ssh keys registered for user %q", username)
	}

	clientKeyBytes := pubKey.Marshal()
	matched := false
	for _, k := range keys {
		parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k.PublicKey))
		if err == nil && string(parsedKey.Marshal()) == string(clientKeyBytes) {
			matched = true
			break
		}
	}

	if !matched {
		return nil, fmt.Errorf("public key rejected for user %q", username)
	}

	jailPath := s.cfg.Paths.ServersDir
	if srv != nil {
		jailPath = filepath.Join(s.cfg.Paths.ServersDir, srv.ID)
	}

	_ = os.MkdirAll(jailPath, 0755)

	return &ssh.Permissions{
		Extensions: map[string]string{
			"jail_path": jailPath,
			"username":  username,
		},
	}, nil
}

func (s *SFTPServer) acceptLoop() {
	defer s.wg.Done()

	for {
		netConn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return
			default:
				log.Printf("[SFTP] Accept connection error: %v", err)
				continue
			}
		}

		go s.handleConnection(netConn)
	}
}

func isSFTPSubsystemRequest(req *ssh.Request) bool {
	if req == nil || req.Type != "subsystem" {
		return false
	}
	if string(req.Payload) == "sftp" {
		return true
	}
	if len(req.Payload) >= 4 {
		nameLen := int(binary.BigEndian.Uint32(req.Payload[:4]))
		if nameLen > 0 && 4+nameLen <= len(req.Payload) {
			return string(req.Payload[4:4+nameLen]) == "sftp"
		}
	}
	return false
}

func (s *SFTPServer) handleConnection(netConn net.Conn) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConf)
	if err != nil {
		log.Printf("[SFTP] Handshake/auth failed from %s: %v", netConn.RemoteAddr(), err)
		return
	}
	defer sshConn.Close()

	log.Printf("[SFTP] Client %s connected as %q", netConn.RemoteAddr(), sshConn.User())

	go ssh.DiscardRequests(reqs)

	jailPath := ""
	if sshConn.Permissions != nil {
		jailPath = sshConn.Permissions.Extensions["jail_path"]
	}
	if jailPath == "" {
		jailPath = s.cfg.Paths.ServersDir
	}

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go func(ch ssh.Channel, in <-chan *ssh.Request, root string) {
			defer ch.Close()

			subsystemAccepted := false
			for req := range in {
				ok := false
				switch req.Type {
				case "subsystem":
					if isSFTPSubsystemRequest(req) {
						ok = true
						subsystemAccepted = true
					}
				case "exec":
					// Fallback for SFTP clients that issue exec sftp-server
					cmd := string(req.Payload)
					if len(req.Payload) >= 4 {
						cmdLen := int(binary.BigEndian.Uint32(req.Payload[:4]))
						if 4+cmdLen <= len(req.Payload) {
							cmd = string(req.Payload[4 : 4+cmdLen])
						}
					}
					if strings.Contains(strings.ToLower(cmd), "sftp") {
						ok = true
						subsystemAccepted = true
					}
				case "pty-req", "env", "window-change":
					// Accept terminal/environment requests from WinSCP/Cyberduck without failing
					ok = true
				case "shell":
					_ = req.Reply(true, nil)
					fmt.Fprintf(ch, "NeoPanel SFTP Service\r\n\r\nThis port is reserved exclusively for SFTP file transfers (FileZilla, WinSCP, Cyberduck, etc.).\r\nDirect interactive shell commands are disabled.\r\n")
					return
				}
				_ = req.Reply(ok, nil)
				if subsystemAccepted {
					break
				}
			}

			if !subsystemAccepted {
				return
			}

			// Drain any remaining out-of-band requests in background
			go func() {
				for req := range in {
					_ = req.Reply(req.Type == "env" || req.Type == "window-change", nil)
				}
			}()

			h := &chrootHandler{rootDir: root}
			handlers := sftp.Handlers{
				FileGet:  h,
				FilePut:  h,
				FileCmd:  h,
				FileList: h,
			}

			server := sftp.NewRequestServer(ch, handlers)
			if err := server.Serve(); err != nil && err != io.EOF {
				log.Printf("[SFTP] Client session finished: %v", err)
			}
			_ = server.Close()
		}(channel, requests, jailPath)
	}
}

func getOrGenerateHostKey(keyPath string) (ssh.Signer, error) {
	if _, err := os.Stat(keyPath); err == nil {
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing host key: %w", err)
		}
		return ssh.ParsePrivateKey(keyBytes)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	_ = os.MkdirAll(filepath.Dir(keyPath), 0700)
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to save host key: %w", err)
	}

	return ssh.ParsePrivateKey(keyPEM)
}

func getOrGenerateEd25519HostKey(keyPath string) (ssh.Signer, error) {
	if _, err := os.Stat(keyPath); err == nil {
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing ed25519 host key: %w", err)
		}
		return ssh.ParsePrivateKey(keyBytes)
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 host key: %w", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ed25519 host key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})

	_ = os.MkdirAll(filepath.Dir(keyPath), 0700)
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to save ed25519 host key: %w", err)
	}

	return ssh.ParsePrivateKey(keyPEM)
}

// -----------------------------------------------------------------------------
// Chroot Handler Implementation
// -----------------------------------------------------------------------------

type chrootHandler struct {
	rootDir string
}

func (h *chrootHandler) RealPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" || p == "." || p == "/" {
		return "/", nil
	}
	clean := path.Clean("/" + strings.ReplaceAll(p, "\\", "/"))
	return clean, nil
}

func (h *chrootHandler) safePath(reqPath string) (string, error) {
	p := path.Clean("/" + strings.ReplaceAll(reqPath, "\\", "/"))
	p = strings.TrimPrefix(p, "/")

	full := filepath.Clean(filepath.Join(h.rootDir, filepath.FromSlash(p)))

	rel, err := filepath.Rel(h.rootDir, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("permission denied: path traversal attempt")
	}

	return full, nil
}

func (h *chrootHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	absPath, err := h.safePath(r.Filepath)
	if err != nil {
		return nil, err
	}
	return os.Open(absPath)
}

func (h *chrootHandler) openFileInternal(r *sftp.Request) (*os.File, error) {
	absPath, err := h.safePath(r.Filepath)
	if err != nil {
		return nil, err
	}

	pflags := r.Pflags()
	var flags int
	wantsWrite := pflags.Write || pflags.Append || pflags.Creat || pflags.Trunc
	switch {
	case wantsWrite:
		flags = os.O_RDWR
	default:
		flags = os.O_RDONLY
	}

	if pflags.Creat {
		flags |= os.O_CREATE
	}
	if pflags.Trunc {
		flags |= os.O_TRUNC
	}
	if pflags.Excl {
		flags |= os.O_EXCL
	}
	// Never pass os.O_APPEND with io.WriterAt (WriteAt ignores/conflicts with O_APPEND offsets).

	if wantsWrite {
		_ = os.MkdirAll(filepath.Dir(absPath), 0755)
	}

	return os.OpenFile(absPath, flags, 0644)
}

func (h *chrootHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	return h.openFileInternal(r)
}

func (h *chrootHandler) OpenFile(r *sftp.Request) (sftp.WriterAtReaderAt, error) {
	return h.openFileInternal(r)
}

func (h *chrootHandler) PosixRename(r *sftp.Request) error {
	return h.Filecmd(r)
}

func (h *chrootHandler) Filecmd(r *sftp.Request) error {
	absPath, err := h.safePath(r.Filepath)
	if err != nil {
		return err
	}

	switch r.Method {
	case "Rename", "PosixRename":
		targetPath, err := h.safePath(r.Target)
		if err != nil {
			return err
		}
		_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
		return os.Rename(absPath, targetPath)
	case "Remove":
		return os.Remove(absPath)
	case "Rmdir":
		return os.RemoveAll(absPath)
	case "Mkdir":
		return os.MkdirAll(absPath, 0755)
	case "Symlink":
		targetPath, err := h.safePath(r.Target)
		if err != nil {
			return err
		}
		_ = os.Remove(targetPath)
		return os.Symlink(absPath, targetPath)
	case "Link":
		targetPath, err := h.safePath(r.Target)
		if err != nil {
			return err
		}
		_ = os.Remove(targetPath)
		return os.Link(absPath, targetPath)
	case "Setstat":
		attrFlags := r.AttrFlags()
		attrs := r.Attributes()
		if attrs != nil {
			if attrFlags.Size {
				if err := os.Truncate(absPath, int64(attrs.Size)); err != nil {
					return err
				}
			}
			if attrFlags.Permissions {
				if mode := attrs.FileMode().Perm(); mode != 0 {
					if err := os.Chmod(absPath, mode); err != nil {
						return err
					}
				}
			}
			if attrFlags.Acmodtime {
				atime := time.Unix(int64(attrs.Atime), 0)
				mtime := time.Unix(int64(attrs.Mtime), 0)
				if err := os.Chtimes(absPath, atime, mtime); err != nil {
					return err
				}
			}
		}
		return nil
	default:
		return nil
	}
}

type fileListerAt []os.FileInfo

func (f fileListerAt) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n := copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

type symlinkFileInfo struct {
	os.FileInfo
	target string
}

func (s symlinkFileInfo) Name() string {
	return s.target
}

func (h *chrootHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	absPath, err := h.safePath(r.Filepath)
	if err != nil {
		return nil, err
	}

	switch r.Method {
	case "List":
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return nil, err
		}
		var infos []os.FileInfo
		for _, e := range entries {
			if info, err := e.Info(); err == nil {
				infos = append(infos, info)
			}
		}
		return fileListerAt(infos), nil
	case "Stat":
		fi, err := os.Stat(absPath)
		if err != nil {
			return nil, err
		}
		return fileListerAt([]os.FileInfo{fi}), nil
	case "Lstat":
		fi, err := os.Lstat(absPath)
		if err != nil {
			return nil, err
		}
		return fileListerAt([]os.FileInfo{fi}), nil
	case "Readlink":
		target, err := os.Readlink(absPath)
		if err != nil {
			return nil, err
		}
		fi, err := os.Lstat(absPath)
		if err != nil {
			return nil, err
		}
		return fileListerAt([]os.FileInfo{symlinkFileInfo{FileInfo: fi, target: target}}), nil
	default:
		return nil, fmt.Errorf("unsupported list method: %s", r.Method)
	}
}
