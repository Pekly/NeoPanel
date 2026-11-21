// Package daemon owns the backend shared by the CLI and desktop hosts.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/api"
	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/Pekly/NeoPanel/internal/sftp"
)

// Service is a running backend. Construct it with Start; its methods are safe
// for concurrent use. The host owns signals and must call Shutdown before exit.
type Service struct {
	url          string
	db           *database.DB
	manager      *runner.ServerManager
	sftp         *sftp.SFTPServer
	http         *http.Server
	listener     *serviceListener
	done         chan struct{}
	cleaned      chan struct{}
	once         sync.Once
	desktopToken string

	mu          sync.Mutex
	stopping    bool
	serveErr    error
	shutdownErr error
	requests    sync.WaitGroup
}

// Start binds HTTP synchronously, including when cfg.Server.Port is zero.
// assets is the same static filesystem used by the CLI; node roles never serve
// it. Callers must not mutate cfg while the service is running.
func Start(cfg *config.Config, assets fs.FS) (*Service, error) {
	return start(cfg, assets, false)
}

// StartDesktop is opt-in and loopback-only; ordinary CLI/node hosts retain DENY framing.
func StartDesktop(cfg *config.Config, assets fs.FS) (*Service, error) {
	if cfg == nil || cfg.IsNode() || cfg.Server.Host != "127.0.0.1" || cfg.SFTP.Enabled {
		return nil, errors.New("desktop requires a loopback panel with SFTP disabled")
	}
	return start(cfg, assets, true)
}

func start(cfg *config.Config, assets fs.FS, desktop bool) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("daemon: configuration is nil")
	}
	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	// Bind before creating workers so a bind failure cannot strand a manager.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not listen on %s: %w", addr, err)
	}
	log.Printf("[INFO] Connecting to SQLite database at: %s", cfg.Paths.DBPath)
	db, err := database.InitDB(cfg.Paths.DBPath)
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}
	if err := os.MkdirAll(cfg.Paths.ServersDir, 0755); err != nil {
		_ = listener.Close()
		_ = db.Close()
		return nil, fmt.Errorf("create servers directory: %w", err)
	}
	var desktopToken string
	if desktop {
		guest, guestErr := db.PrepareDesktopGuest()
		if guestErr == nil && guest != nil {
			desktopToken, guestErr = api.NewAuthHandler(db, cfg).GenerateDesktopGuestToken(guest)
		}
		if guestErr != nil {
			_ = listener.Close()
			_ = db.Close()
			return nil, fmt.Errorf("prepare desktop guest: %w", guestErr)
		}
	}
	log.Printf("[INFO] Initializing server manager with data directory: %s", cfg.Paths.ServersDir)
	manager := runner.NewServerManager(cfg.Paths.ServersDir, db)
	if cfg.Server.StartupTimeoutSec > 0 {
		manager.SetStartupTimeout(time.Duration(cfg.Server.StartupTimeoutSec) * time.Second)
	}
	if cfg.IsNode() {
		assets = nil
	}
	s := &Service{
		url: "http://" + listener.Addr().String(), db: db, manager: manager,
		done: make(chan struct{}), cleaned: make(chan struct{}),
		desktopToken: desktopToken,
	}
	s.listener = &serviceListener{Listener: listener, connections: make(map[*serviceConn]struct{})}
	router := api.NewRouter(cfg, db, manager, assets)
	if desktop {
		router.EnableDesktopFrame()
	}
	s.http = &http.Server{
		Addr: addr, Handler: s.guard(router),
		ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 120 * time.Second,
	}
	if cfg.SFTP.Enabled {
		s.sftp, err = sftp.NewSFTPServer(cfg, db)
		if err != nil {
			log.Printf("[WARN] Failed to configure SFTP server: %v", err)
		} else if err := s.sftp.Start(); err != nil {
			log.Printf("[WARN] Failed to start SFTP server: %v", err)
		}
	}
	// Binding and setup succeeded. Restore once before accepting control requests;
	// clean shutdown saves the running state through the shared manager.
	if err := manager.RestoreRunningServers(); err != nil {
		log.Printf("[WARN] Running-server recovery unavailable: %v", err)
	}

	if cfg.IsNode() || assets == nil {
		log.Printf("[SUCCESS] NeoPanel Cluster Node Daemon running at: %s", s.URL())
	} else {
		log.Printf("[SUCCESS] NeoPanel Master Panel running at: %s", s.URL())
	}
	go s.serve()
	return s, nil
}

// URL returns the actual bound HTTP address, not the configured port.
func (s *Service) URL() string { return s.url }

