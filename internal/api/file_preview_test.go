package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/png"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

const previewURL = "/api/servers/alpha/files/"

func previewPut(t *testing.T, f *reliabilityFixture, name string, data []byte) {
	t.Helper()
	full := filepath.Join(f.manager.GetWorkingDir("alpha"), name)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func previewPNG(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func previewSaveBody(t *testing.T, name, content string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"path": name, "content": content})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFilePreviewClassificationAndEditor(t *testing.T) {
	f := newReliabilityFixture(t)
	tests := []struct {
		name, content, kind string
		editable            bool
	}{
		{"server.properties", "# config\nmotd=Hello 世界\nmax-players=20\n", "text", true},
		{"empty.unknown", "", "text", true},
		{"settings.custom", "enabled: true\n", "text", true},
		{"latest.log", "\x1b[31mError\x1b[0m\r\n", "text", true},
		{"icon.svg", "<svg xmlns=\"http://www.w3.org/2000/svg\"><script>alert(1)</script></svg>", "text", true},
		{"page.html", "<!doctype html><script>alert(1)</script>", "text", true},
		{"page.png", "<!doctype html><script>alert(1)</script>", "binary", false},
		{"late-null.txt", strings.Repeat("a", 4096) + "\x00", "binary", false},
		{"late-control.txt", strings.Repeat("a", 4096) + "\x01", "binary", false},
		{"late-c1.txt", strings.Repeat("a", 4096) + "\u0085", "binary", false},
		{"late-utf8.txt", strings.Repeat("a", 4096) + "\xff", "binary", false},
		{"escape.txt", "log\x1bgarbage", "binary", false},
		{"database.txt", "SQLite format 3\x00binary", "binary", false},
		{"document.txt", "%PDF-1.4\nASCII PDF data", "binary", false},
		{"executable.txt", "MZ" + strings.Repeat("a", 64), "binary", false},
		{"world.mca", "", "binary", false},
		{"renamed.txt", string(previewPNG(t)), "image", false},
		{"plugin.jar", "PK\x03\x04\x00\x00", "zip", false},
		{"overlimit.txt", strings.Repeat("a", fileEditorLimit+1), "text", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			previewPut(t, f, tc.name, []byte(tc.content))
			w := f.request("GET", previewURL+"preview?path="+tc.name, "", "subuser")
			var p filePreview
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
				t.Fatal(w.Code, w.Body.String())
			}
			if p.Path != tc.name || p.Name != tc.name || p.Size != int64(len(tc.content)) || p.Kind != tc.kind || p.Editable != tc.editable || p.MIME == "" {
				t.Fatalf("unexpected metadata: %+v", p)
			}
			if !p.Editable && p.Reason == "" {
				t.Fatal("missing reason")
			}
			w = f.request("GET", previewURL+"content?path="+tc.name, "", "subuser")
			want := 415
			if tc.editable {
				want = 200
			}
			if len(tc.content) > fileEditorLimit {
				want = 400
			}
			if w.Code != want {
				t.Fatal(w.Code, w.Body.String())
			}
			if tc.editable {
				var result map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result["content"] != tc.content {
					t.Fatal("text changed", err)
				}
			}
		})
	}
}

