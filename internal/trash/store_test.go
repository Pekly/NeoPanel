package trash

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func fixture(t *testing.T) (*Store, *rooted.Root, string) {
	t.Helper()
	base := t.TempDir()
	tenantDir := filepath.Join(base, "servers", "alpha")
	if err := os.MkdirAll(tenantDir, 0700); err != nil {
		t.Fatal(err)
	}
	r, err := rooted.Open(tenantDir)
	if err != nil {
		t.Fatal(err)
	}
	s := New(filepath.Join(base, "servers-trash"))
	t.Cleanup(func() { s.Close(); r.Close() })
	return s, r, tenantDir
}
func putFile(t *testing.T, r *rooted.Root, name string) {
	t.Helper()
	if err := r.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := r.WriteFile(name, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
}
func one(t *testing.T, s *Store) Item {
	t.Helper()
	items, err := s.List("alpha")
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	return items[0]
}
func TestLifecyclePersistenceExpiry(t *testing.T) {
	s, r, _ := fixture(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s.now = func() time.Time { return now }
	putFile(t, r, "nested/file.txt")
	if err := s.Put("alpha", r, "nested/file.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Lstat("nested/file.txt"); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	item := one(t, s)
	if item.OriginalPath != "nested/file.txt" || item.Name != "file.txt" || item.IsDir || item.Size != 8 || item.ExpiresAt.Sub(item.DeletedAt) != Retention {
		t.Fatal(item)
	}
	s.Close()
	reopened := New(s.dir)
	defer reopened.Close()
	reopened.now = s.now
	if got := one(t, reopened); got.ID != item.ID || !got.DeletedAt.Equal(item.DeletedAt) {
		t.Fatal(got)
	}
	if err := reopened.Restore("alpha", item.ID, r); err != nil {
		t.Fatal(err)
	}
	if data, err := r.ReadFile("nested/file.txt"); err != nil || string(data) != "original" {
		t.Fatal(string(data), err)
	}
	if err := reopened.Put("alpha", r, "nested"); err != nil {
		t.Fatal(err)
	}
	item = one(t, reopened)
	if !item.IsDir {
		t.Fatal(item)
	}
	now = now.Add(Retention)
	if err := reopened.Restore("alpha", item.ID, r); !errors.Is(err, ErrExpired) {
		t.Fatal(err)
	}
	if items, err := reopened.List("alpha"); err != nil || len(items) != 0 {
		t.Fatal(items, err)
	}
	// Startup-style bounded sweep, persisted timestamps, inclusive expiry boundary.
	if err := reopened.Sweep(128); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.dir, serverKey("alpha"), item.ID)); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
func TestRestoreCollisionMissingParentAndPurge(t *testing.T) {
	s, r, _ := fixture(t)
	putFile(t, r, "dir/file")
	if err := s.Put("alpha", r, "dir/file"); err != nil {
		t.Fatal(err)
	}
	item := one(t, s)
	putFile(t, r, "dir/file")
	if err := s.Restore("alpha", item.ID, r); !errors.Is(err, ErrCollision) {
		t.Fatal(err)
	}
	if err := r.RemoveAll("dir"); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore("alpha", item.ID, r); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if one(t, s).ID != item.ID {
		t.Fatal("lost entry on failed restore")
	}
	if err := s.Purge("alpha", item.ID); err != nil {
		t.Fatal(err)
	}
	if items, err := s.List("alpha"); err != nil || len(items) != 0 {
		t.Fatal(items, err)
	}
}
func TestServerIsolationTraversalAndMetadata(t *testing.T) {
	s, r, _ := fixture(t)
	putFile(t, r, "file")
	for _, name := range []string{".", "/", "../file", "dir/../../file", "C:/file"} {
		if err := s.Put("alpha", r, name); err == nil {
			t.Fatal("accepted", name)
		}
	}
	if err := s.Put("alpha", r, "file"); err != nil {
		t.Fatal(err)
	}
	item := one(t, s)
	if items, err := s.List("beta"); err != nil || len(items) != 0 {
		t.Fatal(items, err)
	}
	for _, id := range []string{item.ID, "../" + item.ID, "../../alpha", "", "payload"} {
		if err := s.Restore("beta", id, r); !os.IsNotExist(err) {
			t.Fatal(id, err)
		}
		if err := s.Purge("beta", id); !os.IsNotExist(err) {
			t.Fatal(id, err)
		}
	}
	item.OriginalPath = "../outside"
	data, _ := json.Marshal(item)
	if err := os.WriteFile(filepath.Join(s.dir, serverKey("alpha"), item.ID, "metadata.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore("alpha", item.ID, r); err == nil {
		t.Fatal("trusted metadata traversal")
	}
	if err := s.Purge("alpha", item.ID); err != nil {
		t.Fatal(err)
	}
}
func TestSymlinkAncestorsAndHardlinks(t *testing.T) {
	s, r, dir := fixture(t)
	outside := t.TempDir()
	target := filepath.Join(outside, "target")
	if err := os.WriteFile(target, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(target, filepath.Join(dir, "hard")); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("alpha", r, "hard"); err != nil {
		t.Fatal(err)
	}
	item := one(t, s)
	if err := s.Restore("alpha", item.ID, r); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("alpha", r, "hard"); err != nil {
		t.Fatal(err)
	}
	if err := s.Purge("alpha", one(t, s).ID); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(target); err != nil || string(data) != "outside" {
		t.Fatal(string(data), err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	for _, name := range []string{"link", "link/target", "link/missing"} {
		if err := s.Put("alpha", r, name); err == nil {
			t.Fatal("accepted", name)
		}
	}
	putFile(t, r, "parent/file")
	if err := s.Put("alpha", r, "parent/file"); err != nil {
		t.Fatal(err)
	}
	item = one(t, s)
	if err := r.RemoveAll("parent"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "parent")); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore("alpha", item.ID, r); err == nil {
		t.Fatal("restored through symlink")
	}
	if _, err := os.Stat(filepath.Join(outside, "file")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
func TestMoveNoReplaceAndStorageFailure(t *testing.T) {
	s, r, _ := fixture(t)
	putFile(t, r, "a")
	putFile(t, r, "b")
	if err := moveNoReplace(r, "a", r, "b"); !os.IsExist(err) {
		t.Fatalf("expected collision: %v", err)
	}
	if _, err := r.Lstat("a"); err != nil {
		t.Fatal("source lost", err)
	}
	if err := os.WriteFile(s.dir, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("alpha", r, "a"); err == nil {
		t.Fatal("storage failure not returned")
	}
	if _, err := r.Lstat("a"); err != nil {
		t.Fatal("source lost", err)
	}
}
func TestBoundedSweepAndServerCleanup(t *testing.T) {
	s, r, _ := fixture(t)
	now := time.Now().UTC()
	s.now = func() time.Time { return now }
	for _, name := range []string{"a", "b", "c"} {
		putFile(t, r, name)
		if err := s.Put("alpha", r, name); err != nil {
			t.Fatal(err)
		}
	}
	items, _ := s.List("alpha")
	now = now.Add(Retention)
	for i := 0; i < 20; i++ {
		if err := s.Sweep(1); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range items {
		if _, err := os.Stat(filepath.Join(s.dir, serverKey("alpha"), item.ID)); !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	putFile(t, r, "fresh")
	if err := s.Put("alpha", r, "fresh"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteServer("alpha"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.List("alpha"); err != nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	s.Close()
	if err := s.Sweep(1); err != nil {
		t.Fatal(err)
	}
	if s.scan != nil {
		t.Fatal("closed store reopened sweep handles")
	}
}

func TestDirectoryRestoreAndSymlinkCollision(t *testing.T) {
	s, r, dir := fixture(t)
	putFile(t, r, "tree/nested/file")
	if err := s.Put("alpha", r, "tree"); err != nil {
		t.Fatal(err)
	}
	item := one(t, s)
	if err := s.Restore("alpha", item.ID, r); err != nil {
		t.Fatal(err)
	}
	if data, err := r.ReadFile("tree/nested/file"); err != nil || string(data) != "original" {
		t.Fatal(string(data), err)
	}
	if err := s.Put("alpha", r, "tree"); err != nil {
		t.Fatal(err)
	}
	item = one(t, s)
	if err := os.Symlink(t.TempDir(), filepath.Join(dir, "tree")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := s.Restore("alpha", item.ID, r); !errors.Is(err, ErrCollision) {
		t.Fatal(err)
	}
	if one(t, s).ID != item.ID {
		t.Fatal("collision lost entry")
	}
}

func TestTrashRejectsInvalidUTF8BeforeAnyMutation(t *testing.T) {
	s, root, _ := fixture(t)
	putFile(t, root, "replacement-\ufffd.txt")
	if err := s.Put("alpha", root, "replacement-\xff.txt"); err == nil {
		t.Fatal("invalid UTF-8 path accepted")
	}
	if _, err := os.Stat(s.dir); !os.IsNotExist(err) {
		t.Fatal("invalid path created trash metadata", err)
	}
	if data, err := root.ReadFile("replacement-\ufffd.txt"); err != nil || string(data) != "original" {
		t.Fatal("replacement-character filename was changed", err)
	}
	putFile(t, root, "世界.txt")
	if err := s.Put("alpha", root, "世界.txt"); err != nil {
		t.Fatal("valid Unicode rejected", err)
	}
	if item := one(t, s); item.OriginalPath != "世界.txt" {
		t.Fatal("Unicode path did not round-trip", item)
	}
}
