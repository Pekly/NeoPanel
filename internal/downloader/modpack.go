package downloader

import (
	"archive/zip"
	"context"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ModpackIndex struct {
	FormatVersion int               `json:"formatVersion"`
	Game          string            `json:"game"`
	VersionID     string            `json:"versionId"`
	Name          string            `json:"name"`
	Summary       string            `json:"summary"`
	Files         []ModpackFile     `json:"files"`
	Dependencies  map[string]string `json:"dependencies"`
}

type ModpackFile struct {
	Path      string            `json:"path"`
	Hashes    map[string]string `json:"hashes"`
	Env       map[string]string `json:"env"`
	Downloads []string          `json:"downloads"`
	FileSize  int64             `json:"fileSize"`
}

type InstalledModpackMetadata struct {
	ProjectID     string            `json:"project_id"`
	Slug          string            `json:"slug"`
	Name          string            `json:"name"`
	VersionID     string            `json:"version_id"`
	VersionNumber string            `json:"version_number"`
	InstalledAt   string            `json:"installed_at"`
	ServerMods    int               `json:"server_mods_count"`
	ClientSkipped int               `json:"client_mods_skipped"`
	Loader        string            `json:"loader"`
	GameVersion   string            `json:"game_version"`
	Dependencies  map[string]string `json:"dependencies"`
}

type ModpackInstallResult struct {
	Name          string            `json:"name"`
	VersionNumber string            `json:"version_number"`
	Loader        string            `json:"loader"`
	GameVersion   string            `json:"game_version"`
	ModsInstalled int               `json:"mods_installed"`
	ModsSkipped   int               `json:"mods_skipped"`
	Overrides     int               `json:"overrides_count"`
	Dependencies  map[string]string `json:"dependencies"`
	Summary       string            `json:"summary"`
}

// InstallModpackArchive extracts and installs an .mrpack file into the destination directory.
func InstallModpackArchive(ctx context.Context, mrpackPath, destDir string, httpClient *http.Client) (*ModpackInstallResult, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}

	zr, err := zip.OpenReader(mrpackPath)
	if err != nil {
		return nil, fmt.Errorf("opening modpack archive: %w", err)
	}
	defer zr.Close()

	// 1. Locate and parse modrinth.index.json
	var indexFile *zip.File
	for _, f := range zr.File {
		if f.Name == "modrinth.index.json" {
			indexFile = f
			break
		}
	}
	if indexFile == nil {
		return nil, fmt.Errorf("modpack archive missing modrinth.index.json")
	}

	rc, err := indexFile.Open()
	if err != nil {
		return nil, fmt.Errorf("reading modrinth.index.json: %w", err)
	}
	defer rc.Close()

	var index ModpackIndex
	if err := json.NewDecoder(rc).Decode(&index); err != nil {
		return nil, fmt.Errorf("parsing modrinth.index.json: %w", err)
	}

	// 2. Extract overrides and server-overrides
	cleanDest, err := filepath.Abs(destDir)
	if err != nil {
		return nil, err
	}

	overridesCount := 0
	// First pass: overrides/
	for _, f := range zr.File {
		var relPath string
		if strings.HasPrefix(f.Name, "overrides/") {
			relPath = strings.TrimPrefix(f.Name, "overrides/")
		} else if strings.HasPrefix(f.Name, "server-overrides/") {
			relPath = strings.TrimPrefix(f.Name, "server-overrides/")
		} else {
			continue
		}

		if relPath == "" || strings.HasSuffix(relPath, "/") {
			// Directory
			targetDir := filepath.Join(cleanDest, filepath.FromSlash(relPath))
			if !isSafePath(cleanDest, targetDir) {
				continue
			}
			_ = os.MkdirAll(targetDir, 0755)
			continue
		}

		targetFile := filepath.Join(cleanDest, filepath.FromSlash(relPath))
		if !isSafePath(cleanDest, targetFile) {
			continue
		}

		if err := extractZipEntry(f, targetFile); err != nil {
			return nil, fmt.Errorf("extracting override %s: %w", relPath, err)
		}
		overridesCount++
	}

	// 3. Download Server Mods (skipping client-only mods)
	var modsToDownload []ModpackFile
	skippedCount := 0

	for _, mf := range index.Files {
		// Env check: if server is unsupported, skip!
		serverEnv := strings.ToLower(mf.Env["server"])
		if serverEnv == "unsupported" {
			skippedCount++
			continue
		}
		modsToDownload = append(modsToDownload, mf)
	}

	// Download with bounded concurrency
	concurrency := 4
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var downloadErr error
	var errMu sync.Mutex

	downloadedCount := 0
	var countMu sync.Mutex

	for _, mf := range modsToDownload {
		if len(mf.Downloads) == 0 {
			continue
		}

		targetPath := filepath.Join(cleanDest, filepath.FromSlash(mf.Path))
		if !isSafePath(cleanDest, targetPath) {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(file ModpackFile, dest string) {
			defer wg.Done()
			defer func() { <-sem }()

			errMu.Lock()
			if downloadErr != nil {
				errMu.Unlock()
				return
			}
			errMu.Unlock()

			if err := downloadModpackFile(ctx, httpClient, file, dest); err != nil {
				errMu.Lock()
				if downloadErr == nil {
					downloadErr = fmt.Errorf("downloading %s: %w", file.Path, err)
				}
				errMu.Unlock()
				return
			}

			countMu.Lock()
			downloadedCount++
			countMu.Unlock()
		}(mf, targetPath)
	}

	wg.Wait()
	if downloadErr != nil {
		return nil, downloadErr
	}

	// 4. Identify loader and Minecraft version
	loader := "fabric"
	if _, ok := index.Dependencies["neoforge"]; ok {
		loader = "neoforge"
	} else if _, ok := index.Dependencies["forge"]; ok {
		loader = "forge"
	} else if _, ok := index.Dependencies["quilt-loader"]; ok {
		loader = "fabric"
	} else if _, ok := index.Dependencies["fabric-loader"]; ok {
		loader = "fabric"
	}

	gameVersion := index.Dependencies["minecraft"]

	// 5. Write .neopanel_modpack.json metadata in server root
	meta := InstalledModpackMetadata{
		Name:          index.Name,
		VersionID:     index.VersionID,
		InstalledAt:   time.Now().UTC().Format(time.RFC3339),
		ServerMods:    downloadedCount,
		ClientSkipped: skippedCount,
		Loader:        loader,
		GameVersion:   gameVersion,
		Dependencies:  index.Dependencies,
	}
	if metaRaw, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(cleanDest, ".neopanel_modpack.json"), metaRaw, 0644)
	}

	return &ModpackInstallResult{
		Name:          index.Name,
		VersionNumber: index.VersionID,
		Loader:        loader,
		GameVersion:   gameVersion,
		ModsInstalled: downloadedCount,
		ModsSkipped:   skippedCount,
		Overrides:     overridesCount,
		Dependencies:  index.Dependencies,
		Summary:       fmt.Sprintf("Installed %s (%s %s): %d server mods downloaded (%d client-only mods omitted), %d overrides applied", index.Name, loader, gameVersion, downloadedCount, skippedCount, overridesCount),
	}, nil
}

