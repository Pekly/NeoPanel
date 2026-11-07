package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pekly/NeoPanel/internal/api"
	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
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

func main() {
	configPath := flag.String("config", "", "Path to YAML configuration file")
	portFlag := flag.Int("port", 0, "Override HTTP listening port")
	hostFlag := flag.String("host", "", "Override HTTP listening host")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("NeoPanel v%s\n", Version)
		return
	}

	fmt.Printf(banner, Version)

	// Determine configuration file path
	targetConfig := *configPath
	if targetConfig == "" {
		if _, err := os.Stat("/etc/neopanel/config.yaml"); err == nil {
			targetConfig = "/etc/neopanel/config.yaml"
		} else {
			targetConfig = "./config.yaml"
		}
	}

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

	// 3. Mount Embedded Frontend & Router
	staticFS := web.GetFS()
	router := api.NewRouter(cfg, db, manager, staticFS)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 4. Graceful Shutdown listener
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-shutdownCh
		log.Printf("[INFO] Received signal %v. Gracefully shutting down NeoPanel...", sig)

		// Create timeout context for HTTP shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("[WARN] HTTP server shutdown error: %v", err)
		}

		log.Printf("[INFO] Stopping running Minecraft servers...")
		manager.ShutdownAll()

		log.Printf("[INFO] Shutdown complete. Goodbye!")
		os.Exit(0)
	}()

	// 5. Start HTTP Server
	log.Printf("[SUCCESS] NeoPanel web dashboard running at: http://%s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server terminated unexpectedly: %v", err)
	}
}

