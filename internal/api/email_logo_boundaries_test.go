package api

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

type emailLogoRoundTrip func(*http.Request) (*http.Response, error)

func (f emailLogoRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestEmailLogoRedirectBoundaries(t *testing.T) {
	for _, target := range []string{"https://127.0.0.1/private", "https://[::ffff:127.0.0.1]/private", "https://user:pass@cdn.example/logo.png", "https://cdn.example:8443/logo.png", "http://cdn.example/logo.png", "file:///secret"} {
		calls := 0
		client := emailLogoHTTPClient(emailLogoRoundTrip(func(r *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {target}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		}))
		response, err := client.Get("https://public.example/logo.png")
		if response != nil {
			response.Body.Close()
		}
		if err == nil || calls != 1 {
			t.Fatalf("unsafe redirect followed: %s, calls=%d", target, calls)
		}
	}
	calls := 0
	client := emailLogoHTTPClient(emailLogoRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://cdn.example/loop"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	}))
	if response, err := client.Get("https://public.example/start"); err == nil {
		response.Body.Close()
		t.Fatal("redirect loop allowed")
	}
	if calls != 4 {
		t.Fatal("redirect budget changed", calls)
	}
	calls = 0
	client = emailLogoHTTPClient(emailLogoRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		response := &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("image")), Request: r}
		if calls == 1 {
			response.StatusCode = 302
			response.Header.Set("Location", "https://cdn.example/logo.png")
		}
		return response, nil
	}))
	response, err := client.Get("http://public.example/start")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if calls != 2 {
		t.Fatal("safe upgrade redirect failed")
	}
}

func TestEmailICOAllocationBoundaries(t *testing.T) {
	pngData := sidebarTestPNG(t, color.RGBA{R: 255, A: 255})
	// Valid PNG header/CRC declaring a huge raster in a tiny ICO entry.
	large := append([]byte{}, pngData...)
	binary.BigEndian.PutUint32(large[16:20], 8192)
	binary.BigEndian.PutUint32(large[20:24], 8192)
	binary.BigEndian.PutUint32(large[29:33], crc32.ChecksumIEEE(large[12:29]))
	if _, err := png.DecodeConfig(bytes.NewReader(large)); err != nil {
		t.Fatal("bad oversized fixture", err)
	}
	if _, err := emailICOToPNG(sidebarTestICO(large)); err == nil {
		t.Fatal("oversized ICO PNG admitted")
	}
	excessive := []byte{0, 0, 1, 0, 65, 0}
	if _, err := emailICOToPNG(excessive); err == nil {
		t.Fatal("too many ICO entries admitted")
	}
	for _, tc := range []struct{ width, bits, size int }{{8192, 32, 40}, {4, 32, 44}, {4, 24, 88}} {
		dib := make([]byte, tc.size)
		binary.LittleEndian.PutUint32(dib[:4], 40)
		binary.LittleEndian.PutUint32(dib[4:8], uint32(tc.width))
		binary.LittleEndian.PutUint32(dib[8:12], 8)
		binary.LittleEndian.PutUint16(dib[12:14], 1)
		binary.LittleEndian.PutUint16(dib[14:16], uint16(tc.bits))
		if _, err := emailICOToPNG(sidebarTestICO(dib)); err == nil {
			t.Fatalf("unsafe/truncated ICO bitmap admitted: %+v", tc)
		}
	}
}

func TestEmailLogoMatchesStaticRouteAlias(t *testing.T) {
	fixture := newReliabilityFixture(t)
	assets := fstest.MapFS{"favicon.svg": {Data: []byte(`<svg viewBox="0 0 4 4"><rect width="4" height="4" fill="#ff0000"/></svg>`)}}
	router := NewRouter(fixture.cfg, fixture.db, fixture.manager, assets)
	for _, source := range []string{"/favicon.svg", "/static/favicon.svg"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest("GET", source, nil))
		if response.Code != 200 {
			t.Fatal("public alias failed", source, response.Code)
		}
		body, logo, warning := router.email.prepareSidebarEmailLogo(renderEmailBrand(emailBrand(PanelBrandingConfig{LogoURL: source}, "")))
		if warning || len(logo) == 0 || !strings.Contains(body, "cid:neopanel-logo") {
			t.Fatal("email differs from sidebar alias", source)
		}
		image, err := png.Decode(bytes.NewReader(logo))
		if err != nil {
			t.Fatal(err)
		}
		r, _, b, _ := image.At(64, 64).RGBA()
		if r < 60000 || b > 1000 {
			t.Fatal("fallback substituted for static logo")
		}
	}
}
