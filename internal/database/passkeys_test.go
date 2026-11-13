package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func passkeyTestDB(t *testing.T) (*DB, *User, *User) {
	t.Helper()
	db, err := InitDB(filepath.Join(t.TempDir(), "passkeys.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	a, err := db.CreateUser("alice", "hash-a", "user")
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.CreateUser("bob", "hash-b", "user")
	if err != nil {
		t.Fatal(err)
	}
	return db, a, b
}

func TestPasskeyStorageOwnershipAndPersistence(t *testing.T) {
	db, a, b := passkeyTestDB(t)
	if err := db.MigratePasskeys(); err != nil {
		t.Fatal(err)
	}
	handle, err := db.GetOrCreateWebAuthnHandle(a.ID)
	if err != nil || len(handle) != 32 {
		t.Fatalf("handle: %v", err)
	}
	again, err := db.GetOrCreateWebAuthnHandle(a.ID)
	if err != nil || !bytes.Equal(handle, again) {
		t.Fatal("handle changed")
	}
	other, err := db.GetOrCreateWebAuthnHandle(b.ID)
	if err != nil || bytes.Equal(handle, other) {
		t.Fatal("handle collision")
	}
	if _, err := db.AddPasskey(a.ID, "wrong", "localhost", "key", []byte("credential"), `{"counter":0}`); err == nil {
		t.Fatal("wrong password proof accepted")
	}
	if rp, _ := db.PasskeyRPID(); rp != "" {
		t.Fatal("failed registration pinned RP")
	}
	p, err := db.AddPasskey(a.ID, a.PasswordHash, "localhost", "key", []byte("credential"), `{"counter":0}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.AddPasskey(b.ID, b.PasswordHash, "localhost", "collision", []byte("credential"), `{}`); err == nil {
		t.Fatal("cross-account credential collision accepted")
	}
	if _, err := db.AddPasskey(b.ID, b.PasswordHash, "other.test", "host", []byte("other"), `{}`); err == nil {
		t.Fatal("RP change accepted")
	}
	if err := db.DeletePasskey(b.ID, p.ID, b.PasswordHash); err == nil {
		t.Fatal("other account removal accepted")
	}
	if err := db.DeletePasskey(a.ID, p.ID, "wrong"); err == nil {
		t.Fatal("wrong password removal accepted")
	}
	if err := db.UpdatePasskey(*p, `{"counter":1}`); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdatePasskey(*p, `{"counter":0}`); err == nil {
		t.Fatal("stale update rolled back counter")
	}
	keys, err := db.ListPasskeys(a.ID)
	if err != nil || len(keys) != 1 || keys[0].LastUsedAt == nil {
		t.Fatal("missing stored metadata", err)
	}
	data, _ := json.Marshal(keys)
	for _, secret := range []string{"credential", "counter", "user_id", "rp_id"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("public JSON leaks %s", secret)
		}
	}
	if err := db.DeletePasskey(a.ID, p.ID, a.PasswordHash); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdatePasskey(keys[0], `{}`); err == nil {
		t.Fatal("deleted key resurrected")
	}
}

func TestPasskeyLimitAtomicAndAccountDeletion(t *testing.T) {
	db, a, _ := passkeyTestDB(t)
	handle, err := db.GetOrCreateWebAuthnHandle(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var successes atomic.Int32
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := db.AddPasskey(a.ID, a.PasswordHash, "localhost", "key", []byte(fmt.Sprint(i)), `{}`); err == nil {
				successes.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if successes.Load() != MaxPasskeysPerUser {
		t.Fatalf("accepted %d keys", successes.Load())
	}
	if err := db.DeleteUser(a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.PasskeyUserByHandle(handle); err == nil {
		t.Fatal("deleted account handle accepted")
	}
	if _, err := db.GetOrCreateWebAuthnHandle(a.ID); err == nil {
		t.Fatal("handle made for missing account")
	}
	if _, err := db.AddPasskey(a.ID, a.PasswordHash, "localhost", "key", []byte("new"), `{}`); err == nil {
		t.Fatal("key made for missing account")
	}
	var n int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM passkeys WHERE user_id=?`, a.ID).Scan(&n); err != nil || n != 0 {
		t.Fatal("account deletion did not clean credentials", err)
	}
}
