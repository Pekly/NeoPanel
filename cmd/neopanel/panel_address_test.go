package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/web"
)

func TestNormalizeRecoveryOrigin(t *testing.T) {
	for _, raw := range []string{"", "http://panel.example.com", "https://user:secret@panel.example.com", "https://panel.example.com/path", "https://panel.example.com?x=1", "https://panel.example.com?", "https://panel.example.com#", "https://*.example.com", "https://bad_host.example", "https://panel.example.com.", "https://panel.example.com:", "https://panel.example.com:0", "https://panel.example.com:65536", "https://panel.example.com:bad", "https://192.0.2.1", "https://[2001:db8::1]", "https://panel.example.com/%2f", "https://panel.example.com/;touch /tmp/no", "https://$(id).example"} {
		t.Run(raw, func(t *testing.T) {
			if _, _, err := normalizeRecoveryOrigin(raw); err == nil {
				t.Fatalf("accepted invalid URL %q", raw)
			}
		})
	}
	for _, tc := range []struct{ raw, origin, host string }{
		{" Panel.Example.com ", "https://panel.example.com", "panel.example.com"},
		{"https://PANEL.example.com:443/", "https://panel.example.com", "panel.example.com"},
		{"https://panel.example.com:8443", "https://panel.example.com:8443", "panel.example.com"},
		{"http://localhost:8080", "http://localhost:8080", "localhost"},
		{"http://127.0.0.1:80/", "http://127.0.0.1", "127.0.0.1"},
		{"http://[::1]:8080", "http://[::1]:8080", "::1"},
	} {
		origin, host, err := normalizeRecoveryOrigin(tc.raw)
		if err != nil || origin != tc.origin || host != tc.host {
			t.Errorf("%q: got %q, %q, %v", tc.raw, origin, host, err)
		}
	}
}

func recoveryFixture(t *testing.T, role string) (string, string) {
	t.Helper()
	t.Setenv("NEOPANEL_ROLE", "")
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "panel.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SetPanelPublicURL("https://old.example.com", "old.example.com"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	// No JWT secret: normal LoadConfig would rewrite this, recovery must not.
	body := fmt.Sprintf("# preserve this config verbatim\nserver:\n  role: %s\n  public_url: https://config.example.com\npaths:\n  data_dir: %q\n  db_path: %q\n", role, dir, dbPath)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path, dbPath
}

func invokeRecovery(path string, args []string, input string) (int, string) {
	var output bytes.Buffer
	code := cliPanelAddress(path, args, strings.NewReader(input), &output, &output)
	return code, output.String()
}

func assertRecoverySetting(t *testing.T, dbPath, want, rpID string) {
	t.Helper()
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, err := db.GetPanelSetting(database.PanelPublicURLKey)
	if err != nil || got != want {
		t.Fatalf("saved URL = %q, %v; want %q", got, err, want)
	}
	got, err = db.PasskeyRPID()
	if err != nil || got != rpID {
		t.Fatalf("RP ID = %q, %v; want %q", got, err, rpID)
	}
}

func TestPanelAddressCancellationAndInvalidInput(t *testing.T) {
	if web.NodeOnlyBuild {
		t.Skip("panel-only recovery")
	}
	for _, tc := range []struct{ name, url, input, message string }{
		{"invalid", "http://new.example.com", "CHANGE\n", "requires HTTPS"},
		{"cancel", "https://new.example.com", "no\n", "Cancelled"},
		{"eof", "https://new.example.com", "", "Cancelled"},
		{"generic yes is not confirmation", "https://new.example.com", "yes\n", "Cancelled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, dbPath := recoveryFixture(t, "panel")
			before, _ := os.ReadFile(path)
			dbBefore, _ := os.ReadFile(dbPath)
			code, output := invokeRecovery(path, []string{"--url", tc.url}, tc.input)
			if code == 0 || !strings.Contains(output, tc.message) {
				t.Fatalf("code %d: %s", code, output)
			}
			after, _ := os.ReadFile(path)
			dbAfter, _ := os.ReadFile(dbPath)
			if !bytes.Equal(before, after) || !bytes.Equal(dbBefore, dbAfter) {
				t.Fatal("cancelled/invalid recovery changed files")
			}
			assertRecoverySetting(t, dbPath, "https://old.example.com", "old.example.com")
		})
	}
}

