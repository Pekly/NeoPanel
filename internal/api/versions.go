package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MinecraftVersionsCatalog struct {
	AllMC    []string            `json:"all_mc"`
	ByEngine map[string][]string `json:"by_engine"`
	CachedAt string              `json:"cached_at"`
}

var (
	versionsMu        sync.RWMutex
	versionsCached    *MinecraftVersionsCatalog
	versionsExpiresAt time.Time
)

const versionsCacheTTL = 30 * time.Minute

var fallbackAllMCVersions = []string{
	"26.3", "26.2", "26.1.2", "26.1.1", "26.1",
	"1.21.11", "1.21.10", "1.21.9", "1.21.8", "1.21.7", "1.21.6", "1.21.5", "1.21.4", "1.21.3", "1.21.2", "1.21.1", "1.21",
	"1.20.6", "1.20.4", "1.20.1",
	"1.19.4", "1.19.2",
	"1.18.2",
	"1.16.5",
	"1.12.2",
	"1.8.9",
	"1.7.10",
}

func filterRecentAndPopular(versions []string) []string {
	popular := map[string]bool{
		"1.20.6": true, "1.20.4": true, "1.20.1": true,
		"1.19.4": true, "1.19.2": true,
		"1.18.2": true,
		"1.16.5": true,
		"1.12.2": true,
		"1.8.9":  true,
		"1.7.10": true,
	}
	var out []string
	seen := make(map[string]bool)
	for _, v := range versions {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		if strings.HasPrefix(v, "1.21") || strings.HasPrefix(v, "26.") || popular[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func HandleGetMinecraftVersions(w http.ResponseWriter, r *http.Request) {
	forceRefresh := r.URL.Query().Get("refresh") == "1"

	versionsMu.RLock()
	if !forceRefresh && versionsCached != nil && time.Now().Before(versionsExpiresAt) {
		cat := versionsCached
		versionsMu.RUnlock()
		writeJSON(w, http.StatusOK, cat)
		return
	}
	versionsMu.RUnlock()

	cat := buildVersionsCatalog(r.Context())

	versionsMu.Lock()
	versionsCached = cat
	versionsExpiresAt = time.Now().Add(versionsCacheTTL)
	versionsMu.Unlock()

	writeJSON(w, http.StatusOK, cat)
}

func buildVersionsCatalog(parentCtx context.Context) *MinecraftVersionsCatalog {
	ctx, cancel := context.WithTimeout(parentCtx, 6*time.Second)
	defer cancel()

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		mojangVers []string
		modrVers   []string
		paperVers  []string
		foliaVers  []string
		velVers    []string
		wfVers     []string
		purpurVers []string
		fabricVers []string
	)

	fetchers := []func(){
		func() {
			list := fetchMojangReleases(ctx)
			mu.Lock()
			mojangVers = list
			mu.Unlock()
		},
		func() {
			list := fetchModrinthGameVersions(ctx)
			mu.Lock()
			modrVers = list
			mu.Unlock()
		},
		func() {
			list := fetchPaperFillProjectVersions(ctx, "paper")
			mu.Lock()
			paperVers = list
			mu.Unlock()
		},
		func() {
			list := fetchPaperFillProjectVersions(ctx, "folia")
			mu.Lock()
			foliaVers = list
			mu.Unlock()
		},
		func() {
			list := fetchPaperFillProjectVersions(ctx, "velocity")
			mu.Lock()
			velVers = list
			mu.Unlock()
		},
		func() {
			list := fetchPaperFillProjectVersions(ctx, "waterfall")
			mu.Lock()
			wfVers = list
			mu.Unlock()
		},
		func() {
			list := fetchPurpurProjectVersions(ctx)
			mu.Lock()
			purpurVers = list
			mu.Unlock()
		},
		func() {
			list := fetchFabricGameVersions(ctx)
			mu.Lock()
			fabricVers = list
			mu.Unlock()
		},
	}

	for _, fn := range fetchers {
		wg.Add(1)
		f := fn
		go func() {
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()

	rawMC := mergeSortedVersions(
		fallbackAllMCVersions,
		mojangVers,
		modrVers,
		paperVers,
		purpurVers,
		fabricVers,
	)
	allMC := filterRecentAndPopular(rawMC)

	byEngine := map[string][]string{
		"vanilla": allMC,
		"paper": mergeSortedVersions(
			filterMinVersion(allMC, 1, 8, 8),
			paperVers,
		),
		"purpur": mergeSortedVersions(
			filterMinVersion(allMC, 1, 14, 1),
			purpurVers,
		),
		"leaf": filterMinVersion(allMC, 1, 20, 4),
		"folia": mergeSortedVersions(
			filterMinVersion(allMC, 1, 19, 4),
			foliaVers,
		),
		"fabric": mergeSortedVersions(
			filterMinVersion(allMC, 1, 14, 0),
			fabricVers,
		),
		"forge":    filterMinVersion(allMC, 1, 5, 2),
		"neoforge": filterMinVersion(allMC, 1, 20, 1),
		"velocity": mergeSortedVersions(
			[]string{"3.4.0-SNAPSHOT", "3.3.0-SNAPSHOT", "3.2.0-SNAPSHOT", "3.1.2-SNAPSHOT", "3.1.1", "3.1.0", "3.0.1", "1.1.9", "1.0.10"},
			velVers,
		),
		"waterfall": mergeSortedVersions(
			[]string{"1.21", "1.20", "1.19", "1.18", "1.17", "1.16", "1.15", "1.14", "1.13", "1.12"},
			wfVers,
		),
		"bungeecord": []string{"latest", "1.21", "1.20", "1.19", "1.18", "1.17", "1.16", "1.12", "1.8"},
	}

	return &MinecraftVersionsCatalog{
		AllMC:    allMC,
		ByEngine: byEngine,
		CachedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func versionGetJSON(ctx context.Context, targetURL string, out interface{}) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "NeoPanel/1.0 (+https://github.com/Pekly/NeoPanel)")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return json.NewDecoder(resp.Body).Decode(out) == nil
}

func fetchMojangReleases(ctx context.Context) []string {
	var manifest struct {
		Versions []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"versions"`
	}
	if !versionGetJSON(ctx, "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json", &manifest) {
		return nil
	}
	var out []string
	for _, v := range manifest.Versions {
		if v.Type == "release" && isCleanMCVersion(v.ID) {
			out = append(out, v.ID)
		}
	}
	return out
}

func fetchModrinthGameVersions(ctx context.Context) []string {
	var items []struct {
		Version     string `json:"version"`
		VersionType string `json:"version_type"`
	}
	if !versionGetJSON(ctx, "https://api.modrinth.com/v2/tag/game_version", &items) {
		return nil
	}
	var out []string
	for _, item := range items {
		if item.VersionType == "release" && isCleanMCVersion(item.Version) {
			out = append(out, item.Version)
		}
	}
	return out
}

func fetchPaperFillProjectVersions(ctx context.Context, project string) []string {
	var payload struct {
		Versions json.RawMessage `json:"versions"`
	}
	if !versionGetJSON(ctx, "https://fill.papermc.io/v3/projects/"+project, &payload) || len(payload.Versions) == 0 {
		return nil
	}

	var out []string
	var asSlice []string
	if json.Unmarshal(payload.Versions, &asSlice) == nil && len(asSlice) > 0 {
		return asSlice
	}

	var asGroups map[string][]string
	if json.Unmarshal(payload.Versions, &asGroups) == nil && len(asGroups) > 0 {
		for _, grp := range asGroups {
			out = append(out, grp...)
		}
		return out
	}

	return out
}

func fetchPurpurProjectVersions(ctx context.Context) []string {
	var payload struct {
		Versions []string `json:"versions"`
	}
	if !versionGetJSON(ctx, "https://api.purpurmc.org/v2/purpur", &payload) {
		return nil
	}
	return payload.Versions
}

func fetchFabricGameVersions(ctx context.Context) []string {
	var items []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if !versionGetJSON(ctx, "https://meta.fabricmc.net/v2/versions/game", &items) {
		return nil
	}
	var out []string
	for _, item := range items {
		if item.Stable && isCleanMCVersion(item.Version) {
			out = append(out, item.Version)
		}
	}
	return out
}

func isCleanMCVersion(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" || strings.ContainsAny(v, " w-") {
		return false
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

type parsedVer struct {
	nums [4]int
	raw  string
}

func parseVerTuple(v string) parsedVer {
	clean := strings.TrimSpace(v)
	if strings.EqualFold(clean, "latest") {
		return parsedVer{nums: [4]int{999999, 0, 0, 0}, raw: clean}
	}
	base := clean
	if idx := strings.IndexAny(base, "-+"); idx >= 0 {
		base = base[:idx]
	}
	parts := strings.Split(base, ".")
	var nums [4]int
	for i := 0; i < 4 && i < len(parts); i++ {
		if n, err := strconv.Atoi(parts[i]); err == nil {
			nums[i] = n
		}
	}
	return parsedVer{nums: nums, raw: clean}
}

func compareVersionsDesc(a, b string) int {
	pa := parseVerTuple(a)
	pb := parseVerTuple(b)
	for i := 0; i < 4; i++ {
		if pa.nums[i] != pb.nums[i] {
			if pa.nums[i] > pb.nums[i] {
				return -1
			}
			return 1
		}
	}
	aSnap := strings.Contains(strings.ToUpper(a), "SNAPSHOT")
	bSnap := strings.Contains(strings.ToUpper(b), "SNAPSHOT")
	if aSnap != bSnap {
		if !aSnap {
			return -1
		}
		return 1
	}
	return strings.Compare(b, a)
}

func mergeSortedVersions(lists ...[]string) []string {
	seen := make(map[string]struct{}, 128)
	var out []string
	for _, list := range lists {
		for _, raw := range list {
			v := strings.TrimSpace(raw)
			if v == "" {
				continue
			}
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return compareVersionsDesc(out[i], out[j]) < 0
	})
	return out
}

func filterMinVersion(vers []string, minMaj, minMin, minPatch int) []string {
	minTuple := [3]int{minMaj, minMin, minPatch}
	var out []string
	for _, v := range vers {
		p := parseVerTuple(v)
		ge := false
		if p.nums[0] > minTuple[0] {
			ge = true
		} else if p.nums[0] == minTuple[0] {
			if p.nums[1] > minTuple[1] {
				ge = true
			} else if p.nums[1] == minTuple[1] && p.nums[2] >= minTuple[2] {
				ge = true
			}
		}
		if ge {
			out = append(out, v)
		}
	}
	return out
}
