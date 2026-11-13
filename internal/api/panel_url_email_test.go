package api

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestPanelURLPasswordResetEmail(t *testing.T) {
	smtp := newInvitationSMTP(t)
	h := newInvitationHarness(t, smtp)
	if err := h.db.SetPanelPublicURL("https://reset.example:8443", "reset.example"); err != nil {
		t.Fatal(err)
	}
	request := func(identity string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "http://untrusted.example/", strings.NewReader(`{"identity":"`+identity+`"}`))
		r.Header.Set("X-Forwarded-Proto", "javascript")
		r.Header.Set("X-Forwarded-Host", "spoof.example")
		w := httptest.NewRecorder()
		h.router.email.RequestPasswordReset(w, r)
		return w
	}
	if w := request("owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case raw := <-smtp.messages:
		message, err := mail.ReadMessage(strings.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		_, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		parts := multipart.NewReader(message.Body, params["boundary"])
		count := 0
		for {
			part, err := parts.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(part)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), "https://reset.example:8443/#reset-token=") || strings.Contains(string(body), "untrusted.example") || strings.Contains(string(body), "spoof.example") {
				t.Fatal("reset link did not use saved origin", string(body))
			}
			count++
		}
		if count != 2 {
			t.Fatal("expected HTML and plain-text email", count)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("reset email not received")
	}
	if err := h.db.SetPanelSetting(database.PanelPublicURLKey, "https://invalid.example/path"); err != nil {
		t.Fatal(err)
	}
	existing, missing := request("owner"), request("unknown")
	if existing.Code != 200 || missing.Code != 200 || existing.Body.String() != missing.Body.String() {
		t.Fatal("configuration error exposed account existence", existing.Body.String(), missing.Body.String())
	}
	select {
	case <-smtp.messages:
		t.Fatal("email sent with invalid configured origin")
	default:
	}
}
