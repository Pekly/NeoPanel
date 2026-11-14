package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSoftwareTraversalRejectedBeforeCleanup(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	victim := filepath.Join(f.manager.GetWorkingDir("beta"), "server.jar")
	os.WriteFile(victim, []byte("secret"), 0600)
	os.WriteFile(filepath.Join(dir, "keep"), []byte("keep"), 0600)
	for _, route := range []string{"software", "reinstall"} {
		for _, version := range []string{"../../../beta/server", `..\..\..\beta\server`} {
			body, _ := json.Marshal(map[string]string{"software": "bungeecord", "version": version})
			w := f.request("POST", "/api/servers/alpha/"+route, string(body), "owner")
			if w.Code != 400 {
				t.Fatal(route, w.Code, w.Body.String())
			}
		}
	}
	b, _ := os.ReadFile(victim)
	if string(b) != "secret" {
		t.Fatal("sibling overwritten")
	}
	if _, e := os.Stat(filepath.Join(dir, "keep")); e != nil {
		t.Fatal("cleanup occurred", e)
	}
}
func TestPropertiesAndEditorHardlinks(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	outside := filepath.Join(t.TempDir(), "victim")
	os.WriteFile(outside, []byte("motd=secret\n"), 0600)
	for _, tc := range []struct{ route, name, body string }{{"properties", "server.properties", `{"motd":"new"}`}, {"configs", "server.properties", `{"motd":"new","eula_accepted":true}`}, {"files/content", "edit.txt", `{"path":"edit.txt","content":"new"}`}} {
		name := filepath.Join(dir, tc.name)
		os.Remove(name)
		if e := os.Link(outside, name); e != nil {
			t.Fatal(e)
		}
		w := f.request("POST", "/api/servers/alpha/"+tc.route, tc.body, "owner")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		b, _ := os.ReadFile(outside)
		if string(b) != "motd=secret\n" {
			t.Fatal("hardlink modified")
		}
	}
	for _, route := range []string{"properties", "configs"} {
		w := f.request("POST", "/api/servers/alpha/"+route, `{"motd":"new\nserver-port=1"}`, "owner")
		if w.Code != 400 {
			t.Fatal("accepted injected line", w.Code)
		}
	}
}
func TestPropertiesSymlinkReadWriteRejected(t *testing.T) {
	f := newReliabilityFixture(t)
	dir := f.manager.GetWorkingDir("alpha")
	outside := filepath.Join(t.TempDir(), "victim")
	os.WriteFile(outside, []byte("motd=secret\n"), 0600)
	if e := os.Symlink(outside, filepath.Join(dir, "server.properties")); e != nil {
		t.Skip(e)
	}
	for _, route := range []string{"properties", "configs"} {
		for _, method := range []string{"GET", "POST"} {
			w := f.request(method, "/api/servers/alpha/"+route, `{"motd":"new"}`, "owner")
			if w.Code < 400 {
				t.Fatal("accepted symlink", method, route, w.Code)
			}
			if strings.Contains(w.Body.String(), "motd=secret") {
				t.Fatal("secret disclosed")
			}
		}
	}
	b, _ := os.ReadFile(outside)
	if string(b) != "motd=secret\n" {
		t.Fatal("outside modified")
	}
}
func TestMultipartTotalBodyCap(t *testing.T) {
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	part, e := mw.CreateFormFile("file", "big.txt")
	if e != nil {
		t.Fatal(e)
	}
	part.Write(bytes.Repeat([]byte("x"), 2048))
	mw.Close()
	r := httptest.NewRequest("POST", "/", &b)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	defer func() {
		if r.MultipartForm != nil {
			r.MultipartForm.RemoveAll()
		}
	}()
	if e := parseUploadForm(w, r, 1024); e == nil {
		t.Fatal("multipart body cap bypassed")
	}
}
