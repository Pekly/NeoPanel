package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestFileMutatorsExcludedByMaintenance(t *testing.T) {
	f := newReliabilityFixture(t)
	previewPut(t, f, "keep.txt", []byte("original"))
	dir := f.manager.GetWorkingDir("alpha")
	release, err := rooted.TryExclusiveMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	h := NewFileHandler(f.db, f.manager)
	for name, handler := range map[string]http.HandlerFunc{
		"save": h.SaveContent, "create": h.Create, "delete": h.Delete,
		"bulk": h.DeleteBulk, "upload": h.Upload, "rename": h.Rename,
		"move": h.Move, "compress": h.Compress, "unzip": h.Unzip,
		"properties": f.router.software.SaveProperties, "quick-config": f.router.servers.UpdateQuickConfigs,
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/?path=keep.txt", nil)
			req.SetPathValue("id", "alpha")
			w := httptest.NewRecorder()
			handler(w, req)
			if w.Code != http.StatusConflict {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	if w := f.request("GET", previewURL+"content?path=keep.txt", "", "owner"); w.Code != 200 {
		t.Fatal("reads blocked", w.Code)
	}
	got, err := os.ReadFile(filepath.Join(dir, "keep.txt"))
	if err != nil || string(got) != "original" {
		t.Fatal(string(got), err)
	}
	release()
	if w := f.request("POST", previewURL+"content", previewSaveBody(t, "keep.txt", "edited"), "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestPortRollbackExcludesCompetingEditor(t *testing.T) {
	f := portAPIFixture(t)
	path := filepath.Join(f.manager.GetWorkingDir("alpha"), "server.properties")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := previewSaveBody(t, "server.properties", "motd=editor\n")
	injected := errors.New("commit failed")
	err = f.router.servers.withPrimaryPortChange(httptest.NewRequest("POST", "/", nil), "alpha", func(apply database.PortConfigApply) error {
		if err := apply(30001); err != nil {
			return err
		}
		applied, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		done := make(chan *httptest.ResponseRecorder, 1)
		go func() { done <- f.request("POST", previewURL+"content", body, "owner") }()
		w := <-done
		if w.Code != http.StatusConflict {
			t.Errorf("competing editor: %d %s", w.Code, w.Body.String())
		}
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(applied) {
			t.Errorf("blocked editor changed config: %v", err)
		}
		return injected
	})
	if !errors.Is(err, injected) {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatal("rollback mismatch", err)
	}
	if w := f.request("POST", previewURL+"content", body, "owner"); w.Code != 200 {
		t.Fatal("editor did not resume", w.Code, w.Body.String())
	}
	after, err = os.ReadFile(path)
	if err != nil || string(after) != "motd=editor\n" {
		t.Fatal(string(after), err)
	}
}
