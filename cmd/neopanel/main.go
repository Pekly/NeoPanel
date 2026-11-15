package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/daemon"
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

	// The CLI owns signals; daemon owns the reusable backend lifecycle.
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdownCh)

	var staticFS fs.FS
	if !cfg.IsNode() {
		staticFS = web.GetFS()
	}
	service, err := daemon.Start(cfg, staticFS)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	select {
	case sig := <-shutdownCh:
		log.Printf("[INFO] Received signal %v. Gracefully shutting down NeoPanel...", sig)
	case <-service.Done():
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := service.Shutdown(ctx); err != nil {
		log.Printf("[WARN] Backend shutdown error: %v", err)
	}
	if err := service.Err(); err != nil {
		log.Fatalf("[FATAL] Server terminated unexpectedly: %v", err)
	}
	log.Printf("[INFO] Shutdown complete. Goodbye!")
}
