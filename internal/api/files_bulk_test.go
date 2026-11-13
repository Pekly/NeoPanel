package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

const bulkDeleteURL = "/api/servers/alpha/files/delete-bulk"

func bulkDeleteBody(t *testing.T, paths []string) string {
	t.Helper()
	body, err := json.Marshal(map[string][]string{"paths": paths})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func bulkDeleteFile(t *testing.T, root, path string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
}

func requireBulkDeleted(t *testing.T, body []byte, paths []string) {
	t.Helper()
	var result struct {
		Deleted []string `json:"deleted"`
		Failed  []struct {
			Path  string `json:"path"`
			Error string `json:"error"`
		} `json:"failed"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Deleted, paths) || result.Failed == nil || len(result.Failed) != 0 {
		t.Fatalf("unexpected response: %s", body)
	}
}

func TestBulkDeleteExceedsBurstInOneRequest(t *testing.T) {
	f := newReliabilityFixture(t)
	// Restore the production limiter disabled by the shared fixture.
	f.router.globalLimiter = NewRouter(f.cfg, f.db, f.manager, nil).globalLimiter
	lim := f.router.globalLimiter
	if lim.ratePerSec != 1 || lim.burstCapacity != 20 {
		t.Fatalf("production limiter changed: rate=%v burst=%v", lim.ratePerSec, lim.burstCapacity)
	}
	root := f.manager.GetWorkingDir("alpha")
	paths := make([]string, 50)
	for i := range paths {
		paths[i] = fmt.Sprintf("file-%d.txt", i)
		bulkDeleteFile(t, root, paths[i])
	}
	w := f.request(http.MethodPost, bulkDeleteURL, bulkDeleteBody(t, paths), "owner")
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	requireBulkDeleted(t, w.Body.Bytes(), paths)
	for _, path := range paths {
		if _, err := os.Stat(filepath.Join(root, path)); !os.IsNotExist(err) {
			t.Fatalf("%s not deleted: %v", path, err)
		}
	}
	logs, err := f.db.ListActivity(0, true, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].Action != "file_delete_bulk" {
		t.Fatalf("expected one bulk audit summary, got %+v", logs)
	}
	lim.mu.Lock()
	if len(lim.buckets) != 1 {
		lim.mu.Unlock()
		t.Fatal("bulk request bypassed global limiter")
	}
	for _, bucket := range lim.buckets {
		if bucket.tokens != 19 {
			lim.mu.Unlock()
			t.Fatalf("bulk request consumed more than one token: %v", bucket.tokens)
		}
		// Exhaust deterministically without relying on test execution speed.
		bucket.tokens = 0
		bucket.lastRefill = time.Now().Add(time.Hour)
	}
	lim.mu.Unlock()
	bulkDeleteFile(t, root, "limited.txt")
	w = f.request(http.MethodPost, bulkDeleteURL, `{"paths":["limited.txt"]}`, "owner")
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
		t.Fatal("limiter no longer enforced", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "limited.txt")); err != nil {
		t.Fatal("rate-limited request deleted file", err)
	}
}

func TestBulkDeletePermissions(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	bulkDeleteFile(t, root, "keep.txt")
	body := `{"paths":["keep.txt"]}`
	for _, tc := range []struct {
		user   string
		status int
	}{{"", 401}, {"stranger", 403}, {"subuser", 403}} {
		if w := f.request("POST", bulkDeleteURL, body, tc.user); w.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.user, w.Code, w.Body.String())
		}
	}
	subs, err := f.db.ListServerSubusers("alpha")
	if err != nil || len(subs) != 1 {
		t.Fatal(subs, err)
	}
	if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, []string{"files.edit"}); err != nil {
		t.Fatal(err)
	}
	if w := f.request("POST", bulkDeleteURL, body, "subuser"); w.Code != 403 {
		t.Fatal("files.edit must not allow deletion", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
		t.Fatal("unauthorized deletion", err)
	}
	if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, []string{"files.delete"}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"subuser", "owner", "admin"} {
		if w := f.request("POST", bulkDeleteURL, body, user); w.Code != 200 {
			t.Fatal(user, w.Code, w.Body.String())
		}
	}
	if w := f.request("POST", "/api/servers/beta/files/delete-bulk", body, "subuser"); w.Code != 403 {
		t.Fatal("permission leaked across servers", w.Code)
	}
}

func TestBulkDeleteInvalidRequestPreflight(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	bulkDeleteFile(t, root, "keep.txt")
	tooMany := make([]string, 1001)
	for i := range tooMany {
		tooMany[i] = "keep.txt"
	}
	cases := map[string]string{
		"empty body": "", "malformed": "{", "missing paths": "{}", "null": "null",
		"empty paths": `{"paths":[]}`, "null paths": `{"paths":null}`,
		"wrong type": `{"paths":"keep.txt"}`, "wrong element": `{"paths":["keep.txt",42]}`,
		"null element": `{"paths":["keep.txt",null]}`, "multiple objects": `{"paths":["keep.txt"]}{}`,
		"trailing garbage":     `{"paths":["keep.txt"]}oops`,
		"too many paths":       bulkDeleteBody(t, tooMany),
		"oversized path":       bulkDeleteBody(t, []string{"keep.txt", strings.Repeat("a", 1<<20)}),
		"oversized whitespace": `{"paths":["keep.txt"]}` + strings.Repeat(" ", 1<<20),
	}
	for _, alias := range []string{"", "/", ".", "./", "folder/..", "folder/../.", `folder\..`, "..", "../keep.txt", `..\keep.txt`, "bad\x00path"} {
		cases["invalid path "+fmt.Sprintf("%q", alias)] = bulkDeleteBody(t, []string{"keep.txt", alias})
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			w := f.request("POST", bulkDeleteURL, body, "owner")
			var result map[string]string
			if w.Code != 400 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result["error"] == "" {
				t.Fatal(w.Code, w.Body.String())
			}
			if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
				t.Fatal("invalid batch partially deleted files", err)
			}
		})
	}
}

func TestBulkDeleteNestedPathsAndIdempotency(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	for _, path := range []string{"parent/child.txt", "other/child.txt", "normal.txt", "keep.txt"} {
		bulkDeleteFile(t, root, path)
	}
	paths := []string{"parent", "parent/child.txt", "other/child.txt", "other", "./normal.txt", "/normal.txt", "normal.txt", "missing"}
	for i := 0; i < 2; i++ {
		w := f.request("POST", bulkDeleteURL, bulkDeleteBody(t, paths), "owner")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		requireBulkDeleted(t, w.Body.Bytes(), paths)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "keep.txt" {
		t.Fatal("unexpected directory contents", entries, err)
	}
	// The upper path-count boundary is accepted; each original entry is reported.
	paths = make([]string, 1000)
	for i := range paths {
		paths[i] = "missing"
	}
	w := f.request("POST", bulkDeleteURL, bulkDeleteBody(t, paths), "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	requireBulkDeleted(t, w.Body.Bytes(), paths)
	// Keep the existing single-delete response contract intact.
	w = f.request("DELETE", "/api/servers/alpha/files?path=keep.txt", "", "owner")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"message":"Deleted successfully"`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestBulkDeleteSymlinkPreflight(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	outside := t.TempDir()
	bulkDeleteFile(t, root, "keep.txt")
	bulkDeleteFile(t, outside, "outside.txt")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	for _, path := range []string{"link", "link/outside.txt", "link/missing.txt"} {
		w := f.request("POST", bulkDeleteURL, bulkDeleteBody(t, []string{"keep.txt", path}), "owner")
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
		for _, full := range []string{filepath.Join(root, "keep.txt"), filepath.Join(outside, "outside.txt")} {
			if _, err := os.Stat(full); err != nil {
				t.Fatal("preflight deleted file", err)
			}
		}
	}
}

func TestBulkDeletePartialFilesystemFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory write permissions unavailable on Windows")
	}
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	bulkDeleteFile(t, root, "locked/keep.txt")
	bulkDeleteFile(t, root, "delete.txt")
	locked := filepath.Join(root, "locked")
	if err := os.Chmod(locked, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(locked, 0755)
	if err := os.WriteFile(filepath.Join(locked, "probe"), nil, 0600); err == nil {
		t.Skip("process can bypass directory write permissions")
	}
	w := f.request("POST", bulkDeleteURL, `{"paths":["locked/keep.txt","delete.txt"]}`, "owner")
	var result struct {
		Deleted []string            `json:"deleted"`
		Failed  []bulkDeleteFailure `json:"failed"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	if !reflect.DeepEqual(result.Deleted, []string{"delete.txt"}) || len(result.Failed) != 1 || result.Failed[0].Path != "locked/keep.txt" || result.Failed[0].Error == "" {
		t.Fatal(w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(locked, "keep.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "delete.txt")); !os.IsNotExist(err) {
		t.Fatal("successful path not deleted", err)
	}
}
