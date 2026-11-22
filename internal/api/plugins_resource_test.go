package api

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestPluginZipDirectoryCapsAndForgedCount(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for i := 0; i < 10001; i++ {
		if _, err := z.Create(fmt.Sprintf("entry%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	original := b.Bytes()
	for _, count := range []uint16{10001, 1} {
		data := append([]byte{}, original...)
		binary.LittleEndian.PutUint16(data[len(data)-22+10:], count)
		path := filepath.Join(t.TempDir(), "hostile.jar")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		_, err = pluginZip(f, int64(len(data)))
		f.Close()
		if err == nil {
			t.Fatal("accepted excessive or forged central-directory entries")
		}
	}
}

func TestPluginDigestPageBudgetAndDiskAdmission(t *testing.T) {
	root, err := rooted.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := root.WriteFile("small.jar", []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	budget := int64(2)
	if _, _, err := pluginDigest(context.Background(), root, "small.jar", &budget); err == nil {
		t.Fatal("aggregate read limit ignored")
	}
	if err := pluginUpdateDiskBudget(context.Background(), root, 1, 1<<20); err == nil {
		t.Fatal("existing disk bytes not accounted")
	}
	if err := pluginUpdateDiskBudget(context.Background(), root, 1, 100); err != nil {
		t.Fatal(err)
	}
}

func TestPluginUpdateIdenticalPayloadAndCancellation(t *testing.T) {
	for _, scenario := range []string{"same", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			f := newReliabilityFixture(t)
			m := installedMock(t)
			old := installedTestJar(t, "plugin.yml", "name: Example\nversion: custom")
			installedTestWrite(t, f, "upload.jar.disabled", old)
			item := pluginIdentityRequest{"upload.jar.disabled", installedTestHash(old), "P1", "next"}
			if err := f.db.ConfirmPluginSource("alpha", item.Filename, item.SHA512, "P1"); err != nil {
				t.Fatal(err)
			}
			target := installedTestVersion(t, "next", "P1", "paper", "2025-01-01T00:00:00Z", old)
			target.Files[0].Hashes = map[string]string{"sha1": fmt.Sprintf("%x", sha1.Sum(old))} // Verify download, then detect identical bytes by SHA512.
			m.versions["next"] = target
			m.downloads["/next.jar"] = old
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if scenario == "cancel" {
				m.onDownload = cancel
			}
			req := httptest.NewRequest("POST", "/api/servers/alpha/plugins/update", strings.NewReader(installedUpdateBody(item))).WithContext(ctx)
			req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, req)
			var result pluginUpdateResponse
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.RestartRequired || len(result.Results) != 1 {
				t.Fatal(w.Code, w.Body.String())
			}
			want := "skipped"
			if scenario == "cancel" {
				want = "failed"
			}
			if result.Results[0].Status != want || result.Results[0].BackupPath != "" {
				t.Fatal(w.Body.String())
			}
			got, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "plugins", item.Filename))
			if err != nil || !bytes.Equal(got, old) {
				t.Fatal("original lost", err)
			}
			unlock, err := rooted.TryExclusiveMutation(f.manager.GetWorkingDir("alpha"))
			if err != nil {
				t.Fatal("gate leaked", err)
			}
			unlock()
		})
	}
}

func TestPluginCatalogFilenameFallbackAndPageCap(t *testing.T) {
	f := newReliabilityFixture(t)
	m := installedMock(t)
	jar := installedTestJar(t, "", "")
	for i := 0; i < 21; i++ {
		installedTestWrite(t, f, fmt.Sprintf("Example-v1.%02d.jar.disabled", i), jar)
	}
	page := installedScan(t, f, "?limit=999")
	if page.Total != 21 || len(page.Items) != 20 || !page.HasMore || m.searches != 20 {
		t.Fatal(page, m.searches)
	}
	for _, item := range page.Items {
		if item.Match != "suggested" || item.Name != "Example" || item.Enabled {
			t.Fatal(item)
		}
	}
}
