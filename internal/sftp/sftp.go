package sftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
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

	signer, err := getOrGenerateHostKey(cfg.SFTP.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("sftp host key setup failed: %w", err)
	}

	s := &SFTPServer{
		cfg:      cfg,
		db:       db,
		stopChan: make(chan struct{}),
	}

	sshConf := &ssh.ServerConfig{
		PasswordCallback: s.authenticate,
	}
	sshConf.AddHostKey(signer)
	s.sshConf = sshConf

	return s, nil
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

func (s *SFTPServer) authenticate(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	rawUser := c.User()
	username := rawUser
	var targetServer string

	if idx := strings.Index(rawUser, "."); idx != -1 {
		username = rawUser[:idx]
		targetServer = rawUser[idx+1:]
	}

	user, err := s.db.GetUserByUsername(username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("authentication failed for user %q", rawUser)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), pass); err != nil {
		return nil, fmt.Errorf("invalid password for user %q", rawUser)
	}

	jailPath := s.cfg.Paths.ServersDir
	if targetServer != "" {
		srv, err := s.db.GetServer(targetServer)
		if err != nil {
			servers, _ := s.db.ListServers()
			for _, sv := range servers {
				if strings.EqualFold(sv.Name, targetServer) || strings.EqualFold(sv.ID, targetServer) {
					srv = sv
					break
				}
			}
		}
		if srv == nil {
			return nil, fmt.Errorf("server %q not found", targetServer)
		}
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

func (s *SFTPServer) handleConnection(netConn net.Conn) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConf)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	jailPath := sshConn.Permissions.Extensions["jail_path"]
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
			return
		}

		go func(ch ssh.Channel, in <-chan *ssh.Request) {
			defer ch.Close()

			for req := range in {
				ok := false
				if req.Type == "subsystem" && len(req.Payload) >= 4 && string(req.Payload[4:]) == "sftp" {
					ok = true
				}
				_ = req.Reply(ok, nil)
			}
		}(channel, requests)

		h := &chrootHandler{rootDir: jailPath}
		handlers := sftp.Handlers{
			FileGet:  h,
			FilePut:  h,
			FileCmd:  h,
			FileList: h,
		}

		server := sftp.NewRequestServer(channel, handlers)
		if err := server.Serve(); err != nil && err != io.EOF {
			log.Printf("[SFTP] Client session finished: %v", err)
		}
		_ = server.Close()
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

func (h *chrootHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	absPath, err := h.safePath(r.Filepath)
	if err != nil {
		return nil, err
	}

	flags := os.O_RDWR | os.O_CREATE
	// Bit 0x00000010 in SFTP is SSH_FXF_TRUNC
	if r.Flags&0x00000010 != 0 {
		flags |= os.O_TRUNC
	}
	// Bit 0x00000004 is SSH_FXF_APPEND
	if r.Flags&0x00000004 != 0 {
		flags |= os.O_APPEND
	}

	dir := filepath.Dir(absPath)
	_ = os.MkdirAll(dir, 0755)

	return os.OpenFile(absPath, flags, 0644)
}

func (h *chrootHandler) Filecmd(r *sftp.Request) error {
	absPath, err := h.safePath(r.Filepath)
	if err != nil {
		return err
	}

	switch r.Method {
	case "Rename":
		targetPath, err := h.safePath(r.Target)
		if err != nil {
			return err
		}
		return os.Rename(absPath, targetPath)
	case "Remove":
		return os.Remove(absPath)
	case "Rmdir":
		return os.RemoveAll(absPath)
	case "Mkdir":
		return os.MkdirAll(absPath, 0755)
	case "Setstat":
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
