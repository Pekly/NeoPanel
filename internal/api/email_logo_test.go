package api

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
)

func TestEmailLogoFormatsBecomePNG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 16, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 16; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	for _, kind := range []string{"png", "jpeg", "gif", "svg"} {
		t.Run(kind, func(t *testing.T) {
			var input bytes.Buffer
			switch kind {
			case "png":
				png.Encode(&input, src)
			case "jpeg":
				jpeg.Encode(&input, src, nil)
			case "gif":
				gif.Encode(&input, src, nil)
			case "svg":
				input.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 8"><rect width="16" height="8" fill="#ff0000"/></svg>`)
			}
			output, err := emailLogoPNG(input.Bytes(), kind == "svg")
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(output))
			if err != nil || img.Bounds() != image.Rect(0, 0, 128, 128) {
				t.Fatal("bad email PNG", err)
			}
			_, _, _, alpha := img.At(64, 64).RGBA()
			if alpha == 0 {
				t.Fatal("blank logo")
			}
			_, _, _, alpha = img.At(64, 0).RGBA()
			if alpha != 0 {
				t.Fatal("aspect ratio not preserved")
			}
		})
	}
}

func TestEmailLogoRejectsUnsafeAndOversizedSVG(t *testing.T) {
	for _, svg := range []string{
		`<svg viewBox="0 0 10 10"><script>alert(1)</script></svg>`,
		`<svg viewBox="0 0 10 10"><image href="http://127.0.0.1/private"/></svg>`,
		`<svg viewBox="0 0 10 10"><use href="#self" id="self"/></svg>`,
		`<!DOCTYPE svg [<!ENTITY external SYSTEM "file:///secret">]><svg>&external;</svg>`,
		`<svg viewBox="0 0 999999999 1"><rect width="1" height="1"/></svg>`,
		`<svg>` + strings.Repeat(`<g>`, 40) + strings.Repeat(`</g>`, 40) + `</svg>`,
		`<svg viewBox="0 0 10 10">` + strings.Repeat(`<rect width="1" height="1"/>`, 513) + `</svg>`,
		`<svg viewBox="0 0 10 10"></svg>`,
	} {
		if _, err := emailLogoPNG([]byte(svg), true); err == nil {
			t.Fatal("unsafe/blank SVG accepted", svg[:min(70, len(svg))])
		}
	}
	if _, err := emailLogoPNG(bytes.Repeat([]byte("x"), emailLogoMaxBytes+1), false); err == nil {
		t.Fatal("oversize accepted")
	}
	if _, err := emailLogoPNG([]byte("not an image"), false); err == nil {
		t.Fatal("invalid image accepted")
	}
}

func TestEmailLogoPrivateNetworkAndSchemesAreBlocked(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.100.100.200", "0.0.0.0", "224.0.0.1", "198.18.0.1", "::1", "::ffff:127.0.0.1", "fe80::1", "fc00::1", "64:ff9b::7f00:1", "2002:7f00:1::1"} {
		if publicEmailLogoIP(netip.MustParseAddr(address)) {
			t.Fatal("unsafe IP accepted", address)
		}
	}
	for _, address := range []string{"8.8.8.8", "2606:4700:4700::1111"} {
		if !publicEmailLogoIP(netip.MustParseAddr(address)) {
			t.Fatal("public IP rejected", address)
		}
	}
	for _, raw := range []string{"http://example.test/logo.png", "https://user:pass@example.test/logo.png", "file:///secret", "https://127.0.0.1/logo.png", "https://example.test:8443/logo.png"} {
		u, _ := url.Parse(raw)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := fetchEmailLogo(ctx, u)
		cancel()
		if err == nil {
			t.Fatal("unsafe URL accepted", raw)
		}
	}
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1) }))
	defer server.Close()
	u, _ := url.Parse(server.URL + "/private")
	if _, err := fetchEmailLogo(context.Background(), u); err == nil || requests.Load() != 0 {
		t.Fatal("private endpoint contacted")
	}
}

func TestEmailLogoLocalSVGAndMIMEAttachment(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = t.TempDir()
	cfg.Server.PublicURL = "https://panel.example"
	h := &EmailHandler{cfg: cfg}
	dir := filepath.Join(cfg.Paths.DataDir, "branding")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><circle cx="10" cy="10" r="8" fill="#ff0000"/></svg>`
	if err := os.WriteFile(filepath.Join(dir, "logo.svg"), []byte(svg), 0600); err != nil {
		t.Fatal(err)
	}
	link := cfg.Server.PublicURL + "/#invite-token=" + strings.Repeat("a", 64)
	raw, plain, err := invitationEmail(PanelBrandingConfig{BrandName: "Test Brand", LogoURL: "/branding/logo.svg"}, &database.Invitation{Username: "preview-member", Permissions: []string{"console.view", "console.command"}, ExpiresAt: time.Date(2030, 6, 12, 15, 30, 0, 0, time.UTC)}, "Survival server", link)
	if err != nil {
		t.Fatal(err)
	}
	body, logo := h.prepareEmailLogo(raw)
	if len(logo) == 0 || !strings.Contains(body, `src="cid:neopanel-logo"`) || strings.Contains(body, `src="https://`) || strings.Contains(body, "data-neopanel-logo") {
		t.Fatal("logo not embedded", body)
	}
	if !strings.Contains(body, ">"+link+"</a>") || !strings.Contains(plain, link) {
		t.Fatal("full invitation link missing")
	}
	contentType, mimeBody, err := emailMIME(body, plain, logo)
	if err != nil {
		t.Fatal(err)
	}
	wire := "Content-Type: " + contentType + "\r\n\r\n" + string(mimeBody)
	if !strings.Contains(strings.ToLower(wire), "content-id: <neopanel-logo>") || !strings.Contains(wire, "inline; filename=logo.png") {
		t.Fatal("missing inline metadata")
	}
	for _, line := range strings.Split(wire, "\r\n") {
		if len(line) > 998 {
			t.Fatal("SMTP line too long")
		}
	}
	parts := decodedEmailParts(t, wire)
	if parts["text/plain"] != plain || strings.ReplaceAll(parts["text/html"], "\r\n", "\n") != strings.ReplaceAll(body, "\r\n", "\n") || !bytes.Equal([]byte(parts["image/png"]), logo) {
		t.Fatal("MIME body did not round trip")
	}
	if _, err := png.Decode(strings.NewReader(parts["image/png"])); err != nil {
		t.Fatal("invalid attached PNG", err)
	}
	for _, path := range []string{"/branding/missing.png", "/branding/../secret.png", "/branding/%2e%2e%2fsecret.png"} {
		source := renderEmailBrand(emailBrand(PanelBrandingConfig{BrandName: "Test Brand", LogoURL: cfg.Server.PublicURL + path}, cfg.Server.PublicURL))
		fallback, image := h.prepareEmailLogo(source)
		if len(image) != 0 || strings.Contains(fallback, "<img") || !strings.Contains(fallback, "Test Brand") {
			t.Fatal("bad fallback", fallback)
		}
	}
	if preview := os.Getenv("NEOPANEL_EMAIL_LOGO_PREVIEW_DIR"); preview != "" {
		if err := os.WriteFile(filepath.Join(preview, "logo.png"), logo, 0600); err != nil {
			t.Fatal(err)
		}
		previewHTML := strings.Replace(body, "cid:neopanel-logo", "logo.png", 1)
		if err := os.WriteFile(filepath.Join(preview, "invitation.html"), []byte(previewHTML), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(preview, "invitation.eml"), []byte(wire), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
