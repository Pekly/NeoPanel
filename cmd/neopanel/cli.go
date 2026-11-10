package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
)

func runCLI(defaultConfigPath string, args []string) int {
	if len(args) == 0 {
		printCLIUsage()
		return 0
	}

	sub := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]

	switch sub {
	case "help", "-h", "--help":
		printCLIUsage()
		return 0
	case "version":
		fmt.Printf("NeoPanel v%s\n", Version)
		return 0
	case "node-token":
		return cliNodeToken(defaultConfigPath, rest)
	case "api-key", "apikey", "key":
		return cliAPIKey(defaultConfigPath, rest)
	case "node", "nodes":
		return cliNode(defaultConfigPath, rest)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
		printCLIUsage()
		return 1
	}
}

func printCLIUsage() {
	fmt.Printf(`NeoPanel v%s — Panel & Cluster Node CLI

Usage:
  neopanel [server flags]                   Start the HTTP/SFTP daemon
  neopanel [global flags] <command> ...     Run management CLI commands

Daemon Flags:
  -config <path>    Path to config.yaml (auto-detects /etc/neopanel/config.yaml)
  -host <ip>        Override HTTP listening host
  -port <port>      Override HTTP listening port
  -role <role>      Override runtime role: "panel" (Master Panel) or "node" (Cluster Node Daemon)
  -node             Shorthand for -role node

CLI Commands:
  node-token [--rotate] [--raw]
      Print (or rotate) this daemon's security.node_token (neo_node_...).
      Use this token on a Master Panel to authenticate this cluster node.

  api-key create [--name <label>] [--user <username>] [--raw]
      Generate a persistent neo_... API key directly in SQLite.

  api-key list [--user <username>]
      List existing API keys (prefixes & creation dates).

  api-key delete <key-id> [--user <username>]
      Revoke an API key by numeric ID.

  node connect --name <name> --url <http://ip:8080> [--token <neo_...>]
               [--fqdn <host>] [--location <region>] [--force]
               [--master <http://panel-url> --api-key <neo_...>] [--insecure]
      Connect/pair a remote NeoPanel node either directly into the local Master
      database OR remotely against a Master Panel via API key.

  node list [--master <url> --api-key <neo_...>] [--insecure]
      List connected cluster nodes.

  node remove <id-or-name> [--master <url> --api-key <neo_...>] [--insecure]
      Disconnect a cluster node.

  node test --url <http://ip:8080> --token <neo_...> [--insecure]
      Probe a remote node daemon API without saving it.

`, Version)
}

func cliNodeToken(defaultConfigPath string, args []string) int {
	fs := flag.NewFlagSet("node-token", flag.ContinueOnError)
	cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
	rotate := fs.Bool("rotate", false, "Generate and save a new node_token")
	raw := fs.Bool("raw", false, "Print only the raw token value")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfgPath := resolveConfigPath(*cfgFlag)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config (%s): %v\n", cfgPath, err)
		return 1
	}

	changed := false
	if *rotate || strings.TrimSpace(cfg.Security.NodeToken) == "" {
		cfg.Security.NodeToken = config.GenerateNodeToken()
		changed = true
	}

	if changed {
		if err := config.SaveConfig(cfgPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config (%s): %v\n", cfgPath, err)
			return 1
		}
	}

	if *raw {
		fmt.Println(cfg.Security.NodeToken)
		return 0
	}

	role := cfg.Server.Role
	if role == "" {
		role = "panel"
	}
	fmt.Printf("Config File : %s\n", cfgPath)
	fmt.Printf("Daemon Role : %s\n", role)
	fmt.Printf("Node Token  : %s\n", cfg.Security.NodeToken)
	if *rotate {
		fmt.Println("\n[INFO] Token rotated. Restart the service for live changes: sudo systemctl restart neopanel")
	}
	fmt.Println("\nConnect this node from your Master Panel (UI or CLI):")
	fmt.Printf("  neopanel node connect --name my-node --url http://<THIS_NODE_IP>:%d --token %s\n", cfg.Server.Port, cfg.Security.NodeToken)
	return 0
}

