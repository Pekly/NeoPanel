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
	"strconv"
	"strings"
	"sync"

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

// Exact IDs are canonical; non-exact aliases must identify exactly one server.
func (s *SFTPServer) findServer(targetServer string) *database.Server {
	targetServer = strings.TrimSpace(targetServer)
	if targetServer == "" {
		return nil
	}
	if srv, err := s.db.GetServer(targetServer); err == nil && srv != nil {
		return srv
	}
	servers, err := s.db.ListServers()
	if err != nil {
		return nil
	}
	var match *database.Server
	for _, sv := range servers {
		if strings.EqualFold(sv.ID, targetServer) || strings.EqualFold(sv.Name, targetServer) ||
			strings.EqualFold(sv.ID, "srv-"+targetServer) || strings.HasPrefix(strings.ToLower(sv.ID), strings.ToLower(targetServer)) {
			if match != nil {
				return nil
			}
			match = sv
		}
	}
	return match
}

func (s *SFTPServer) authorize(userID int, serverID, permission string) error {
	u, err := s.db.GetUserByID(userID)
	if err != nil || u == nil || serverID == "" {
		return os.ErrPermission
	}
	srv, err := s.db.GetServer(serverID)
	if err != nil || srv == nil {
		return os.ErrPermission
	}
	if u.Role == "admin" || srv.OwnerID == u.ID {
		return nil
	}
	perms, err := s.db.GetUserServerPermissions(serverID, u.ID)
	if err != nil {
		return os.ErrPermission
	}
	access, operation := false, permission == ""
	for _, p := range perms {
		access = access || p == "sftp.access"
		operation = operation || p == permission
	}
	if !access || !operation {
		return os.ErrPermission
	}
	return nil
}

func (s *SFTPServer) resolveUsernameAndServer(rawUser string) (string, *database.Server, error) {
	rawUser = strings.TrimSpace(rawUser)
	if rawUser == "" {
		return "", nil, os.ErrPermission
	}
	candUser, candSrv := rawUser, ""
	if idx := strings.IndexAny(rawUser, "#/"); idx >= 0 {
		if idx == 0 || idx == len(rawUser)-1 {
			return "", nil, os.ErrPermission
		}
		candUser, candSrv = rawUser[:idx], rawUser[idx+1:]
	} else if u, _ := s.db.GetUserByEmailOrUsername(rawUser); u == nil {
		if idx := strings.LastIndex(rawUser, "."); idx > 0 && idx < len(rawUser)-1 {
			candUser, candSrv = rawUser[:idx], rawUser[idx+1:]
		}
	}
	u, err := s.db.GetUserByEmailOrUsername(candUser)
	if err != nil || u == nil {
		return "", nil, os.ErrPermission
	}
	if candSrv != "" {
		srv := s.findServer(candSrv)
		if srv == nil || s.authorize(u.ID, srv.ID, "") != nil {
			return "", nil, os.ErrPermission
		}
		return u.Username, srv, nil
	}
	servers, err := s.db.ListServersForUser(u.ID, u.Role)
	if err != nil {
		return "", nil, os.ErrPermission
	}
	var permitted *database.Server
	for _, srv := range servers {
		if s.authorize(u.ID, srv.ID, "") != nil {
			continue
		}
		if permitted != nil {
			return "", nil, fmt.Errorf("multiple SFTP servers: specify username#server-id")
		}
		permitted = srv
	}
	// Every account, including administrators, gets an explicit per-server jail.
	if permitted == nil {
		return "", nil, os.ErrPermission
	}
	return u.Username, permitted, nil
}

