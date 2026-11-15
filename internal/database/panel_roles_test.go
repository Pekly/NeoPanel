package database

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// Reconstruct the retired schema only in disposable test databases.
func seedLegacyPanelRoles(t *testing.T, db *DB, users ...int) {
	t.Helper()
	lifecycleExec(t, db, `
 CREATE TABLE panel_roles (
  id TEXT PRIMARY KEY, name TEXT NOT NULL COLLATE NOCASE UNIQUE,
  description TEXT NOT NULL DEFAULT '', color TEXT NOT NULL DEFAULT '',
  permissions TEXT NOT NULL DEFAULT '[]'
 );
 CREATE TABLE panel_user_roles (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  role_id TEXT NOT NULL REFERENCES panel_roles(id) ON DELETE RESTRICT
 );
 CREATE INDEX idx_panel_user_roles_role ON panel_user_roles(role_id);
 CREATE TRIGGER panel_role_delete_guard BEFORE DELETE ON panel_roles
 WHEN EXISTS (SELECT 1 FROM panel_user_roles WHERE role_id = OLD.id)
 BEGIN SELECT RAISE(ABORT, 'panel role assigned'); END;
 CREATE TRIGGER panel_user_roles_insert_guard BEFORE INSERT ON panel_user_roles
 WHEN NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id)
 OR NOT EXISTS (SELECT 1 FROM panel_roles WHERE id = NEW.role_id)
 BEGIN SELECT RAISE(ABORT, 'invalid panel role assignment'); END;
 CREATE TRIGGER panel_user_roles_update_guard BEFORE UPDATE ON panel_user_roles
 WHEN NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id)
 OR NOT EXISTS (SELECT 1 FROM panel_roles WHERE id = NEW.role_id)
 BEGIN SELECT RAISE(ABORT, 'invalid panel role assignment'); END;
 CREATE TRIGGER panel_user_roles_cleanup AFTER DELETE ON users
 BEGIN DELETE FROM panel_user_roles WHERE user_id = OLD.id; END;
 INSERT INTO panel_roles(id,name,permissions) VALUES('creator','Creators','["servers.create"]');`)
	for _, id := range users {
		lifecycleExec(t, db, `INSERT INTO panel_user_roles(user_id,role_id) VALUES(?,'creator')`, id)
	}
}

