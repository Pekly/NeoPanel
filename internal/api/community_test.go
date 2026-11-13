package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type communityTransport func(*http.Request) (*http.Response, error)

func (f communityTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCommunityAdminValidationAndPersistence(t *testing.T) {
	f := newReliabilityFixture(t)
	body := `{"discord":{"enabled":false,"guild_id":"","title":"Join us"},"announcement":{"enabled":true,"text":" Maintenance <script>not HTML</script> ","tone":"warning","dismissible":true}}`
	for _, user := range []string{"", "owner", "subuser"} {
		want := 403
		if user == "" {
			want = 401
		}
		for _, method := range []string{"GET", "PUT"} {
			if w := f.request(method, "/api/admin/community", body, user); w.Code != want {
				t.Fatal(user, method, w.Code, w.Body.String())
			}
		}
	}
	if w := f.request("PUT", "/api/admin/community", body, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	h := NewCommunityHandler(f.db)
	settings, err := h.settings()
	if err != nil || settings.Announcement.Text != "Maintenance <script>not HTML</script>" {
		t.Fatal(settings, err)
	}
	public := f.request("GET", "/api/community", "", "")
	if public.Code != 200 {
		t.Fatal(public.Code, public.Body.String())
	}
	var response struct {
		Announcement struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		Discord publicDiscord
	}
	if err := json.Unmarshal(public.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Announcement.ID == "" || response.Discord.Status != "disabled" {
		t.Fatal(response)
	}
	firstID := response.Announcement.ID
	changed := strings.Replace(body, "Maintenance", "Updated maintenance", 1)
	if w := f.request("PUT", "/api/admin/community", changed, "admin"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	public = f.request("GET", "/api/community", "", "")
	json.Unmarshal(public.Body.Bytes(), &response)
	if response.Announcement.ID == firstID {
		t.Fatal("updated message has same dismissal identity")
	}
	for _, bad := range []string{`null`, `{}`, `{"discord":{},"announcement":{},"unknown":true}`, `{"discord":{"enabled":true,"guild_id":"https://evil.example"},"announcement":{}}`, `{"discord":{},"announcement":{"enabled":true}}`, `{"discord":{},"announcement":{"tone":"html"}}`, body + ` {}`} {
		if w := f.request("PUT", "/api/admin/community", bad, "admin"); w.Code != 400 {
			t.Fatal("accepted invalid settings", bad, w.Code, w.Body.String())
		}
	}
}
func TestDiscordWidgetUsesFixedOriginCachesAndSanitizes(t *testing.T) {
	f := newReliabilityFixture(t)
	h := NewCommunityHandler(f.db)
	calls := 0
	h.client.Transport = communityTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != "https://discord.com/api/guilds/123456789012345678/widget.json" || r.Header.Get("Authorization") != "" {
			t.Fatal("unexpected outbound request", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"123456789012345678","name":"Community","instant_invite":"https://discord.com/invite/abc123","presence_count":42,"members":[{"username":"private detail"}]}`))}, nil
	})
	s := DiscordSettings{Enabled: true, GuildID: "123456789012345678", Title: "Join"}
	first := h.discord(s)
	second := h.discord(s)
	if first.Status != "ready" || second.InviteURL != "https://discord.gg/abc123" || first.MemberCount != 42 || calls != 1 {
		t.Fatal(first, second, calls)
	}
	s.GuildID = "../../localhost"
	if result := h.discord(s); result.Status != "unavailable" || calls != 1 {
		t.Fatal("invalid guild reached network", result, calls)
	}
	for _, raw := range []string{"javascript:alert(1)", "https://evil.example/abc", "http://discord.gg/abc", "https://user@discord.gg/abc", "https://discord.gg:443/abc", "https://discord.gg/abc?redirect=evil", "https://discord.com/other/abc"} {
		if got := safeDiscordInvite(raw); got != "" {
			t.Fatal("unsafe link accepted", raw, got)
		}
	}
}
func TestDiscordUnavailableIsSafeAndNegativelyCached(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, payload := range []string{`{"id":"wrong","instant_invite":"https://discord.gg/abc"}`, `{"id":"123456789012345678","instant_invite":"https://evil.example/abc"}`, `{"id":"123456789012345678","instant_invite":null}`, `not json`} {
		h := NewCommunityHandler(f.db)
		calls := 0
		h.client.Transport = communityTransport(func(*http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
		})
		s := DiscordSettings{Enabled: true, GuildID: "123456789012345678"}
		for i := 0; i < 2; i++ {
			if got := h.discord(s); got.Status != "unavailable" || got.InviteURL != "" {
				t.Fatal(got)
			}
		}
		if calls != 1 {
			t.Fatal("failed widget not cached", calls)
		}
	}
	h := NewCommunityHandler(f.db)
	if h.client.Timeout == 0 || h.client.CheckRedirect(httptest.NewRequest("GET", "https://evil.example", nil), nil) != http.ErrUseLastResponse {
		t.Fatal("unbounded/redirecting HTTP client")
	}
}
