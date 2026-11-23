package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModpackEndpoints(t *testing.T) {
	f := newReliabilityFixture(t)

	// 1. Check initially not installed
	w := f.request("GET", "/api/servers/alpha/modpack", "", "owner")
	if w.Code != 200 {
		t.Fatalf("GET modpack failed: %d %s", w.Code, w.Body.String())
	}
	var res struct {
		Installed bool `json:"installed"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil || res.Installed {
		t.Fatalf("Expected not installed, got %+v", res)
	}

	// 2. Write fake .neopanel_modpack.json into alpha server directory
	liveDir := f.manager.GetWorkingDir("alpha")
	fakeMeta := `{
		"name": "Cobblemon Official",
		"loader": "fabric",
		"game_version": "1.20.1",
		"server_mods_count": 35
	}`
	_ = os.WriteFile(filepath.Join(liveDir, ".neopanel_modpack.json"), []byte(fakeMeta), 0644)

	// 3. Verify it reports installed
	w = f.request("GET", "/api/servers/alpha/modpack", "", "owner")
	if w.Code != 200 {
		t.Fatalf("GET modpack failed: %d %s", w.Code, w.Body.String())
	}
	var res2 struct {
		Installed bool `json:"installed"`
		Metadata  struct {
			Name        string `json:"name"`
			Loader      string `json:"loader"`
			GameVersion string `json:"game_version"`
			ServerMods  int    `json:"server_mods_count"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res2); err != nil || !res2.Installed || res2.Metadata.Name != "Cobblemon Official" {
		t.Fatalf("Expected Cobblemon installed, got %+v", res2)
	}

	// 4. Test delete / unlink
	w = f.request("DELETE", "/api/servers/alpha/modpack", "", "owner")
	if w.Code != 200 {
		t.Fatalf("DELETE modpack failed: %d %s", w.Code, w.Body.String())
	}

	// 5. Check it is unlinked
	w = f.request("GET", "/api/servers/alpha/modpack", "", "owner")
	var res3 struct {
		Installed bool `json:"installed"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res3)
	if res3.Installed {
		t.Errorf("Expected uninstalled after DELETE")
	}

	// 6. Test install validation (invalid slug)
	w = f.request("POST", "/api/servers/alpha/modpack/install", `{"slug":"invalid$$slug!"}`, "owner")
	if w.Code != 400 || !strings.Contains(w.Body.String(), "Valid modpack identifier") {
		t.Errorf("Expected 400 for invalid slug, got %d %s", w.Code, w.Body.String())
	}
}
