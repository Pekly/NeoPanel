package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

const (
	defaultUpdateRepoURL = "https://github.com/Pekly/NeoPanel.git"
	defaultUpdateBranch  = "main"
	updateLogPath        = "/tmp/neopanel-install.log"
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
	if err != nil {
		return "", "", false
	}
	path := strings.Trim(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1], true
	}
	return "", "", false
}

// getLocalCommit attempts to find the current git commit hash if running in a git workspace
func getLocalCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	// Fallback to reading from .git/HEAD directly if git binary is not installed
	if headData, err := os.ReadFile(".git/HEAD"); err == nil {
		ref := strings.TrimSpace(string(headData))
		if strings.HasPrefix(ref, "ref: ") {
			refPath := strings.TrimSpace(strings.TrimPrefix(ref, "ref: "))
			if commitData, err := os.ReadFile(".git/" + refPath); err == nil {
				return strings.TrimSpace(string(commitData))
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
		localCommit := getLocalCommit()

		status := UpdateStatus{
			RepoURL:        cfg.RepoURL,
			Branch:         cfg.Branch,
			CurrentVersion: "1.0.0",
			CurrentCommit:  localCommit,
			IsUpdating:     updateRunning,
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
			status.UpdateAvailable = !strings.HasPrefix(status.RemoteCommit, localCommit) && !strings.HasPrefix(localCommit, status.RemoteCommit)
		} else if status.RemoteCommit != "" {
			status.UpdateAvailable = true
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
		if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Repository URL must start with http:// or https://"})
			return
		}

		branch := strings.TrimSpace(req.Branch)
		if branch == "" {
			branch = defaultUpdateBranch
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

		go executeSystemUpdate(cfg.RepoURL, cfg.Branch)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":   "updating",
			"message":  "Panel update started in background. Logs are streaming to " + updateLogPath,
			"repo_url": cfg.RepoURL,
			"branch":   cfg.Branch,
		})
	}
}

func executeSystemUpdate(repoURL, branch string) {
	defer func() {
		updateMu.Lock()
		updateRunning = false
		updateMu.Unlock()
	}()

	logFile, err := os.OpenFile(updateLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	header := fmt.Sprintf("\n\n================================================================\n  NeoPanel Update Triggered via Web Admin [%s]\n  Repository: %s\n  Branch:     %s\n================================================================\n\n",
		time.Now().Format("2006-01-02 15:04:05"), repoURL, branch)
	_, _ = logFile.WriteString(header)

	owner, repo, ok := parseGithubRepo(repoURL)
	rawInstallURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/install.sh", owner, repo, branch)
	if !ok {
		rawInstallURL = "https://raw.githubusercontent.com/Pekly/NeoPanel/main/install.sh"
	}

	// Prefer /usr/local/bin/neopanel-update if present; otherwise stream install.sh via curl
	var scriptCmd string
	if _, err := os.Stat("/usr/local/bin/neopanel-update"); err == nil {
		scriptCmd = fmt.Sprintf("NEOPANEL_REPO_URL=%q NEOPANEL_REPO_BRANCH=%q /usr/local/bin/neopanel-update --repo %q --branch %q",
			repoURL, branch, repoURL, branch)
	} else {
		scriptCmd = fmt.Sprintf("curl -sSL %q | NEOPANEL_REPO_URL=%q NEOPANEL_REPO_BRANCH=%q bash -s -- --update --repo %q --branch %q",
			rawInstallURL, repoURL, branch, repoURL, branch)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", scriptCmd)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("\n[NeoPanel Update ERROR] Process exited with error: %v\n", err)
		_, _ = logFile.WriteString(errMsg)
	} else {
		successMsg := fmt.Sprintf("\n[NeoPanel Update SUCCESS] Finished update successfully at %s\n", time.Now().Format("2006-01-02 15:04:05"))
		_, _ = logFile.WriteString(successMsg)
	}
}

func HandleGetUpdateLogs(w http.ResponseWriter, r *http.Request) {
	updateMu.Lock()
	running := updateRunning
	updateMu.Unlock()

	var lines []string
	if file, err := os.Open(updateLogPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		// Keep ring buffer of last 250 lines
		const maxLines = 250
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
			if len(lines) > maxLines {
				lines = lines[1:]
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"running": running,
		"logs":    strings.Join(lines, "\n"),
		"lines":   len(lines),
	})
}
