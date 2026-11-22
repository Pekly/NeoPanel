package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestPluginMutationsRespectExclusiveServerLease(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	dir := filepath.Join(root, "plugins")
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	calls := 0
	failProvider := false
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		// Install must retain the shared lease throughout metadata and JAR requests.
		release, err := rooted.TryExclusiveMutation(root)
		if err == nil {
			release()
			t.Fatal("install released its mutation lease before download completed")
		}
		if !errors.Is(err, rooted.ErrMutationBusy) {
			t.Fatal(err)
		}
		if failProvider {
			return nil, fmt.Errorf("mock provider unavailable")
		}
		data := ""
		switch r.URL.String() {
		case "https://api.modrinth.com/v2/project/example":
			data = `{"id":"P1","slug":"example"}`
		case "https://api.modrinth.com/v2/project/P1/version":
			data = `[{"id":"v1","project_id":"P1","loaders":["paper"],"files":[{"filename":"installed.jar","url":"https://cdn.modrinth.com/installed.jar","size":3}]}]`
		case "https://cdn.modrinth.com/installed.jar":
			data = "jar"
		default:
			t.Fatalf("unexpected provider request: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	assertLeaseReleased := func() {
		t.Helper()
		release, err := rooted.TryExclusiveMutation(root)
		if err != nil {
			t.Fatalf("handler leaked mutation lease: %v", err)
		}
		release()
	}
	assertConflict := func(method, path, body string) {
		t.Helper()
		w := f.request(method, path, body, "owner")
		if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "server files are busy") || !strings.Contains(w.Body.String(), "retry") {
			t.Fatalf("expected friendly conflict: %d %s", w.Code, w.Body.String())
		}
		if calls != 0 {
			t.Fatalf("provider contacted during exclusive mutation: %d calls", calls)
		}
	}
	const base = "/api/servers/alpha/plugins"
	release, err := rooted.TryExclusiveMutation(root)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	assertConflict("POST", base+"/install", `{"slug":"example"}`)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("blocked install created plugin directory: %v", err)
	}
	// Seed files directly as the exclusive operation, then ensure handlers cannot
	// remove or rename them while that operation still owns the server root.
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"delete.jar", "toggle.jar"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("unchanged"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	assertConflict("DELETE", base+"?filename=delete.jar", "")
	assertConflict("POST", base+"/toggle", `{"filename":"toggle.jar","enabled":false}`)
	for _, name := range []string{"delete.jar", "toggle.jar"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(data) != "unchanged" {
			t.Fatalf("blocked mutation changed %s: %q %v", name, data, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("blocked mutation left artifacts: %v %v", entries, err)
	}
	release()
	for _, req := range []struct{ method, path, body string }{
		{"POST", base + "/install", `{"slug":"example"}`},
		{"POST", base + "/toggle", `{"filename":"toggle.jar","enabled":false}`},
		{"DELETE", base + "?filename=delete.jar", ""},
	} {
		w := f.request(req.method, req.path, req.body, "owner")
		if w.Code != http.StatusOK {
			t.Fatalf("after release: %d %s", w.Code, w.Body.String())
		}
		assertLeaseReleased()
	}
	if calls != 3 {
		t.Fatalf("expected project, versions and JAR fetch: %d calls", calls)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "installed.jar")); err != nil || string(data) != "jar" {
		t.Fatalf("install: %q %v", data, err)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "toggle.jar.disabled")); err != nil || string(data) != "unchanged" {
		t.Fatalf("toggle: %q %v", data, err)
	}
	for _, name := range []string{"delete.jar", "toggle.jar"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("mutation did not remove %s: %v", name, err)
		}
	}
	// Failure returns must release the shared lease as well.
	failProvider = true
	for _, req := range []struct {
		method, path, body string
		status             int
	}{
		{"POST", base + "/install", `{"slug":"example"}`, http.StatusBadRequest},
		{"DELETE", base + "?filename=missing.jar", "", http.StatusInternalServerError},
		{"POST", base + "/toggle", `{"filename":"missing.jar","enabled":false}`, http.StatusConflict},
	} {
		w := f.request(req.method, req.path, req.body, "owner")
		if w.Code != req.status {
			t.Fatalf("failure response: %d %s", w.Code, w.Body.String())
		}
		assertLeaseReleased()
	}
}