func cliAPIKey(defaultConfigPath string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: neopanel api-key <create|list|delete> [flags]")
		return 1
	}
	action := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]

	switch action {
	case "create", "new", "generate":
		fs := flag.NewFlagSet("api-key create", flag.ContinueOnError)
		cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
		nameFlag := fs.String("name", "CLI Node Pairing Key", "Label/description for the API key")
		userFlag := fs.String("user", "", "Username to attach key to (defaults to first admin)")
		raw := fs.Bool("raw", false, "Print only the generated neo_... token")
		if err := fs.Parse(rest); err != nil {
			return 1
		}

		cfgPath := resolveConfigPath(*cfgFlag)
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config (%s): %v\n", cfgPath, err)
			return 1
		}
		db, err := database.InitDB(cfg.Paths.DBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database (%s): %v\n", cfg.Paths.DBPath, err)
			return 1
		}
		defer db.Close()

		targetUser, err := resolveCLIUser(db, *userFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving user: %v\n", err)
			return 1
		}

		label := strings.TrimSpace(*nameFlag)
		if label == "" {
			label = "CLI Node Pairing Key"
		}

		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate random key: %v\n", err)
			return 1
		}
		rawToken := "neo_" + hex.EncodeToString(tokenBytes)
		sum := sha256.Sum256([]byte(rawToken))
		keyHash := hex.EncodeToString(sum[:])
		keyPrefix := rawToken[:12] + "..."

		created, err := db.AddAPIKey(targetUser.ID, label, keyHash, keyPrefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to store API key: %v\n", err)
			return 1
		}
		_ = db.LogActivity(targetUser.ID, "api_key_create", fmt.Sprintf("Created API key '%s' (%s) via CLI", label, keyPrefix), "127.0.0.1")

		if *raw {
			fmt.Println(rawToken)
			return 0
		}

		fmt.Printf("API Key Created Successfully!\n")
		fmt.Printf("  ID       : %d\n", created.ID)
		fmt.Printf("  Owner    : %s (%s)\n", targetUser.Username, targetUser.Role)
		fmt.Printf("  Name     : %s\n", created.Name)
		fmt.Printf("  Prefix   : %s\n", created.KeyPrefix)
		fmt.Printf("  Token    : %s\n", rawToken)
		fmt.Println("\nCopy this secret token now; only its SHA-256 hash is stored.")
		return 0

	case "list", "ls":
		fs := flag.NewFlagSet("api-key list", flag.ContinueOnError)
		cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
		userFlag := fs.String("user", "", "Filter keys by username")
		if err := fs.Parse(rest); err != nil {
			return 1
		}

		cfgPath := resolveConfigPath(*cfgFlag)
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			return 1
		}
		db, err := database.InitDB(cfg.Paths.DBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			return 1
		}
		defer db.Close()

		var users []*database.User
		if strings.TrimSpace(*userFlag) != "" {
			u, err := db.GetUserByUsername(strings.TrimSpace(*userFlag))
			if err != nil {
				fmt.Fprintf(os.Stderr, "User not found: %s\n", *userFlag)
				return 1
			}
			users = []*database.User{u}
		} else {
			users, err = db.ListUsers()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list users: %v\n", err)
				return 1
			}
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tUSER\tNAME\tPREFIX\tCREATED")
		count := 0
		for _, u := range users {
			keys, err := db.ListAPIKeys(u.ID)
			if err != nil {
				continue
			}
			for _, k := range keys {
				count++
				fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", k.ID, u.Username, k.Name, k.KeyPrefix, k.CreatedAt.Format(time.RFC3339))
			}
		}
		_ = tw.Flush()
		if count == 0 {
			fmt.Println("(no API keys stored)")
		}
		return 0

	case "delete", "revoke", "rm":
		if len(rest) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: neopanel api-key delete <key-id> [--user <username>]")
			return 1
		}
		fs := flag.NewFlagSet("api-key delete", flag.ContinueOnError)
		cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
		userFlag := fs.String("user", "", "Owner username (optional)")
		idArg, parseArgs := splitFirstPositional(rest)
		if err := fs.Parse(parseArgs); err != nil {
			return 1
		}
		if idArg == "" && fs.NArg() > 0 {
			idArg = fs.Arg(0)
		}
		keyID, err := strconv.Atoi(strings.TrimSpace(idArg))
		if err != nil || keyID <= 0 {
			fmt.Fprintln(os.Stderr, "Error: valid numeric key ID is required")
			return 1
		}

		cfgPath := resolveConfigPath(*cfgFlag)
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			return 1
		}
		db, err := database.InitDB(cfg.Paths.DBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			return 1
		}
		defer db.Close()

		var ownerID int
		if strings.TrimSpace(*userFlag) != "" {
			u, err := db.GetUserByUsername(strings.TrimSpace(*userFlag))
			if err != nil {
				fmt.Fprintf(os.Stderr, "User not found: %s\n", *userFlag)
				return 1
			}
			ownerID = u.ID
		} else {
			users, err := db.ListUsers()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to query users: %v\n", err)
				return 1
			}
			for _, u := range users {
				keys, _ := db.ListAPIKeys(u.ID)
				for _, k := range keys {
					if k.ID == keyID {
						ownerID = u.ID
						break
					}
				}
				if ownerID > 0 {
					break
				}
			}
		}

		if ownerID == 0 {
			fmt.Fprintf(os.Stderr, "API key ID %d not found\n", keyID)
			return 1
		}

		if err := db.DeleteAPIKey(ownerID, keyID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete API key: %v\n", err)
			return 1
		}
		_ = db.LogActivity(ownerID, "api_key_delete", fmt.Sprintf("Deleted API key ID %d via CLI", keyID), "127.0.0.1")
		fmt.Printf("API key ID %d revoked.\n", keyID)
		return 0

	default:
		fmt.Fprintf(os.Stderr, "Unknown api-key subcommand: %s\n", action)
		return 1
	}
}

