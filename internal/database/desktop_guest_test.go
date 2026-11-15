package database

import (
	"golang.org/x/crypto/bcrypt"
	"path/filepath"
	"testing"
)

func guestDB(t *testing.T) *DB {
	t.Helper()
	db, err := InitDB(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestDesktopGuestFreshAndPersistent(t *testing.T) {
	db := guestDB(t)
	u, err := db.PrepareDesktopGuest()
	if err != nil || u == nil {
		t.Fatalf("guest: %v %v", u, err)
	}
	if u.Username != "Guest" || u.Role != "admin" {
		t.Fatal(u)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("")) == nil {
		t.Fatal("guest has usable password")
	}
	for range 3 {
		next, err := db.PrepareDesktopGuest()
		if err != nil || next == nil || next.ID != u.ID {
			t.Fatalf("reuse: %v %v", next, err)
		}
	}
	count, _ := db.GetUserCount()
	if count != 1 {
		t.Fatal(count)
	}
}
func TestDesktopGuestNeverAdoptsExistingAccount(t *testing.T) {
	db := guestDB(t)
	u, err := db.CreateUser("Guest", "existing-password-hash", "admin")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := db.PrepareDesktopGuest()
	if err != nil || guest != nil {
		t.Fatalf("adopted: %v %v", guest, err)
	}
	saved, _ := db.GetUserByID(u.ID)
	if saved.PasswordHash != u.PasswordHash {
		t.Fatal("changed existing account")
	}
	if _, err := db.conn.Exec(`DELETE FROM users`); err != nil {
		t.Fatal(err)
	}
	guest, err = db.PrepareDesktopGuest()
	if err != nil || guest != nil {
		t.Fatalf("adopted after deletion: %v %v", guest, err)
	}
}
func TestDesktopGuestSecurityChangesPermanentlyRevoke(t *testing.T) {
	changes := map[string]string{
		"password":   `UPDATE users SET password_hash='new'`,
		"version":    `UPDATE users SET auth_version=auth_version+1`,
		"two factor": `UPDATE users SET two_factor_enabled=1, two_factor_secret='secret'`,
		"role":       `UPDATE users SET role='user'`,
		"passkey":    `INSERT INTO passkeys(id,user_id,credential_id,credential_json,rp_id,label,created_at) SELECT 'test',id,x'01','{}','localhost','key','today' FROM users`,
		"delete":     `DELETE FROM users`,
	}
	for name, query := range changes {
		t.Run(name, func(t *testing.T) {
			db := guestDB(t)
			if _, err := db.PrepareDesktopGuest(); err != nil {
				t.Fatal(err)
			}
			if _, err := db.conn.Exec(query); err != nil {
				t.Fatal(err)
			}
			// Restoring the old credential fields must not re-enable automatic access.
			if _, err := db.conn.Exec(`DELETE FROM passkeys; UPDATE users SET password_hash='!desktop-guest-no-password',auth_version=0,two_factor_enabled=0,two_factor_secret='',role='admin'`); err != nil {
				t.Fatal(err)
			}
			guest, err := db.PrepareDesktopGuest()
			if err != nil || guest != nil {
				t.Fatalf("revoked guest: %v %v", guest, err)
			}
		})
	}
}
