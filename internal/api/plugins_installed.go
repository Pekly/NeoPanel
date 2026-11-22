package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

type installedProject struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	IconURL string `json:"icon_url"`
}
type installedCandidate struct {
	installedProject
	Description string `json:"description"`
}
type installedRelease struct {
	ID            string   `json:"id"`
	VersionNumber string   `json:"version_number"`
	Name          string   `json:"name"`
	DatePublished string   `json:"date_published"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string `json:"loaders"`
}
type installedCatalogItem struct {
	InstalledPlugin
	SHA512             string               `json:"sha512"`
	Name               string               `json:"name"`
	InstalledVersion   string               `json:"installed_version"`
	Match              string               `json:"match"`
	Project            *installedProject    `json:"project"`
	InstalledVersionID string               `json:"installed_version_id"`
	Latest             *installedRelease    `json:"latest"`
	UpdateAvailable    bool                 `json:"update_available"`
	Candidates         []installedCandidate `json:"candidates"`
	Message            string               `json:"message"`
}
type installedCatalog struct {
	Items   []installedCatalogItem `json:"items"`
	Offset  int                    `json:"offset"`
	Limit   int                    `json:"limit"`
	Total   int                    `json:"total"`
	HasMore bool                   `json:"has_more"`
	Folder  string                 `json:"folder"`
	Kind    string                 `json:"kind"`
}

func pluginFolder(software string) (string, string) {
	if downloader.IsModdedSoftware(software) {
		return "mods", "mod"
	}
	return "plugins", "plugin"
}
func catalogProject(p *downloader.ModrinthProject) *installedProject {
	icon := ""
	if p.IconURL != nil {
		icon = *p.IconURL
	}
	return &installedProject{p.ID, p.Slug, p.Title, icon}
}
func catalogRelease(v *downloader.ModrinthVersion) *installedRelease {
	if v == nil {
		return nil
	}
	games, loaders := append([]string{}, v.GameVersions...), append([]string{}, v.Loaders...)
	return &installedRelease{v.ID, v.VersionNumber, v.Name, v.DatePublished, games, loaders}
}
func pluginNewer(installed, target *downloader.ModrinthVersion, hash string) bool {
	if target == nil || downloader.ModrinthVersionHasHash(target, hash) {
		return false
	}
	if installed == nil {
		return true
	} // explicitly user-confirmed replacement, not a verified upgrade
	if installed.ID == target.ID {
		return false
	}
	a, ea := time.Parse(time.RFC3339Nano, installed.DatePublished)
	b, eb := time.Parse(time.RFC3339Nano, target.DatePublished)
	return ea == nil && eb == nil && b.After(a)
}
func itemError(item *installedCatalogItem, message string) {
	item.Match, item.Message, item.UpdateAvailable = "error", message, false
}

// Catalog is nonmutating and keeps the canonical shared root lease through the
// complete scan. A successful rescan cannot overlap an uncertain bulk update.
func (h *PluginHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	srv, err := h.db.GetServer(id)
	if err != nil {
		writeError(w, 404, "Server not found")
		return
	}
	if rejectSteelPlugins(srv, w) {
		return
	}
	offset, limit := 0, 20
	for key, dest := range map[string]*int{"offset": &offset, "limit": &limit} {
		if s := r.URL.Query().Get(key); s != "" {
			v, e := strconv.Atoi(s)
			if e != nil || v < 0 {
				writeError(w, 400, "Invalid catalog pagination")
				return
			}
			*dest = v
		}
	}
	if limit == 0 || limit > 20 {
		limit = 20
	}
	metadataMode := r.URL.Query().Get("metadata_only")
	if metadataMode != "" && metadataMode != "true" && metadataMode != "false" {
		writeError(w, 400, "Invalid metadata mode")
		return
	}
	metadataOnly := metadataMode == "true"
	folder, kind := pluginFolder(srv.Type)
	page := installedCatalog{Items: []installedCatalogItem{}, Offset: offset, Limit: limit, Folder: folder, Kind: kind}
	release, err := rooted.TryMutation(h.manager.GetWorkingDir(id))
	if err != nil {
		writeError(w, 409, "Server files are busy; retry the scan")
		return
	}
	defer release()
	root, err := rooted.Open(h.manager.GetWorkingDir(id))
	if err != nil {
		writeError(w, 400, "Cannot open server files")
		return
	}
	defer root.Close()
	dir, err := root.Sub(folder)
	if os.IsNotExist(err) {
		writeJSON(w, 200, page)
		return
	}
	if err != nil {
		writeError(w, 400, "Cannot open plugin directory")
		return
	}
	defer dir.Close()
	entries, err := dir.ReadDir(".")
	if err != nil {
		writeError(w, 400, "Cannot list plugin directory")
		return
	}
	names := []string{}
	for _, entry := range entries {
		if strictPluginFilename(entry.Name()) && entry.Type().IsRegular() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	page.Total = len(names)
	start := min(offset, len(names))
	end := start + min(limit, len(names)-start)
	page.HasMore = end < len(names)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	ctx = downloader.ModrinthMetadataBudgetContext(ctx)
	remaining := pluginBatchLimit
	hashes := []string{}
	for _, name := range names[start:end] {
		item := installedCatalogItem{InstalledPlugin: InstalledPlugin{Filename: name, Enabled: !strings.HasSuffix(strings.ToLower(name), ".disabled")}, Match: "unknown", Candidates: []installedCandidate{}}
		item.Name = pluginFilenameQuery(name)
		hash, st, err := pluginDigest(ctx, dir, name, &remaining)
		if st != nil {
			item.Size, item.Modified = st.Size(), st.ModTime()
		}
		if err != nil {
			itemError(&item, err.Error())
		} else {
			item.SHA512 = hash
			hashes = append(hashes, hash)
		}
		page.Items = append(page.Items, item)
	}
	matches, lookupErr := downloader.LookupModrinthHashes(ctx, hashes)
	projectCache := make(map[string]*downloader.ModrinthProject)
	for i := range page.Items {
		item := &page.Items[i]
		if item.Match == "error" {
			continue
		}
		if lookupErr != nil {
			itemError(item, "Hash lookup unavailable; retry the scan")
			continue
		}
		var installed *downloader.ModrinthVersion
		projectID := ""
		if v, ok := matches[item.SHA512]; ok {
			installed = &v
			projectID = v.ProjectID
			item.Match = "hash"
			item.InstalledVersion, item.InstalledVersionID = v.VersionNumber, v.ID
		} else {
			projectID, err = h.db.ConfirmedPluginSource(id, item.Filename, item.SHA512)
			if err != nil {
				itemError(item, "Cannot read confirmed association")
				continue
			}
			item.Name, item.InstalledVersion, err = pluginDescriptor(dir, item.Filename)
			if err != nil {
				itemError(item, err.Error())
				continue
			}
			if item.Name == "" {
				item.Name = pluginFilenameQuery(item.Filename)
			}
			if projectID != "" {
				item.Match = "confirmed"
			}
		}
		if projectID != "" {
			p := projectCache[projectID]
			var projectErr error
			if p == nil {
				p, projectErr = downloader.GetModrinthProjectContext(ctx, projectID)
				if projectErr == nil {
					projectCache[projectID] = p
				}
			}
			if projectErr != nil {
				itemError(item, "Project metadata unavailable; retry the scan")
				continue
			}
			item.Project = catalogProject(p)
			item.Name = p.Title
			if metadataOnly {
				if item.Match == "confirmed" {
					item.Message = "User-confirmed project; installed release unverified. Updates not checked."
				} else {
					item.Message = "Exact hash identity; updates not checked."
				}
				continue
			}
			latest, err := downloader.ResolveModrinthVersionContext(ctx, p.ID, srv.Type, "")
			if err != nil {
				itemError(item, "Latest compatible release unavailable; retry the scan")
				continue
			}
			item.Latest = catalogRelease(latest)
			item.UpdateAvailable = pluginNewer(installed, latest, item.SHA512)
			if item.Match == "confirmed" {
				item.Message = "User-confirmed project, not a verified installed release. Updating replaces this uploaded JAR."
			}
			if installed != nil && latest != nil && !item.UpdateAvailable && installed.ID != latest.ID {
				item.Message = "No verified newer compatible release; automatic downgrade is disabled."
			}
			continue
		}
		if metadataOnly {
			item.Message = "Unverified embedded name; check for updates to search possible sources."
			continue
		}
		if !crediblePluginName(item.Name, item.Name, "") {
			item.Message = "No credible plugin name found"
			continue
		}
		candidates, err := downloader.SearchModrinthPluginsPage(ctx, item.Name, 5, 0, srv.Type, "", "relevance")
		if err != nil {
			itemError(item, "Project search unavailable; retry the scan")
			continue
		}
		for _, c := range candidates.Hits {
			if downloader.ValidModrinthIdentifier(c.ProjectID) && crediblePluginName(item.Name, c.Title, c.Slug) {
				item.Candidates = append(item.Candidates, installedCandidate{installedProject{c.ProjectID, c.Slug, c.Title, c.IconURL}, c.Description})
			}
		}
		if len(item.Candidates) > 0 {
			item.Match = "suggested"
			item.Message = "Name match only; explicitly confirm a project before replacement."
		} else {
			item.Message = "No credible project match found"
		}
	}
	writeJSON(w, 200, page)
}

type pluginIdentityRequest struct {
	Filename  string `json:"filename"`
	SHA512    string `json:"sha512"`
	ProjectID string `json:"project_id"`
	VersionID string `json:"version_id,omitempty"`
}

func (v pluginIdentityRequest) valid() bool {
	return strictPluginFilename(v.Filename) && downloader.ValidPluginSHA512(v.SHA512) && downloader.ValidModrinthIdentifier(v.ProjectID)
}
func decodePluginRequest(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("expected one request")
	}
	return nil
}

func (h *PluginHandler) Identify(w http.ResponseWriter, r *http.Request) {
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
		pluginIdentityRequest
		Confirm bool `json:"confirm"`
	}
	if decodePluginRequest(w, r, &req) != nil || !req.Confirm || !req.valid() || req.VersionID != "" {
		writeError(w, 400, "Explicit confirmation, filename, SHA512 and project ID are required")
		return
	}
	release, err := rooted.TryExclusiveMutation(h.manager.GetWorkingDir(id))
	if err != nil {
		writeError(w, 409, "Server files are busy")
		return
	}
	defer release()
	root, err := rooted.Open(h.manager.GetWorkingDir(id))
	if err != nil {
		writeError(w, 409, "Plugin file unavailable")
		return
	}
	defer root.Close()
	folder, _ := pluginFolder(srv.Type)
	dir, err := root.Sub(folder)
	if err != nil {
		writeError(w, 409, "Plugin file unavailable")
		return
	}
	defer dir.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	ctx = downloader.ModrinthMetadataBudgetContext(ctx)
	remaining := pluginFileLimit
	hash, _, err := pluginDigest(ctx, dir, req.Filename, &remaining)
	if err != nil || hash != req.SHA512 {
		writeError(w, 409, "Plugin changed; rescan before confirming")
		return
	}
	p, err := downloader.GetModrinthProjectContext(ctx, req.ProjectID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Project metadata unavailable", "code": "plugin_catalog_unavailable"})
		return
	}
	if p.ID != req.ProjectID {
		writeError(w, 400, "Use the exact project ID")
		return
	}
	v, err := downloader.ResolveModrinthVersionContext(ctx, p.ID, srv.Type, "")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Release metadata unavailable", "code": "plugin_catalog_unavailable"})
		return
	}
	if v == nil {
		writeError(w, 400, "Project has no compatible JAR release")
		return
	}
	if err := h.db.ConfirmPluginSource(id, req.Filename, hash, p.ID); err != nil {
		writeError(w, 500, "Cannot save confirmed association")
		return
	}
	logAudit(h.db, r, "plugin_identify", fmt.Sprintf("Confirmed project %s for %s on %s", p.ID, req.Filename, id))
	writeJSON(w, 200, map[string]string{"message": "Project association saved for this exact file hash; this is user confirmation, not hash verification."})
}
