package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Pekly/NeoPanel/internal/api"
	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/Pekly/NeoPanel/internal/sftp"
	"github.com/Pekly/NeoPanel/web"
)

const Version = "1.0.0"

const banner = `
  _   _             ____                  _ 
 | \ | | ___  ___  |  _ \ __ _ _ __   ___| |
 |  \| |/ _ \/ _ \ | |_) / _` + "`" + ` | '_ \ / _ \ |
 | |\  |  __/ (_) ||  __/ (_| | | | |  __/ |
 |_| \_|\___|\___/ |_|   \__,_|_| |_|\___|_|
   >> Ultra-Lightweight Minecraft Server Manager (v%s) <<
`

func resolveConfigPath(flagVal string) string {
	role := "panel"
	if web.NodeOnlyBuild {
		role = "node"
	}
	return resolveConfigPathForRole(flagVal, role)
}

func resolveConfigPathForRole(flagVal, role string) string {
	if strings.TrimSpace(flagVal) != "" {
		return strings.TrimSpace(flagVal)
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(config.DefaultConfigForRole(role).Paths.DataDir, "config.yaml")
	}
	if _, err := os.Stat("/etc/neopanel/config.yaml"); err == nil {
		return "/etc/neopanel/config.yaml"
	}
	return "./config.yaml"
}

func main() {
	configPath := flag.String("config", "", "Path to YAML configuration file")
	portFlag := flag.Int("port", 0, "Override HTTP listening port")
	hostFlag := flag.String("host", "", "Override HTTP listening host")
	roleFlag := flag.String("role", "", "Override server role ('panel' or 'node')")
	nodeOnlyFlag := flag.Bool("node", false, "Run as a NeoPanel Cluster Node Daemon")
	showVersion := flag.Bool("version", false, "Show version and exit")
	initOnly := flag.Bool("init", false, "Create/validate configuration and exit without starting servers")
	flag.Parse()

	if *showVersion {
		fmt.Printf("NeoPanel v%s\n", Version)
		return
	}

	requestedRole := strings.ToLower(strings.TrimSpace(*roleFlag))
	if requestedRole != "" && requestedRole != "panel" && requestedRole != "node" {
		log.Fatalf("Invalid role %q", requestedRole)
	}
	if web.NodeOnlyBuild && requestedRole == "panel" {
		log.Fatal("This is a node-only build; use the panel binary to host the web UI")
	}
	if *nodeOnlyFlag || web.NodeOnlyBuild {
		requestedRole = "node"
	}
	pathRole := requestedRole
	if pathRole == "" {
		pathRole = strings.ToLower(strings.TrimSpace(os.Getenv("NEOPANEL_ROLE")))
	}
	targetConfig := resolveConfigPathForRole(*configPath, pathRole)

	// Handle CLI subcommands (e.g. `neopanel node-token`, `neopanel api-key`, `neopanel node`)
	if args := flag.Args(); len(args) > 0 {
		// CLI helpers use LoadConfig; preserve explicit flag/build role precedence.
		if requestedRole != "" {
			_ = os.Setenv("NEOPANEL_ROLE", requestedRole)
		}
		os.Exit(runCLI(targetConfig, args))
	}
	fmt.Printf(banner, Version)

	cfg, err := config.LoadConfigForRole(targetConfig, requestedRole)
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	if *portFlag > 0 {
		cfg.Server.Port = *portFlag
	}
	if *hostFlag != "" {
		cfg.Server.Host = *hostFlag
	}
	if *portFlag < 0 || *portFlag > 65535 {
		log.Fatal("HTTP port must be between 1 and 65535")
	}
	if *initOnly {
		if err := config.SaveConfig(cfg.FilePath, cfg); err != nil {
			log.Fatalf("Could not save configuration: %v", err)
		}
		fmt.Printf("Configuration ready: %s\nRole: %s\nData: %s\n", cfg.FilePath, cfg.Server.Role, cfg.Paths.DataDir)
		return
	}
	if runtime.GOOS == "windows" {
		log.Printf("[INFO] Windows %s: local workloads run without filesystem/network sandboxing; use isolated Linux nodes for untrusted workloads.", runtime.GOARCH)
		log.Printf("[INFO] Configuration: %s", cfg.FilePath)
	}

	// 1. Initialize SQLite Database
	log.Printf("[INFO] Connecting to SQLite database at: %s", cfg.Paths.DBPath)
	db, err := database.InitDB(cfg.Paths.DBPath)
	if err != nil {
		log.Fatalf("[FATAL] Database initialization failed: %v", err)
	}
	defer db.Close()

	// 2. Initialize Process / Server Manager
	log.Printf("[INFO] Initializing server manager with data directory: %s", cfg.Paths.ServersDir)
	manager := runner.NewServerManager(cfg.Paths.ServersDir, db)
	manager.SetAllowInsecureGameFallback(cfg.Security.AllowInsecureGameFallback)
	if cfg.Server.StartupTimeoutSec > 0 {
		manager.SetStartupTimeout(time.Duration(cfg.Server.StartupTimeoutSec) * time.Second)
	}

	// 3. Mount Router (cluster node daemons expose daemon & control APIs only)
	var staticFS fs.FS
	if !cfg.IsNode() {
		staticFS = web.GetFS()
	}
	router := api.NewRouter(cfg, db, manager, staticFS)

	// 4. Initialize & Start Built-in SFTP Server (Port 2022)
	var sftpServer *sftp.SFTPServer
	if cfg.SFTP.Enabled {
		var err error
		sftpServer, err = sftp.NewSFTPServer(cfg, db)
		if err != nil {
			log.Printf("[WARN] Failed to configure SFTP server: %v", err)
		} else {
			if err := sftpServer.Start(); err != nil {
				log.Printf("[WARN] Failed to start SFTP server: %v", err)
			}
		}
	}

	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    32 << 10,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		if sftpServer != nil {
			sftpServer.Stop()
		}
		manager.ShutdownAll()
		log.Fatalf("[FATAL] Could not listen on %s: %v", addr, err)
	}
	addr = listener.Addr().String()

	// 5. Graceful Shutdown listener
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdownCh)
	shutdownDone := make(chan struct{})

	go func() {
		sig := <-shutdownCh
		log.Printf("[INFO] Received signal %v. Gracefully shutting down NeoPanel...", sig)

		// Create timeout context for HTTP shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if sftpServer != nil {
			sftpServer.Stop()
		}

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("[WARN] HTTP server shutdown error: %v", err)
			_ = srv.Close()
		}

		log.Printf("[INFO] Stopping running Minecraft servers...")
		manager.ShutdownAll()

		log.Printf("[INFO] Shutdown complete. Goodbye!")
		close(shutdownDone)
	}()

	// Restore only after successful binding/setup; failed startup leaves the
	// previous snapshot available for a later successful launch.
	if err := manager.RestoreRunningServers(); err != nil {
		log.Printf("[WARN] Running-server recovery unavailable: %v", err)
	}
	manager.StartWakeListeners()

	// 6. Start HTTP Server
	if cfg.IsNode() || staticFS == nil {
		log.Printf("[SUCCESS] NeoPanel Cluster Node Daemon running at: http://%s", addr)
	} else {
		log.Printf("[SUCCESS] NeoPanel Master Panel running at: http://%s", addr)
	}
	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		if sftpServer != nil {
			sftpServer.Stop()
		}
		manager.ShutdownAll()
		log.Fatalf("[FATAL] Server terminated unexpectedly: %v", err)
	}
	// Serve returns as soon as the listener closes, before game shutdown finishes.
	// Keep the process alive until the graceful shutdown goroutine has completed.
	<-shutdownDone
}
