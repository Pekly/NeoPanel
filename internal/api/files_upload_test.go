package api

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

// Match transferUpload's actual wire contract: file first, path in the body,
// POST /files/upload with no query string. No live listener or game is started.
func uploadRequest(t *testing.T, query, name string, paths ...string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write([]byte("uploaded content")); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if err = form.WriteField("path", path); err != nil {
			t.Fatal(err)
		}
	}
	if err = form.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/servers/alpha/files/upload"+query, &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.SetPathValue("id", "alpha")
	return req
}

func TestUploadMultipartDestination(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	for _, tc := range []struct {
		name, query, dest string
		paths             []string
	}{
		{"nested-body", "", "plugins/config", []string{"plugins/config"}},
		{"unicode-spaces", "", "世界/My folder", []string{"世界/My folder"}},
		{"explicit-root", "", "", []string{""}},
		{"legacy-root", "", "", nil},
		{"legacy-query", "?path=" + url.QueryEscape("世界/My folder"), "世界/My folder", nil},
		{"matching-query-body", "?path=plugins/config", "plugins/config", []string{"plugins/config"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.MkdirAll(filepath.Join(root, tc.dest), 0755); err != nil {
				t.Fatal(err)
			}
			req := uploadRequest(t, tc.query, tc.name+".txt", tc.paths...)
			req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, req)
			if w.Code != 200 {
				t.Fatalf("%d: %s", w.Code, w.Body.String())
			}
			got, err := os.ReadFile(filepath.Join(root, tc.dest, tc.name+".txt"))
			if err != nil || string(got) != "uploaded content" {
				t.Fatalf("destination %q: %q, %v", tc.dest, got, err)
			}
			if tc.dest != "" {
				if _, err := os.Stat(filepath.Join(root, tc.name+".txt")); !os.IsNotExist(err) {
					t.Fatal("upload leaked to root", err)
				}
			}
		})
	}
}

func TestUploadRejectsAmbiguousUnsafeAndExistingDestinations(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	h := NewFileHandler(f.db, f.manager)
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "keep.txt"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, query, file string
		paths             []string
		code              int
	}{
		{"conflict", "?path=nested", "new.txt", []string{""}, 400},
		{"duplicate-body", "", "new.txt", []string{"nested", ""}, 400},
		{"duplicate-query", "?path=nested&path=", "new.txt", nil, 400},
		{"traversal", "", "new.txt", []string{"../beta"}, 400},
		{"cleaned-traversal", "", "new.txt", []string{"nested/../"}, 400},
		{"backslash-traversal", "", "new.txt", []string{`..\beta`}, 400},
		{"nul", "", "new.txt", []string{"nested\x00"}, 400},
		{"drive", "", "new.txt", []string{"C:/temp"}, 400},
		{"filename", "", "bad:name", []string{"nested"}, 400},
		{"existing", "", "keep.txt", []string{"nested"}, 409},
		{"missing-folder", "", "new.txt", []string{"missing"}, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.Upload(w, uploadRequest(t, tc.query, tc.file, tc.paths...))
			if w.Code != tc.code {
				t.Fatalf("%d: %s", w.Code, w.Body.String())
			}
		})
	}
	got, err := os.ReadFile(filepath.Join(dir, "nested", "keep.txt"))
	if err != nil || string(got) != "original" {
		t.Fatal("existing file changed", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "new.txt")); !os.IsNotExist(err) {
		t.Fatal("rejected upload fell back to root", err)
	}
	stages, _ := filepath.Glob(filepath.Join(dir, ".neopanel-*"))
	if len(stages) != 0 {
		t.Fatal("staging leaked", stages)
	}
}

func TestUploadPermissionsLeaseCancellationAndLinks(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	h := NewFileHandler(f.db, f.manager)
	for _, user := range []string{"subuser", "stranger", ""} {
		req := uploadRequest(t, "", "denied.txt", "")
		if user != "" {
			req.Header.Set("Authorization", "Bearer "+f.tokens[user])
		}
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, req)
		if w.Code != 401 && w.Code != 403 {
			t.Fatal(user, w.Code, w.Body.String())
		}
	}
	release, err := rooted.TryExclusiveMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.Upload(w, uploadRequest(t, "", "locked.txt", ""))
	release()
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w = httptest.NewRecorder()
	h.Upload(w, uploadRequest(t, "", "cancelled.txt", "").WithContext(ctx))
	if w.Code < 400 {
		t.Fatal("cancelled upload published")
	}
	outside := filepath.Join(t.TempDir(), "victim")
	if err = os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Link(outside, filepath.Join(dir, "hard.txt")); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.Upload(w, uploadRequest(t, "", "hard.txt", ""))
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	got, err := os.ReadFile(outside)
	if err != nil || string(got) != "secret" {
		t.Fatal("outside hardlink changed", err)
	}
	for _, name := range []string{"denied.txt", "locked.txt", "cancelled.txt"} {
		if _, err = os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatal(name, "published", err)
		}
	}
	t.Run("symlink", func(t *testing.T) {
		if err := os.Symlink(filepath.Dir(outside), filepath.Join(dir, "link")); err != nil {
			t.Skip(err)
		}
		w := httptest.NewRecorder()
		h.Upload(w, uploadRequest(t, "", "victim", "link"))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	})
}

// Optional end-to-end replay of the exact body serialized by real Chrome in
// web/tests/file-drag.chrome.test.cjs; still entirely in-process with temp roots.
func TestUploadChromeMultipartWire(t *testing.T) {
	fixture := os.Getenv("NEOPANEL_UPLOAD_WIRE")
	if fixture == "" {
		t.Skip("set NEOPANEL_UPLOAD_WIRE to upload-chrome-wire.json from the Chrome test")
	}
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct{ ContentType, Body, Destination, Filename string }
	if err = json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.Destination != "nested/My space/世界 folder" || wire.Filename != "wire.txt" {
		t.Fatal("unexpected wire fixture")
	}
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	if err = os.MkdirAll(filepath.Join(dir, wire.Destination), 0755); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/servers/alpha/files/upload", strings.NewReader(wire.Body))
	req.Header.Set("Content-Type", wire.ContentType)
	req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("%d: %s", w.Code, w.Body.String())
	}
	if got, err := os.ReadFile(filepath.Join(dir, wire.Destination, wire.Filename)); err != nil || string(got) != "uploaded content" {
		rootBytes, rootErr := os.ReadFile(filepath.Join(dir, wire.Filename))
		t.Fatalf("chosen destination missing: %v; root file = %q (%v)", err, rootBytes, rootErr)
	}
	if _, err = os.Stat(filepath.Join(dir, wire.Filename)); !os.IsNotExist(err) {
		t.Fatal("upload leaked to root", err)
	}
}
