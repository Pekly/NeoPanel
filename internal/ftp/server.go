package ftp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"golang.org/x/crypto/bcrypt"
)

type FTPServer struct {
	cfg      *config.Config
	db       *database.DB
	listener net.Listener
	stopChan chan struct{}
	wg       sync.WaitGroup
	stopOnce sync.Once
}

func NewFTPServer(cfg *config.Config, db *database.DB) (*FTPServer, error) {
	if cfg == nil || !cfg.FTP.Enabled {
		return nil, nil
	}
	return &FTPServer{
		cfg:      cfg,
		db:       db,
		stopChan: make(chan struct{}),
	}, nil
}

func (s *FTPServer) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.FTP.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on FTP port %d: %w", s.cfg.FTP.Port, err)
	}
	s.listener = l
	log.Printf("Built-in FTP Server listening on port %d", s.cfg.FTP.Port)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					continue
				}
			}
			s.wg.Add(1)
			go func(c net.Conn) {
				defer s.wg.Done()
				s.handleConn(c)
			}(conn)
		}
	}()
	return nil
}

func (s *FTPServer) Addr() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *FTPServer) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		if s.listener != nil {
			_ = s.listener.Close()
		}
		s.wg.Wait()
	})
}

type ftpSession struct {
	srv      *FTPServer
	conn     net.Conn
	writer   *bufio.Writer
	user     *database.User
	server   *database.Server
	root     *rooted.Root
	cwd      string // always POSIX relative, e.g. "/" or "/plugins"
	pendUser string
	authed   bool
	pasvL    net.Listener
	renameFrom string
}

func (sess *ftpSession) reply(code int, msg string) {
	_, _ = fmt.Fprintf(sess.writer, "%d %s\r\n", code, msg)
	_ = sess.writer.Flush()
}

func (sess *ftpSession) closePasv() {
	if sess.pasvL != nil {
		_ = sess.pasvL.Close()
		sess.pasvL = nil
	}
}

func (sess *ftpSession) openDataConn() (net.Conn, error) {
	if sess.pasvL == nil {
		return nil, errors.New("use PASV or EPSV first")
	}
	if tl, ok := sess.pasvL.(*net.TCPListener); ok {
		_ = tl.SetDeadline(time.Now().Add(15 * time.Second))
	}
	dc, err := sess.pasvL.Accept()
	sess.closePasv()
	return dc, err
}

func (sess *ftpSession) checkPerm(perm string) error {
	if !sess.authed || sess.user == nil || sess.server == nil {
		return os.ErrPermission
	}
	return sess.srv.authorize(sess.user.ID, sess.server.ID, perm)
}

func (sess *ftpSession) resolvePath(arg string) (string, error) {
	p := strings.ReplaceAll(strings.TrimSpace(arg), "\\", "/")
	var combined string
	if strings.HasPrefix(p, "/") {
		combined = path.Clean(p)
	} else {
		combined = path.Clean(path.Join(sess.cwd, p))
	}
	rel := strings.TrimPrefix(combined, "/")
	if rel == "" || rel == "." {
		return ".", nil
	}
	if sess.root == nil {
		return "", os.ErrPermission
	}
	return sess.root.Path(rel)
}