// TakeDesktopSession transfers the initial Guest session to the native shell
// once. It is never exposed by an HTTP route or included in a URL/log.
func (s *Service) TakeDesktopSession() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	token := s.desktopToken
	s.desktopToken = ""
	return token
}

// Done closes when HTTP serving stops. Call Shutdown to wait for cleanup too.
func (s *Service) Done() <-chan struct{} { return s.done }

// Err reports an unexpected Serve error, or nil on normal shutdown.
func (s *Service) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.serveErr
}

// ActiveServerCount counts local managed servers that are starting, online, or
// stopping. Remote nodes are not processes owned by this service. After shutdown
// it returns zero. Database failures are logged, since this API has no error result.
func (s *Service) ActiveServerCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.cleaned:
		return 0
	default:
	}
	servers, err := s.db.ListServers()
	if err != nil {
		log.Printf("[WARN] Could not count active servers: %v", err)
		return 0
	}
	count := 0
	for _, server := range servers {
		if server.NodeID != "" && server.NodeID != "local" {
			continue
		}
		switch s.manager.GetStatus(server.ID) {
		case "starting", "online", "stopping":
			count++
		}
	}
	return count
}

func (s *Service) serve() {
	err := s.http.Serve(s.listener)
	s.mu.Lock()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.serveErr = err
	}
	s.mu.Unlock()
	close(s.done)
	// Also release resources on an unexpected listener failure, without relying
	// on a host signal. The first shutdown caller owns the HTTP drain context.
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		go func() { defer cancel(); s.shutdown(ctx) }()
	})
}

// Shutdown is idempotent and waits for request handlers, ShutdownAll and DB
// closure. ctx bounds graceful HTTP draining only: on expiry connections are
// closed, but cleanup is still joined before returning. Concurrent callers all
// wait for the same cleanup and receive its result; the first context wins.
func (s *Service) Shutdown(ctx context.Context) error {
	s.once.Do(func() { go s.shutdown(ctx) })
	<-s.cleaned
	return s.shutdownErr
}

func (s *Service) shutdown(ctx context.Context) {
	// Close admission before draining existing requests or stopping processes.
	s.mu.Lock()
	s.stopping = true
	s.mu.Unlock()
	err := s.http.Shutdown(ctx)
	if err != nil {
		log.Printf("[WARN] HTTP server shutdown error: %v", err)
		_ = s.http.Close()
	}
	// net/http does not close hijacked WebSocket connections. Close those too,
	// then join handlers even when the HTTP drain deadline has expired.
	s.listener.closeConnections()
	s.requests.Wait()
	if s.sftp != nil {
		s.sftp.Stop()
	}
	log.Printf("[INFO] Stopping running Minecraft servers...")
	// ShutdownAll joins process Stop calls. Runner currently does not expose a
	// join for its scheduler/watchdog/retry workers; that limitation belongs in
	// runner rather than guessing worker completion with sleeps here.
	s.manager.ShutdownAll()
	<-s.done
	s.mu.Lock()
	s.shutdownErr = errors.Join(err, s.db.Close())
	close(s.cleaned)
	s.mu.Unlock()
}

func (s *Service) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		if s.stopping {
			s.mu.Unlock()
			http.Error(w, "NeoPanel is shutting down", http.StatusServiceUnavailable)
			return
		}
		s.requests.Add(1)
		s.mu.Unlock()
		defer s.requests.Done()
		next.ServeHTTP(w, r)
	})
}

// Track connections at the listener rather than wrapping ResponseWriter, so
// streaming, flushing and WebSocket hijacking retain their original behavior.
type serviceListener struct {
	net.Listener
	mu          sync.Mutex
	connections map[*serviceConn]struct{}
	closed      bool
}

func (l *serviceListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	c := &serviceConn{Conn: conn, owner: l}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		_ = conn.Close()
		return nil, net.ErrClosed
	}
	l.connections[c] = struct{}{}
	return c, nil
}

func (l *serviceListener) closeConnections() {
	l.mu.Lock()
	l.closed = true
	connections := make([]*serviceConn, 0, len(l.connections))
	for c := range l.connections {
		connections = append(connections, c)
	}
	l.mu.Unlock()
	for _, c := range connections {
		_ = c.Close()
	}
}

type serviceConn struct {
	net.Conn
	owner *serviceListener
}

func (c *serviceConn) Close() error {
	err := c.Conn.Close()
	c.owner.mu.Lock()
	delete(c.owner.connections, c)
	c.owner.mu.Unlock()
	return err
}