// All authentication methods finish here; extensions carry identity, never grants.
func (s *SFTPServer) sessionPermissions(username string, srv *database.Server) (*ssh.Permissions, error) {
	u, err := s.db.GetUserByEmailOrUsername(username)
	if err != nil || u == nil || srv == nil {
		return nil, os.ErrPermission
	}
	if err := s.authorize(u.ID, srv.ID, ""); err != nil {
		return nil, err
	}
	if !validServerID(srv.ID) {
		return nil, os.ErrPermission
	}
	root, err := filepath.Abs(s.cfg.Paths.ServersDir)
	if err != nil {
		return nil, err
	}
	jail := filepath.Join(root, srv.ID)
	if err := prepareJail(jail); err != nil {
		return nil, err
	}
	return &ssh.Permissions{Extensions: map[string]string{
		"jail_path": jail, "user_id": strconv.Itoa(u.ID), "server_id": srv.ID,
	}}, nil
}

func validServerID(id string) bool {
	return id != "" && id != "." && id != ".." && !strings.ContainsAny(id, "/\\:\x00") &&
		filepath.IsLocal(id) && filepath.Base(id) == id && strings.TrimRight(id, ". ") == id
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

	return s.sessionPermissions(username, srv)
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

	return s.sessionPermissions(username, srv)
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

	if sshConn.Permissions == nil {
		return
	}
	ext := sshConn.Permissions.Extensions
	jailPath, serverID := ext["jail_path"], ext["server_id"]
	userID, err := strconv.Atoi(ext["user_id"])
	if err != nil || jailPath == "" || s.authorize(userID, serverID, "") != nil {
		return
	}

	h := &chrootHandler{rootDir: jailPath, server: s, userID: userID, serverID: serverID}
	rootHandle, err := h.fsRoot()
	if err != nil {
		return
	}
	defer rootHandle.Close()
	h.rootInfo, err = rootHandle.Stat()
	if !pinRootHandle {
		_ = rootHandle.Close()
	}
	if err != nil {
		return
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

		go func(ch ssh.Channel, in <-chan *ssh.Request) {
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
					// Shell, exec (including sftp-server), forwarding, and environment
					// requests are deliberately not supported.
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
					_ = req.Reply(false, nil)
				}
			}()

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
		}(channel, requests)
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
	rootDir  string
	server   *SFTPServer
	userID   int
	serverID string
	rootInfo os.FileInfo // immutable identity; Linux also pins the root handle
}

func (h *chrootHandler) check(permission string) error {
	if h.server == nil {
		return sftp.ErrSSHFxPermissionDenied
	}
	if h.server.authorize(h.userID, h.serverID, permission) != nil {
		return sftp.ErrSSHFxPermissionDenied
	}
	return nil
}

func (h *chrootHandler) RealPath(p string) (string, error) {
	if err := h.check("files.view"); err != nil {
		return "", err
	}
	if _, err := h.fsStat(p); err != nil {
		return "", err
	}
	return path.Clean("/" + strings.ReplaceAll(p, "\\", "/")), nil
}

func (h *chrootHandler) safePath(reqPath string) (string, error) {
	p := strings.ReplaceAll(reqPath, "\\", "/")
	for _, part := range strings.Split(p, "/") {
		if part == ".." || strings.ContainsAny(part, ":\x00") {
			return "", sftp.ErrSSHFxPermissionDenied
		}
		// Reject Windows path aliases, ADS, device names, and trailing-dot/space aliases.
		if part != "" && part != "." && (!filepath.IsLocal(part) || strings.TrimRight(part, ". ") != part) {
			return "", sftp.ErrSSHFxPermissionDenied
		}
	}
	p = strings.TrimPrefix(path.Clean("/"+p), "/")
	full := filepath.Join(h.rootDir, filepath.FromSlash(p))
	rel, err := filepath.Rel(h.rootDir, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", sftp.ErrSSHFxPermissionDenied
	}
	return full, nil
}

// Never return a raw file: READ/WRITE on an already-open handle must consult the
// database too, not only OPEN. Do not embed os.File and accidentally expose an
// unguarded ReadFrom/WriteTo/Stat shortcut.
type guardedFile struct {
	file               *os.File
	h                  *chrootHandler
	requestPath        string
	readable, writable bool
}

func (f *guardedFile) check(permission string) error {
	if err := f.h.check(permission); err != nil {
		return err
	}
	current, err := f.h.fsStat(f.requestPath)
	if err != nil {
		return err
	}
	opened, err := f.file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(current, opened) {
		return sftp.ErrSSHFxPermissionDenied
	}
	return nil
}

