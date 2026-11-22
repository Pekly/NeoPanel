package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFilesContentModifiedAt(t *testing.T) {
	f := newReliabilityFixture(t)
	previewPut(t, f, "timestamp.txt", []byte("original"))
	full := filepath.Join(f.manager.GetWorkingDir("alpha"), "timestamp.txt")
	old := time.Date(2020, 2, 3, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes(full, old, old); err != nil {
		t.Fatal(err)
	}
	read := func() map[string]string {
		t.Helper()
		w := f.request(http.MethodGet, previewURL+"content?path=timestamp.txt", "", "owner")
		var result map[string]string
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &result) != nil {
			t.Fatal(w.Code, w.Body.String())
		}
		return result
	}
	loaded := read()
	if loaded["path"] != "timestamp.txt" || loaded["content"] != "original" || loaded["modified_at"] != old.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected read: %#v", loaded)
	}
	w := f.request(http.MethodPost, previewURL+"content", previewSaveBody(t, "timestamp.txt", "edited"), "owner")
	var saved map[string]string
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &saved) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	if saved["message"] != "File saved successfully" {
		t.Fatal(saved)
	}
	stamp, err := time.Parse(time.RFC3339Nano, saved["modified_at"])
	if err != nil || !stamp.After(old) {
		t.Fatal(saved, err)
	}
	loaded = read()
	if loaded["content"] != "edited" || loaded["modified_at"] != saved["modified_at"] {
		t.Fatal(loaded, saved)
	}
	w = f.request(http.MethodPost, previewURL+"content", previewSaveBody(t, "timestamp.txt", "\x00"), "owner")
	if w.Code == http.StatusOK {
		t.Fatal("binary save unexpectedly succeeded")
	}
	if loadedAfter := read(); loadedAfter["modified_at"] != loaded["modified_at"] || loadedAfter["content"] != "edited" {
		t.Fatal(loadedAfter)
	}
}
