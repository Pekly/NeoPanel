package api

import (
	"encoding/json"
	"fmt"
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

func invitationBudgetResponse(t *testing.T, w *httptest.ResponseRecorder, scope string, want time.Time) {
	t.Helper()
	invitationWant(t, w, 429)
	var result struct {
		Scope       string    `json:"scope"`
		AvailableAt time.Time `json:"send_available_at"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Scope != scope || !result.AvailableAt.Equal(want) {
		t.Fatalf("budget response: %s, %v", w.Body.String(), err)
	}
	seconds, err := strconv.Atoi(w.Header().Get("Retry-After"))
	if err != nil || seconds < 1 || math.Abs(float64(seconds)-time.Until(want).Seconds()) > 2 {
		t.Fatalf("Retry-After: %q", w.Header().Get("Retry-After"))
	}
	for _, secret := range []string{"token_hash", "actor_id", "inviter_id", "attempted_at", "@example.test"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("private ledger field in response: %s", w.Body.String())
		}
	}
}
func invitationBudgetBody(name string) map[string]any {
	return map[string]any{"server_id": "server", "username": name, "email": name + "@example.test"}
}
func invitationBudgetSession(t *testing.T, h *invitationHarness, name, role string) (*database.User, string) {
	t.Helper()
	u, err := h.db.CreateUserWithDetails(name, name+"@example.test", "hash", role, "")
	if err != nil {
		t.Fatal(err)
	}
	session, err := h.router.auth.GenerateToken(u)
	if err != nil {
		t.Fatal(err)
	}
	return u, session
}

func TestInvitationSendBudgetHTTPRevokeRecreateFailedSMTP(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("smtpFailure=%v", fail), func(t *testing.T) {
			s := newInvitationSMTP(t)
			h := newInvitationHarness(t, s)
			s.fail.Store(fail)
			var first *database.Invitation
			for n := 0; n < 5; n++ {
				body := invitationBudgetBody(fmt.Sprintf("budget%d", n))
				w := h.request("POST", "/api/team/invitations", body, h.session)
				if fail {
					invitationWant(t, w, 502)
				} else {
					invitationWant(t, w, 201)
					s.token(t)
				}
				list, err := h.db.ListInvitations("server", "local")
				if err != nil {
					t.Fatal(err)
				}
				i := list[0]
				if n == 0 {
					first = i
				}
				invitationWant(t, h.request("DELETE", "/api/team/invitations/"+i.ID, nil, h.session), 200)
				if n == 0 {
					// New ID after revoke cannot resend to the same normalized mailbox.
					body["email"] = "  BUDGET0@EXAMPLE.TEST  "
					invitationBudgetResponse(t, h.request("POST", "/api/team/invitations", body, h.session), "recipient", first.ResendAvailableAt())
				}
			}
			invitationBudgetResponse(t, h.request("POST", "/api/team/invitations", invitationBudgetBody("sixth"), h.session), "sender", first.ResendAvailableAt())
			w := h.request("GET", "/api/team/invitations?server_id=server", nil, h.session)
			invitationWant(t, w, 200)
			var result struct {
				Availability database.InvitationSendAvailability `json:"send_availability"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Availability.Scope != "sender" || !result.Availability.AvailableAt.Equal(first.ResendAvailableAt()) {
				t.Fatalf("list availability: %s, %v", w.Body.String(), err)
			}
		})
	}
}

