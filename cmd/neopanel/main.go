package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	if strings.TrimSpace(flagVal) != "" {
		return strings.TrimSpace(flagVal)
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
	nodeOnlyFlag := flag.Bool("node", false, "Run in headless Node Daemon mode (no web panel UI)")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("NeoPanel v%s\n", Version)
		return
	}

	targetConfig := resolveConfigPath(*configPath)

	// Handle headless CLI subcommands (e.g. `neopanel node-token`, `neopanel api-key`, `neopanel node`)
	if args := flag.Args(); len(args) > 0 {
		os.Exit(runCLI(targetConfig, args))
	}

	fmt.Printf(banner, Version)

	cfg, err := config.LoadConfig(targetConfig)
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	if *portFlag > 0 {
		cfg.Server.Port = *portFlag
	}
	if *hostFlag != "" {
		cfg.Server.Host = *hostFlag
	}
	if *nodeOnlyFlag {
		cfg.Server.Role = "node"
	} else if r := strings.ToLower(strings.TrimSpace(*roleFlag)); r == "node" || r == "panel" {
		cfg.Server.Role = r
	}
	if cfg.IsNode() && strings.TrimSpace(cfg.Security.NodeToken) == "" {
		cfg.Security.NodeToken = config.GenerateNodeToken()
		if cfg.FilePath != "" {
			_ = config.SaveConfig(cfg.FilePath, cfg)
		}
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
	if cfg.Server.StartupTimeoutSec > 0 {
		manager.SetStartupTimeout(time.Duration(cfg.Server.StartupTimeoutSec) * time.Second)
	}

	// 3. Mount Router (omit embedded web UI completely in node builds or headless node role)
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

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Graceful Shutdown listener
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

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
		}

		log.Printf("[INFO] Stopping running Minecraft servers...")
		manager.ShutdownAll()

		log.Printf("[INFO] Shutdown complete. Goodbye!")
		os.Exit(0)
	}()

	// 6. Start HTTP Server
	if cfg.IsNode() || staticFS == nil {
		log.Printf("[SUCCESS] NeoPanel Node Daemon (API-only, no web UI) running at: http://%s", addr)
	} else {
		log.Printf("[SUCCESS] NeoPanel web dashboard running at: http://%s", addr)
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server terminated unexpectedly: %v", err)
	}
}
