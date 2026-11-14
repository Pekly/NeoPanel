package database

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func lifecycleInvite(t *testing.T, db *DB, owner *User, name string) (*User, int) {
	t.Helper()
	i := seedInvitation(t, db, owner, name)
	u, err := db.AcceptInvitation(i.ID, i.TokenHash, 0, "hash", nil)
	if err != nil {
		t.Fatal(err)
	}
	var mid int
	if err = db.conn.QueryRow(`SELECT id FROM server_subusers WHERE user_id=? AND server_id='server'`, u.ID).Scan(&mid); err != nil {
		t.Fatal(err)
	}
	return u, mid
}
func lifecycleExec(t *testing.T, db *DB, query string, args ...any) {
	t.Helper()
	if _, err := db.conn.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}
func TestInvitationLifecycleRemovalAndCredentials(t *testing.T) {
	for _, fk := range []bool{false, true} {
		t.Run(fmt.Sprint(fk), func(t *testing.T) {
			db, owner := invitationDB(t)
			if fk {
				lifecycleExec(t, db, `PRAGMA foreign_keys=ON`)
			}
			u, mid := lifecycleInvite(t, db, owner, "created")
			if !u.InvitationCreated {
				t.Fatal("missing creation provenance")
			}
			now := time.Now().UTC()
			lifecycleExec(t, db, `INSERT INTO ssh_keys(user_id,name,fingerprint,public_key,created_at) VALUES(?,'key','fp','pub',?)`, u.ID, now)
			lifecycleExec(t, db, `INSERT INTO api_keys(user_id,name,key_hash,key_prefix,created_at) VALUES(?,'key','hash','prefix',?)`, u.ID, now)
			if err := db.CreatePasswordResetToken(u.ID, "reset", now.Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
			lifecycleExec(t, db, `INSERT INTO webauthn_users(user_id,handle) VALUES(?,?)`, u.ID, []byte("handle"))
			lifecycleExec(t, db, `INSERT INTO passkeys(id,user_id,credential_id,credential_json,rp_id,label,created_at) VALUES('pk',?,?,'{}','test','key','now')`, u.ID, []byte("credential"))
			if err := db.DeleteServerSubuser("wrong", mid); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal(err)
			}
			if _, err := db.GetUserByID(u.ID); err != nil {
				t.Fatal("failed deletion changed account", err)
			}
			if err := db.DeleteServerSubuser("server", mid); err != nil {
				t.Fatal(err)
			}
			if _, err := db.GetUserByID(u.ID); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("account survived", err)
			}
			for _, table := range []string{"ssh_keys", "api_keys", "password_resets", "passkeys", "webauthn_users"} {
				var n int
				if err := db.conn.QueryRow("SELECT count(*) FROM "+table+" WHERE user_id=?", u.ID).Scan(&n); err != nil || n != 0 {
					t.Fatal(table, n, err)
				}
			}
			if err := db.UpsertSyncedUser(u.ID, u.Username, u.Email, "hash", "user"); err == nil {
				t.Fatal("sync resurrected identity")
			}
			if _, err := db.AddServerSubuser("server", u.ID, nil); err == nil {
				t.Fatal("late membership resurrected access")
			}
		})
	}
}

