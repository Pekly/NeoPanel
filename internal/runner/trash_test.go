package runner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/trash"
)

func TestTrashStartupCleanupAndPersistence(t *testing.T) {
	base := filepath.Join(t.TempDir(), "servers")
	m := NewServerManager(base, nil)
	defer m.ShutdownAll()
	dir := m.GetWorkingDir("alpha")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := rooted.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for _, name := range []string{"expired", "fresh"} {
		if err := root.WriteFile(name, []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
		if err := m.Trash().Put("alpha", root, name); err != nil {
			t.Fatal(err)
		}
	}
	items, err := m.Trash().List("alpha")
	if err != nil || len(items) != 2 {
		t.Fatal(items, err)
	}
	m.ShutdownAll()
	var expiredID string
	for _, item := range items {
		if item.Name != "expired" {
			continue
		}
		expiredID = item.ID
		item.DeletedAt = time.Now().Add(-25 * time.Hour)
		// A forged long expires_at must not override the fixed retention duration.
		item.ExpiresAt = time.Now().Add(365 * 24 * time.Hour)
		data, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(base+"-trash", fmt.Sprintf("%x", sha256.Sum256([]byte("alpha"))), item.ID, "metadata.json")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	restarted := NewServerManager(base, nil)
	defer restarted.ShutdownAll()
	got, err := restarted.Trash().List("alpha")
	if err != nil || len(got) != 1 || got[0].Name != "fresh" {
		t.Fatal(got, err)
	}
	if err := restarted.Trash().Restore("alpha", expiredID, root); !os.IsNotExist(err) {
		t.Fatal("startup did not physically purge expired entry", err)
	}
	if got[0].ExpiresAt.Sub(got[0].DeletedAt) != trash.Retention {
		t.Fatal(got[0])
	}
	if err := restarted.Trash().Restore("alpha", got[0].ID, root); err != nil {
		t.Fatal(err)
	}
}
