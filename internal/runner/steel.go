package runner

import (
	"context"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Pekly/NeoPanel/internal/downloader"
)

// Steel's logger optionally prefixes the INFO level with wall time or uptime,
// and can include its Rust module path before the vanilla join/leave message.
var steelLogPrefix = regexp.MustCompile(`(?i)^(?:(?:\d{2}:\d{2}:\d{2}:\d+|\d+(?:\.\d+)?s)\s+)?(?:INFO|\[INFO\])\s+(?:steel(?:_\w+)?(?:::\w+)+\s+)?`)

func steelExecutable(dir, name string) (string, error) {
	if name == "" {
		name = downloader.ServerFileName("steel")
	}
	if filepath.Base(name) != name || name == "." || name == ".." || strings.ContainsAny(name, `/\\:`) {
		return "", fmt.Errorf("Steel executable must be a file in the server root")
	}
	path, err := filepath.Abs(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	root, err := rooted.Open(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	f, err := root.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("Steel executable %q not found or not a regular file; reinstall Steel or upload its native binary", name)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
		if _, err := root.Copy(context.Background(), name, f, 1<<30, info.Mode().Perm()|0700); err != nil {
			return "", fmt.Errorf("make Steel executable: %w", err)
		}
	}
	return path, nil
}

// SyncSteelPort changes only the server table's port, preserving other TOML settings.
// The release's complete default config is installed by the downloader because
// Steel requires more fields than just the port on first boot.
func SyncSteelPort(dir string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid Steel port: %d", port)
	}
	root, err := rooted.Open(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	path := filepath.Join("config", "config.toml")
	raw, err := root.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read Steel config/config.toml (reinstall Steel or upload its release configuration): %w", err)
	}
	lines := strings.Split(string(raw), "\n")
	inServer, found := false, false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			header, _, _ := strings.Cut(trimmed, "#")
			inServer = strings.TrimSpace(header) == "[server]"
			continue
		}
		if !inServer {
			continue
		}
		key, _, ok := strings.Cut(trimmed, "=")
		if ok && strings.TrimSpace(key) == "server_port" {
			lines[i] = fmt.Sprintf("server_port = %d", port)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Steel config/config.toml is missing [server].server_port; restore the release configuration")
	}
	updated := strings.Join(lines, "\n")
	if updated == string(raw) {
		return nil
	}
	return root.WriteFile(path, []byte(updated), 0644)
}

func serverReadyLine(serverType, line string) bool {
	if downloader.IsNativeSoftware(serverType) {
		return strings.Contains(ansiEscapeRe.ReplaceAllString(line, ""), "Started Steel Server")
	}
	return strings.Contains(line, "Done (") || strings.Contains(line, "For help, type \"help\"") || strings.Contains(line, "Listening on /")
}