func TestInvitationLifecyclePreservesIndependentAndAssociatedAccounts(t *testing.T) {
	for _, kind := range []string{"existing-invite", "direct-add", "membership", "ownership", "admin", "builtin", "panel-role", "unknown-role", "unresolved", "exported", "configured-node"} {
		t.Run(kind, func(t *testing.T) {
			db, owner := invitationDB(t)
			var u *User
			var mid int
			if kind == "existing-invite" || kind == "direct-add" {
				var err error
				u, err = db.CreateUserWithDetails("existing", "existing@example.test", "hash", "user", "")
				if err != nil {
					t.Fatal(err)
				}
				if kind == "existing-invite" {
					i := seedInvitation(t, db, owner, "existing")
					u, err = db.AcceptInvitation(i.ID, i.TokenHash, u.ID, "", nil)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					_, err = db.AddServerSubuser("server", u.ID, nil)
					if err != nil {
						t.Fatal(err)
					}
				}
				if u.InvitationCreated {
					t.Fatal("independent account marked invitation-created")
				}
				if err = db.conn.QueryRow(`SELECT id FROM server_subusers WHERE user_id=?`, u.ID).Scan(&mid); err != nil {
					t.Fatal(err)
				}
			} else {
				u, mid = lifecycleInvite(t, db, owner, "created")
			}
			switch kind {
			case "membership":
				if err := db.CreateServer(&Server{ID: "other", Name: "other", OwnerID: owner.ID}); err != nil {
					t.Fatal(err)
				}
				if _, err := db.AddServerSubuser("other", u.ID, nil); err != nil {
					t.Fatal(err)
				}
			case "ownership":
				if err := db.CreateServer(&Server{ID: "other", Name: "other", OwnerID: u.ID}); err != nil {
					t.Fatal(err)
				}
			case "admin":
				if err := db.UpdateUserRole(u.ID, "admin"); err != nil {
					t.Fatal(err)
				}
			case "builtin":
				lifecycleExec(t, db, `UPDATE users SET invitation_created=1 WHERE id=?`, owner.ID)
				su, err := db.AddServerSubuser("server", owner.ID, nil)
				if err != nil {
					t.Fatal(err)
				}
				u = owner
				mid = su.ID
			case "panel-role":
				lifecycleExec(t, db, `INSERT INTO panel_roles(id,name) VALUES('r','Manager')`)
				lifecycleExec(t, db, `INSERT INTO panel_user_roles(user_id,role_id) VALUES(?,'r')`, u.ID)
			case "unknown-role":
				lifecycleExec(t, db, `UPDATE users SET role='operator' WHERE id=?`, u.ID)
			case "unresolved":
				lifecycleExec(t, db, `UPDATE team_invitations SET status='granting',node_id='remote' WHERE account_id=?`, u.ID)
			case "exported":
				if err := db.PinExportedIdentity(u.ID); err != nil {
					t.Fatal(err)
				}
			case "configured-node":
				if err := db.CreateNode(&ClusterNode{ID: "offline", Name: "offline", BaseURL: "https://offline.invalid"}); err != nil {
					t.Fatal(err)
				}
			}
			if err := db.DeleteServerSubuser("server", mid); err != nil {
				t.Fatal(err)
			}
			if _, err := db.GetUserByID(u.ID); err != nil {
				t.Fatal("protected account deleted", err)
			}
		})
	}
}

func TestInvitationLifecycleServerDeletionAndRollback(t *testing.T) {
	db, owner := invitationDB(t)
	first, _ := lifecycleInvite(t, db, owner, "first")
	second, _ := lifecycleInvite(t, db, owner, "second")
	kept, _ := lifecycleInvite(t, db, owner, "kept")
	if err := db.CreateServer(&Server{ID: "other", Name: "other", OwnerID: kept.ID}); err != nil {
		t.Fatal(err)
	}
	independent, err := db.CreateUser("independent", "hash", "user")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.AddServerSubuser("server", independent.ID, nil); err != nil {
		t.Fatal(err)
	}
	lifecycleExec(t, db, `CREATE TRIGGER fail_server_delete BEFORE DELETE ON servers BEGIN SELECT RAISE(ABORT,'test failure'); END`)
	if err = db.DeleteServer("server"); err == nil {
		t.Fatal("expected failure")
	}
	members, err := db.ListServerSubusers("server")
	if err != nil || len(members) != 4 {
		t.Fatal("failed deletion changed members", len(members), err)
	}
	if _, err = db.GetUserByID(first.ID); err != nil {
		t.Fatal("failed deletion changed users", err)
	}
	lifecycleExec(t, db, `DROP TRIGGER fail_server_delete`)
	if err = db.DeleteServer("server"); err != nil {
		t.Fatal(err)
	}
	for _, u := range []*User{first, second} {
		if _, err = db.GetUserByID(u.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatal(u.Username, err)
		}
	}
	for _, u := range []*User{owner, kept, independent} {
		if _, err = db.GetUserByID(u.ID); err != nil {
			t.Fatal(u.Username, err)
		}
	}
	members, err = db.ListServerSubusers("server")
	if err != nil || len(members) != 0 {
		t.Fatal(members, err)
	}
}