func (f *guardedFile) ReadAt(p []byte, off int64) (int, error) {
	if !f.readable {
		return 0, sftp.ErrSSHFxPermissionDenied
	}
	if err := f.check("files.view"); err != nil {
		return 0, err
	}
	return f.file.ReadAt(p, off)
}

func (f *guardedFile) WriteAt(p []byte, off int64) (int, error) {
	if !f.writable {
		return 0, sftp.ErrSSHFxPermissionDenied
	}
	if err := f.check("files.edit"); err != nil {
		return 0, err
	}
	return f.file.WriteAt(p, off)
}

func (f *guardedFile) Close() error { return f.file.Close() }

func (h *chrootHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	if err := h.check("files.view"); err != nil {
		return nil, err
	}
	file, err := h.fsOpen(r.Filepath, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		file.Close()
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return &guardedFile{file: file, h: h, requestPath: r.Filepath, readable: true}, nil
}

func (h *chrootHandler) openFileInternal(r *sftp.Request) (*guardedFile, error) {
	pf := r.Pflags()
	writes := pf.Write || pf.Append || pf.Creat || pf.Trunc
	if err := h.check(""); err != nil {
		return nil, err
	}
	if writes {
		if err := h.check("files.edit"); err != nil {
			return nil, err
		}
	}
	if pf.Read {
		if err := h.check("files.view"); err != nil {
			return nil, err
		}
	}
	flags := os.O_RDONLY
	if writes {
		flags = os.O_WRONLY
	}
	if writes && pf.Read {
		flags = os.O_RDWR
	}
	if pf.Creat {
		flags |= os.O_CREATE
	}
	if pf.Trunc {
		flags |= os.O_TRUNC
	}
	if pf.Excl {
		flags |= os.O_EXCL
	}
	// O_APPEND conflicts with WriterAt; SFTP supplies the write offset.
	file, err := h.fsOpen(r.Filepath, flags)
	if err != nil {
		return nil, err
	}
	return &guardedFile{file: file, h: h, requestPath: r.Filepath, readable: pf.Read, writable: writes}, nil
}

func (h *chrootHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) { return h.openFileInternal(r) }
func (h *chrootHandler) OpenFile(r *sftp.Request) (sftp.WriterAtReaderAt, error) {
	return h.openFileInternal(r)
}
func (h *chrootHandler) PosixRename(r *sftp.Request) error { return h.Filecmd(r) }

func (h *chrootHandler) Filecmd(r *sftp.Request) error {
	permission := "files.edit"
	if r.Method == "Remove" || r.Method == "Rmdir" {
		permission = "files.delete"
	}
	if err := h.check(permission); err != nil {
		return err
	}
	if r.Method == "Rename" || r.Method == "PosixRename" {
		if err := h.check("files.delete"); err != nil {
			return err
		}
	}
	return h.fsCommand(r)
}

type fileListerAt struct {
	infos       []os.FileInfo
	h           *chrootHandler
	requestPath string
}

func (f *fileListerAt) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if err := f.h.check("files.view"); err != nil {
		return 0, err
	}
	if _, err := f.h.fsStat(f.requestPath); err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, os.ErrInvalid
	}
	if offset >= int64(len(f.infos)) {
		return 0, io.EOF
	}
	n := copy(ls, f.infos[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

func (h *chrootHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	if err := h.check("files.view"); err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	switch r.Method {
	case "List":
		var err error
		infos, err = h.fsList(r.Filepath)
		if err != nil {
			return nil, err
		}
	case "Stat", "Lstat":
		info, err := h.fsStat(r.Filepath)
		if err != nil {
			return nil, err
		}
		infos = []os.FileInfo{info}
	default:
		return nil, sftp.ErrSSHFxOpUnsupported
	}
	return &fileListerAt{infos: infos, h: h, requestPath: r.Filepath}, nil
}
