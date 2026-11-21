package api

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func avatarRequest(t *testing.T, name string, payload []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("avatar", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/api/account/avatar", bytes.NewReader(body.Bytes()))
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}
func avatarPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestAvatarRejectsActiveSVGAndDisguisedContent(t *testing.T) {
	t.Chdir(t.TempDir())
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(document.domain)</script></svg>`)
	for _, name := range []string{"avatar.svg", "avatar.png", "avatar.jpg"} {
		if _, err := saveUploadedAvatarFile(avatarRequest(t, name, svg), 1); err == nil {
			t.Fatal("accepted active/disguised SVG", name)
		}
	}
}
func TestAvatarRejectsOversizedAndExcessiveDimensions(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := saveUploadedAvatarFile(avatarRequest(t, "large.png", bytes.Repeat([]byte{'x'}, (5<<20)+1)), 1); err == nil {
		t.Fatal("accepted over-5MiB avatar")
	}
	if _, err := saveUploadedAvatarFile(avatarRequest(t, "wide.png", avatarPNG(t, 8192, 1)), 1); err == nil {
		t.Fatal("accepted excessive dimensions")
	}
}
func TestAvatarValidRasterIsCanonicalAndOldFileNotFollowed(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	url, err := saveUploadedAvatarFile(avatarRequest(t, "photo.png", avatarPNG(t, 2, 2)), 42)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "/api/avatars/avatar_42.png?") {
		t.Fatal(url)
	}
	data, err := os.ReadFile(filepath.Join("data", "avatars", "avatar_42.png"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(root, "outside")
	if err := os.WriteFile(sentinel, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join("data", "avatars", "avatar_42.png")); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(sentinel, filepath.Join("data", "avatars", "avatar_42.png")); err != nil {
		t.Skip("hardlinks unavailable:", err)
	}
	if _, err := saveUploadedAvatarFile(avatarRequest(t, "photo.png", avatarPNG(t, 3, 3)), 42); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(sentinel); string(got) != "unchanged" {
		t.Fatal("avatar overwrite followed hardlink")
	}
}
func TestLegacyAvatarSVGIsSandboxedAndDownloaded(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Join("data", "avatars"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("data", "avatars", "avatar_7.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), 0600); err != nil {
		t.Fatal(err)
	}
	h := &AccountHandler{}
	w := httptest.NewRecorder()
	h.ServeAvatar(w, httptest.NewRequest("GET", "/api/avatars/avatar_7.svg", nil))
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatal(w.Code, w.Header())
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal(w.Header())
	}
}
