package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSecurityLegacyAuthVersionMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', created_at DATETIME NOT NULL); INSERT INTO users (username,password_hash,role,created_at) VALUES ('legacy','hash','admin',?)`, time.Now().UTC())
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	user, err := db.GetUserByUsername("legacy")
	if err != nil || user.AuthVersion != 0 {
		t.Fatal("legacy version", err)
	}
	if err := db.UpdateUserPassword(user.ID, "newhash"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, lookup := range []func() (*User, error){func() (*User, error) { return db.GetUserByID(user.ID) }, func() (*User, error) { return db.GetUserByUsername("legacy") }, func() (*User, error) { return db.GetUserByEmailOrUsername("legacy") }} {
		got, err := lookup()
		if err != nil || got.AuthVersion != 1 {
			t.Fatal("version not persisted/scanned", err)
		}
		data, _ := json.Marshal(got)
		if strings.Contains(string(data), "auth_version") || strings.Contains(string(data), "newhash") {
			t.Fatal("private user fields leaked")
		}
	}
	users, err := db.ListUsers()
	if err != nil || users[0].AuthVersion != 1 {
		t.Fatal("list version", err)
	}
	if _, err := db.AddAPIKey(user.ID, "test", "keyhash", "prefix"); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetUserByAPIKeyHash("keyhash")
	if err != nil || got.AuthVersion != 1 {
		t.Fatal("API-key lookup version", err)
	}
	if err := db.AdminUpdateUser(user.ID, "legacy", "", "", "", "admin", ""); err != nil {
		t.Fatal(err)
	}
	got, _ = db.GetUserByID(user.ID)
	if got.AuthVersion != 1 {
		t.Fatal("profile-only update revoked session")
	}
}

func TestSecurityInitialAdminAtomicAcrossConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setup.db")
	first, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var wg sync.WaitGroup
	var successes atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db := first
			if i%2 == 0 {
				db = second
			}
			_, err := db.CreateInitialAdmin(fmt.Sprintf("admin%d", i), fmt.Sprintf("admin%d@example.test", i), "hash")
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, ErrSetupCompleted) {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()
	n, err := first.GetUserCount()
	if err != nil || n != 1 || successes.Load() != 1 {
		t.Fatalf("count=%d successes=%d err=%v", n, successes.Load(), err)
	}
}

func TestSecurityConcurrentFactorEnrollmentAndReset(t *testing.T) {
	db, user, _ := passkeyTestDB(t)
	var wg sync.WaitGroup
	var successes atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := db.EnableUserTwoFactor(user.ID, fmt.Sprintf("secret%d", i), fmt.Sprintf("recovery%d", i))
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, ErrFactorAlreadyEnabled) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatal("multiple factors enrolled")
	}
	saved, _ := db.GetUserByID(user.ID)
	if err := db.EnableUserTwoFactor(user.ID, "attacker", ""); !errors.Is(err, ErrFactorAlreadyEnabled) {
		t.Fatal("replacement allowed", err)
	}
	got, _ := db.GetUserByID(user.ID)
	if got.TwoFactorSecret != saved.TwoFactorSecret || got.RecoveryCodes != saved.RecoveryCodes {
		t.Fatal("factor changed")
	}
	if err := db.CreatePasswordResetToken(user.ID, "reset", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	successes.Store(0)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if db.ResetUserPassword("reset", fmt.Sprintf("hash%d", i)) == nil {
				successes.Add(1)
			}
		}(i)
	}
	wg.Wait()
	got, _ = db.GetUserByID(user.ID)
	if successes.Load() != 1 || got.AuthVersion != 1 {
		t.Fatal("reset replay incremented version", successes.Load(), got.AuthVersion)
	}
	if err := db.CreatePasswordResetToken(user.ID, "expired", time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if db.ResetUserPassword("expired", "bad") == nil {
		t.Fatal("expired credential accepted")
	}
}

func TestSecurityProfileWriteChecksAuthVersion(t *testing.T) {
	db, user, _ := passkeyTestDB(t)
	if err := db.UpdateUserPassword(user.ID, "newhash"); err != nil {
		t.Fatal(err)
	}
	avatar := "/new-avatar.png"
	err := db.UpdateUserProfileForSession(user.ID, &user.AuthVersion, "changed", "First", "Last", "changed@example.test", &avatar)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("stale profile write accepted", err)
	}
	saved, err := db.GetUserByID(user.ID)
	if err != nil || saved.Username != user.Username || saved.Email != user.Email || saved.AvatarURL != user.AvatarURL {
		t.Fatal("stale write mutated profile", saved, err)
	}
	if err := db.UpdateUserProfileForSession(user.ID, &saved.AuthVersion, "changed", "First", "Last", "changed@example.test", &avatar); err != nil {
		t.Fatal(err)
	}
	saved, err = db.GetUserByID(user.ID)
	if err != nil || saved.Username != "changed" || saved.Email != "changed@example.test" || saved.AvatarURL != avatar || saved.AuthVersion != 1 {
		t.Fatal("atomic profile update failed", saved, err)
	}
	if err := db.UpdateUserProfileForSession(user.ID, nil, "changed", "", "", "api@example.test", nil); err != nil {
		t.Fatal("API key mutation rejected", err)
	}
	saved, err = db.GetUserByID(user.ID)
	if err != nil || saved.AvatarURL != avatar {
		t.Fatal("omitted avatar changed", saved, err)
	}
}