func TestFilePreviewSaveSafety(t *testing.T) {
	f := newReliabilityFixture(t)
	original := append(bytes.Repeat([]byte("x"), 1024), 0)
	previewPut(t, f, "renamed.txt", original)
	if w := f.request("POST", previewURL+"content", previewSaveBody(t, "renamed.txt", "replacement"), "owner"); w.Code != 415 {
		t.Fatal(w.Code, w.Body.String())
	}
	got, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "renamed.txt"))
	if err != nil || !bytes.Equal(got, original) {
		t.Fatal("binary was modified", err)
	}
	for _, tc := range []struct {
		name, content string
		code          int
	}{
		{"bad.txt", "a\x00b", 415}, {"bad.txt", "a\x01b", 415}, {"bad.txt", string(previewPNG(t)), 415},
		{"new.jar", "plain text", 415}, {"new.mca", "", 415}, {"new.db", "plain", 415},
		{"new.txt", strings.Repeat("x", fileEditorLimit+1), 413},
		{"new.custom", "key=世界\n", 200}, {"new.custom", "short", 200}, {"new.custom", "", 200},
	} {
		w := f.request("POST", previewURL+"content", previewSaveBody(t, tc.name, tc.content), "owner")
		if w.Code != tc.code {
			t.Fatalf("%s: %d %s", tc.name, w.Code, w.Body.String())
		}
		if tc.code == 200 {
			data, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), tc.name))
			if err != nil || string(data) != tc.content {
				t.Fatal("save mismatch", err)
			}
		}
	}
	for _, body := range []string{"{\"path\":\"bad.txt\",\"content\":\"\xff\"}", `{"path":"new.custom","content":"keep"} {}`} {
		if w := f.request("POST", previewURL+"content", body, "owner"); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := f.request("POST", previewURL+"content", strings.Repeat(" ", 6*fileEditorLimit+(64<<10)+1), "owner"); w.Code != 413 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestFilePreviewMediaAndPermissions(t *testing.T) {
	f := newReliabilityFixture(t)
	data := previewPNG(t)
	previewPut(t, f, "image.txt", data)
	mp4 := append([]byte{0, 0, 0, 24}, []byte("ftypmp42\x00\x00\x00\x00mp42isom")...)
	previewPut(t, f, "video.data", mp4)
	for _, tc := range []struct {
		name, mime string
		data       []byte
	}{{"image.txt", "image/png", data}, {"video.data", "video/mp4", mp4}} {
		for _, method := range []string{"GET", "HEAD"} {
			req := httptest.NewRequest(method, previewURL+"media?path="+tc.name+"&token="+url.QueryEscape(f.tokens["subuser"]), nil)
			req.Header.Set("Range", "bytes=2-5")
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, req)
			if w.Code != 206 || w.Header().Get("Content-Type") != tc.mime || w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("Cache-Control") != "private, no-store" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "inline;") || w.Header().Get("Content-Range") == "" {
				t.Fatal(w.Code, w.Header(), w.Body.String())
			}
			if method == "GET" && !bytes.Equal(w.Body.Bytes(), tc.data[2:6]) || method == "HEAD" && w.Body.Len() != 0 {
				t.Fatal("incorrect range/HEAD body")
			}
		}
	}
	previewPut(t, f, "café;data.png", data)
	w := f.request("GET", previewURL+"media?path="+url.QueryEscape("café;data.png"), "", "owner")
	_, params, err := mime.ParseMediaType(w.Header().Get("Content-Disposition"))
	if w.Code != 200 || err != nil || params["filename"] != "café;data.png" || strings.ContainsAny(w.Header().Get("Content-Disposition"), "\r\n") {
		t.Fatal("unsafe filename header", w.Code, w.Header(), err)
	}
	req := httptest.NewRequest("GET", previewURL+"media?path=image.txt", nil)
	req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
	req.Header.Set("Range", "bytes=999999-9999999")
	w = httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	if w.Code != 416 || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("range error caching", w.Code, w.Header())
	}
	for _, name := range []string{"active.html", "active.svg", "active.xml", "spoof.png", "spoof.mp4"} {
		previewPut(t, f, name, []byte("<html><script>alert(1)</script></html>"))
		if w := f.request("GET", previewURL+"media?path="+name, "", "owner"); w.Code != 415 {
			t.Fatal(name, w.Code)
		}
	}
	for _, route := range []string{"preview", "media", "zip", "zip-entry"} {
		for _, method := range []string{"GET", "HEAD"} {
			for _, user := range []string{"", "stranger"} {
				want := 403
				if user == "" {
					want = 401
				}
				if w := f.request(method, previewURL+route+"?path=image.txt&index=0", "", user); w.Code != want {
					t.Fatal(route, method, user, w.Code)
				}
			}
		}
	}
}

