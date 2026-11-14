package api

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPanelEmailSharedThemeAndLogoLayout(t *testing.T) {
	cases := []struct {
		name                              string
		config                            PanelBrandingConfig
		background, surface, text, accent string
	}{
		{"dark", PanelBrandingConfig{DefaultTheme: "dark", Preset: "obsidian", AccentColor: "#387aff"}, "#101318", "#181d25", "#e9edf4", "#2f6feb"},
		{"light", PanelBrandingConfig{DefaultTheme: "light", Preset: "obsidian", AccentColor: "#387aff"}, "#f5f6f8", "#ffffff", "#0f172a", "#2563eb"},
		{"crimson", PanelBrandingConfig{DefaultTheme: "dark", BgDeep: "#12080c", BgCanvas: "#1a0c12", CardBg: "#231118", NavBg: "#0e0609", TextColor: "#fff1f2", AccentColor: "#f43f5e", BorderRadiusPx: 12}, "#12080c", "#231118", "#fff1f2", "#f43f5e"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.BrandName, tc.config.LogoURL = "NeoPanel", "/favicon.svg"
			brand := emailBrand(tc.config, "http://127.0.0.1:8097")
			if brand.Background != tc.background || brand.Surface != tc.surface || brand.Text != tc.text || brand.Accent != tc.accent {
				t.Fatalf("palette mismatch: %+v", brand)
			}
			html, plain, err := renderPanelEmail(panelEmail{Branding: brand, Title: "Password reset", Heading: "Reset your password", Greeting: "Hello Alex,", Intro: "We received a request to reset your account password.", ActionText: "Reset Password", ActionURL: "https://panel.example/#reset-token=" + strings.Repeat("a", 64), Code: strings.Repeat("a", 64), Notice: "Valid for 1 hour. This code can be used only once.", Footer: "If you did not request this, you can safely ignore this email."})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{`class="np-email-shell"`, `class="np-email-brand"`, `width="24" height="24"`, "font-size:18px", "vertical-align:middle", tc.background, tc.surface, tc.text, tc.accent} {
				if !strings.Contains(html, want) {
					t.Errorf("missing %q", want)
				}
			}
			if strings.Index(html, "<img ") > strings.Index(html, "font-size:18px") {
				t.Fatal("logo must precede name")
			}
			if strings.Contains(html, "ZgotmplZ") || strings.Contains(plain, "<img") {
				t.Fatal("invalid HTML/plain rendering")
			}
			// Optional local preview artifact; normal test runs remain side-effect free.
			if dir := os.Getenv("NEOPANEL_EMAIL_PREVIEW_DIR"); dir != "" {
				if err := os.WriteFile(filepath.Join(dir, "email-"+tc.name+".html"), []byte(html), 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestEmailThemeRejectsUnsafeCSSAndChoosesReadableButtons(t *testing.T) {
	brand := emailBrand(PanelBrandingConfig{BgDeep: "red;display:none", CardBg: "url(https://evil.test)", TextColor: "expression(x)", NavBg: "javascript:evil", AccentColor: "#ffffff", CustomCSS: "<script>bad</script>", BorderRadiusPx: 999}, "https://panel.example")
	if brand.Background != "#101318" || brand.Surface != "#181d25" || brand.ButtonText != "#000000" || brand.Radius != 14 {
		t.Fatal(brand)
	}
	html, _, err := renderPanelEmail(panelEmail{Branding: brand, Title: "Test", Heading: `<img src=x onerror=alert(1)>`, Intro: "Safe text", Footer: "Footer"})
	if err != nil || strings.Contains(html, "evil.test") || strings.Contains(html, "<script>") || strings.Contains(html, "<img src=x") || !strings.Contains(html, "&lt;img") {
		t.Fatal(err, html)
	}
	if emailButtonText("#000000") != "#ffffff" {
		t.Fatal("dark button needs light text")
	}
}

func TestPasswordResetEmailToNewPassword(t *testing.T) {
	smtp := newInvitationSMTP(t)
	h := newInvitationHarness(t, smtp)
	before, err := h.db.GetUserByID(h.owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.router.email.RequestPasswordReset(w, httptest.NewRequest("POST", "https://spoof.example/", strings.NewReader(`{"identity":"owner"}`)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	parts := brandedEmailParts(t, smtp)
	match := regexp.MustCompile(`#reset-token=([a-f0-9]{64})`).FindStringSubmatch(parts["text/plain"])
	if len(match) != 2 {
		t.Fatal("reset email has no usable code")
	}
	if !strings.Contains(parts["text/html"], `class="np-email-shell"`) {
		t.Fatal("reset does not use shared frame")
	}
	body, _ := json.Marshal(map[string]string{"token": match[1], "new_password": "a-new-safe-password"})
	confirm := func() *httptest.ResponseRecorder {
		out := httptest.NewRecorder()
		h.router.email.ConfirmPasswordReset(out, httptest.NewRequest("POST", "/", strings.NewReader(string(body))))
		return out
	}
	if out := confirm(); out.Code != 200 {
		t.Fatal(out.Code, out.Body.String())
	}
	after, err := h.db.GetUserByID(h.owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(after.PasswordHash), []byte("a-new-safe-password")) != nil || after.AuthVersion <= before.AuthVersion {
		t.Fatal("password/session version did not update")
	}
	if out := confirm(); out.Code != 400 {
		t.Fatal("reset code reused", out.Code)
	}
}

func TestPasswordResetRejectsBcryptOversizedPassword(t *testing.T) {
	h := newInvitationHarness(t, nil)
	body, _ := json.Marshal(map[string]string{"token": strings.Repeat("a", 64), "new_password": strings.Repeat("é", 37)})
	w := httptest.NewRecorder()
	h.router.email.ConfirmPasswordReset(w, httptest.NewRequest("POST", "/", strings.NewReader(string(body))))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "72 UTF-8 bytes") {
		t.Fatal(w.Code, w.Body.String())
	}
}