func (s *FTPServer) handleConn(c net.Conn) {
	defer c.Close()
	sess := &ftpSession{
		srv:    s,
		conn:   c,
		writer: bufio.NewWriter(c),
		cwd:    "/",
	}
	defer func() {
		sess.closePasv()
		if sess.root != nil {
			_ = sess.root.Close()
		}
	}()

	sess.reply(220, "NeoPanel FTP Server Ready")
	scanner := bufio.NewScanner(io.LimitReader(c, 1<<20))
	for scanner.Scan() {
		_ = c.SetReadDeadline(time.Now().Add(5 * time.Minute))
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(strings.TrimSpace(parts[0]))
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		switch cmd {
		case "USER":
			sess.pendUser = strings.TrimSpace(arg)
			sess.authed = false
			sess.reply(331, "Password required")

		case "PASS":
			if sess.pendUser == "" {
				sess.reply(503, "Login with USER first")
				continue
			}
			u, srv, root, err := s.authenticate(sess.pendUser, arg)
			if err != nil {
				sess.reply(530, "Login incorrect")
				continue
			}
			if sess.root != nil {
				_ = sess.root.Close()
			}
			sess.user = u
			sess.server = srv
			sess.root = root
			sess.cwd = "/"
			sess.authed = true
			sess.reply(230, "Login successful")

		case "QUIT":
			sess.reply(221, "Goodbye")
			return

		case "NOOP":
			sess.reply(200, "OK")

		case "SYST":
			sess.reply(215, "UNIX Type: L8")

		case "FEAT":
			_, _ = fmt.Fprintf(sess.writer, "211-Features:\r\n UTF8\r\n SIZE\r\n MDTM\r\n PASV\r\n EPSV\r\n MLSD\r\n211 End\r\n")
			_ = sess.writer.Flush()

		case "OPTS":
			sess.reply(200, "OK")

		case "TYPE":
			sess.reply(200, "Type set")

		case "PWD", "XPWD":
			if !sess.authed {
				sess.reply(530, "Not logged in")
				continue
			}
			sess.reply(257, fmt.Sprintf("%q is current directory", sess.cwd))

		case "CWD":
			if err := sess.checkPerm("files.view"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil {
				sess.reply(550, "Invalid path")
				continue
			}
			st, err := sess.root.Lstat(rel)
			if err != nil || !st.IsDir() {
				sess.reply(550, "Directory not found")
				continue
			}
			if rel == "." {
				sess.cwd = "/"
			} else {
				sess.cwd = "/" + filepath.ToSlash(rel)
			}
			sess.reply(250, "Directory changed")

		case "CDUP":
			if !sess.authed {
				sess.reply(530, "Not logged in")
				continue
			}
			sess.cwd = path.Dir(sess.cwd)
			if sess.cwd == "." || sess.cwd == "" {
				sess.cwd = "/"
			}
			sess.reply(250, "Directory changed")

		case "PASV":
			if !sess.authed {
				sess.reply(530, "Not logged in")
				continue
			}
			sess.closePasv()
			hostIP := "127.0.0.1"
			if tcpAddr, ok := c.LocalAddr().(*net.TCPAddr); ok {
				if ip4 := tcpAddr.IP.To4(); ip4 != nil && !ip4.IsUnspecified() {
					hostIP = ip4.String()
				}
			}
			l, err := net.Listen("tcp4", "0.0.0.0:0")
			if err != nil {
				sess.reply(425, "Cannot open passive connection")
				continue
			}
			sess.pasvL = l
			port := l.Addr().(*net.TCPAddr).Port
			ipParts := strings.Split(hostIP, ".")
			sess.reply(227, fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", ipParts[0], ipParts[1], ipParts[2], ipParts[3], port/256, port%256))

		case "EPSV":
			if !sess.authed {
				sess.reply(530, "Not logged in")
				continue
			}
			sess.closePasv()
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				sess.reply(425, "Cannot open passive connection")
				continue
			}
			sess.pasvL = l
			port := l.Addr().(*net.TCPAddr).Port
			sess.reply(229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", port))

		case "LIST", "NLST", "MLSD":
			if err := sess.checkPerm("files.view"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			target := arg
			if strings.HasPrefix(target, "-") {
				target = ""
			}
			rel, err := sess.resolvePath(target)
			if err != nil {
				sess.reply(550, "Invalid path")
				continue
			}
			f, err := sess.root.OpenFile(rel, os.O_RDONLY, 0)
			if err != nil {
				sess.reply(550, "Cannot open directory")
				continue
			}
			entries, err := f.ReadDir(-1)
			_ = f.Close()
			if err != nil {
				sess.reply(550, "Failed to read directory")
				continue
			}
			sess.reply(150, "Opening data connection for directory list")
			dc, err := sess.openDataConn()
			if err != nil {
				sess.reply(425, "Can't open data connection")
				continue
			}
			for _, ent := range entries {
				if strings.HasPrefix(ent.Name(), ".neopanel-") {
					continue
				}
				info, err := ent.Info()
				if err != nil {
					continue
				}
				if cmd == "NLST" {
					_, _ = fmt.Fprintf(dc, "%s\r\n", ent.Name())
				} else if cmd == "MLSD" {
					typ := "file"
					if info.IsDir() {
						typ = "dir"
					}
					_, _ = fmt.Fprintf(dc, "type=%s;size=%d;modify=%s; %s\r\n", typ, info.Size(), info.ModTime().UTC().Format("20060102150405"), ent.Name())
				} else {
					perm := "-rw-r--r--"
					if info.IsDir() {
						perm = "drwxr-xr-x"
					}
					_, _ = fmt.Fprintf(dc, "%s 1 owner group %12d %s %s\r\n", perm, info.Size(), info.ModTime().Format("Jan 02 15:04"), ent.Name())
				}
			}
			_ = dc.Close()
			sess.reply(226, "Transfer complete")

		case "SIZE":
			if err := sess.checkPerm("files.view"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil {
				sess.reply(550, "Invalid path")
				continue
			}
			st, err := sess.root.Lstat(rel)
			if err != nil || st.IsDir() {
				sess.reply(550, "File not found")
				continue
			}
			sess.reply(213, strconv.FormatInt(st.Size(), 10))

		case "MDTM":
			if err := sess.checkPerm("files.view"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil {
				sess.reply(550, "Invalid path")
				continue
			}
			st, err := sess.root.Lstat(rel)
			if err != nil {
				sess.reply(550, "File not found")
				continue
			}
			sess.reply(213, st.ModTime().UTC().Format("20060102150405"))

		case "RETR":
			if err := sess.checkPerm("files.view"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil {
				sess.reply(550, "Invalid path")
				continue
			}
			f, err := sess.root.Open(rel)
			if err != nil {
				sess.reply(550, "File not found")
				continue
			}
			sess.reply(150, "Opening data connection for file download")
			dc, err := sess.openDataConn()
			if err != nil {
				_ = f.Close()
				sess.reply(425, "Can't open data connection")
				continue
			}
			_, copyErr := io.Copy(dc, f)
			_ = f.Close()
			_ = dc.Close()
			if copyErr != nil {
				sess.reply(426, "Connection closed; transfer aborted")
			} else {
				sess.reply(226, "Transfer complete")
			}

		case "STOR":
			if err := sess.checkPerm("files.edit"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil || rel == "." {
				sess.reply(550, "Invalid destination path")
				continue
			}
			sess.reply(150, "Opening data connection for file upload")
			dc, err := sess.openDataConn()
			if err != nil {
				sess.reply(425, "Can't open data connection")
				continue
			}
			limit := int64(2 << 30)
			if sess.server.DiskLimitMB > 0 {
				limit = sess.server.DiskLimitMB * (1 << 20)
			}
			_, copyErr := sess.root.Copy(context.Background(), rel, dc, limit, 0644)
			_ = dc.Close()
			if copyErr != nil {
				sess.reply(552, fmt.Sprintf("Upload failed: %v", copyErr))
			} else {
				sess.reply(226, "Transfer complete")
			}

		case "DELE":
			if err := sess.checkPerm("files.delete"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil || rel == "." {
				sess.reply(550, "Invalid path")
				continue
			}
			if err := sess.root.Remove(rel); err != nil {
				sess.reply(550, "Delete failed")
			} else {
				sess.reply(250, "File deleted")
			}

		case "MKD", "XMKD":
			if err := sess.checkPerm("files.edit"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil || rel == "." {
				sess.reply(550, "Invalid path")
				continue
			}
			if err := sess.root.Mkdir(rel, 0755); err != nil {
				sess.reply(550, "Create directory failed")
			} else {
				sess.reply(257, fmt.Sprintf("%q created", arg))
			}

		case "RMD", "XRMD":
			if err := sess.checkPerm("files.delete"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil || rel == "." {
				sess.reply(550, "Invalid path")
				continue
			}
			if err := sess.root.Remove(rel); err != nil {
				sess.reply(550, "Remove directory failed")
			} else {
				sess.reply(250, "Directory removed")
			}

		case "RNFR":
			if err := sess.checkPerm("files.edit+files.delete"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			rel, err := sess.resolvePath(arg)
			if err != nil || rel == "." {
				sess.reply(550, "Invalid path")
				continue
			}
			if _, err := sess.root.Lstat(rel); err != nil {
				sess.reply(550, "Source not found")
				continue
			}
			sess.renameFrom = rel
			sess.reply(350, "Ready for RNTO")

		case "RNTO":
			if err := sess.checkPerm("files.edit+files.delete"); err != nil {
				sess.reply(550, "Permission denied")
				continue
			}
			if sess.renameFrom == "" {
				sess.reply(503, "RNFR required first")
				continue
			}
			from := sess.renameFrom
			sess.renameFrom = ""
			to, err := sess.resolvePath(arg)
			if err != nil || to == "." {
				sess.reply(550, "Invalid destination")
				continue
			}
			if err := sess.root.Rename(from, to); err != nil {
				sess.reply(550, "Rename failed")
			} else {
				sess.reply(250, "Renamed")
			}

		default:
			sess.reply(502, "Command not implemented")
		}
	}
}

func (s *FTPServer) authenticate(rawUser, password string) (*database.User, *database.Server, *rooted.Root, error) {
	rawUser = strings.TrimSpace(rawUser)
	if rawUser == "" || password == "" {
		return nil, nil, nil, os.ErrPermission
	}
	candUser, candSrv := rawUser, ""
	if idx := strings.IndexAny(rawUser, "#/"); idx >= 0 {
		if idx == 0 || idx == len(rawUser)-1 {
			return nil, nil, nil, os.ErrPermission
		}
		candUser, candSrv = rawUser[:idx], rawUser[idx+1:]
	} else if u, _ := s.db.GetUserByEmailOrUsername(rawUser); u == nil {
		if idx := strings.LastIndex(rawUser, "."); idx > 0 && idx < len(rawUser)-1 {
			candUser, candSrv = rawUser[:idx], rawUser[idx+1:]
		}
	}

	u, err := s.db.GetUserByEmailOrUsername(candUser)
	if err != nil || u == nil {
		return nil, nil, nil, os.ErrPermission
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, nil, os.ErrPermission
	}

	var target *database.Server
	if candSrv != "" {
		srv, _ := s.db.GetServer(candSrv)
		if srv == nil {
			servers, _ := s.db.ListServers()
			for _, candidate := range servers {
				if strings.EqualFold(candidate.Name, candSrv) {
					srv = candidate
					break
				}
			}
		}
		if srv == nil || s.authorize(u.ID, srv.ID, "") != nil {
			return nil, nil, nil, os.ErrPermission
		}
		target = srv
	} else {
		servers, err := s.db.ListServersForUser(u.ID, u.Role)
		if err != nil {
			return nil, nil, nil, os.ErrPermission
		}
		for _, srv := range servers {
			if s.authorize(u.ID, srv.ID, "") == nil {
				if target != nil {
					return nil, nil, nil, errors.New("multiple servers: specify username#server-id")
				}
				target = srv
			}
		}
	}
	if target == nil {
		return nil, nil, nil, os.ErrPermission
	}

	serverDir := filepath.Join(s.cfg.Paths.ServersDir, target.ID)
	_ = os.MkdirAll(serverDir, 0755)
	root, err := rooted.Open(serverDir)
	if err != nil {
		return nil, nil, nil, err
	}
	return u, target, root, nil
}

func (s *FTPServer) authorize(userID int, serverID, perm string) error {
	if !s.db.FeatureEnabled(serverID, "sftp") {
		return os.ErrPermission
	}
	u, err := s.db.GetUserByID(userID)
	if err != nil || u == nil {
		return os.ErrPermission
	}
	srv, err := s.db.GetServer(serverID)
	if err != nil || srv == nil {
		return os.ErrPermission
	}
	if u.Role == "admin" || srv.OwnerID == u.ID {
		return nil
	}
	perms, err := s.db.GetUserServerPermissions(serverID, userID)
	if err != nil {
		return os.ErrPermission
	}
	has := func(p string) bool {
		for _, granted := range perms {
			if granted == p {
				return true
			}
		}
		return false
	}
	if !has("sftp.access") {
		return os.ErrPermission
	}
	if perm == "" {
		return nil
	}
	for _, req := range strings.Split(perm, "+") {
		if !has(req) {
			return os.ErrPermission
		}
	}
	return nil
}
