package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/web"
	"gopkg.in/yaml.v3"
)

// This is deliberately address-only recovery, not certificate/proxy provisioning.
// Filesystem access to the installed configuration and database is the authority.
func cliPanelAddress(defaultConfigPath string, args []string, in io.Reader, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("panel-address", flag.ContinueOnError)
	fs.SetOutput(errOut)
	cfgPath := fs.String("config", defaultConfigPath, "Existing installed config.yaml")
	rawURL := fs.String("url", "", "New HTTPS panel origin (a hostname defaults to HTTPS)")
	confirm := fs.Bool("confirm", false, "Confirm address-only recovery and that trusted HTTPS is already configured (unverified)")
	supported := fs.Bool("supported", false, "Print recovery interface marker without reading or changing installation")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(errOut, "Unexpected arguments; use panel-address --url <https://panel.example.com> [--confirm]")
		return 1
	}
	if *supported {
		fmt.Fprintln(out, "neopanel-panel-address-v1")
		return 0
	}
	origin, host, err := normalizeRecoveryOrigin(*rawURL)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	path := resolveConfigPath(*cfgPath)
	dbPath, err := recoveryDatabasePath(path)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	fmt.Fprintf(out, "Manual / unverified panel address recovery\nConfig: %s\nDatabase: %s\nNew panel address: %s\n", path, dbPath, origin)
	fmt.Fprintln(out, "Before continuing, configure DNS and a trusted HTTPS reverse proxy for this address (HTTP is allowed only for localhost). Verify access in your browser first.")
	fmt.Fprintln(out, "This saves only the panel URL and passkey RP hostname. It does NOT verify connectivity, obtain certificates, edit config/proxy files, restart services, or change node/game addresses.")
	if !*confirm {
		fmt.Fprint(out, "Type CHANGE to confirm these prerequisites and save this address (anything else cancels): ")
		line, readErr := bufio.NewReader(io.LimitReader(in, 4096)).ReadString('\n')
		if (readErr != nil && readErr != io.EOF) || strings.TrimSpace(line) != "CHANGE" {
			fmt.Fprintln(errOut, "Cancelled; panel address unchanged. For unattended recovery, explicitly pass --confirm after configuring trusted HTTPS.")
			return 1
		}
	}
	// Do not open/migrate SQLite until the operator explicitly confirms.
	db, err := database.InitDB(dbPath)
	if err != nil {
		fmt.Fprintf(errOut, "Could not open installed panel database: %v\n", err)
		return 1
	}
	defer db.Close()
	if err := db.SetPanelPublicURL(origin, host); err != nil {
		if errors.Is(err, database.ErrPublicURLPasskeys) {
			fmt.Fprintln(errOut, "Blocked: registered passkeys belong to another hostname. No passkeys were removed and the panel address was not changed.")
			if accounts, lookupErr := db.PanelDomainPasskeyConflicts(host); lookupErr == nil {
				for _, account := range accounts {
					fmt.Fprintf(errOut, "  Account %q (ID %d): %d passkey(s) for %q\n", account.Username, account.UserID, account.Count, account.Hostname)
				}
			}
			fmt.Fprintln(errOut, "Keep or restore access at the original hostname. Have each affected user sign in, verify password/2FA recovery, then remove their old passkeys in Account settings. Retry this command only after all incompatible passkeys are removed; enroll new passkeys at the new address. There is no force/bypass option.")
		} else {
			fmt.Fprintf(errOut, "Could not save panel address: %v\n", err)
		}
		return 1
	}
	fmt.Fprintf(out, "Saved panel address: %s (database override; no service restart needed).\nHTTPS remains manually configured and unverified by this command. Open the new address and sign in again.\n", origin)
	return 0
}

// Load only role/path metadata, read-only. config.LoadConfig bootstraps directories
// and may rewrite secrets, which is inappropriate for a bounded recovery action.
// Path defaults and Windows anchoring match config.LoadConfigForRole.
func recoveryDatabasePath(path string) (string, error) {
	if web.NodeOnlyBuild || strings.EqualFold(strings.TrimSpace(os.Getenv("NEOPANEL_ROLE")), "node") {
		return "", errors.New("panel address recovery is unavailable in node mode; run it on the Master Panel host")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read existing installed config %q: %w", path, err)
	}
	var installed struct {
		Server struct {
			Role string `yaml:"role"`
		} `yaml:"server"`
		Paths struct {
			DataDir *string `yaml:"data_dir"`
			DBPath  *string `yaml:"db_path"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &installed); err != nil {
		return "", fmt.Errorf("parse installed config: %w", err)
	}
	role := strings.ToLower(strings.TrimSpace(installed.Server.Role))
	if role != "" && role != "panel" {
		return "", errors.New("panel address recovery requires a Master Panel config, not node mode or an unknown role")
	}
	dataDir := config.DefaultConfigForRole("panel").Paths.DataDir
	if installed.Paths.DataDir != nil {
		dataDir = *installed.Paths.DataDir
	}
	dbPath := filepath.Join(dataDir, "neopanel.db")
	if installed.Paths.DBPath != nil {
		dbPath = *installed.Paths.DBPath
	}
	if strings.TrimSpace(dataDir) == "" || strings.TrimSpace(dbPath) == "" {
		return "", errors.New("installed data/database paths must not be empty")
	}
	if runtime.GOOS == "windows" && !filepath.IsAbs(dbPath) {
		absConfig, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		dbPath = filepath.Join(filepath.Dir(absConfig), dbPath)
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		return "", fmt.Errorf("existing panel database %q required (use the installed config and service working directory): %w", dbPath, err)
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("installed database must be an existing regular file")
	}
	return dbPath, nil
}

// Keep the same origin rules as API canonicalPasskeyOrigin (not exported).
// No request headers, network access or shell interpretation is involved.
func normalizeRecoveryOrigin(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("enter a panel URL, for example https://panel.example.com")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" || u.Host == "" || u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") || (u.Path != "" && u.Path != "/") || u.RawPath != "" {
		return "", "", errors.New("enter the exact panel origin without credentials, path, query, or fragment")
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || len(host) > 253 || strings.HasSuffix(host, ".") || strings.HasSuffix(u.Host, ":") {
		return "", "", errors.New("invalid panel hostname")
	}
	local := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return "", "", errors.New("panel address requires HTTPS (HTTP is allowed only on localhost)")
	}
	if ip := net.ParseIP(host); ip != nil && !local {
		return "", "", errors.New("use a stable panel hostname, not a public IP address")
	}
	if net.ParseIP(host) == nil {
		for _, label := range strings.Split(host, ".") {
			if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return "", "", errors.New("invalid panel hostname")
			}
			for _, c := range label {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
					return "", "", errors.New("invalid panel hostname")
				}
			}
		}
	}
	port := u.Port()
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return "", "", errors.New("invalid panel port")
		}
	}
	if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		port = ""
	}
	authority := host
	if strings.Contains(host, ":") {
		authority = "[" + host + "]"
	}
	if port != "" {
		authority = net.JoinHostPort(host, port)
	}
	return u.Scheme + "://" + authority, host, nil
}
