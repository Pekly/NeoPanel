package downloader

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func ValidPluginSHA512(s string) bool {
	if len(s) != 128 || s != strings.ToLower(s) {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// LookupModrinthHashes sends only SHA512 digests, never JAR contents. An exact
// requested key AND a matching file digest are required for authoritative linkage.
func LookupModrinthHashes(ctx context.Context, hashes []string) (map[string]ModrinthVersion, error) {
	result := make(map[string]ModrinthVersion)
	if len(hashes) == 0 {
		return result, nil
	}
	if len(hashes) > 20 {
		return nil, fmt.Errorf("too many plugin hashes")
	}
	wanted := make(map[string]bool)
	for _, hash := range hashes {
		if !ValidPluginSHA512(hash) {
			return nil, fmt.Errorf("invalid SHA512")
		}
		wanted[hash] = true
	}
	ctx, cancel := modrinthMetadataContext(ctx)
	defer cancel()
	body, _ := json.Marshal(struct {
		Hashes    []string `json:"hashes"`
		Algorithm string   `json:"algorithm"`
	}{hashes, "sha512"})
	var response map[string]ModrinthVersion
	if err := modrinthMetadataRequest(ctx, http.MethodPost, "/version_files", bytes.NewReader(body), &response); err != nil {
		return nil, err
	}
	for hash, v := range response {
		if !wanted[hash] {
			continue
		}
		if !ValidModrinthIdentifier(v.ID) || !ValidModrinthIdentifier(v.ProjectID) || !ModrinthVersionHasHash(&v, hash) {
			return nil, fmt.Errorf("Modrinth returned inconsistent hash metadata")
		}
		result[hash] = v
	}
	return result, nil
}

func ModrinthVersionHasHash(v *ModrinthVersion, hash string) bool {
	if v == nil {
		return false
	}
	for _, f := range v.Files {
		if strings.EqualFold(f.Hashes["sha512"], hash) && hash != "" {
			return true
		}
	}
	return false
}

func GetModrinthProjectContext(ctx context.Context, id string) (*ModrinthProject, error) {
	ctx, cancel := modrinthMetadataContext(ctx)
	defer cancel()
	return fetchModrinthProject(ctx, id)
}

func updateFileIndex(v *ModrinthVersion) int {
	if v == nil {
		return -1
	}
	found := -1
	for i, f := range v.Files {
		if safeModrinthJar(f.Filename, f.URL, f.Size) && f.Size > 0 && (found < 0 || f.Primary) {
			found = i
			if f.Primary {
				break
			}
		}
	}
	return found
}

func ModrinthUpdateSize(v *ModrinthVersion) (int64, error) {
	i := updateFileIndex(v)
	if i < 0 {
		return 0, fmt.Errorf("release has no bounded JAR download")
	}
	return v.Files[i].Size, nil
}

// StageModrinthUpdate writes only to a caller-owned isolated directory. It never
// changes generic installation's no-overwrite semantics. Budget includes failed
// download bytes as well as successful ones and is shared sequentially.
func StageModrinthUpdate(ctx context.Context, v *ModrinthVersion, stage *rooted.Root, remaining *int64) (string, error) {
	i := updateFileIndex(v)
	if i < 0 {
		return "", fmt.Errorf("release has no bounded JAR download")
	}
	f := v.Files[i]
	if f.Hashes["sha512"] == "" && f.Hashes["sha1"] == "" {
		return "", fmt.Errorf("release has no verification checksum")
	}
	if f.Size > *remaining {
		return "", fmt.Errorf("batch download budget exceeded")
	}
	resp, err := modrinthGet(ctx, f.URL)
	if err != nil {
		return "", modrinthRequestError(ctx, "download")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Modrinth download failed")
	}
	out, err := stage.OpenFile("plugin.jar", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("cannot create staged download")
	}
	h512, h1 := sha512.New(), sha1.New()
	n, copyErr := rooted.Copy(ctx, io.MultiWriter(out, h512, h1), resp.Body, f.Size)
	*remaining -= n
	syncErr := out.Sync()
	closeErr := out.Close()
	if copyErr != nil || syncErr != nil || closeErr != nil || n != f.Size {
		return "", fmt.Errorf("plugin download incomplete or exceeds size limit")
	}
	for algorithm, actual := range map[string]string{"sha512": hex.EncodeToString(h512.Sum(nil)), "sha1": hex.EncodeToString(h1.Sum(nil))} {
		if expected := f.Hashes[algorithm]; expected != "" && !strings.EqualFold(expected, actual) {
			return "", fmt.Errorf("plugin download checksum mismatch")
		}
	}
	return "plugin.jar", nil
}
