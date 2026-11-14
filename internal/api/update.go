package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

const (
	defaultUpdateRepoURL = "https://github.com/Pekly/NeoPanel.git"
	defaultUpdateBranch  = "main"
	updateLogPath        = "/var/lib/neopanel/installer/install.log"
)

var (
	updateMu      sync.Mutex
	updateRunning bool
)

type UpdateConfig struct {
	RepoURL string `json:"repo_url"`
	Branch  string `json:"branch"`
}

type UpdateStatus struct {
	RepoURL         string `json:"repo_url"`
	Branch          string `json:"branch"`
	CurrentVersion  string `json:"current_version"`
	CurrentCommit   string `json:"current_commit"`
	RemoteCommit    string `json:"remote_commit"`
	RemoteMessage   string `json:"remote_message"`
	RemoteDate      string `json:"remote_date"`
	UpdateAvailable bool   `json:"update_available"`
	IsUpdating      bool   `json:"is_updating"`
	CanAutoUpdate   bool   `json:"can_auto_update"`
	Error           string `json:"error,omitempty"`
}

// parseGithubRepo extracts owner and repo name from URLs like:
// https://github.com/owner/repo or https://github.com/owner/repo.git
func parseGithubRepo(rawURL string) (owner, repo string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", "", false
	}
	path := strings.Trim(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) == 2 && validUpdateRepoPart(parts[0]) && validUpdateRepoPart(parts[1]) {
		return parts[0], parts[1], true
	}
	return "", "", false
}

// BuildCommit can be injected at compile time via:
// -ldflags "-X 'github.com/Pekly/NeoPanel/internal/api.BuildCommit=<sha>'"
var BuildCommit string

// getLocalCommit attempts to find the current git commit hash through all available sources
func getLocalCommit(db *database.DB) string {
	// 1. Injected via -ldflags at compile time
	if c := strings.TrimSpace(BuildCommit); len(c) >= 7 {
		return c
	}

	// 2. Go runtime VCS revision (embedded automatically by Go 1.18+ when built inside a git repo)
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(strings.TrimSpace(s.Value)) >= 7 {
				return strings.TrimSpace(s.Value)
			}
		}
	}

	// 3. Persistent commit marker written by installer in /etc/neopanel/current_commit
	if data, err := os.ReadFile("/etc/neopanel/current_commit"); err == nil {
		if c := strings.TrimSpace(string(data)); len(c) >= 7 {
			return c
		}
	}

	// 4. Persistent commit marker in /var/lib/neopanel/current_commit
	if data, err := os.ReadFile("/var/lib/neopanel/current_commit"); err == nil {
		if c := strings.TrimSpace(string(data)); len(c) >= 7 {
			return c
		}
	}

	// 5. Database setting saved after successful update
	if db != nil {
		if c, _ := db.GetPanelSetting("installed_git_commit"); len(strings.TrimSpace(c)) >= 7 {
			return strings.TrimSpace(c)
		}
	}

	// 6. Check the protected installer workspace if present.
	if out, err := exec.Command("git", "-C", filepath.Join(filepath.Dir(updateLogPath), "build"), "rev-parse", "HEAD").Output(); err == nil {
		if c := strings.TrimSpace(string(out)); len(c) >= 7 {
			return c
		}
	}

	// 7. Check current working directory git workspace
	cmd := exec.Command("git", "rev-parse", "HEAD")
	if out, err := cmd.Output(); err == nil {
		if c := strings.TrimSpace(string(out)); len(c) >= 7 {
			return c
		}
	}

	// 8. Fallback to reading .git/HEAD
	if headData, err := os.ReadFile(".git/HEAD"); err == nil {
		ref := strings.TrimSpace(string(headData))
		if strings.HasPrefix(ref, "ref: ") {
			refPath := strings.TrimSpace(strings.TrimPrefix(ref, "ref: "))
			if commitData, err := os.ReadFile(".git/" + refPath); err == nil {
				if c := strings.TrimSpace(string(commitData)); len(c) >= 7 {
					return c
				}
			}
		} else if len(ref) >= 7 {
			return ref
		}
	}

	return ""
}