func TestInvitationSendBudgetHTTPActualResenderAcrossServers(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	_, session := invitationBudgetSession(t, h, "resender", "admin")
	var first *database.Invitation
	for n := 0; n < 4; n++ {
		w := h.request("POST", "/api/team/invitations", invitationBudgetBody(fmt.Sprintf("actor%d", n)), session)
		invitationWant(t, w, 201)
		s.token(t)
		if n == 0 {
			var i database.Invitation
			if err := json.Unmarshal(w.Body.Bytes(), &i); err != nil {
				t.Fatal(err)
			}
			first, _ = h.db.GetInvitation(i.ID)
		}
	}
	past := time.Now().UTC().Add(-3 * time.Hour)
	i := &database.Invitation{ID: "oldinvite", ServerID: "server", NodeID: "local", Username: "oldinvite", Email: "oldinvite@example.test", Permissions: []string{}, InviterID: h.owner.ID, TokenHash: "old-hash", Status: "pending", DeliveryStatus: "sent", CreatedAt: past, LastSentAt: past, ExpiresAt: past.Add(72 * time.Hour)}
	if err := h.db.CreateInvitation(i); err != nil {
		t.Fatal(err)
	}
	invitationWant(t, h.request("POST", "/api/team/invitations/oldinvite/resend", nil, session), 200)
	s.token(t)
	if err := h.db.CreateServer(&database.Server{ID: "second", Name: "Second", OwnerID: h.owner.ID}); err != nil {
		t.Fatal(err)
	}
	body := invitationBudgetBody("crossserver")
	body["server_id"] = "second"
	invitationBudgetResponse(t, h.request("POST", "/api/team/invitations", body, session), "sender", first.ResendAvailableAt())
	invitationWant(t, h.request("POST", "/api/team/invitations", body, h.session), 201)
	s.token(t)
}

func TestInvitationSendBudgetHTTPServerAndPermissionGuards(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	_, outsider := invitationBudgetSession(t, h, "outsider", "user")
	body := invitationBudgetBody("forbidden")
	invitationWant(t, h.request("POST", "/api/team/invitations", body, outsider), 403)
	body["permissions"] = []string{"invalid.permission"}
	invitationWant(t, h.request("POST", "/api/team/invitations", body, h.session), 400)
	// Rejected permission/identity/conflict requests leave all five owner slots.
	body = invitationBudgetBody("identity")
	body["email"] = h.owner.Email
	invitationWant(t, h.request("POST", "/api/team/invitations", body, h.session), 409)
	var first *database.Invitation
	for n := 0; n < 5; n++ {
		i, _ := h.create(t, s, fmt.Sprintf("owner%d", n), "local")
		if n == 0 {
			first, _ = h.db.GetInvitation(i.ID)
			invitationWant(t, h.request("POST", "/api/team/invitations", invitationBudgetBody("owner0"), h.session), 409)
		}
	}
	_, second := invitationBudgetSession(t, h, "secondactor", "admin")
	for n := 0; n < 5; n++ {
		invitationWant(t, h.request("POST", "/api/team/invitations", invitationBudgetBody(fmt.Sprintf("other%d", n)), second), 201)
		s.token(t)
	}
	_, third := invitationBudgetSession(t, h, "thirdactor", "admin")
	invitationBudgetResponse(t, h.request("POST", "/api/team/invitations", invitationBudgetBody("serverblocked"), third), "server", first.ResendAvailableAt())
	invitationWant(t, h.request("DELETE", "/api/team/invitations/"+first.ID, nil, third), 200)
	invitationBudgetResponse(t, h.request("POST", "/api/team/invitations", invitationBudgetBody("stillblocked"), third), "server", first.ResendAvailableAt())
}

func TestInvitationSendBudgetHTTPConcurrentCreate(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	for n := 0; n < 4; n++ {
		h.create(t, s, fmt.Sprintf("seed%d", n), "local")
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			w := h.request("POST", "/api/team/invitations", invitationBudgetBody(fmt.Sprintf("parallel%d", n)), h.session)
			if w.Code == 201 {
				successes.Add(1)
			} else if w.Code != 429 || w.Header().Get("Retry-After") == "" {
				t.Errorf("concurrent create: %d %s", w.Code, w.Body.String())
			}
		}(n)
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful creates with one slot left: %d", successes.Load())
	}
	s.token(t)
	select {
	case <-s.messages:
		t.Fatal("blocked request sent mail")
	default:
	}
}
