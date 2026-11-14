package database

import (
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func invitationDB(t *testing.T) (*DB, *User) {
	t.Helper()
	db, e := InitDB(filepath.Join(t.TempDir(), "invites.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e = db.MigrateInvitations(); e != nil {
		t.Fatal(e)
	}
	owner, e := db.CreateUserWithDetails("owner", "owner@example.test", "owner-hash", "user", "")
	if e != nil {
		t.Fatal(e)
	}
	if e = db.CreateServer(&Server{ID: "server", Name: "Test", OwnerID: owner.ID}); e != nil {
		t.Fatal(e)
	}
	return db, owner
}
func seedInvitation(t *testing.T, db *DB, owner *User, id string) *Invitation {
	t.Helper()
	now := time.Now().UTC()
	i := &Invitation{ID: id, ServerID: "server", NodeID: "local", Username: id, Email: id + "@example.test", Permissions: []string{"console.view", "files.view"}, InviterID: owner.ID, TokenHash: "hash-" + id, Status: "pending", DeliveryStatus: "sent", CreatedAt: now, LastSentAt: now, ExpiresAt: now.Add(time.Hour)}
	if e := db.CreateInvitation(i); e != nil {
		t.Fatal(e)
	}
	return i
}
func TestInvitationLocalAtomicConcurrent(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "newuser")
	var success atomic.Int32
	var wg sync.WaitGroup
	for x := 0; x < 12; x++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := db.AcceptInvitation(i.ID, i.TokenHash, 0, "hash", nil)
			if e == nil {
				success.Add(1)
			} else if !errors.Is(e, ErrInvitationGone) {
				t.Errorf("unexpected: %v", e)
			}
		}()
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatalf("accepted %d times", success.Load())
	}
	u, e := db.GetUserByUsername(i.Username)
	if e != nil {
		t.Fatal(e)
	}
	if u.Role != "user" || u.Email != i.Email {
		t.Fatalf("bad account: %+v", u)
	}
	p, e := db.GetUserServerPermissions(i.ServerID, u.ID)
	if e != nil || len(p) != 2 {
		t.Fatalf("permissions %v %v", p, e)
	}
}
func TestInvitationExpiryRevocationRotationAndCollisions(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "expires")
	if _, e := db.RotateInvitation(i.ID, "new-hash", time.Now()); !errors.Is(e, ErrInvitationCooldown) {
		t.Fatal(e)
	}
	rotated, e := db.RotateInvitation(i.ID, "new-hash", time.Now().Add(2*time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.AcceptInvitation(i.ID, i.TokenHash, 0, "hash", nil); !errors.Is(e, ErrInvitationGone) {
		t.Fatal(e)
	}
	if e = db.RevokeInvitation(i.ID, nil); e != nil {
		t.Fatal(e)
	}
	if _, e = db.AcceptInvitation(i.ID, rotated.TokenHash, 0, "hash", nil); !errors.Is(e, ErrInvitationGone) {
		t.Fatal(e)
	}
	i = seedInvitation(t, db, owner, "expired")
	if _, e = db.conn.Exec(`UPDATE team_invitations SET expires_at=? WHERE id=?`, time.Now().Add(-time.Hour), i.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = db.AcceptInvitation(i.ID, i.TokenHash, 0, "hash", nil); !errors.Is(e, ErrInvitationGone) {
		t.Fatal(e)
	}
	i = seedInvitation(t, db, owner, "duplicate")
	copy := *i
	copy.ID = "another"
	copy.TokenHash = "another"
	if e = db.CreateInvitation(&copy); !errors.Is(e, ErrInvitationConflict) {
		t.Fatal(e)
	}
	copy.Email = "different@example.test"
	if e = db.CreateInvitation(&copy); !errors.Is(e, ErrInvitationConflict) {
		t.Fatal(e)
	}
}
func TestInvitationStaleInviterAndDeletedServerRollback(t *testing.T) {
	for _, kind := range []string{"deleted", "demoted", "permission"} {
		t.Run(kind, func(t *testing.T) {
			db, owner := invitationDB(t)
			i := seedInvitation(t, db, owner, "invitee")
			switch kind {
			case "deleted":
				db.DeleteServer(i.ServerID)
			case "demoted":
				db.UpdateServerOwner(i.ServerID, 999)
			case "permission":
				db.UpdateServerOwner(i.ServerID, 999)
				su, e := db.AddServerSubuser(i.ServerID, owner.ID, []string{"subusers.manage"})
				if e != nil {
					t.Fatal(e)
				}
				db.DeleteServerSubuser(i.ServerID, su.ID)
			}
			wantErr := ErrInvitationForbidden
			wantStatus := "pending"
			if kind == "deleted" {
				wantErr = ErrInvitationGone
				wantStatus = "revoked"
			}
			if _, e := db.AcceptInvitation(i.ID, i.TokenHash, 0, "hash", nil); !errors.Is(e, wantErr) {
				t.Fatal(e)
			}
			if u, _ := db.GetUserByUsername(i.Username); u != nil {
				t.Fatal("account leaked from failed local transaction")
			}
			stored, _ := db.GetInvitation(i.ID)
			if stored.Status != wantStatus {
				t.Fatal(stored.Status)
			}
		})
	}
}
func TestInvitationExistingIdentityAndReservedRecovery(t *testing.T) {
	db, owner := invitationDB(t)
	existing, e := db.CreateUserWithDetails("existing", "existing@example.test", "original", "user", "")
	if e != nil {
		t.Fatal(e)
	}
	i := seedInvitation(t, db, owner, "existing")
	for _, session := range []int{0, owner.ID} {
		if _, e = db.AcceptInvitation(i.ID, i.TokenHash, session, "attacker", nil); !errors.Is(e, ErrInvitationIdentity) {
			t.Fatal(e)
		}
	}
	if _, e = db.AcceptInvitation(i.ID, i.TokenHash, existing.ID, "attacker", nil); e != nil {
		t.Fatal(e)
	}
	existing, _ = db.GetUserByID(existing.ID)
	if existing.PasswordHash != "original" {
		t.Fatal("existing password reset")
	}
	i = seedInvitation(t, db, owner, "recover")
	if e = db.ReserveInvitationAccount(i.ID, i.TokenHash, 0, "first-password"); e != nil {
		t.Fatal(e)
	}
	offline := errors.New("offline")
	if _, e = db.AcceptInvitation(i.ID, i.TokenHash, 0, "changed", func(*Invitation, *User, *User) error { return offline }); !errors.Is(e, offline) {
		t.Fatal(e)
	}
	if _, e = db.RotateInvitation(i.ID, "changed", time.Now().Add(time.Hour)); !errors.Is(e, ErrInvitationBusy) {
		t.Fatal(e)
	}
	var recoveredID int
	if _, e = db.AcceptInvitation(i.ID, i.TokenHash, 0, "changed", func(_ *Invitation, u, _ *User) error {
		recoveredID = u.ID
		if u.PasswordHash != "first-password" {
			t.Fatal("retry reset password")
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	stored, _ := db.GetInvitation(i.ID)
	if stored.AccountID != recoveredID || stored.Status != "accepted" {
		t.Fatalf("bad recovery %+v", stored)
	}
}
func TestInvitationWorkerReceiptsTombstonesAndProvisioning(t *testing.T) {
	db, owner := invitationDB(t)
	u := &User{ID: 10, Username: "workeruser", Email: "worker@example.test", PasswordHash: "first", Role: "user"}
	if e := db.ProvisionInvitationUser(u); e != nil {
		t.Fatal(e)
	}
	u.PasswordHash = "changed"
	if e := db.ProvisionInvitationUser(u); e != nil {
		t.Fatal(e)
	}
	stored, _ := db.GetUserByID(u.ID)
	if stored.PasswordHash != "first" {
		t.Fatal("provision reset password")
	}
	collision := *u
	collision.ID = 11
	if e := db.ProvisionInvitationUser(&collision); !errors.Is(e, ErrInvitationConflict) {
		t.Fatal(e)
	}
	ps := []string{"console.view"}
	grant := func(id string, revoke bool) error {
		return db.InvitationMembership(id, "server", u.ID, owner.ID, "user", ps, time.Now().Add(time.Hour), revoke)
	}
	if e := grant("grant", false); e != nil {
		t.Fatal(e)
	}
	if e := grant("grant", false); e != nil {
		t.Fatal(e)
	}
	subs, _ := db.ListServerSubusers("server")
	if len(subs) != 1 {
		t.Fatal(len(subs))
	}
	db.DeleteServerSubuser("server", subs[0].ID)
	if e := grant("grant", false); !errors.Is(e, ErrInvitationGone) {
		t.Fatal(e)
	}
	subs, _ = db.ListServerSubusers("server")
	if len(subs) != 0 {
		t.Fatal("retry resurrected deleted membership")
	}
	if e := grant("revoked", true); e != nil {
		t.Fatal(e)
	}
	if e := grant("revoked", false); e == nil {
		t.Fatal("late grant passed revoke tombstone")
	}
	if e := db.InvitationMembership("expired", "server", u.ID, owner.ID, "user", ps, time.Now().Add(-time.Hour), false); !errors.Is(e, ErrInvitationGone) {
		t.Fatal(e)
	}
}

func TestInvitationOverlapTransactionalChecks(t *testing.T) {
	for _, kind := range []string{"member", "owner", "admin", "identity"} {
		t.Run(kind, func(t *testing.T) {
			db, owner := invitationDB(t)
			u, err := db.CreateUserWithDetails("pending", "pending@example.test", "original", "user", "")
			if err != nil {
				t.Fatal(err)
			}
			i := seedInvitation(t, db, owner, "pending")
			switch kind {
			case "member":
				_, err = db.AddServerSubuser("server", u.ID, []string{"files.edit"})
			case "owner":
				// Keep the original inviter authorized after ownership changes.
				_, err = db.conn.Exec(`UPDATE users SET role='admin' WHERE id=?`, owner.ID)
				if err == nil {
					err = db.UpdateServerOwner("server", u.ID)
				}
			case "admin":
				_, err = db.conn.Exec(`UPDATE users SET role='admin' WHERE id=?`, u.ID)
			case "identity":
				_, err = db.conn.Exec(`UPDATE users SET email='changed@example.test' WHERE id=?`, u.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.RotateInvitation(i.ID, "rotated", time.Now().Add(2*time.Minute)); !errors.Is(err, ErrInvitationConflict) {
				t.Fatalf("rotate: %v", err)
			}
			if _, err = db.AcceptInvitation(i.ID, i.TokenHash, u.ID, "replacement", nil); !errors.Is(err, ErrInvitationConflict) {
				t.Fatalf("accept: %v", err)
			}
			stored, err := db.GetInvitation(i.ID)
			if err != nil || stored.TokenHash != i.TokenHash || stored.Status != "pending" {
				t.Fatalf("conflict mutated invite: %+v %v", stored, err)
			}
			user, err := db.GetUserByID(u.ID)
			if err != nil || user.PasswordHash != "original" {
				t.Fatal("password reset")
			}
			if kind != "identity" {
				if err = db.InvitationMembership("worker-grant", "server", u.ID, owner.ID, "admin", i.Permissions, i.ExpiresAt, false); !errors.Is(err, ErrInvitationMember) {
					t.Fatalf("worker accepted overlap: %v", err)
				}
			}
			if kind == "member" {
				ps, err := db.GetUserServerPermissions("server", u.ID)
				if err != nil || len(ps) != 1 || ps[0] != "files.edit" {
					t.Fatalf("permissions changed: %v %v", ps, err)
				}
			}
		})
	}
}

func TestInvitationExpiredResendCannotStealNewReservation(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "reserved")
	if _, err := db.conn.Exec(`UPDATE team_invitations SET expires_at=?,last_sent_at=? WHERE id=?`, time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(-2*time.Hour), i.ID); err != nil {
		t.Fatal(err)
	}
	copy := *i
	copy.ID, copy.TokenHash, copy.Email = "new-reservation", "new-hash", "different@example.test"
	if err := db.CreateInvitation(&copy); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RotateInvitation(i.ID, "rotated", time.Now()); !errors.Is(err, ErrInvitationConflict) {
		t.Fatalf("expired resend stole identity: %v", err)
	}
}

func TestInvitationWorkerCheckUsesOriginalCaller(t *testing.T) {
	db, _ := invitationDB(t)
	u, err := db.CreateUserWithDetails("outsider", "outsider@example.test", "hash", "user", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.CheckInvitationTargetOnNode("server", "available", "available@example.test", 0, u.ID, u.Role); !errors.Is(err, ErrInvitationForbidden) {
		t.Fatalf("unauthorized target check: %v", err)
	}
	if _, err = db.AddServerSubuser("server", u.ID, []string{"subusers.manage"}); err != nil {
		t.Fatal(err)
	}
	if err = db.CheckInvitationTargetOnNode("server", "available", "available@example.test", 0, u.ID, u.Role); err != nil {
		t.Fatal(err)
	}
}
