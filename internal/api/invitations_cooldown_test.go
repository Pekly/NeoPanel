package api

import (
	"encoding/json"
	"math"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func invitationCooldownTimestamp(t *testing.T, raw []byte, want time.Time) {
	t.Helper()
	var response struct {
		AvailableAt time.Time `json:"resend_available_at"`
	}
	if err := json.Unmarshal(raw, &response); err != nil || !response.AvailableAt.Equal(want) {
		t.Fatalf("cooldown timestamp: %s, want %v, error %v", raw, want, err)
	}
	for _, secret := range []string{"token_hash", "last_sent_at", "inviter_id", "account_id"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("private field %q in response", secret)
		}
	}
}

func TestInvitationResendCooldownResponsesAndAuthorization(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	w := h.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "username": "cooldown", "email": "cooldown@example.test"}, h.session)
	invitationWant(t, w, 201)
	token := s.token(t)
	var created database.Invitation
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	i, err := h.db.GetInvitation(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	invitationCooldownTimestamp(t, w.Body.Bytes(), i.ResendAvailableAt())
	w = h.request("GET", "/api/team/invitations?server_id=server", nil, h.session)
	invitationWant(t, w, 200)
	var list struct {
		Invitations []json.RawMessage `json:"invitations"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list.Invitations) != 1 {
		t.Fatalf("list: %s, %v", w.Body.String(), err)
	}
	invitationCooldownTimestamp(t, list.Invitations[0], i.ResendAvailableAt())
	url := "/api/team/invitations/" + i.ID + "/resend"
	invitationWant(t, h.request("POST", url, nil, ""), 401)
	user, err := h.db.CreateUserWithDetails("outsider", "outsider@example.test", "hash", "user", "")
	if err != nil {
		t.Fatal(err)
	}
	session, err := h.router.auth.GenerateToken(user)
	if err != nil {
		t.Fatal(err)
	}
	invitationWant(t, h.request("POST", url, nil, session), 403)
	before := time.Now()
	w = h.request("POST", url, nil, h.session)
	invitationWant(t, w, 429)
	after := time.Now()
	seconds, err := strconv.Atoi(w.Header().Get("Retry-After"))
	if err != nil || seconds < max(1, int(math.Ceil(i.ResendAvailableAt().Sub(after).Seconds()))) || seconds > int(math.Ceil(i.ResendAvailableAt().Sub(before).Seconds())) {
		t.Fatalf("inaccurate Retry-After: %q, %v", w.Header().Get("Retry-After"), err)
	}
	invitationCooldownTimestamp(t, w.Body.Bytes(), i.ResendAvailableAt())
	if !strings.Contains(w.Body.String(), "2 hours") {
		t.Fatal(w.Body.String())
	}
	current, err := h.db.GetInvitation(i.ID)
	if err != nil || current.TokenHash != invitationHash(token) || !current.LastSentAt.Equal(i.LastSentAt) {
		t.Fatalf("blocked request rotated token: %+v, %v", current, err)
	}
	select {
	case <-s.messages:
		t.Fatal("blocked resend sent email")
	default:
	}
}

func TestInvitationResendCooldownRetryAfterRounding(t *testing.T) {
	for _, tc := range []struct {
		remaining time.Duration
		want      string
	}{
		{1500 * time.Millisecond, "2"},
		{500 * time.Millisecond, "1"},
		{-time.Second, "1"},
	} {
		w := httptest.NewRecorder()
		invitationError(w, &database.InvitationCooldownError{AvailableAt: time.Now().Add(tc.remaining)})
		invitationWant(t, w, 429)
		if w.Header().Get("Retry-After") != tc.want {
			t.Fatalf("Retry-After: %q, want %q", w.Header().Get("Retry-After"), tc.want)
		}
	}
}

func TestInvitationResendCooldownConcurrentHTTP(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	past := time.Now().UTC().Add(-database.InvitationResendCooldown - time.Minute)
	i := &database.Invitation{ID: "parallel", ServerID: "server", NodeID: "local", Username: "parallel", Email: "parallel@example.test", Permissions: []string{}, InviterID: h.owner.ID, TokenHash: "old-hash", Status: "pending", DeliveryStatus: "sent", CreatedAt: past, LastSentAt: past, ExpiresAt: past.Add(72 * time.Hour)}
	if err := h.db.CreateInvitation(i); err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 6; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := h.request("POST", "/api/team/invitations/parallel/resend", nil, h.session)
			if w.Code == 200 {
				successes.Add(1)
			} else if w.Code != 429 || w.Header().Get("Retry-After") == "" {
				t.Errorf("resend: %d %s", w.Code, w.Body.String())
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful resends: %d", successes.Load())
	}
	token := s.token(t)
	current, err := h.db.GetInvitation(i.ID)
	if err != nil || current.TokenHash != invitationHash(token) {
		t.Fatalf("winning token: %+v, %v", current, err)
	}
	select {
	case <-s.messages:
		t.Fatal("multiple resend emails")
	default:
	}
}

func TestInvitationResendCooldownAfterDeliveryFailure(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	s.fail.Store(true)
	w := h.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "username": "failedcooldown", "email": "failedcooldown@example.test"}, h.session)
	invitationWant(t, w, 502)
	var response struct {
		Invitation json.RawMessage `json:"invitation"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	list, err := h.db.ListInvitations("server", "local")
	if err != nil || len(list) != 1 {
		t.Fatalf("failed delivery list: %v, %v", list, err)
	}
	i := list[0]
	invitationCooldownTimestamp(t, response.Invitation, i.ResendAvailableAt())
	if i.DeliveryStatus != "failed" {
		t.Fatal(i.DeliveryStatus)
	}
	s.fail.Store(false)
	w = h.request("POST", "/api/team/invitations/"+i.ID+"/resend", nil, h.session)
	invitationWant(t, w, 429)
	current, err := h.db.GetInvitation(i.ID)
	if err != nil || current.TokenHash != i.TokenHash || !current.LastSentAt.Equal(i.LastSentAt) {
		t.Fatalf("blocked failed-delivery resend mutated invitation: %+v, %v", current, err)
	}
}
