package api

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"
)

func TestEmailLogoURL(t *testing.T) {
	for _, tc := range []struct{ logo, base, want string }{
		{"/branding/logo.png", "https://panel.example:8443/#reset-token=secret", "https://panel.example:8443/branding/logo.png"},
		{"assets/logo.png?v=2", "https://panel.example/path?token=secret#invite-token=secret", "https://panel.example/assets/logo.png?v=2"},
		{"https://cdn.example/logo.png?size=64", "", "https://cdn.example/logo.png?size=64"},
		{"//cdn.example/logo.png", "https://panel.example", "https://cdn.example/logo.png"},
		{"http://cdn.example/logo.png", "", "http://cdn.example/logo.png"},
		{"/branding/logo.png", "", ""},
		{"/branding/logo.png", "https://user:pass@panel.example", ""},
		{"", "https://panel.example", ""},
		{"#reset-token=secret", "https://panel.example", ""},
		{"?secret=value", "https://panel.example", ""},
		{"javascript:alert(1)", "https://panel.example", ""},
		{"data:image/svg+xml,<svg onload=alert(1)>", "https://panel.example", ""},
		{"file:///etc/passwd", "https://panel.example", ""},
		{"https://user:pass@cdn.example/logo.png", "", ""},
		{"https://cdn.example\\evil/logo.png", "", ""},
		{"https://cdn.example/\nlogo.png", "", ""},
		{"https:///missing-host.png", "", ""},
	} {
		if got := emailLogoURL(tc.logo, tc.base); got != tc.want {
			t.Errorf("logo=%q base=%q: got %q want %q", tc.logo, tc.base, got, tc.want)
		}
	}
}

func TestEmailBrandHeaderEscapingAndFallback(t *testing.T) {
	brand := PanelBrandingConfig{BrandName: `Acme "<img src=x onerror=alert(1)>`, LogoURL: `https://cdn.example/logo.png?q=" onerror="alert(1)`, AccentColor: "red;display:none"}
	output := renderEmailBrand(emailBrand(brand, "https://panel.example"))
	if strings.Count(output, "<img ") != 1 || strings.Contains(output, `onerror="`) || strings.Contains(output, "<img src=x") || strings.Contains(output, "ZgotmplZ") || !strings.Contains(output, "&lt;img") || !strings.Contains(output, `width="24" height="24"`) || !strings.Contains(output, "font-size:18px") {
		t.Fatal("unsafe or missing logo markup", output)
	}
	for _, logo := range []string{"", "javascript:alert(1)", "data:image/png;base64,bad"} {
		output := renderEmailBrand(emailBrand(PanelBrandingConfig{LogoURL: logo}, "https://panel.example"))
		if strings.Contains(output, "<img") || !strings.Contains(output, "NeoPanel") {
			t.Fatal("fallback must show brand without a broken image", output)
		}
	}
}

func brandedEmailParts(t *testing.T, smtp *invitationSMTP) map[string]string {
	t.Helper()
	var raw string
	select {
	case raw = <-smtp.messages:
	case <-time.After(3 * time.Second):
		t.Fatal("no email received")
	}
	message, err := mail.ReadMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	kind, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || kind != "multipart/alternative" {
		t.Fatal(kind, err)
	}
	reader := multipart.NewReader(message.Body, params["boundary"])
	parts := map[string]string{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		kind, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		parts[kind] = string(body)
	}
	return parts
}

func TestPanelLogoInEveryEmailType(t *testing.T) {
	smtp := newInvitationSMTP(t)
	h := newInvitationHarness(t, smtp)
	branding, _ := json.Marshal(PanelBrandingConfig{BrandName: "Acme & Cloud", LogoURL: "/branding/logo.png", AccentColor: "#123456"})
	if err := h.db.SetPanelSetting(panelSettingBrandingKey, string(branding)); err != nil {
		t.Fatal(err)
	}
	if err := h.db.SetPanelPublicURL("https://mail.example:8443", "mail.example"); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"invitation", "reset", "test"} {
		t.Run(kind, func(t *testing.T) {
			switch kind {
			case "invitation":
				w := h.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "username": "logo-guest", "email": "guest@example.test", "permissions": []string{"files.view"}}, h.session)
				invitationWant(t, w, 201)
			default:
				body := `{"identity":"owner"}`
				if kind == "test" {
					body = `{"to_email":"admin@example.test"}`
				}
				req := httptest.NewRequest("POST", "http://untrusted.example/", strings.NewReader(body))
				req.Header.Set("X-Forwarded-Host", "spoof.example")
				w := httptest.NewRecorder()
				if kind == "test" {
					h.router.email.HandleTestEmail(w, req)
				} else {
					h.router.email.RequestPasswordReset(w, req)
				}
				if w.Code != 200 {
					t.Fatal(w.Code, w.Body.String())
				}
			}
			parts := brandedEmailParts(t, smtp)
			html := parts["text/html"]
			if len(parts) != 2 || parts["text/plain"] == "" || strings.Contains(parts["text/plain"], "<img") {
				t.Fatal("plain alternative changed", parts)
			}
			if strings.Count(html, "<img ") != 1 || !strings.Contains(html, `src="https://mail.example:8443/branding/logo.png"`) || !strings.Contains(html, `alt="Acme &amp; Cloud logo"`) || !strings.Contains(html, "#123456") {
				t.Fatal("missing branded logo", html)
			}
			if strings.Contains(html, "untrusted.example") || strings.Contains(html, "spoof.example") || strings.Contains(html, "ZgotmplZ") {
				t.Fatal("untrusted email URL", html)
			}
		})
	}
	// A subsequent delivery reflects saved branding changes without restarting.
	branding, _ = json.Marshal(PanelBrandingConfig{BrandName: "Changed", LogoURL: "https://cdn.example/new-logo.png"})
	if err := h.db.SetPanelSetting(panelSettingBrandingKey, string(branding)); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.router.email.HandleTestEmail(w, httptest.NewRequest("POST", "/", strings.NewReader(`{"to_email":"admin@example.test"}`)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	parts := brandedEmailParts(t, smtp)
	if !strings.Contains(parts["text/html"], `src="https://cdn.example/new-logo.png"`) || strings.Contains(parts["text/html"], "Acme") {
		t.Fatal("stale branding", parts["text/html"])
	}
}

func TestSMTPTestLogoWithoutPublicURLFallsBackSafely(t *testing.T) {
	smtp := newInvitationSMTP(t)
	h := newInvitationHarness(t, smtp)
	if err := h.db.SetPanelSetting("panel_public_url", ""); err != nil {
		t.Fatal(err)
	}
	branding, _ := json.Marshal(PanelBrandingConfig{BrandName: "My panel", LogoURL: "/branding/logo.png"})
	if err := h.db.SetPanelSetting(panelSettingBrandingKey, string(branding)); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.router.email.HandleTestEmail(w, httptest.NewRequest("POST", "https://untrusted.example/", strings.NewReader(`{"to_email":"admin@example.test"}`)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	parts := brandedEmailParts(t, smtp)
	if strings.Contains(parts["text/html"], "<img") || !strings.Contains(parts["text/html"], "My panel") {
		t.Fatal("unsafe missing-origin fallback", parts["text/html"])
	}
}
