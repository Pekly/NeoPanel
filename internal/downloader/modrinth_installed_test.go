package downloader

import (
	"context"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestModrinthInstalledExactHashesAndBounds(t *testing.T) {
	hash := strings.Repeat("a", 128)
	v := catalogVersion("v1", "P1", "paper", "any", "2025-01-01T00:00:00Z")
	v.Files[0].Hashes = map[string]string{"sha512": hash}
	mismatch := false
	mockModrinth(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != "POST" || r.URL.Path != "/v2/version_files" {
			t.Fatal(r.Method, r.URL)
		}
		var req struct {
			Hashes    []string `json:"hashes"`
			Algorithm string   `json:"algorithm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Algorithm != "sha512" || len(req.Hashes) != 1 || req.Hashes[0] != hash {
			t.Fatal(req, err)
		}
		if mismatch {
			v.Files[0].Hashes["sha512"] = strings.Repeat("b", 128)
		}
		b, _ := json.Marshal(map[string]ModrinthVersion{hash: v, strings.Repeat("c", 128): v})
		return modrinthResponse(string(b)), nil
	})
	found, err := LookupModrinthHashes(context.Background(), []string{hash})
	if err != nil || len(found) != 1 || found[hash].ID != "v1" {
		t.Fatal(found, err)
	}
	mismatch = true
	if _, err := LookupModrinthHashes(context.Background(), []string{hash}); err == nil {
		t.Fatal("accepted mismatched file digest")
	}
	if _, err := LookupModrinthHashes(context.Background(), make([]string, 21)); err == nil {
		t.Fatal("unbounded hashes")
	}
	if _, err := LookupModrinthHashes(context.Background(), []string{"invalid"}); err == nil {
		t.Fatal("invalid hash")
	}
}

func TestModrinthInstalledSharedMetadataBudget(t *testing.T) {
	calls := 0
	mockModrinth(t, func(r *http.Request) (*http.Response, error) { calls++; return modrinthResponse(`{"id":"P1"}`), nil })
	ctx := ModrinthMetadataBudgetContext(context.Background())
	budget := ctx.Value(modrinthBudgetKey{}).(*modrinthBudget)
	budget.remaining = 1
	if _, err := GetModrinthProjectContext(ctx, "P1"); err == nil {
		t.Fatal("nested call reset budget")
	}
	if _, err := GetModrinthProjectContext(ctx, "P1"); err == nil || calls != 1 {
		t.Fatal("exhausted budget still requested", calls, err)
	}
}

func TestModrinthInstalledStageChecksumsSizeAndBudget(t *testing.T) {
	for _, scenario := range []string{"ok", "sha512", "sha1", "short", "long", "budget", "missing-checksum"} {
		t.Run(scenario, func(t *testing.T) {
			v := catalogVersion("v1", "P1", "paper", "any", "2025-01-01T00:00:00Z")
			h512, h1 := sha512.Sum512([]byte("jar")), sha1.Sum([]byte("jar"))
			v.Files[0].Hashes = map[string]string{"sha512": hex.EncodeToString(h512[:]), "sha1": hex.EncodeToString(h1[:])}
			data := "jar"
			budget := int64(10)
			switch scenario {
			case "missing-checksum":
				v.Files[0].Hashes = nil
			case "sha512":
				v.Files[0].Hashes["sha512"] = strings.Repeat("a", 128)
			case "sha1":
				v.Files[0].Hashes["sha1"] = strings.Repeat("b", 40)
			case "short":
				data = "ja"
			case "long":
				data = "jars"
			case "budget":
				budget = 2
			}
			calls := 0
			mockModrinth(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "cdn.modrinth.com" {
					t.Fatal(r.URL)
				}
				return modrinthResponse(data), nil
			})
			root, err := rooted.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			_, err = StageModrinthUpdate(context.Background(), &v, root, &budget)
			if (err == nil) != (scenario == "ok") {
				t.Fatal(scenario, err)
			}
			if (scenario == "budget" || scenario == "missing-checksum") && calls != 0 {
				t.Fatal("download exceeded admitted budget")
			}
			if scenario == "ok" && budget != 7 {
				t.Fatal("download budget not charged", budget)
			}
		})
	}
}