func getEffectiveUpdateConfig(db *database.DB) UpdateConfig {
	repoURL, _ := db.GetPanelSetting("update_repo_url")
	if strings.TrimSpace(repoURL) == "" {
		repoURL = defaultUpdateRepoURL
	}
	branch, _ := db.GetPanelSetting("update_repo_branch")
	if strings.TrimSpace(branch) == "" {
		branch = defaultUpdateBranch
	}
	return UpdateConfig{
		RepoURL: strings.TrimSpace(repoURL),
		Branch:  strings.TrimSpace(branch),
	}
}

func HandleGetUpdateStatus(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := getEffectiveUpdateConfig(db)
		localCommit := getLocalCommit(db)
		updateMu.Lock()
		running := updateRunning
		updateMu.Unlock()

		status := UpdateStatus{
			RepoURL:        cfg.RepoURL,
			Branch:         cfg.Branch,
			CurrentVersion: "1.0.0",
			CurrentCommit:  localCommit,
			IsUpdating:     running,
			CanAutoUpdate:  runtime.GOOS == "linux",
		}

		owner, repo, ok := parseGithubRepo(cfg.RepoURL)
		if !ok {
			status.Error = fmt.Sprintf("Could not parse GitHub owner/repo from %q", cfg.RepoURL)
			writeJSON(w, http.StatusOK, status)
			return
		}

		// Query GitHub REST API with 6s timeout
		client := &http.Client{Timeout: 6 * time.Second}
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, url.PathEscape(cfg.Branch))
		req, err := http.NewRequestWithContext(r.Context(), "GET", apiURL, nil)
		if err != nil {
			status.Error = err.Error()
			writeJSON(w, http.StatusOK, status)
			return
		}
		req.Header.Set("User-Agent", "NeoPanel-Updater/1.0")
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			status.Error = "Failed to query GitHub API: " + err.Error()
			writeJSON(w, http.StatusOK, status)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			status.Error = fmt.Sprintf("GitHub API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			writeJSON(w, http.StatusOK, status)
			return
		}

		var ghResp struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message   string `json:"message"`
				Committer struct {
					Date string `json:"date"`
				} `json:"committer"`
			} `json:"commit"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
			status.Error = "Failed to parse GitHub response: " + err.Error()
			writeJSON(w, http.StatusOK, status)
			return
		}

		status.RemoteCommit = ghResp.SHA
		// First line of commit message only
		msg := ghResp.Commit.Message
		if idx := strings.Index(msg, "\n"); idx > 0 {
			msg = msg[:idx]
		}
		status.RemoteMessage = strings.TrimSpace(msg)
		status.RemoteDate = ghResp.Commit.Committer.Date

		if localCommit != "" && status.RemoteCommit != "" {
			loc := strings.ToLower(strings.TrimSpace(localCommit))
			rem := strings.ToLower(strings.TrimSpace(status.RemoteCommit))
			status.UpdateAvailable = (loc != rem) && !strings.HasPrefix(rem, loc) && !strings.HasPrefix(loc, rem)
		} else {
			status.UpdateAvailable = false
		}

		writeJSON(w, http.StatusOK, status)
	}
}

func HandleSaveUpdateConfig(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
			return
		}

		repoURL := strings.TrimSpace(req.RepoURL)
		if repoURL == "" {
			repoURL = defaultUpdateRepoURL
		}
		if _, _, ok := parseGithubRepo(repoURL); !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Repository must be an HTTPS github.com owner/repository URL"})
			return
		}

		branch := strings.TrimSpace(req.Branch)
		if branch == "" {
			branch = defaultUpdateBranch
		}
		if !validUpdateBranch(branch) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid Git branch or tag"})
			return
		}

		if err := db.SetPanelSetting("update_repo_url", repoURL); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save repository URL: " + err.Error()})
			return
		}
		if err := db.SetPanelSetting("update_repo_branch", branch); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save repository branch: " + err.Error()})
			return
		}

		logAudit(db, r, "admin_update_config", fmt.Sprintf("Configured GitHub update repository to %s (branch: %s)", repoURL, branch))

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message":  "Update repository settings saved successfully",
			"repo_url": repoURL,
			"branch":   branch,
		})
	}
}

func HandleApplyUpdate(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if runtime.GOOS != "linux" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Automatic panel update is only supported on Linux VPS hosts. On other platforms, please rebuild from git.",
			})
			return
		}

		updateMu.Lock()
		if updateRunning {
			updateMu.Unlock()
			writeJSON(w, http.StatusConflict, map[string]string{"error": "An update is already running in the background"})
			return
		}
		updateRunning = true
		updateMu.Unlock()

		cfg := getEffectiveUpdateConfig(db)
		logAudit(db, r, "admin_update_apply", fmt.Sprintf("Triggered panel binary update from %s (%s)", cfg.RepoURL, cfg.Branch))

		go executeSystemUpdate(db, cfg.RepoURL, cfg.Branch)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":   "updating",
			"message":  "Panel update started in background. Logs are streaming to " + updateLogPath,
			"repo_url": cfg.RepoURL,
			"branch":   cfg.Branch,
		})
	}
}

func executeSystemUpdate(db *database.DB, repoURL, branch string) {
	defer func() {
		updateMu.Lock()
		updateRunning = false
		updateMu.Unlock()
	}()

	logFile, err := openUpdateLog(updateLogPath)
	if err != nil {
		return
	}
	defer logFile.Close()

	header := fmt.Sprintf("\n\n================================================================\n  NeoPanel Update Triggered via Web Admin [%s]\n  Repository: %s\n  Branch:     %s\n================================================================\n\n",
		time.Now().Format("2006-01-02 15:04:05"), repoURL, branch)
	_, _ = logFile.WriteString(header)

	owner, repo, ok := parseGithubRepo(repoURL)
	if !ok || !validUpdateBranch(branch) {
		_, _ = logFile.WriteString("\n[NeoPanel Update ERROR] Invalid HTTPS GitHub repository or branch; update refused.\n")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	useHelper := false
	if info, err := os.Stat("/usr/local/bin/neopanel-update"); err == nil && !info.IsDir() {
		useHelper = true
	}
	var script []byte
	if !useHelper {
		endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/install.sh", owner, repo, url.PathEscape(branch))
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return
		}
		response, err := (&http.Client{Timeout: time.Minute}).Do(request)
		if err != nil {
			_, _ = fmt.Fprintf(logFile, "\n[NeoPanel Update ERROR] Download failed: %v\n", err)
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			_, _ = fmt.Fprintf(logFile, "\n[NeoPanel Update ERROR] Download HTTP %d\n", response.StatusCode)
			return
		}
		script, err = io.ReadAll(io.LimitReader(response.Body, (2<<20)+1))
		if err != nil || len(script) == 0 || len(script) > 2<<20 {
			_, _ = logFile.WriteString("\n[NeoPanel Update ERROR] Invalid or oversized installer response.\n")
			return
		}
	}
	cmd := updateCommand(ctx, repoURL, branch, script, useHelper)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("\n[NeoPanel Update ERROR] Process exited with error: %v\n", err)
		_, _ = logFile.WriteString(errMsg)
	} else {
		successMsg := fmt.Sprintf("\n[NeoPanel Update SUCCESS] Finished update successfully at %s\n", time.Now().Format("2006-01-02 15:04:05"))
		_, _ = logFile.WriteString(successMsg)
		if db != nil {
			if commitData, err := os.ReadFile("/etc/neopanel/current_commit"); err == nil {
				c := strings.TrimSpace(string(commitData))
				if len(c) >= 7 {
					_ = db.SetPanelSetting("installed_git_commit", c)
				}
			} else if commitData, err := os.ReadFile("/var/lib/neopanel/current_commit"); err == nil {
				c := strings.TrimSpace(string(commitData))
				if len(c) >= 7 {
					_ = db.SetPanelSetting("installed_git_commit", c)
				}
			}
		}
	}
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\([a-zA-Z]|\r`)

func cleanANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func parseUpdateProgress(logs string) (int, string) {
	if strings.Contains(logs, "Successfully Updated") || strings.Contains(logs, "Finished update successfully") {
		return 100, "Successfully updated"
	}

	lines := strings.Split(logs, "\n")
	rePct := regexp.MustCompile(`(\d{1,3})%\s*(?:✔|✖|⠋|⠙|⠹|⠸|⠼|⠴|⠦|⠧|⠇|⠏)?\s*(.*)`)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		matches := rePct.FindStringSubmatch(line)
		if len(matches) >= 3 {
			pct, err := strconv.Atoi(matches[1])
			if err == nil && pct >= 0 && pct <= 100 {
				step := strings.TrimSpace(matches[2])
				if step == "" {
					step = "Updating..."
				}
				return pct, step
			}
		}
	}
	return 0, ""
}

func HandleGetUpdateLogs(w http.ResponseWriter, r *http.Request) {
	updateMu.Lock()
	running := updateRunning
	updateMu.Unlock()

	var rawLines []string
	if file, err := openPreviewPath(filepath.Dir(updateLogPath), filepath.Base(updateLogPath), os.O_RDONLY); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		// Keep ring buffer of last 250 lines
		const maxLines = 250
		for scanner.Scan() {
			rawLines = append(rawLines, scanner.Text())
			if len(rawLines) > maxLines {
				rawLines = rawLines[1:]
			}
		}
	}

	cleanedLogs := cleanANSI(strings.Join(rawLines, "\n"))
	pct, step := parseUpdateProgress(cleanedLogs)
	if !running && pct == 100 {
		step = "Update complete"
	} else if running && pct < 5 {
		pct = 5
		step = "Starting update..."
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"running":  running,
		"progress": pct,
		"step":     step,
		"logs":     cleanedLogs,
		"lines":    len(rawLines),
	})
}

var updateRepoPartRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,100}$`)
var updateBranchRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,127}$`)

func validUpdateRepoPart(part string) bool {
	return part != "." && part != ".." && updateRepoPartRe.MatchString(part)
}

func validUpdateBranch(branch string) bool {
	return updateBranchRe.MatchString(branch) && !strings.Contains(branch, "..") && !strings.Contains(branch, "//") && !strings.HasSuffix(branch, "/") && !strings.HasSuffix(branch, ".lock")
}

func openUpdateLog(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if info, err := os.Lstat(dir); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return nil, fmt.Errorf("unsafe update log directory")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return nil, err
	}
	return openPreviewPath(dir, filepath.Base(path), os.O_CREATE|os.O_WRONLY|os.O_APPEND)
}

func updateCommand(ctx context.Context, repo, branch string, script []byte, helper bool) *exec.Cmd {
	var cmd *exec.Cmd
	if helper {
		cmd = exec.CommandContext(ctx, "/usr/local/bin/neopanel-update", "--repo", repo, "--branch", branch)
	} else {
		cmd = exec.CommandContext(ctx, "/bin/bash", "-s", "--", "--update", "--repo", repo, "--branch", branch)
		cmd.Stdin = bytes.NewReader(script)
	}
	cmd.Env = append(os.Environ(), "NEOPANEL_WEB_UPDATE=1", "NEOPANEL_REPO_URL="+repo, "NEOPANEL_REPO_BRANCH="+branch)
	return cmd
}
