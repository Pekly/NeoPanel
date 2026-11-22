package database

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestCredentialEnrollmentAtomicAuthority(t *testing.T) {
	for _, kind := range []string{"api", "ssh"} {
		t.Run(kind, func(t *testing.T) {
			db, err := InitDB(filepath.Join(t.TempDir(), "enrollment.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			user, err := db.CreateUser("owner", "hash", "user")
			if err != nil {
				t.Fatal(err)
			}
			other, err := db.CreateUser("other", "hash", "admin")
			if err != nil {
				t.Fatal(err)
			}
			source, err := db.AddAPIKey(user.ID, "source", "source-hash", "prefix")
			if err != nil {
				t.Fatal(err)
			}
			version := user.AuthVersion
			jwt := CredentialEnrollmentAuth{UserID: user.ID, AuthVersion: &version}
			api := CredentialEnrollmentAuth{UserID: user.ID, APIKeyHash: "source-hash"}
			insert := func(auth CredentialEnrollmentAuth, wantSuccess bool) {
				t.Helper()
				var err error
				if kind == "api" {
					key, e := db.AddAPIKeyForAuth(auth, "enrolled", "new-hash", "prefix")
					err = e
					if (key != nil) != wantSuccess {
						t.Fatal("unexpected returned API credential")
					}
				} else {
					key, e := db.AddSSHKeyForAuth(auth, "enrolled", "fingerprint", "public")
					err = e
					if (key != nil) != wantSuccess {
						t.Fatal("unexpected returned SSH credential")
					}
				}
				if wantSuccess && err != nil || !wantSuccess && !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("success=%v, error=%v", wantSuccess, err)
				}
				// Remove successful test enrollments so hash uniqueness is immaterial.
				var count int
				table := "ssh_keys"
				if kind == "api" {
					table = "api_keys"
				}
				if err := db.conn.QueryRow("SELECT COUNT(*) FROM " + table + " WHERE name = 'enrolled'").Scan(&count); err != nil {
					t.Fatal(err)
				}
				if (count == 1) != wantSuccess || count > 1 {
					t.Fatalf("unexpected inserted row count %d", count)
				}
				if _, err := db.conn.Exec("DELETE FROM " + table + " WHERE name = 'enrolled'"); err != nil {
					t.Fatal(err)
				}
			}
			insert(jwt, true)
			insert(api, true)
			insert(CredentialEnrollmentAuth{UserID: other.ID, APIKeyHash: api.APIKeyHash}, false)
			insert(CredentialEnrollmentAuth{UserID: user.ID}, false)
			insert(CredentialEnrollmentAuth{UserID: user.ID, AuthVersion: &version, APIKeyHash: api.APIKeyHash}, false)
			insert(CredentialEnrollmentAuth{AuthVersion: &version}, false)
			insert(CredentialEnrollmentAuth{UserID: -1, APIKeyHash: api.APIKeyHash}, false)
			insert(CredentialEnrollmentAuth{UserID: 99999, AuthVersion: &version}, false)
			if err := db.UpdateUserPassword(user.ID, "newhash"); err != nil {
				t.Fatal(err)
			}
			insert(jwt, false)
			insert(api, true) // Password changes do not revoke independent API keys.
			version++
			insert(jwt, true)
			if err := db.DeleteAPIKey(user.ID, source.ID); err != nil {
				t.Fatal(err)
			}
			insert(api, false)
		})
	}
}

func TestAccountMutationAtomicAuthority(t *testing.T) {
	for _, kind := range []string{"profile", "totp"} {
		for _, mode := range []string{"jwt", "api", "stale-jwt", "api-reset", "deleted-key", "other-owner", "missing", "ambiguous", "existing-factor", "revoked-existing-factor"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				db, user, other := passkeyTestDB(t)
				source, err := db.AddAPIKey(user.ID, "source", "source-hash", "prefix")
				if err != nil {
					t.Fatal(err)
				}
				auth := CredentialEnrollmentAuth{UserID: user.ID, AuthVersion: &user.AuthVersion}
				var want error
				switch mode {
				case "api", "api-reset", "deleted-key", "other-owner":
					auth = CredentialEnrollmentAuth{UserID: user.ID, APIKeyHash: "source-hash"}
				case "missing":
					auth.AuthVersion = nil
					want = sql.ErrNoRows
				case "ambiguous":
					auth.APIKeyHash = "source-hash"
					want = sql.ErrNoRows
				}
				if mode == "stale-jwt" || mode == "api-reset" || mode == "revoked-existing-factor" {
					if err := db.UpdateUserPassword(user.ID, "newhash"); err != nil {
						t.Fatal(err)
					}
					if mode != "api-reset" {
						want = sql.ErrNoRows
					}
				}
				if mode == "deleted-key" {
					if err := db.DeleteAPIKey(user.ID, source.ID); err != nil {
						t.Fatal(err)
					}
					want = sql.ErrNoRows
				}
				if mode == "other-owner" {
					auth.APIKeyHash = "other-hash"
					if _, err := db.AddAPIKey(other.ID, "other", "other-hash", "prefix"); err != nil {
						t.Fatal(err)
					}
					want = sql.ErrNoRows
				}
				if mode == "existing-factor" || mode == "revoked-existing-factor" {
					if err := db.EnableUserTwoFactor(user.ID, "existing", "codes"); err != nil {
						t.Fatal(err)
					}
					if kind == "totp" && want == nil {
						want = ErrFactorAlreadyEnabled
					}
				}
				before, err := db.GetUserByID(user.ID)
				if err != nil {
					t.Fatal(err)
				}
				if kind == "profile" {
					avatar := "/changed.png"
					var saved *User
					saved, err = db.UpdateUserProfileForAuth(auth, "changed", "First", "Last", "changed@example.test", &avatar)
					if want != nil && saved != nil {
						t.Fatal("failed write returned a token snapshot")
					}
					if want == nil && (saved == nil || saved.AuthVersion != before.AuthVersion || saved.Username != "changed" || saved.AvatarURL != avatar) {
						t.Fatal("incorrect committed profile snapshot")
					}
				} else {
					err = db.EnableUserTwoFactorForAuth(auth, "new-secret", "new-codes")
				}
				if !errors.Is(err, want) {
					t.Fatalf("error=%v, want %v", err, want)
				}
				after, err := db.GetUserByID(user.ID)
				if err != nil {
					t.Fatal(err)
				}
				if want != nil && *before != *after {
					t.Fatal("rejected mutation changed user")
				}
				if want == nil && kind == "totp" && (!after.TwoFactorEnabled || after.TwoFactorSecret != "new-secret" || after.RecoveryCodes != "new-codes") {
					t.Fatal("factor not persisted")
				}
			})
		}
	}
}

func TestProfileTokenSnapshotDoesNotUpgradeAfterReset(t *testing.T) {
	db, user, _ := passkeyTestDB(t)
	auth := CredentialEnrollmentAuth{UserID: user.ID, AuthVersion: &user.AuthVersion}
	snapshot, err := db.UpdateUserProfileForAuth(auth, "changed", "", "", "changed@example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateUserPassword(user.ID, "newhash"); err != nil {
		t.Fatal(err)
	}
	fresh, err := db.GetUserByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.AuthVersion != user.AuthVersion || fresh.AuthVersion != snapshot.AuthVersion+1 {
		t.Fatal("profile token snapshot upgraded after revocation")
	}
}
