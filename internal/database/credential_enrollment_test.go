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
