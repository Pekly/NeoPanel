package database

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func budgetInvitation(owner *User, id string, at time.Time) *Invitation {
	return &Invitation{ID: id, ServerID: "server", NodeID: "local", Username: id, Email: id + "@example.test", Permissions: []string{"console.view"}, InviterID: owner.ID, TokenHash: "hash-" + id, Status: "pending", DeliveryStatus: "sending", CreatedAt: at, LastSentAt: at, ExpiresAt: at.Add(72 * time.Hour)}
}
func budgetCount(t *testing.T, db *DB) int {
	t.Helper()
	var count int
	if err := db.conn.QueryRow(`SELECT count(*) FROM invitation_send_attempts`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
func wantBudget(t *testing.T, err error, scope string, at time.Time) {
	t.Helper()
	var limit *InvitationSendBudgetError
	if !errors.As(err, &limit) || limit.Scope != scope || !limit.AvailableAt.Equal(at) {
		t.Fatalf("budget error = %v, want %s at %v", err, scope, at)
	}
}
func budgetAdmin(t *testing.T, db *DB, name string) *User {
	t.Helper()
	u, err := db.CreateUserWithDetails(name, name+"@example.test", "hash", "admin", "")
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestInvitationSendBudgetRevokeRecreateAndWindow(t *testing.T) {
	db, owner := invitationDB(t)
	now := time.Now().UTC()
	for n := 0; n < InvitationSenderLimit; n++ {
		i := budgetInvitation(owner, fmt.Sprintf("recreate%d", n), now.Add(time.Duration(n)*time.Minute))
		if err := db.CreateInvitationForSend(i); err != nil {
			t.Fatal(err)
		}
		if err := db.RevokeInvitation(i.ID, nil); err != nil {
			t.Fatal(err)
		}
	}
	i := budgetInvitation(owner, "blocked", now.Add(10*time.Minute))
	wantBudget(t, db.CreateInvitationForSend(i), "sender", now.Add(InvitationSendWindow))
	if _, err := db.GetInvitation(i.ID); err == nil {
		t.Fatal("blocked create persisted")
	}
	if budgetCount(t, db) != 5 {
		t.Fatal("blocked request consumed quota")
	}
	i.LastSentAt = now.Add(InvitationSendWindow)
	if err := db.CreateInvitationForSend(i); err != nil {
		t.Fatal("window boundary:", err)
	}
	if budgetCount(t, db) != 5 {
		t.Fatal("expired attempt was not pruned")
	}
}

func TestInvitationSendBudgetRecipientPanelwide(t *testing.T) {
	db, owner := invitationDB(t)
	now := time.Now().UTC()
	i := budgetInvitation(owner, "mailbox", now)
	if err := db.CreateInvitationForSend(i); err != nil {
		t.Fatal(err)
	}
	if err := db.RevokeInvitation(i.ID, nil); err != nil {
		t.Fatal(err)
	}
	other := budgetInvitation(budgetAdmin(t, db, "otheractor"), "otherinvite", now.Add(time.Minute))
	other.NodeID, other.ServerID, other.Email = "othernode", "otherserver", "  MAILBOX@EXAMPLE.TEST  "
	wantBudget(t, db.CreateInvitationForSend(other), "recipient", now.Add(InvitationSendWindow))
	if budgetCount(t, db) != 1 {
		t.Fatal("recipient rejection consumed quota")
	}
}

func TestInvitationSendBudgetServerAcrossAccountsAndSenderAcrossNodes(t *testing.T) {
	t.Run("server", func(t *testing.T) {
		db, _ := invitationDB(t)
		now := time.Now().UTC()
		for account := 0; account < 2; account++ {
			actor := budgetAdmin(t, db, fmt.Sprintf("actor%d", account))
			for n := 0; n < 5; n++ {
				if err := db.CreateInvitationForSend(budgetInvitation(actor, fmt.Sprintf("server%d_%d", account, n), now)); err != nil {
					t.Fatal(err)
				}
			}
		}
		third := budgetAdmin(t, db, "thirdactor")
		i := budgetInvitation(third, "serverblocked", now)
		wantBudget(t, db.CreateInvitationForSend(i), "server", now.Add(InvitationSendWindow))
		availability, err := db.InvitationSendAvailability(third.ID, "local", "server", now)
		if err != nil || availability.Scope != "server" || !availability.AvailableAt.Equal(now.Add(InvitationSendWindow)) {
			t.Fatalf("availability: %+v, %v", availability, err)
		}
		// A different node with the same server ID has a separate server budget.
		i.NodeID = "anothernode"
		if err = db.CreateInvitationForSend(i); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("sender", func(t *testing.T) {
		db, _ := invitationDB(t)
		actor := budgetAdmin(t, db, "sender")
		now := time.Now().UTC()
		for n := 0; n < 6; n++ {
			i := budgetInvitation(actor, fmt.Sprintf("cross%d", n), now)
			i.NodeID, i.ServerID = fmt.Sprintf("node%d", n), fmt.Sprintf("server%d", n)
			err := db.CreateInvitationForSend(i)
			if n < 5 && err != nil {
				t.Fatal(err)
			}
			if n == 5 {
				wantBudget(t, err, "sender", now.Add(InvitationSendWindow))
			}
		}
	})
}

func TestInvitationSendBudgetActualResenderAndRollback(t *testing.T) {
	db, owner := invitationDB(t)
	actor := budgetAdmin(t, db, "resender")
	now := time.Now().UTC()
	i := budgetInvitation(owner, "resendtarget", now.Add(-3*time.Hour))
	if err := db.CreateInvitation(i); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 4; n++ {
		if err := db.CreateInvitationForSend(budgetInvitation(actor, fmt.Sprintf("resender%d", n), now)); err != nil {
			t.Fatal(err)
		}
	}
	rotated, err := db.RotateInvitationForSend(i.ID, "rotated", now, actor.ID)
	if err != nil || rotated.InviterID != owner.ID {
		t.Fatalf("rotate: %+v %v", rotated, err)
	}
	var charged int
	if err := db.conn.QueryRow(`SELECT actor_id FROM invitation_send_attempts WHERE invitation_id=?`, i.ID).Scan(&charged); err != nil || charged != actor.ID {
		t.Fatalf("charged %d, %v", charged, err)
	}
	wantBudget(t, db.CreateInvitationForSend(budgetInvitation(actor, "resenderblocked", now)), "sender", now.Add(InvitationSendWindow))
	// The original inviter still has their full sender budget.
	if err := db.CreateInvitationForSend(budgetInvitation(owner, "ownerstillfree", now)); err != nil {
		t.Fatal(err)
	}
	blocked := budgetInvitation(owner, "blockedrotation", now.Add(-3*time.Hour))
	if err := db.CreateInvitation(blocked); err != nil {
		t.Fatal(err)
	}
	_, err = db.RotateInvitationForSend(blocked.ID, "must-not-rotate", now, actor.ID)
	wantBudget(t, err, "sender", now.Add(InvitationSendWindow))
	unchanged, err := db.GetInvitation(blocked.ID)
	if err != nil || unchanged.TokenHash != blocked.TokenHash || !unchanged.LastSentAt.Equal(blocked.LastSentAt) {
		t.Fatal("blocked rotation changed record", err)
	}
}

func TestInvitationSendBudgetInvalidAndConflictDoNotConsume(t *testing.T) {
	db, owner := invitationDB(t)
	now := time.Now().UTC()
	valid := budgetInvitation(owner, "valid", now)
	if err := db.CreateInvitationForSend(valid); err != nil {
		t.Fatal(err)
	}
	conflict := budgetInvitation(owner, "conflicting", now)
	conflict.Email = valid.Email
	if err := db.CreateInvitationForSend(conflict); !errors.Is(err, ErrInvitationConflict) {
		t.Fatal(err)
	}
	// Failure after reservation (unique token constraint) rolls back the ledger.
	conflict.Email, conflict.TokenHash = "distinct@example.test", valid.TokenHash
	if err := db.CreateInvitationForSend(conflict); !errors.Is(err, ErrInvitationConflict) {
		t.Fatal(err)
	}
	identity := budgetInvitation(owner, "identity", now)
	identity.Email = owner.Email
	if err := db.CreateInvitationForSend(identity); !errors.Is(err, ErrInvitationConflict) {
		t.Fatal(err)
	}
	outsider, err := db.CreateUserWithDetails("outsider", "outsider@example.test", "hash", "user", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateInvitationForSend(budgetInvitation(outsider, "forbidden", now)); !errors.Is(err, ErrInvitationForbidden) {
		t.Fatal(err)
	}
	if _, err := db.RotateInvitationForSend(valid.ID, "bad", now.Add(3*time.Hour), outsider.ID); !errors.Is(err, ErrInvitationForbidden) {
		t.Fatal(err)
	}
	if budgetCount(t, db) != 1 {
		t.Fatal("rejected request consumed budget")
	}
}

func TestInvitationSendBudgetPersistenceBackfillAndDeletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err := db.MigrateInvitations(); err != nil {
		t.Fatal(err)
	}
	owner := budgetAdmin(t, db, "persistactor")
	if err := db.CreateServer(&Server{ID: "server", Name: "Server", OwnerID: owner.ID}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	// Simulate an upgrade: historical failed/revoked sends and an expired row.
	for n := 0; n < 7; n++ {
		at := now.Add(time.Duration(n) * time.Minute)
		if n == 6 {
			at = now.Add(-3 * time.Hour)
		}
		i := budgetInvitation(owner, fmt.Sprintf("historical%d", n), at)
		i.Status, i.DeliveryStatus = "revoked", "failed"
		if err := db.CreateInvitation(i); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.conn.Exec(`DELETE FROM invitation_send_budget_migrations WHERE version=1`); err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateInvitations(); err != nil {
		t.Fatal(err)
	}
	if budgetCount(t, db) != 6 {
		t.Fatal("historical backfill missing")
	}
	// Above-limit backfill requires the second-oldest attempt to expire.
	want := now.Add(time.Minute).Add(InvitationSendWindow)
	i := budgetInvitation(owner, "persistblocked", now.Add(10*time.Minute))
	wantBudget(t, db.CreateInvitationForSend(i), "sender", want)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateInvitations(); err != nil {
		t.Fatal(err)
	}
	if budgetCount(t, db) != 6 {
		t.Fatal("reopen duplicated/reset budget")
	}
	wantBudget(t, db.CreateInvitationForSend(i), "sender", want)
	if _, err := db.conn.Exec(`DELETE FROM team_invitations; DELETE FROM servers; DELETE FROM users`); err != nil {
		t.Fatal(err)
	}
	if budgetCount(t, db) != 6 {
		t.Fatal("deletion refunded budget")
	}
	if err := db.MigrateInvitations(); err != nil {
		t.Fatal(err)
	}
	if budgetCount(t, db) != 6 {
		t.Fatal("migration altered durable ledger")
	}
}

func TestInvitationSendBudgetConcurrentConnections(t *testing.T) {
	db, owner := invitationDB(t)
	var seq int
	var name, path string
	if err := db.conn.QueryRow(`PRAGMA database_list`).Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	other, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if err := other.MigrateInvitations(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for n := 0; n < 4; n++ {
		if err := db.CreateInvitationForSend(budgetInvitation(owner, fmt.Sprintf("before%d", n), now)); err != nil {
			t.Fatal(err)
		}
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			conn := db
			if n%2 == 0 {
				conn = other
			}
			err := conn.CreateInvitationForSend(budgetInvitation(owner, fmt.Sprintf("concurrent%d", n), now))
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, ErrInvitationCooldown) {
				t.Errorf("concurrent: %v", err)
			}
		}(n)
	}
	wg.Wait()
	if successes.Load() != 1 || budgetCount(t, db) != 5 {
		t.Fatalf("remaining budget exceeded: %d successes", successes.Load())
	}
}

func TestInvitationSendBudgetMaximumDeadlineAndFailClosed(t *testing.T) {
	db, owner := invitationDB(t)
	now := time.Now().UTC()
	// Sender expires sooner than recipient; preserve the larger wait.
	for n := 0; n < 5; n++ {
		if err := db.CreateInvitationForSend(budgetInvitation(owner, fmt.Sprintf("maximum%d", n), now.Add(-time.Hour))); err != nil {
			t.Fatal(err)
		}
	}
	other := budgetAdmin(t, db, "maximumactor")
	mail := budgetInvitation(other, "maxmail", now)
	if err := db.CreateInvitationForSend(mail); err != nil {
		t.Fatal(err)
	}
	if err := db.RevokeInvitation(mail.ID, nil); err != nil {
		t.Fatal(err)
	}
	i := budgetInvitation(owner, "maxblocked", now)
	i.Email = mail.Email
	wantBudget(t, db.CreateInvitationForSend(i), "recipient", now.Add(InvitationSendWindow))
	// Missing/corrupt quota storage must not permit creation or rotation.
	old := budgetInvitation(owner, "failclosed", now.Add(-3*time.Hour))
	if err := db.CreateInvitation(old); err != nil {
		t.Fatal(err)
	}
	if _, err := db.conn.Exec(`DROP TABLE invitation_send_attempts`); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateInvitationForSend(budgetInvitation(owner, "storagefailure", now)); err == nil {
		t.Fatal("creation failed open")
	}
	if _, err := db.RotateInvitationForSend(old.ID, "bad", now, owner.ID); err == nil {
		t.Fatal("rotation failed open")
	}
	current, err := db.GetInvitation(old.ID)
	if err != nil || current.TokenHash != old.TokenHash {
		t.Fatal("failed quota lookup rotated token", err)
	}
	if _, err := db.InvitationSendAvailability(owner.ID, "local", "server", now); err == nil {
		t.Fatal("availability failed open")
	}
}