func TestSecurityConcurrentRecoveryConsumption(t *testing.T) {
	for _, sameCode := range []bool{true, false} {
		t.Run(fmt.Sprintf("same-code=%v", sameCode), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "recovery.db")
			first, err := InitDB(path)
			if err != nil {
				t.Fatal(err)
			}
			defer first.Close()
			second, err := InitDB(path)
			if err != nil {
				t.Fatal(err)
			}
			defer second.Close()
			user, err := first.CreateUser("alice", "hash", "user")
			if err != nil {
				t.Fatal(err)
			}
			if err := first.EnableUserTwoFactor(user.ID, "secret", "aaaa-bbbb,cccc-dddd"); err != nil {
				t.Fatal(err)
			}
			// Both requests have verified their codes against the same snapshot
			// before either write is allowed to begin, independent of scheduling.
			one, err := first.GetUserByID(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			two, err := second.GetUserByID(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			remaining := []string{"cccc-dddd", "aaaa-bbbb"}
			if sameCode {
				remaining[1] = remaining[0]
			}
			type result struct {
				index int
				err   error
			}
			results := make(chan result, 2)
			start := make(chan struct{})
			for i, db := range []*DB{first, second} {
				snapshot := []*User{one, two}[i]
				go func(i int, db *DB, snapshot *User) {
					<-start
					results <- result{i, db.ConsumeUserRecoveryCode(snapshot.ID, snapshot.AuthVersion, snapshot.TwoFactorSecret, snapshot.RecoveryCodes, remaining[i])}
				}(i, db, snapshot)
			}
			close(start)
			winner, successes := -1, 0
			for i := 0; i < 2; i++ {
				result := <-results
				if result.err == nil {
					winner = result.index
					successes++
				} else if !errors.Is(result.err, sql.ErrNoRows) {
					t.Fatal(result.err)
				}
			}
			if successes != 1 {
				t.Fatalf("successful stale-snapshot consumptions: %d", successes)
			}
			saved, err := first.GetUserByID(user.ID)
			if err != nil || saved.RecoveryCodes != remaining[winner] || !saved.TwoFactorEnabled || saved.TwoFactorSecret != "secret" {
				t.Fatal("consumed code restored or factor changed", saved, err)
			}
			if err := first.ConsumeUserRecoveryCode(saved.ID, saved.AuthVersion, saved.TwoFactorSecret, saved.RecoveryCodes, ""); err != nil {
				t.Fatal(err)
			}
			if err := second.ConsumeUserRecoveryCode(two.ID, two.AuthVersion, two.TwoFactorSecret, two.RecoveryCodes, remaining[1]); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("old snapshot restored consumed codes", err)
			}
			saved, err = first.GetUserByID(user.ID)
			if err != nil || saved.RecoveryCodes != "" {
				t.Fatal("codes restored", saved, err)
			}
		})
	}
}

func TestSecurityRecoveryConsumptionRejectsChangedCredentials(t *testing.T) {
	for _, change := range []string{"version", "disabled", "secret", "codes", "error"} {
		t.Run(change, func(t *testing.T) {
			db, user, _ := passkeyTestDB(t)
			if err := db.EnableUserTwoFactor(user.ID, "secret", "aaaa-bbbb,cccc-dddd"); err != nil {
				t.Fatal(err)
			}
			snapshot, err := db.GetUserByID(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "version":
				err = db.UpdateUserPassword(user.ID, "newhash")
			case "disabled":
				err = db.UpdateUserTwoFactor(user.ID, false, "secret", snapshot.RecoveryCodes)
			case "secret":
				err = db.UpdateUserTwoFactor(user.ID, true, "replacement", snapshot.RecoveryCodes)
			case "codes":
				err = db.UpdateUserTwoFactor(user.ID, true, "secret", "new-codes")
			case "error":
				_, err = db.conn.Exec(`CREATE TRIGGER reject_recovery BEFORE UPDATE OF recovery_codes ON users BEGIN SELECT RAISE(ABORT, 'write failed'); END`)
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := db.GetUserByID(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			err = db.ConsumeUserRecoveryCode(snapshot.ID, snapshot.AuthVersion, snapshot.TwoFactorSecret, snapshot.RecoveryCodes, "cccc-dddd")
			if err == nil || (change != "error" && !errors.Is(err, sql.ErrNoRows)) {
				t.Fatal("changed credentials accepted", err)
			}
			after, err := db.GetUserByID(user.ID)
			if err != nil || after.RecoveryCodes != before.RecoveryCodes || after.TwoFactorEnabled != before.TwoFactorEnabled || after.TwoFactorSecret != before.TwoFactorSecret || after.AuthVersion != before.AuthVersion {
				t.Fatal("failed consumption mutated credentials", after, err)
			}
		})
	}
}