func TestInvitationLifecycleRemoteReceiptAndIdentitySafety(t *testing.T) {
	master, owner := invitationDB(t)
	worker, workerOwner := invitationDB(t)
	i := seedInvitation(t, master, owner, "remoteuser")
	lifecycleExec(t, master, `UPDATE team_invitations SET node_id='worker' WHERE id=?`, i.ID)
	if err := master.ReserveInvitationAccount(i.ID, i.TokenHash, 0, "hash"); err != nil {
		t.Fatal(err)
	}
	u, err := master.AcceptInvitation(i.ID, i.TokenHash, 0, "", func(i *Invitation, u, inviter *User) error {
		if err := worker.ProvisionInvitationUser(u); err != nil {
			return err
		}
		return worker.InvitationMembership(i.ID, i.ServerID, u.ID, workerOwner.ID, "user", i.Permissions, i.ExpiresAt, false)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !u.InvitationCreated {
		t.Fatal("lost provenance during reservation recovery")
	}
	members, err := worker.ListServerSubusers("server")
	if err != nil || len(members) != 1 {
		t.Fatal(members, err)
	}
	if err = worker.DeleteServerSubuser("server", members[0].ID); err != nil {
		t.Fatal(err)
	}
	ids, err := worker.RevokedInvitationGrants("server")
	if err != nil || len(ids) != 1 {
		t.Fatal(ids, err)
	}
	if err = master.ReconcileInvitationRemovals("wrong-node", "server", ids); err != nil {
		t.Fatal(err)
	}
	saved, _ := master.GetInvitation(i.ID)
	if saved.Status != "accepted" {
		t.Fatal("wrong-node reconciliation")
	}
	if err = master.ReconcileInvitationRemovals("worker", "server", ids); err != nil {
		t.Fatal(err)
	}
	saved, _ = master.GetInvitation(i.ID)
	if saved.Status != "revoked" {
		t.Fatal(saved.Status)
	}
	for _, db := range []*DB{master, worker} {
		if _, err = db.GetUserByID(u.ID); err != nil {
			t.Fatal("remote identity unsafely deleted", err)
		}
	}
	if err = worker.InvitationMembership(i.ID, "server", u.ID, workerOwner.ID, "user", i.Permissions, i.ExpiresAt, false); !errors.Is(err, ErrInvitationGone) {
		t.Fatal("late grant resurrected", err)
	}
}

func TestInvitationLifecycleLastAssociationRace(t *testing.T) {
	db, owner := invitationDB(t)
	if err := db.CreateServer(&Server{ID: "other", Name: "other", OwnerID: owner.ID}); err != nil {
		t.Fatal(err)
	}
	var sequence int
	var name, path string
	if err := db.conn.QueryRow(`PRAGMA database_list`).Scan(&sequence, &name, &path); err != nil {
		t.Fatal(err)
	}
	other, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	for n := 0; n < 20; n++ {
		u, mid := lifecycleInvite(t, db, owner, fmt.Sprintf("race%d", n))
		var wg sync.WaitGroup
		wg.Add(2)
		var addErr, deleteErr error
		go func() { defer wg.Done(); _, addErr = other.AddServerSubuser("other", u.ID, nil) }()
		go func() { defer wg.Done(); deleteErr = db.DeleteServerSubuser("server", mid) }()
		wg.Wait()
		if deleteErr != nil {
			t.Fatal(deleteErr)
		}
		_, getErr := db.GetUserByID(u.ID)
		if addErr == nil && getErr != nil {
			t.Fatal("new grant lost its user", getErr)
		}
		if addErr != nil && !errors.Is(getErr, sql.ErrNoRows) {
			t.Fatal("unexpected race outcome", addErr, getErr)
		}
	}
}

func TestInvitationLifecycleLegacyMigration(t *testing.T) {
	db, owner := invitationDB(t)
	proven, _ := lifecycleInvite(t, db, owner, "proven")
	ambiguous, _ := lifecycleInvite(t, db, owner, "ambiguous")
	independent, err := db.CreateUserWithDetails("existing", "existing@example.test", "hash", "user", "")
	if err != nil {
		t.Fatal(err)
	}
	i := seedInvitation(t, db, owner, "existing")
	if _, err = db.AcceptInvitation(i.ID, i.TokenHash, independent.ID, "", nil); err != nil {
		t.Fatal(err)
	}
	lifecycleExec(t, db, `UPDATE users SET invitation_created=0`)
	lifecycleExec(t, db, `UPDATE users SET email='changed@example.test' WHERE id=?`, ambiguous.ID)
	lifecycleExec(t, db, `DELETE FROM invitation_lifecycle_migrations`)
	if err = db.migrateInvitationLifecycle(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id   int
		want bool
	}{{proven.ID, true}, {ambiguous.ID, false}, {independent.ID, false}, {owner.ID, false}} {
		u, err := db.GetUserByID(test.id)
		if err != nil || u.InvitationCreated != test.want {
			t.Fatal(test, u, err)
		}
	}
}

func TestInvitationLifecycleLocalNodeSettingsDoNotPreventCleanup(t *testing.T) {
	for _, operation := range []string{"membership", "server"} {
		t.Run(operation, func(t *testing.T) {
			db, owner := invitationDB(t)
			if err := db.CreateNode(&ClusterNode{ID: "local", Name: "Local machine"}); err != nil {
				t.Fatal(err)
			}
			u, membership := lifecycleInvite(t, db, owner, "localcreated")
			// A local settings row is not an exported identity, including at upgrade.
			lifecycleExec(t, db, `DELETE FROM invitation_lifecycle_migrations WHERE id=1`)
			if err := db.migrateInvitationLifecycle(); err != nil {
				t.Fatal(err)
			}
			var err error
			if operation == "membership" {
				err = db.DeleteServerSubuser("server", membership)
			} else {
				err = db.DeleteServer("server")
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.GetUserByID(u.ID); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("local settings disabled account cleanup", err)
			}
		})
	}
}