func resolveCLIUser(db *database.DB, preferredUsername string) (*database.User, error) {
	if strings.TrimSpace(preferredUsername) != "" {
		u, err := db.GetUserByUsername(strings.TrimSpace(preferredUsername))
		if err != nil || u == nil {
			return nil, fmt.Errorf("user '%s' not found", preferredUsername)
		}
		return u, nil
	}
	users, err := db.ListUsers()
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no users exist in %s yet (complete panel setup first, or use node-token on cluster nodes)", "database")
	}
	for _, u := range users {
		if strings.EqualFold(u.Role, "admin") {
			return u, nil
		}
	}
	return users[0], nil
}

func cliNode(defaultConfigPath string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: neopanel node <connect|list|remove|test> [flags]")
		return 1
	}
	action := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]

	switch action {
	case "connect", "add", "pair":
		fs := flag.NewFlagSet("node connect", flag.ContinueOnError)
		cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
		nameFlag := fs.String("name", "", "Display name for the node (e.g. EU-Frankfurt-01)")
		urlFlag := fs.String("url", "", "Reachable base URL of the node daemon (e.g. http://192.168.1.50:8080)")
		nodeURLAlias := fs.String("node-url", "", "Alias for --url")
		tokenFlag := fs.String("token", "", "Remote node's neo_node_... or neo_... API token (defaults to local config node_token when pairing to --master)")
		nodeTokenAlias := fs.String("node-token", "", "Alias for --token")
		fqdnFlag := fs.String("fqdn", "", "Public game IP / hostname shown to players (defaults to URL host)")
		locFlag := fs.String("location", "", "Region or location label (e.g. Frankfurt, DE)")
		masterFlag := fs.String("master", "", "Master Panel base URL for remote API pairing (e.g. https://panel.example.com)")
		apiKeyFlag := fs.String("api-key", os.Getenv("NEOPANEL_API_KEY"), "Master Panel Admin API key (neo_... or Master node_token)")
		keyAlias := fs.String("key", "", "Alias for --api-key")
		forceFlag := fs.Bool("force", false, "Save node even if initial connectivity probe fails")
		skipProbe := fs.Bool("skip-probe", false, "Skip live connectivity test before saving to local database")
		insecure := fs.Bool("insecure", false, "Skip TLS certificate verification for self-signed HTTPS")
		if err := fs.Parse(rest); err != nil {
			return 1
		}

		nodeName := strings.TrimSpace(*nameFlag)
		rawURL := strings.TrimSpace(*urlFlag)
		if rawURL == "" {
			rawURL = strings.TrimSpace(*nodeURLAlias)
		}
		nodeTok := strings.TrimSpace(*tokenFlag)
		if nodeTok == "" {
			nodeTok = strings.TrimSpace(*nodeTokenAlias)
		}
		masterURL := strings.TrimSpace(*masterFlag)
		masterKey := strings.TrimSpace(*apiKeyFlag)
		if strings.TrimSpace(*keyAlias) != "" {
			masterKey = strings.TrimSpace(*keyAlias)
		}

		if nodeName == "" {
			if hn, err := os.Hostname(); err == nil && hn != "" {
				nodeName = hn
			} else {
				fmt.Fprintln(os.Stderr, "Error: --name is required")
				return 1
			}
		}

		baseURL, err := cliNormalizeURL(rawURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --url (%v)\n", err)
			return 1
		}

		fqdn := strings.TrimSpace(*fqdnFlag)
		if fqdn == "" {
			if u, err := url.Parse(baseURL); err == nil {
				fqdn = u.Hostname()
			}
		}

		// If pairing to a remote Master Panel via HTTP API:
		if masterURL != "" {
			normMaster, err := cliNormalizeURL(masterURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid --master URL (%v)\n", err)
				return 1
			}
			if masterKey == "" {
				fmt.Fprintln(os.Stderr, "Error: --api-key (or NEOPANEL_API_KEY) is required when pairing to --master")
				return 1
			}
			if nodeTok == "" {
				cfgPath := resolveConfigPath(*cfgFlag)
				if cfg, err := config.LoadConfig(cfgPath); err == nil {
					nodeTok = strings.TrimSpace(cfg.Security.NodeToken)
				}
			}
			if nodeTok == "" {
				fmt.Fprintln(os.Stderr, "Error: --token is required (no local security.node_token found)")
				return 1
			}

			payload := map[string]interface{}{
				"name":      nodeName,
				"base_url":  baseURL,
				"api_token": nodeTok,
				"fqdn":      fqdn,
				"location":  strings.TrimSpace(*locFlag),
				"upsert":    true,
			}
			bodyBytes, _ := json.Marshal(payload)
			client := cliHTTPClient(*insecure, 20*time.Second)
			req, err := http.NewRequest(http.MethodPost, normMaster+"/api/nodes", bytes.NewReader(bodyBytes))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
				return 1
			}
			req.Header.Set("Authorization", "Bearer "+masterKey)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to reach Master Panel (%s): %v\n", normMaster, err)
				return 1
			}
			defer resp.Body.Close()

			respData, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				fmt.Fprintf(os.Stderr, "Master Panel rejected pairing (HTTP %d): %s\n", resp.StatusCode, strings.TrimSpace(string(respData)))
				return 1
			}

			var view struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				BaseURL     string `json:"base_url"`
				Online      bool   `json:"online"`
				LatencyMS   int64  `json:"latency_ms"`
				ServerCount int    `json:"server_count"`
				Error       string `json:"error"`
			}
			_ = json.Unmarshal(respData, &view)
			fmt.Printf("Paired node '%s' (%s) with Master Panel %s\n", view.Name, view.ID, normMaster)
			if view.Online {
				fmt.Printf("  Reachability : ONLINE (%dms, %d server(s))\n", view.LatencyMS, view.ServerCount)
			} else if view.Error != "" {
				fmt.Printf("  Reachability : WARNING — %s\n", view.Error)
			}
			return 0
		}

		// Local Master database direct registration:
		if nodeTok == "" {
			fmt.Fprintln(os.Stderr, "Error: --token (remote node_token or neo_... API key) is required")
			return 1
		}

		if !*skipProbe {
			online, cpus, ramMB, srvCnt, latMS, probeErr := cliProbeNode(baseURL, nodeTok, *insecure)
			if !online {
				if !*forceFlag {
					fmt.Fprintf(os.Stderr, "Probe failed for %s: %s\n(Re-run with --force to register offline node anyway)\n", baseURL, probeErr)
					return 1
				}
				fmt.Fprintf(os.Stderr, "[WARN] Probe failed (%s), saving node anyway due to --force.\n", probeErr)
			} else {
				fmt.Printf("[OK] Reached %s in %dms (%d vCPU, %d MB RAM, %d server(s))\n", baseURL, latMS, cpus, ramMB, srvCnt)
			}
		}

		cfgPath := resolveConfigPath(*cfgFlag)
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config (%s): %v\n", cfgPath, err)
			return 1
		}
		db, err := database.InitDB(cfg.Paths.DBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database (%s): %v\n", cfg.Paths.DBPath, err)
			return 1
		}
		defer db.Close()

		existingList, err := db.ListNodes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying existing nodes: %v\n", err)
			return 1
		}
		for _, ex := range existingList {
			if strings.EqualFold(strings.TrimRight(ex.BaseURL, "/"), baseURL) || strings.EqualFold(ex.Name, nodeName) {
				ex.Name = nodeName
				ex.BaseURL = baseURL
				ex.APIToken = nodeTok
				ex.FQDN = fqdn
				ex.Location = strings.TrimSpace(*locFlag)
				if err := db.UpdateNode(ex); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update existing node: %v\n", err)
					return 1
				}
				_ = db.LogActivity(0, "node_update", fmt.Sprintf("Re-paired cluster node '%s' (%s) at %s via CLI", ex.Name, ex.ID, ex.BaseURL), "127.0.0.1")
				fmt.Printf("Updated existing cluster node '%s' (%s) -> %s\n", ex.Name, ex.ID, ex.BaseURL)
				return 0
			}
		}

		idBytes := make([]byte, 3)
		_, _ = rand.Read(idBytes)
		nodeID := "node-" + hex.EncodeToString(idBytes)

		node := &database.ClusterNode{
			ID:       nodeID,
			Name:     nodeName,
			BaseURL:  baseURL,
			APIToken: nodeTok,
			FQDN:     fqdn,
			Location: strings.TrimSpace(*locFlag),
		}
		if err := db.CreateNode(node); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save node: %v\n", err)
			return 1
		}
		_ = db.LogActivity(0, "node_add", fmt.Sprintf("Connected cluster node '%s' (%s) at %s via CLI", node.Name, node.ID, node.BaseURL), "127.0.0.1")
		fmt.Printf("Connected new cluster node '%s' (%s) -> %s\n", node.Name, node.ID, node.BaseURL)
		return 0

	case "list", "ls":
		fs := flag.NewFlagSet("node list", flag.ContinueOnError)
		cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
		masterFlag := fs.String("master", "", "Query remote Master Panel instead of local DB")
		apiKeyFlag := fs.String("api-key", os.Getenv("NEOPANEL_API_KEY"), "Master Panel API key")
		insecure := fs.Bool("insecure", false, "Skip TLS certificate verification")
		if err := fs.Parse(rest); err != nil {
			return 1
		}

		if strings.TrimSpace(*masterFlag) != "" {
			normMaster, err := cliNormalizeURL(*masterFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid --master URL: %v\n", err)
				return 1
			}
			if strings.TrimSpace(*apiKeyFlag) == "" {
				fmt.Fprintln(os.Stderr, "Error: --api-key is required with --master")
				return 1
			}
			client := cliHTTPClient(*insecure, 15*time.Second)
			req, _ := http.NewRequest(http.MethodGet, normMaster+"/api/nodes", nil)
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(*apiKeyFlag))
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to contact Master Panel: %v\n", err)
				return 1
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				raw, _ := io.ReadAll(resp.Body)
				fmt.Fprintf(os.Stderr, "HTTP %d: %s\n", resp.StatusCode, strings.TrimSpace(string(raw)))
				return 1
			}
			var views []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				BaseURL     string `json:"base_url"`
				FQDN        string `json:"fqdn"`
				Location    string `json:"location"`
				Online      bool   `json:"online"`
				LatencyMS   int64  `json:"latency_ms"`
				ServerCount int    `json:"server_count"`
				IsLocal     bool   `json:"is_local"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&views); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to decode response: %v\n", err)
				return 1
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tNAME\tSTATUS\tSERVERS\tENDPOINT\tLOCATION")
			for _, v := range views {
				st := "OFFLINE"
				if v.Online {
					if v.IsLocal {
						st = "ONLINE"
					} else {
						st = fmt.Sprintf("ONLINE (%dms)", v.LatencyMS)
					}
				}
				ep := v.BaseURL
				if v.IsLocal {
					ep = "(local master)"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n", v.ID, v.Name, st, v.ServerCount, ep, v.Location)
			}
			_ = tw.Flush()
			return 0
		}

		cfgPath := resolveConfigPath(*cfgFlag)
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			return 1
		}
		db, err := database.InitDB(cfg.Paths.DBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			return 1
		}
		defer db.Close()

		nodes, err := db.ListNodes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing nodes: %v\n", err)
			return 1
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tNAME\tBASE URL\tFQDN\tLOCATION")
		for _, n := range nodes {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", n.ID, n.Name, n.BaseURL, n.FQDN, n.Location)
		}
		_ = tw.Flush()
		if len(nodes) == 0 {
			fmt.Println("(no remote cluster nodes connected)")
		}
		return 0

	case "remove", "rm", "delete", "disconnect":
		if len(rest) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: neopanel node remove <node-id-or-name> [--master <url> --api-key <key>]")
			return 1
		}
		fs := flag.NewFlagSet("node remove", flag.ContinueOnError)
		cfgFlag := fs.String("config", defaultConfigPath, "Path to config.yaml")
		masterFlag := fs.String("master", "", "Master Panel base URL")
		apiKeyFlag := fs.String("api-key", os.Getenv("NEOPANEL_API_KEY"), "Master Panel API key")
		insecure := fs.Bool("insecure", false, "Skip TLS verification")
		targetArg, parseArgs := splitFirstPositional(rest)
		if err := fs.Parse(parseArgs); err != nil {
			return 1
		}
		if targetArg == "" && fs.NArg() > 0 {
			targetArg = fs.Arg(0)
		}
		target := strings.TrimSpace(targetArg)
		if target == "" || target == "local" {
			fmt.Fprintln(os.Stderr, "Error: specify a valid remote node ID or name")
			return 1
		}

		if strings.TrimSpace(*masterFlag) != "" {
			normMaster, err := cliNormalizeURL(*masterFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid --master URL: %v\n", err)
				return 1
			}
			if strings.TrimSpace(*apiKeyFlag) == "" {
				fmt.Fprintln(os.Stderr, "Error: --api-key is required with --master")
				return 1
			}
			client := cliHTTPClient(*insecure, 15*time.Second)
			req, _ := http.NewRequest(http.MethodDelete, normMaster+"/api/nodes/"+url.PathEscape(target), nil)
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(*apiKeyFlag))
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
				return 1
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				raw, _ := io.ReadAll(resp.Body)
				fmt.Fprintf(os.Stderr, "HTTP %d: %s\n", resp.StatusCode, strings.TrimSpace(string(raw)))
				return 1
			}
			fmt.Printf("Removed node '%s' from %s\n", target, normMaster)
			return 0
		}

		cfgPath := resolveConfigPath(*cfgFlag)
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			return 1
		}
		db, err := database.InitDB(cfg.Paths.DBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			return 1
		}
		defer db.Close()

		node, err := db.GetNode(target)
		if err != nil || node == nil {
			nodes, _ := db.ListNodes()
			for _, n := range nodes {
				if strings.EqualFold(n.Name, target) {
					node = n
					break
				}
			}
		}
		if node == nil {
			fmt.Fprintf(os.Stderr, "Cluster node '%s' not found\n", target)
			return 1
		}
		if err := db.DeleteNode(node.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete node: %v\n", err)
			return 1
		}
		_ = db.LogActivity(0, "node_delete", fmt.Sprintf("Disconnected cluster node '%s' (%s) via CLI", node.Name, node.ID), "127.0.0.1")
		fmt.Printf("Disconnected cluster node '%s' (%s).\n", node.Name, node.ID)
		return 0

	case "test", "probe":
		fs := flag.NewFlagSet("node test", flag.ContinueOnError)
		urlFlag := fs.String("url", "", "Node base URL")
		tokenFlag := fs.String("token", "", "Node API token or node_token")
		insecure := fs.Bool("insecure", false, "Skip TLS verification")
		if err := fs.Parse(rest); err != nil {
			return 1
		}
		baseURL, err := cliNormalizeURL(*urlFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		tok := strings.TrimSpace(*tokenFlag)
		if tok == "" {
			fmt.Fprintln(os.Stderr, "Error: --token is required")
			return 1
		}
		online, cpus, ramMB, srvCnt, latMS, probeErr := cliProbeNode(baseURL, tok, *insecure)
		if !online {
			fmt.Fprintf(os.Stderr, "OFFLINE: %s (%s)\n", baseURL, probeErr)
			return 1
		}
		fmt.Printf("ONLINE: %s (%dms) — %d vCPU, %d MB RAM, %d server(s)\n", baseURL, latMS, cpus, ramMB, srvCnt)
		return 0

	default:
		fmt.Fprintf(os.Stderr, "Unknown node subcommand: %s\n", action)
		return 1
	}
}

func splitFirstPositional(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

func cliNormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("URL must begin with http:// or https://")
	}
	if strings.TrimSpace(u.Host) == "" {
		return "", fmt.Errorf("URL must include host")
	}
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/"), nil
}

func cliHTTPClient(insecure bool, timeout time.Duration) *http.Client {
	tr := &http.Transport{}
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
}

func cliProbeNode(baseURL, token string, insecure bool) (bool, int, int64, int, int64, string) {
	client := cliHTTPClient(insecure, 6*time.Second)
	started := time.Now()

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/system", nil)
	if err != nil {
		return false, 0, 0, 0, 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, 0, 0, 0, err.Error()
	}
	defer resp.Body.Close()

	latMS := time.Since(started).Milliseconds()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, 0, 0, 0, latMS, fmt.Sprintf("authentication failed (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return false, 0, 0, 0, latMS, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	var sys struct {
		CPUs       int   `json:"cpus"`
		TotalRAMMB int64 `json:"total_ram_mb"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sys); err != nil {
		return false, 0, 0, 0, latMS, "invalid /api/system JSON response"
	}

	srvCnt := 0
	srvReq, err := http.NewRequest(http.MethodGet, baseURL+"/api/servers", nil)
	if err == nil {
		srvReq.Header.Set("Authorization", "Bearer "+token)
		srvReq.Header.Set("Accept", "application/json")
		if sResp, err := client.Do(srvReq); err == nil {
			defer sResp.Body.Close()
			if sResp.StatusCode == http.StatusOK {
				var arr []json.RawMessage
				if json.NewDecoder(sResp.Body).Decode(&arr) == nil {
					srvCnt = len(arr)
				}
			}
		}
	}

	return true, sys.CPUs, sys.TotalRAMMB, srvCnt, latMS, ""
}
