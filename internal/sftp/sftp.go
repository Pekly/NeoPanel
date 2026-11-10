package sftp

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
	sshConf.AddHostKey(rsaSigner)
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
			strings.EqualFold(sv.ID, "srv-"+targetServer) {
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

	if idx := strings.LastIndex(rawUser, "."); idx > 0 && idx < len(rawUser)-1 {
		candUser := rawUser[:idx]
		candSrv := rawUser[idx+1:]
		if srv := s.findServer(candSrv); srv != nil {
			return candUser, srv, nil
		}
		if u, err := s.db.GetUserByUsername(rawUser); err == nil && u != nil {
			return rawUser, nil, nil
		}
		return candUser, nil, fmt.Errorf("server %q not found", candSrv)
	}

	return rawUser, nil, nil
}

func (s *SFTPServer) authenticate(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	rawUser := c.User()
	username, srv, srvErr := s.resolveUsernameAndServer(rawUser)
	if srvErr != nil {
		return nil, srvErr
	}

	authenticated := false
	if s.cfg != nil && s.cfg.IsNode() {
		if nodeTok := strings.TrimSpace(s.cfg.Security.NodeToken); nodeTok != "" && len(pass) == len(nodeTok) &&
			subtle.ConstantTimeCompare(pass, []byte(nodeTok)) == 1 {
			authenticated = true
		}
	}

	if !authenticated {
		user, err := s.db.GetUserByUsername(username)
		if err != nil || user == nil {
			return nil, fmt.Errorf("authentication failed for user %q", rawUser)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), pass); err != nil {
			return nil, fmt.Errorf("invalid password for user %q", rawUser)
		}
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
	if req == nil || req.Type != "subsystem" || len(req.Payload) < 4 {
		return false
	}
	nameLen := int(binary.BigEndian.Uint32(req.Payload[:4]))
	if nameLen <= 0 || 4+nameLen > len(req.Payload) {
		return false
	}
	return string(req.Payload[4:4+nameLen]) == "sftp"
}

func (s *SFTPServer) handleConnection(netConn net.Conn) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConf)
	if err != nil {
		return
	}
	defer sshConn.Close()

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

			go func() {
				for req := range in {
					ok := false
					switch req.Type {
					case "subsystem":
						ok = isSFTPSubsystemRequest(req)
					case "env":
						ok = true
					}
					_ = req.Reply(ok, nil)
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

func (h *chrootHandler) safePath(reqPath string) (string, error) {
	cleanRel := filepath.Clean("/" + reqPath)
	cleanRel = strings.TrimPrefix(cleanRel, "/")
	cleanRel = strings.TrimPrefix(cleanRel, "\\")

	full := filepath.Clean(filepath.Join(h.rootDir, cleanRel))

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
	case pflags.Read && wantsWrite:
		flags = os.O_RDWR
	case wantsWrite:
		flags = os.O_WRONLY
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
	default:
		return nil, fmt.Errorf("unsupported list method: %s", r.Method)
	}
}