func assertPanelRolesRetired(t *testing.T, db *DB) {
	t.Helper()
	var n int
	if err := db.conn.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name IN
 ('panel_roles','panel_user_roles','idx_panel_user_roles_role','panel_role_delete_guard',
 'panel_user_roles_insert_guard','panel_user_roles_update_guard','panel_user_roles_cleanup')`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("retired schema remains: %d %v", n, err)
	}
}

func retirementSnapshot(t *testing.T, db *DB, table string) [][]any {
	t.Helper()
	rows, err := db.conn.Query("SELECT * FROM " + table + " ORDER BY 1")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var result [][]any
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			t.Fatal(err)
		}
		result = append(result, values)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestPanelRoleRetirementFreshReopenAndDataRetention(t *testing.T) {
	for _, fk := range []bool{false, true} {
		t.Run(fmt.Sprint(fk), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "retirement.db")
			db, err := InitDB(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { db.Close() }()
			assertPanelRolesRetired(t, db)
			admin, err := db.CreateUser("admin", "admin-hash", "admin")
			if err != nil {
				t.Fatal(err)
			}
			member, err := db.CreateUserWithDetails("member", "member@example.test", "member-hash", "user", "avatar")
			if err != nil {
				t.Fatal(err)
			}
			unassigned, err := db.CreateUser("unassigned", "unassigned-hash", "user")
			if err != nil {
				t.Fatal(err)
			}
			lifecycleExec(t, db, `UPDATE users SET invitation_created=1 WHERE id IN (?,?)`, member.ID, unassigned.ID)
			lifecycleExec(t, db, `UPDATE users SET two_factor_enabled=1,two_factor_secret='secret',recovery_codes='codes',auth_version=7 WHERE id=?`, member.ID)
			if _, err := db.AddAPIKey(member.ID, "existing", "key-hash", "neo_key"); err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC()
			lifecycleExec(t, db, `INSERT INTO ssh_keys(user_id,name,fingerprint,public_key,created_at) VALUES(?,'key','fp','pub',?)`, member.ID, now)
			if err := db.CreatePasswordResetToken(member.ID, "reset", now.Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
			lifecycleExec(t, db, `INSERT INTO webauthn_users(user_id,handle) VALUES(?,?)`, member.ID, []byte("handle"))
			lifecycleExec(t, db, `INSERT INTO passkeys(id,user_id,credential_id,credential_json,rp_id,label,created_at) VALUES('pk',?,?,'{}','test','key','now')`, member.ID, []byte("credential"))
			if err := db.CreateServer(&Server{ID: "owned", Name: "Owned", OwnerID: member.ID}); err != nil {
				t.Fatal(err)
			}
			if err := db.CreateServer(&Server{ID: "shared", Name: "Shared", OwnerID: admin.ID}); err != nil {
				t.Fatal(err)
			}
			if _, err := db.AddServerSubuser("shared", member.ID, []string{"files.view", "console.view"}); err != nil {
				t.Fatal(err)
			}
			seedLegacyPanelRoles(t, db, member.ID, admin.ID)
			beforeUsers, err := db.ListUsers()
			if err != nil {
				t.Fatal(err)
			}
			beforeMember, err := db.GetUserByID(member.ID)
			if err != nil {
				t.Fatal(err)
			}
			beforeMember.InvitationCreated = false
			for _, user := range beforeUsers {
				if user.ID == member.ID {
					user.InvitationCreated = false
				}
			}
			snapshots := map[string][][]any{}
			for _, table := range []string{"servers", "server_subusers", "api_keys", "ssh_keys", "password_resets", "passkeys", "webauthn_users"} {
				snapshots[table] = retirementSnapshot(t, db, table)
			}
			if fk {
				lifecycleExec(t, db, `PRAGMA foreign_keys=ON`)
				if err := db.retirePanelRoles(); err != nil {
					t.Fatal(err)
				}
			}
			for startup := 0; startup < 2; startup++ {
				if err := db.Close(); err != nil {
					t.Fatal(err)
				}
				db, err = InitDB(path)
				if err != nil {
					t.Fatal(err)
				}
				assertPanelRolesRetired(t, db)
				afterUsers, err := db.ListUsers()
				if err != nil || !reflect.DeepEqual(beforeUsers, afterUsers) {
					t.Fatalf("accounts changed: %+v %v", afterUsers, err)
				}
				for _, lookup := range []func() (*User, error){
					func() (*User, error) { return db.GetUserByID(member.ID) },
					func() (*User, error) { return db.GetUserByUsername(member.Username) },
					func() (*User, error) { return db.GetUserByEmailOrUsername(member.Email) },
				} {
					after, err := lookup()
					if err != nil || !reflect.DeepEqual(beforeMember, after) {
						t.Fatalf("identity/credentials changed: %+v %v", after, err)
					}
				}
				keyUser, err := db.GetUserByAPIKeyHash("key-hash")
				if err != nil || keyUser.ID != member.ID || keyUser.Role != "user" {
					t.Fatalf("API key lost: %+v %v", keyUser, err)
				}
				for table, before := range snapshots {
					if !reflect.DeepEqual(before, retirementSnapshot(t, db, table)) {
						t.Fatalf("%s changed", table)
					}
				}
				permissions, err := db.GetUserServerPermissions("shared", member.ID)
				if err != nil || !reflect.DeepEqual(permissions, []string{"files.view", "console.view"}) {
					t.Fatalf("team rights changed: %v %v", permissions, err)
				}
			}
		})
	}
}

func TestPanelRoleRetirementRollback(t *testing.T) {
	db, owner := invitationDB(t)
	u, _ := lifecycleInvite(t, db, owner, "protected")
	seedLegacyPanelRoles(t, db, u.ID)
	lifecycleExec(t, db, `PRAGMA foreign_keys=ON`)
	// Force the final DROP to fail after the provenance update and earlier DDL.
	lifecycleExec(t, db, `CREATE TABLE retirement_blocker(role_id TEXT REFERENCES panel_roles(id)); INSERT INTO retirement_blocker VALUES('creator')`)
	if err := db.retirePanelRoles(); err == nil {
		t.Fatal("expected retirement failure")
	}
	loaded, err := db.GetUserByID(u.ID)
	if err != nil || !loaded.InvitationCreated {
		t.Fatalf("provenance update was not rolled back: %+v %v", loaded, err)
	}
	var count int
	if err := db.conn.QueryRow(`SELECT count(*) FROM panel_user_roles WHERE user_id=?`, u.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("assignment not rolled back: %d %v", count, err)
	}
	if err := db.conn.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name IN ('panel_role_delete_guard','panel_user_roles_insert_guard','panel_user_roles_update_guard','panel_user_roles_cleanup')`).Scan(&count); err != nil || count != 4 {
		t.Fatalf("triggers not rolled back: %d %v", count, err)
	}
	lifecycleExec(t, db, `DROP TABLE retirement_blocker`)
	if err := db.retirePanelRoles(); err != nil {
		t.Fatal(err)
	}
	assertPanelRolesRetired(t, db)
}

func TestPanelRoleRetirementPreservesFormerInvitationUserAfterReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invitation-retirement.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	owner, err := db.CreateUser("owner", "owner-hash", "user")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateServer(&Server{ID: "server", Name: "Shared", OwnerID: owner.ID}); err != nil {
		t.Fatal(err)
	}
	retained, retainedMembership := lifecycleInvite(t, db, owner, "retained")
	ordinary, ordinaryMembership := lifecycleInvite(t, db, owner, "ordinary")
	seedLegacyPanelRoles(t, db, retained.ID)
	// Simulate an upgrade that also needs the invitation provenance backfill.
	lifecycleExec(t, db, `DELETE FROM invitation_lifecycle_migrations; UPDATE users SET invitation_created=0`)
	invitations := retirementSnapshot(t, db, "team_invitations")
	for startup := 0; startup < 2; startup++ {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		db, err = InitDB(path)
		if err != nil {
			t.Fatal(err)
		}
		assertPanelRolesRetired(t, db)
		u, err := db.GetUserByID(retained.ID)
		if err != nil || u.InvitationCreated || u.Role != "user" || u.PasswordHash != retained.PasswordHash {
			t.Fatalf("former invitation user changed: %+v %v", u, err)
		}
		u, err = db.GetUserByID(ordinary.ID)
		if err != nil || !u.InvitationCreated {
			t.Fatalf("unassigned invitation provenance changed: %+v %v", u, err)
		}
		if !reflect.DeepEqual(invitations, retirementSnapshot(t, db, "team_invitations")) {
			t.Fatal("retirement changed invitations")
		}
	}
	if err := db.DeleteServerSubuser("server", retainedMembership); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetUserByID(retained.ID); err != nil {
		t.Fatal("last membership removal deleted former assigned user", err)
	}
	if err := db.DeleteServerSubuser("server", ordinaryMembership); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetUserByID(ordinary.ID); err == nil {
		t.Fatal("ordinary invitation cleanup stopped working")
	}
}