func downloadModpackFile(ctx context.Context, client *http.Client, file ModpackFile, targetPath string) error {
	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)

	var lastErr error
	for _, urlStr := range file.Downloads {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "NeoPanel-Modpack-Installer/1.0")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, urlStr)
			continue
		}

		tmpFile := targetPath + ".tmp"
		f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			resp.Body.Close()
			return err
		}

		hSha1 := sha1.New()
		hSha512 := sha512.New()
		w := io.MultiWriter(f, hSha1, hSha512)

		_, copyErr := io.Copy(w, resp.Body)
		resp.Body.Close()
		f.Close()

		if copyErr != nil {
			_ = os.Remove(tmpFile)
			lastErr = copyErr
			continue
		}

		// Verify hashes if provided
		if expectedSha512, ok := file.Hashes["sha512"]; ok && expectedSha512 != "" {
			actual := hex.EncodeToString(hSha512.Sum(nil))
			if !strings.EqualFold(actual, expectedSha512) {
				_ = os.Remove(tmpFile)
				lastErr = fmt.Errorf("sha512 mismatch for %s: expected %s, got %s", file.Path, expectedSha512, actual)
				continue
			}
		} else if expectedSha1, ok := file.Hashes["sha1"]; ok && expectedSha1 != "" {
			actual := hex.EncodeToString(hSha1.Sum(nil))
			if !strings.EqualFold(actual, expectedSha1) {
				_ = os.Remove(tmpFile)
				lastErr = fmt.Errorf("sha1 mismatch for %s: expected %s, got %s", file.Path, expectedSha1, actual)
				continue
			}
		}

		// Success! Rename tmp to final
		if err := os.Rename(tmpFile, targetPath); err != nil {
			_ = os.Remove(tmpFile)
			return err
		}
		return nil
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("no valid download URLs for %s", file.Path)
}

func extractZipEntry(f *zip.File, targetFile string) error {
	_ = os.MkdirAll(filepath.Dir(targetFile), 0755)

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()&0777|0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, io.LimitReader(rc, 100<<20)) // 100 MB max per override file
	return err
}

func isSafePath(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	if strings.HasPrefix(rel, "..") || rel == "." {
		return false
	}
	return true
}
