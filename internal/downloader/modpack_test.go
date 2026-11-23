package downloader

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallModpackArchive(t *testing.T) {
	// 1. Mock HTTP server serving mod jar downloads
	modContent := []byte("fake-server-mod-jar-content")
	h := sha1.New()
	h.Write(modContent)
	modSha1 := hex.EncodeToString(h.Sum(nil))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(modContent)
	}))
	defer mockServer.Close()

	// 2. Build mock .mrpack zip in memory
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add modrinth.index.json
	indexJSON := `{
		"formatVersion": 1,
		"game": "minecraft",
		"versionId": "v1.0.0",
		"name": "Test Survival Modpack",
		"summary": "Awesome modpack",
		"files": [
			{
				"path": "mods/server-mod.jar",
				"hashes": {"sha1": "` + modSha1 + `"},
				"env": {"client": "required", "server": "required"},
				"downloads": ["` + mockServer.URL + `/server-mod.jar"],
				"fileSize": 27
			},
			{
				"path": "mods/client-only-mod.jar",
				"hashes": {"sha1": "abcdef"},
				"env": {"client": "required", "server": "unsupported"},
				"downloads": ["` + mockServer.URL + `/client-only-mod.jar"],
				"fileSize": 50
			}
		],
		"dependencies": {
			"minecraft": "1.20.1",
			"fabric-loader": "0.15.11"
		}
	}`

	iw, _ := zw.Create("modrinth.index.json")
	_, _ = iw.Write([]byte(indexJSON))

	// Add an override config file
	cw, _ := zw.Create("overrides/config/test.toml")
	_, _ = cw.Write([]byte("enabled = true\n"))

	_ = zw.Close()

	tmpDir := t.TempDir()
	mrpackPath := filepath.Join(tmpDir, "test.mrpack")
	if err := os.WriteFile(mrpackPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write mock mrpack: %v", err)
	}

	destDir := filepath.Join(tmpDir, "server")
	_ = os.MkdirAll(destDir, 0755)

	res, err := InstallModpackArchive(context.Background(), mrpackPath, destDir, mockServer.Client())
	if err != nil {
		t.Fatalf("InstallModpackArchive failed: %v", err)
	}

	if res.Name != "Test Survival Modpack" {
		t.Errorf("Expected pack name 'Test Survival Modpack', got %s", res.Name)
	}
	if res.ModsInstalled != 1 {
		t.Errorf("Expected 1 server mod installed, got %d", res.ModsInstalled)
	}
	if res.ModsSkipped != 1 {
		t.Errorf("Expected 1 client mod skipped, got %d", res.ModsSkipped)
	}
	if res.Loader != "fabric" {
		t.Errorf("Expected loader 'fabric', got %s", res.Loader)
	}
	if res.GameVersion != "1.20.1" {
		t.Errorf("Expected game version '1.20.1', got %s", res.GameVersion)
	}

	// Verify server-mod.jar exists and client-only mod does NOT exist
	serverModPath := filepath.Join(destDir, "mods", "server-mod.jar")
	if _, err := os.Stat(serverModPath); os.IsNotExist(err) {
		t.Errorf("Expected server-mod.jar to exist at %s", serverModPath)
	}

	clientModPath := filepath.Join(destDir, "mods", "client-only-mod.jar")
	if _, err := os.Stat(clientModPath); !os.IsNotExist(err) {
		t.Errorf("Expected client-only-mod.jar to NOT exist, but it was found at %s", clientModPath)
	}

	// Verify override was extracted
	overridePath := filepath.Join(destDir, "config", "test.toml")
	if _, err := os.Stat(overridePath); os.IsNotExist(err) {
		t.Errorf("Expected override at %s", overridePath)
	}

	// Verify metadata file was written
	metaPath := filepath.Join(destDir, ".neopanel_modpack.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("Expected .neopanel_modpack.json at %s", metaPath)
	}
}
