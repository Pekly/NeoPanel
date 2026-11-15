package api

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Pekly/NeoPanel/internal/config"
)

func sidebarTestPNG(t *testing.T, fill color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func sidebarTestICO(payload []byte) []byte {
	data := make([]byte, 22)
	copy(data, []byte{0, 0, 1, 0, 1, 0, 4, 4})
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint16(data[12:14], 32)
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
	binary.LittleEndian.PutUint32(data[18:22], 22)
	return append(data, payload...)
}

func TestEmailUsesSavedSidebarLogoNotFavicon(t *testing.T) {
	red := sidebarTestPNG(t, color.RGBA{R: 255, A: 255})
	blue := sidebarTestPNG(t, color.RGBA{B: 255, A: 255})
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><style>.logo { fill: #ff0000; }</style><rect class="logo" width="4" height="4"/></svg>`
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = t.TempDir()
	cfg.Server.PublicURL = "https://configured.example"
	writeEmailLogoFixture(t, cfg.Paths.DataDir, "sidebar.png", color.RGBA{R: 255, A: 255})
	h := &EmailHandler{cfg: cfg, assets: fstest.MapFS{"sidebar.svg": {Data: []byte(svg)}, "sidebar.ico": {Data: sidebarTestICO(red)}, "favicon.png": {Data: blue}}}
	for _, source := range []string{"/branding/sidebar.png", "branding/sidebar.png", "https://configured.example/branding/sidebar.png?v=2", "/sidebar.svg", "/sidebar.ico", "data:image/png;base64," + base64.StdEncoding.EncodeToString(red), "data:image/svg+xml," + url.PathEscape(svg)} {
		brand := emailBrand(PanelBrandingConfig{BrandName: "Example", LogoURL: source, FaviconURL: "/favicon.png"}, "https://different-email-origin.example/#invite-token=secret")
		if brand.LogoSource != source {
			t.Fatal("sidebar source was changed")
		}
		body, logo, warning := h.prepareSidebarEmailLogo(renderEmailBrand(brand))
		if warning || len(logo) == 0 {
			t.Fatalf("sidebar logo not used: %s", source[:min(len(source), 60)])
		}
		image, err := png.Decode(bytes.NewReader(logo))
		if err != nil {
			t.Fatal(err)
		}
		r, _, b, a := image.At(64, 64).RGBA()
		if r < 60000 || b > 1000 || a == 0 {
			t.Fatal("favicon or fallback used instead of red sidebar logo")
		}
		if !strings.Contains(body, `src="cid:neopanel-logo"`) || strings.Contains(body, "data-logo-") || strings.Contains(body, "data:image/") {
			t.Fatal("logo metadata leaked", body)
		}
	}
	// Uploaded branding permits 12 MiB; a valid larger source is still normalized.
	large := append(append([]byte{}, red...), make([]byte, (2<<20)+1)...)
	if err := os.WriteFile(filepath.Join(cfg.Paths.DataDir, "branding", "large.png"), large, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := h.loadEmailLogo("/branding/large.png"); err != nil {
		t.Fatal("valid upload rejected by remote-image limit", err)
	}
}

func TestEmailSidebarCubeAndFallbackDiagnostic(t *testing.T) {
	h := &EmailHandler{}
	for _, source := range []string{"", "javascript:alert(1)", "data:image/png;base64,bad"} {
		body, logo, warning := h.prepareSidebarEmailLogo(renderEmailBrand(emailBrand(PanelBrandingConfig{LogoURL: source, AccentColor: "#ff0000"}, "")))
		if len(logo) == 0 || warning != (source != "") || !strings.Contains(body, `src="cid:neopanel-logo"`) {
			t.Fatal("wrong sidebar fallback", warning, body)
		}
		image, err := png.Decode(bytes.NewReader(logo))
		if err != nil {
			t.Fatal(err)
		}
		r, _, b, _ := image.At(64, 64).RGBA()
		if r < 60000 || b > 1000 {
			t.Fatal("default cube ignored sidebar accent")
		}
	}
	smtp := newInvitationSMTP(t)
	fixture := newInvitationHarness(t, smtp)
	branding, _ := json.Marshal(PanelBrandingConfig{LogoURL: "/branding/missing.png"})
	if err := fixture.db.SetPanelSetting(panelSettingBrandingKey, string(branding)); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	fixture.router.email.HandleTestEmail(response, httptest.NewRequest("POST", "/", strings.NewReader(`{"to_email":"test@example.test"}`)))
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result["logo_warning"], "Sidebar logo") || !strings.Contains(result["logo_warning"], "default cube") {
		t.Fatal("silent logo failure", result)
	}
	parts := brandedEmailParts(t, smtp)
	if parts["image/png"] == "" {
		t.Fatal("fallback not attached")
	}
}

func TestEmailICOBitmapAndInvalidOffsets(t *testing.T) {
	// Four-by-four bottom-up BGRA DIB with its AND mask.
	dib := make([]byte, 40+4*4*4+4*4)
	binary.LittleEndian.PutUint32(dib[:4], 40)
	binary.LittleEndian.PutUint32(dib[4:8], 4)
	binary.LittleEndian.PutUint32(dib[8:12], 8)
	binary.LittleEndian.PutUint16(dib[12:14], 1)
	binary.LittleEndian.PutUint16(dib[14:16], 32)
	for i := 40; i < 104; i += 4 {
		dib[i+2] = 255
		dib[i+3] = 255
	}
	data := sidebarTestICO(dib)
	output, err := emailLogoPNG(data, false)
	if err != nil {
		t.Fatal(err)
	}
	image, err := png.Decode(bytes.NewReader(output))
	if err != nil {
		t.Fatal(err)
	}
	r, _, _, _ := image.At(64, 64).RGBA()
	if r < 60000 {
		t.Fatal("ICO colors changed")
	}
	for _, offset := range []uint32{0, 21, 0xffffffff} {
		broken := append([]byte{}, data...)
		binary.LittleEndian.PutUint32(broken[18:22], offset)
		if _, err := emailICOToPNG(broken); err == nil {
			t.Fatal("invalid ICO offset accepted", offset)
		}
	}
	for i := 0; i < 22; i++ {
		if _, err := emailICOToPNG(data[:i]); err == nil {
			t.Fatal("truncated ICO accepted")
		}
	}
}

func TestEmailSVGStyleBudgets(t *testing.T) {
	for _, svg := range []string{
		`<svg viewBox="0 0 4 4"><style>.a{stroke-width:1e100}</style><rect class="a" width="4" height="4"/></svg>`,
		`<svg viewBox="0 0 4 4"><style>@import url(https://example.test/x)</style></svg>`,
		`<svg viewBox="0 0 4 4"><rect style="stroke-width:Infinity" width="4" height="4"/></svg>`,
		`<svg viewBox="0 0 4 4">` + strings.Repeat(`<style>.a,.b,.c,.d,.e,.f,.g,.h{fill:red;stroke:red;opacity:1;stroke-width:1;stroke-opacity:1;fill-opacity:1;stroke-linecap:round;stroke-linejoin:round;}</style>`, 64) + `</svg>`,
	} {
		if safeEmailSVG([]byte(svg)) {
			t.Fatal("unbounded SVG styles accepted")
		}
	}
}