func TestPanelAddressConfirmationSavesOverride(t *testing.T) {
	if web.NodeOnlyBuild {
		t.Skip("panel-only recovery")
	}
	for _, explicitFlag := range []bool{false, true} {
		t.Run(fmt.Sprint(explicitFlag), func(t *testing.T) {
			path, dbPath := recoveryFixture(t, "panel")
			before, _ := os.ReadFile(path)
			args := []string{"--url", "New.Example.com:443/"}
			input := "CHANGE\n"
			if explicitFlag {
				args = append(args, "--confirm")
				input = ""
			}
			code, output := invokeRecovery(path, args, input)
			if code != 0 || !strings.Contains(output, "Saved panel address: https://new.example.com") || !strings.Contains(output, "unverified") {
				t.Fatalf("code %d: %s", code, output)
			}
			after, _ := os.ReadFile(path)
			if !bytes.Equal(before, after) {
				t.Fatal("recovery rewrote config")
			}
			assertRecoverySetting(t, dbPath, "https://new.example.com", "new.example.com")
		})
	}
}

func TestPanelAddressNodeRejection(t *testing.T) {
	for _, role := range []string{"node", "panel"} {
		t.Run(role, func(t *testing.T) {
			path, dbPath := recoveryFixture(t, role)
			if role == "panel" {
				t.Setenv("NEOPANEL_ROLE", "node")
			}
			before, _ := os.ReadFile(path)
			code, output := invokeRecovery(path, []string{"--url", "https://new.example.com", "--confirm"}, "")
			if code == 0 || !strings.Contains(output, "node mode") {
				t.Fatalf("code %d: %s", code, output)
			}
			after, _ := os.ReadFile(path)
			if !bytes.Equal(before, after) {
				t.Fatal("node config changed")
			}
			assertRecoverySetting(t, dbPath, "https://old.example.com", "old.example.com")
		})
	}
	if web.NodeOnlyBuild {
		path, _ := recoveryFixture(t, "panel")
		code, output := invokeRecovery(path, []string{"--url", "https://new.example.com", "--confirm"}, "")
		if code == 0 || !strings.Contains(output, "node mode") {
			t.Fatalf("node-only binary accepted recovery: %s", output)
		}
	}
}

func TestPanelAddressPasskeysGuardAndSameHost(t *testing.T) {
	if web.NodeOnlyBuild {
		t.Skip("panel-only recovery")
	}
	path, dbPath := recoveryFixture(t, "panel")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	user, err := db.CreateUser("affected-user", "password-hash", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.AddPasskey(user.ID, user.PasswordHash, "old.example.com", "security key", []byte("credential"), "{}"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	code, output := invokeRecovery(path, []string{"--url", "https://new.example.com", "--confirm"}, "")
	if code == 0 || !strings.Contains(output, "affected-user") || !strings.Contains(output, "Account settings") || !strings.Contains(output, "no force/bypass") {
		t.Fatalf("code %d: %s", code, output)
	}
	assertRecoverySetting(t, dbPath, "https://old.example.com", "old.example.com")
	code, output = invokeRecovery(path, []string{"--url", "https://old.example.com:8443", "--confirm"}, "")
	if code != 0 {
		t.Fatalf("same-host change rejected: %s", output)
	}
	assertRecoverySetting(t, dbPath, "https://old.example.com:8443", "old.example.com")
	db, err = database.InitDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	keys, err := db.ListPasskeys(user.ID)
	if err != nil || len(keys) != 1 {
		t.Fatalf("passkeys altered: %v, %v", keys, err)
	}
}

func TestPanelAddressProbeAndMissingInstallation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")
	code, output := invokeRecovery(path, []string{"--supported"}, "")
	if code != 0 || output != "neopanel-panel-address-v1\n" {
		t.Fatalf("probe: %d %s", code, output)
	}
	code, _ = invokeRecovery(path, []string{"--url", "https://new.example.com", "--confirm"}, "")
	if code == 0 {
		t.Fatal("accepted missing installation")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("created missing config")
	}
}

func TestPanelAddressInstalledPaths(t *testing.T) {
	if web.NodeOnlyBuild {
		t.Skip("panel-only recovery")
	}
	t.Setenv("NEOPANEL_ROLE", "")
	dir := t.TempDir()
	t.Setenv("NEOPANEL_DATA_DIR", dir)
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("server:\n  role: panel\n"), 0600); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dir, "neopanel.db")
	if _, err := recoveryDatabasePath(path); err == nil {
		t.Fatal("accepted a missing database")
	}
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatal("created a missing database")
	}
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	got, err := recoveryDatabasePath(path)
	if err != nil || got != dbPath {
		t.Fatalf("NEOPANEL_DATA_DIR ignored: %q, %v", got, err)
	}
	code, output := invokeRecovery(filepath.Join(dir, "missing.yaml"), []string{"--config", path, "--url", "http://localhost:8080", "--confirm"}, "")
	if code != 0 {
		t.Fatalf("explicit config override failed: %s", output)
	}
	assertRecoverySetting(t, dbPath, "http://localhost:8080", "localhost")
}
