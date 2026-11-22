package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

type pluginUpdateResult struct {
	Filename   string `json:"filename"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	BackupPath string `json:"backup_path"`
}
type pluginUpdateResponse struct {
	Results         []pluginUpdateResult `json:"results"`
	RestartRequired bool                 `json:"restart_required"`
}

// Account conservatively for retained originals and staging, including existing
// backups. This is a bounded application-level admission check, not an OS quota.
func pluginUpdateDiskBudget(ctx context.Context, root *rooted.Root, limitMB, additional int64) error {
	if limitMB <= 0 {
		return nil
	}
	limit := limitMB * (1 << 20)
	if limit < 0 || additional > limit {
		return fmt.Errorf("plugin update exceeds assigned disk budget")
	}
	used, entries := int64(0), 0
	err := root.Walk(".", func(_ string, st os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		entries++
		if entries > 100000 {
			return fmt.Errorf("disk budget scan limit exceeded")
		}
		if st.Mode().IsRegular() {
			if st.Size() > limit-used {
				return fmt.Errorf("assigned disk budget exceeded")
			}
			used += st.Size()
		}
		return nil
	})
	if err != nil || additional > limit-used {
		return fmt.Errorf("cannot fit plugin staging and retained backups within assigned disk budget")
	}
	return nil
}

// Retain the original before replacing it. Publication/rollback is a short,
// non-cancellable critical section under the exclusive mutation and power gates.
// Link refuses to overwrite a preexisting backup or rollback destination.
func publishPluginUpdate(live, staged, backup string, link func(string, string) error, remove func(string) error, rename func(string, string) error) (bool, error) {
	if err := link(live, backup); err != nil {
		return false, fmt.Errorf("cannot retain original plugin backup")
	}
	if err := remove(live); err != nil {
		return false, fmt.Errorf("cannot replace original plugin; backup retained")
	}
	if err := rename(staged, live); err != nil {
		if err := link(backup, live); err != nil {
			return true, fmt.Errorf("publication and rollback failed; manual repair required before starting")
		}
		return false, fmt.Errorf("publication failed; original plugin restored")
	}
	return false, nil
}

func (h *PluginHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	srv, err := h.db.GetServer(id)
	if err != nil {
		writeError(w, 404, "Server not found")
		return
	}
	if rejectSteelPlugins(srv, w) {
		return
	}
	var req struct {
		Items   []pluginIdentityRequest `json:"items"`
		Confirm bool                    `json:"confirm"`
	}
	if decodePluginRequest(w, r, &req) != nil || !req.Confirm || len(req.Items) == 0 || len(req.Items) > 10 {
		writeError(w, 400, "Explicit confirmation and 1-10 selected plugins are required")
		return
	}
	seen := map[string]bool{}
	for _, item := range req.Items {
		// Conservative cross-platform uniqueness: never touch the same Windows file twice.
		key := strings.ToLower(item.Filename)
		if !item.valid() || !downloader.ValidModrinthIdentifier(item.VersionID) || seen[key] {
			writeError(w, 400, "Invalid or duplicate plugin update selection")
			return
		}
		seen[key] = true
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	ctx, release, err := h.manager.BeginPluginUpdateOperation(ctx, id)
	if err != nil {
		writeError(w, 409, "Stop the server and wait for active file or power operations before updating plugins")
		return
	}
	defer release()
	ctx = downloader.ModrinthMetadataBudgetContext(ctx)
	root, err := rooted.Open(h.manager.GetWorkingDir(id))
	if err != nil {
		writeError(w, 409, "Server files unavailable")
		return
	}
	defer root.Close()
	// Re-read software while gated rather than trust a preview or request field.
	srv, err = h.db.GetServer(id)
	if err != nil {
		writeError(w, 409, "Server changed; rescan plugins")
		return
	}
	folder, _ := pluginFolder(srv.Type)
	dir, err := root.Sub(folder)
	if err != nil {
		writeError(w, 409, "Plugin directory unavailable")
		return
	}
	defer dir.Close()
	hashes := []string{}
	originalBytes := int64(0)
	for _, item := range req.Items {
		hashBudget := pluginFileLimit
		hash, st, err := pluginDigest(ctx, dir, item.Filename, &hashBudget)
		if err != nil || hash != item.SHA512 {
			writeError(w, 409, "A selected plugin changed; rescan before updating")
			return
		}
		hashes = append(hashes, hash)
		originalBytes += st.Size()
	}
	matches, err := downloader.LookupModrinthHashes(ctx, hashes)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Hash lookup unavailable; no plugins changed", "code": "plugin_catalog_unavailable"})
		return
	}
	targets := make([]*downloader.ModrinthVersion, len(req.Items))
	response := pluginUpdateResponse{Results: make([]pluginUpdateResult, len(req.Items))}
	downloadBytes := int64(0)
	for i, item := range req.Items {
		var installed *downloader.ModrinthVersion
		projectID := ""
		if v, ok := matches[item.SHA512]; ok {
			installed = &v
			projectID = v.ProjectID
		} else {
			projectID, err = h.db.ConfirmedPluginSource(id, item.Filename, item.SHA512)
			if err != nil {
				writeError(w, 500, "Cannot read confirmed plugin association")
				return
			}
		}
		if projectID == "" || projectID != item.ProjectID {
			writeError(w, 409, "Selected project is not identified for this exact plugin hash; rescan or explicitly identify it")
			return
		}
		target, err := downloader.ResolveModrinthVersionContext(ctx, projectID, srv.Type, item.VersionID)
		if err != nil || target == nil {
			writeError(w, 400, "Selected release cannot be verified for this project and server software; no plugins changed")
			return
		}
		response.Results[i] = pluginUpdateResult{Filename: item.Filename, Status: "skipped", Message: "No verified newer release; unchanged files and downgrades are skipped."}
		if !pluginNewer(installed, target, item.SHA512) {
			continue
		}
		size, err := downloader.ModrinthUpdateSize(target)
		if err != nil || size > pluginBatchLimit-downloadBytes {
			writeError(w, 400, "Selected downloads exceed the bounded batch budget or have invalid size")
			return
		}
		downloadBytes += size
		targets[i] = target
	}
	if downloadBytes == 0 {
		writeJSON(w, 200, response)
		return
	}
	if err := pluginUpdateDiskBudget(ctx, root, srv.DiskLimitMB, downloadBytes+originalBytes); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := ctx.Err(); err != nil {
		writeError(w, 408, "Plugin update preflight timed out; no plugins changed")
		return
	}
	// All identities/targets are now preflighted; only isolated staging writes follow.
	stageName, stage, err := root.TempDir()
	if err != nil {
		writeError(w, 500, "Cannot stage plugin updates")
		return
	}
	defer root.RemoveAll(stageName)
	defer stage.Close()
	staged := make([]string, len(req.Items))
	downloadBudget := pluginBatchLimit
	for i, target := range targets {
		if target == nil {
			continue
		}
		result := &response.Results[i]
		result.Status = "failed"
		result.Message = "Download failed; original plugin retained."
		childName, child, err := stage.TempDir()
		if err != nil {
			continue
		}
		name, err := downloader.StageModrinthUpdate(ctx, target, child, &downloadBudget)
		if err == nil {
			// Validate the archive without extraction or executing any plugin code.
			f, e := child.Open(name)
			if e != nil {
				err = e
			} else {
				st, e := f.Stat()
				if e == nil {
					_, e = pluginZip(f, st.Size())
				}
				f.Close()
				err = e
			}
		}
		if err == nil {
			budget := pluginFileLimit
			digest, _, e := pluginDigest(ctx, child, name, &budget)
			if e != nil {
				err = e
			} else if digest == req.Items[i].SHA512 {
				result.Status = "skipped"
				result.Message = "Downloaded JAR is identical to the installed file; no replacement needed."
				child.Close()
				continue
			}
		}
		child.Close()
		if err != nil {
			continue
		}
		staged[i] = filepath.Join(stageName, childName, name)
	}
	ready := false
	for _, path := range staged {
		ready = ready || path != ""
	}
	if !ready {
		writeJSON(w, 200, response)
		return
	}
	// Create backups only after staging is complete. Never use plugins/ or mods/.
	backupBase := "plugin-backups"
	if err := root.Check(backupBase); err != nil {
		finishFailedPluginBatch(w, response)
		return
	}
	if err := root.MkdirAll(backupBase, 0700); err != nil {
		finishFailedPluginBatch(w, response)
		return
	}
	backups, err := root.Sub(backupBase)
	if err != nil {
		finishFailedPluginBatch(w, response)
		return
	}
	defer backups.Close()
	batchName, batch, err := backups.TempDir()
	if err != nil {
		finishFailedPluginBatch(w, response)
		return
	}
	batch.Close()
	blocked := false
	for i, stagedPath := range staged {
		if stagedPath == "" {
			continue
		}
		result := &response.Results[i]
		if blocked || ctx.Err() != nil {
			result.Message = "Update interrupted; original plugin retained."
			continue
		}
		item := req.Items[i]
		// Rehash before publication too; unsupported external host edits are not leased.
		checkBudget := pluginFileLimit
		hash, _, err := pluginDigest(ctx, dir, item.Filename, &checkBudget)
		if err != nil || hash != item.SHA512 {
			result.Message = "Plugin changed after preflight; original retained. Rescan."
			continue
		}
		backup := filepath.Join(backupBase, batchName, item.Filename)
		live := filepath.Join(folder, item.Filename)
		if root.Check(live) != nil || root.Check(backup) != nil {
			result.Message = "Unsafe plugin publication path; original retained."
			continue
		}
		recovery, err := publishPluginUpdate(live, stagedPath, backup, root.Link, root.Remove, root.Rename)
		if _, e := root.Lstat(backup); e == nil {
			result.BackupPath = filepath.ToSlash(backup)
		}
		if recovery {
			h.manager.BlockPluginUpdateRecovery(id)
			blocked = true
		}
		if err != nil {
			result.Message = err.Error()
			continue
		}
		result.Status = "updated"
		result.Message = "Plugin updated; server remains stopped. Original backup retained."
		response.RestartRequired = true
		logAudit(h.db, r, "plugin_update", fmt.Sprintf("Updated %s to %s on %s; backup %s", item.Filename, item.VersionID, id, result.BackupPath))
	}
	writeJSON(w, 200, response)
}

func finishFailedPluginBatch(w http.ResponseWriter, response pluginUpdateResponse) {
	for i := range response.Results {
		if response.Results[i].Status == "failed" {
			response.Results[i].Message = "Cannot create safe plugin backups; original plugin retained."
		}
	}
	writeJSON(w, 200, response)
}