func TestFilePreviewSignatures(t *testing.T) {
	ogv := make([]byte, 28)
	copy(ogv, "OggS")
	ogv[26] = 1
	ogv[27] = 7
	ogv = append(ogv, []byte("\x80theora")...)
	mp4 := make([]byte, 24)
	binary.BigEndian.PutUint32(mp4, 24)
	copy(mp4[4:], "ftypisom\x00\x00\x00\x00isomavc1")
	for _, tc := range []struct {
		data       []byte
		kind, mime string
	}{
		{[]byte("\x89PNG\r\n\x1a\n"), "image", "image/png"},
		{[]byte("\xff\xd8\xff\xe0"), "image", "image/jpeg"},
		{[]byte("GIF89a"), "image", "image/gif"},
		{[]byte("RIFF\x00\x00\x00\x00WEBPVP8 "), "image", "image/webp"},
		{[]byte("BM"), "image", "image/bmp"},
		{[]byte{0, 0, 1, 0}, "image", "image/x-icon"},
		{mp4, "video", "video/mp4"},
		{[]byte{0x1a, 0x45, 0xdf, 0xa3, 0x87, 0x42, 0x82, 0x84, 'w', 'e', 'b', 'm'}, "video", "video/webm"},
		{[]byte{0x1a, 0x45, 0xdf, 0xa3, 0x80, 0x42, 0x82, 0x84, 'w', 'e', 'b', 'm'}, "binary", "application/octet-stream"},
		{ogv, "video", "video/ogg"},
		{[]byte("OggS\x00\x00audio"), "binary", "application/octet-stream"},
		{[]byte("\x1a\x45\xdf\xa3matroska"), "binary", "application/octet-stream"},
	} {
		kind, m := sniffFileKind(tc.data)
		if kind != tc.kind || m != tc.mime {
			t.Fatalf("%x: %s %s", tc.data, kind, m)
		}
	}
}

func TestFilePreviewPaths(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	if err := os.Mkdir(filepath.Join(root, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"../secret.txt", "folder/../../secret.txt", "..\\secret.txt", "folder", "/"} {
		for _, route := range []string{"preview", "media", "zip", "zip-entry", "content"} {
			if w := f.request("GET", previewURL+route+"?index=0&path="+url.QueryEscape(name), "", "owner"); w.Code < 400 {
				t.Fatal(route, name, w.Code)
			}
		}
		if w := f.request("POST", previewURL+"content", previewSaveBody(t, name, "text"), "owner"); w.Code < 400 {
			t.Fatal("saved unsafe path", name)
		}
	}
	t.Run("symlinks", func(t *testing.T) {
		outside := t.TempDir()
		target := filepath.Join(outside, "secret.txt")
		if err := os.WriteFile(target, []byte("keep"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(root, "link.txt")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"link.txt", "linked/secret.txt"} {
			for _, route := range []string{"preview", "media", "zip", "zip-entry", "content"} {
				if w := f.request("GET", previewURL+route+"?index=0&path="+name, "", "owner"); w.Code < 400 {
					t.Fatal("followed link", route, name)
				}
			}
			if w := f.request("POST", previewURL+"content", previewSaveBody(t, name, "changed"), "owner"); w.Code < 400 {
				t.Fatal("saved link")
			}
		}
		data, _ := os.ReadFile(target)
		if string(data) != "keep" {
			t.Fatal("outside file modified")
		}
	})
}

type previewTransport func(*http.Request) (*http.Response, error)

func (f previewTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFilePreviewProxyRangeAndAuth(t *testing.T) {
	master, node := newReliabilityFixture(t), newReliabilityFixture(t)
	previewPut(t, node, "image.png", previewPNG(t))
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "remote", BaseURL: "http://node.invalid", APIToken: node.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	master.router.nodes.httpClient = &http.Client{Transport: previewTransport(func(req *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		node.router.ServeHTTP(w, req)
		return w.Result(), nil
	})}
	for _, method := range []string{"GET", "HEAD"} {
		req := httptest.NewRequest(method, "/api/nodes/remote/proxy/api/servers/alpha/files/media?path=image.png&token="+url.QueryEscape(master.tokens["subuser"]), nil)
		req.Header.Set("Range", "bytes=0-3")
		w := httptest.NewRecorder()
		master.router.ServeHTTP(w, req)
		if w.Code != 206 || w.Header().Get("Content-Range") == "" || w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("Content-Security-Policy") == "" {
			t.Fatal(method, w.Code, w.Header(), w.Body.String())
		}
		if method == "HEAD" && w.Body.Len() != 0 {
			t.Fatal("proxy HEAD has body")
		}
		for _, route := range []string{"preview", "media", "zip", "zip-entry"} {
			for _, user := range []string{"", "stranger"} {
				want := 403
				if user == "" {
					want = 401
				}
				if w := master.request(method, "/api/nodes/remote/proxy/api/servers/alpha/files/"+route+"?path=image.png&index=0", "", user); w.Code != want {
					t.Fatal(route, method, user, w.Code, w.Body.String())
				}
			}
		}
	}
}
