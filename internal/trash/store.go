// Package trash stores deleted tenant files outside their rooted filesystem.
package trash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

const Retention = 24 * time.Hour

var (
	ErrExpired   = errors.New("trash entry expired")
	ErrCollision = errors.New("restore destination already exists")
)

type Item struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	Name         string    `json:"name"`
	IsDir        bool      `json:"is_dir"`
	Size         int64     `json:"size"`
	DeletedAt    time.Time `json:"deleted_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Store struct {
	mu     sync.Mutex
	closed bool
	dir    string
	now    func() time.Time
	// Sweep retains directory cursors, not goroutines, to bound work and avoid
	// starving newer entries behind a large collection of unexpired entries.
	scanRoot   *rooted.Root
	scan       *os.File
	scanServer *rooted.Root
	entries    *os.File
}

func New(dir string) *Store      { return &Store{dir: dir, now: time.Now} }
func serverKey(id string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(id))) }
func validID(id string) bool {
	b, err := hex.DecodeString(id)
	return err == nil && len(b) == 16 && id == strings.ToLower(id)
}
func (s *Store) open() (*rooted.Root, error) {
	// dir is panel configuration, never a request path. Refuse a replaced root.
	st, err := os.Lstat(s.dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("invalid trash root")
	}
	r, err := rooted.Open(s.dir)
	if err != nil {
		return nil, err
	}
	actual, err := r.Stat(".")
	if err != nil || !os.SameFile(st, actual) {
		r.Close()
		return nil, fmt.Errorf("trash root changed")
	}
	return r, nil
}
func (s *Store) server(id string, create bool) (*rooted.Root, error) {
	if create {
		if err := os.MkdirAll(s.dir, 0700); err != nil {
			return nil, err
		}
	}
	r, err := s.open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	key := serverKey(id)
	if create {
		if err := r.Mkdir(key, 0700); err != nil && !os.IsExist(err) {
			return nil, err
		}
	}
	return pinDir(r, key)
}

// pinDir opens each component independently and verifies its identity. Once
// pinned, renames of ancestors cannot redirect any subsequent operation.
func pinDir(r *rooted.Root, name string) (*rooted.Root, error) {
	if !filepath.IsLocal(name) || strings.ContainsAny(name, ":\x00") {
		return nil, fmt.Errorf("invalid directory")
	}
	cur, err := r.Sub(".")
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(filepath.Clean(name), string(filepath.Separator)) {
		if part == "." {
			continue
		}
		st, e := cur.Lstat(part)
		if e != nil {
			cur.Close()
			return nil, e
		}
		if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			cur.Close()
			return nil, fmt.Errorf("symbolic links are not supported")
		}
		next, e := cur.Sub(part)
		cur.Close()
		if e != nil {
			return nil, e
		}
		actual, e := next.Stat(".")
		if e != nil || !os.SameFile(st, actual) {
			next.Close()
			return nil, fmt.Errorf("directory changed")
		}
		cur = next
	}
	return cur, nil
}
func canonical(r *rooted.Root, name string) (string, error) {
	// JSON replaces invalid UTF-8. Refuse it before moving data so the recorded
	// original path always identifies exactly the file that was deleted.
	if !utf8.ValidString(name) {
		return "", fmt.Errorf("trash paths must be valid UTF-8")
	}
	// Do not let a cleaned traversal in untrusted metadata become a new path.
	for _, p := range strings.Split(strings.ReplaceAll(name, "\\", "/"), "/") {
		if p == ".." {
			return "", fmt.Errorf("parent traversal is not supported")
		}
	}
	return r.Destructive(name)
}

// Put is idempotent for missing files. Metadata is durable before moving the
// payload; a failed move leaves the original untouched and removes metadata.
func (s *Store) Put(serverID string, tenant *rooted.Root, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, err := canonical(tenant, name)
	if err != nil {
		return err
	}
	parent, err := pinDir(tenant, filepath.Dir(name))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer parent.Close()
	leaf := filepath.Base(name)
	st, err := parent.Lstat(leaf)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 || (!st.IsDir() && !st.Mode().IsRegular()) {
		return fmt.Errorf("only files and directories can be trashed")
	}
	r, err := s.server(serverID, true)
	if err != nil {
		return err
	}
	defer r.Close()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return err
	}
	id := hex.EncodeToString(b[:])
	if err := r.Mkdir(id, 0700); err != nil {
		return err
	}
	entry, err := pinDir(r, id)
	if err != nil {
		r.RemoveAll(id)
		return err
	}
	defer entry.Close()
	now := s.now().UTC()
	item := Item{ID: id, OriginalPath: filepath.ToSlash(name), Name: leaf, IsDir: st.IsDir(), Size: st.Size(), DeletedAt: now, ExpiresAt: now.Add(Retention)}
	data, err := json.Marshal(item)
	if err == nil {
		var f *os.File
		f, err = entry.OpenFile("metadata.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			_, err = f.Write(data)
			if err == nil {
				err = f.Sync()
			}
			if e := f.Close(); err == nil {
				err = e
			}
		}
	}
	if err == nil {
		err = moveNoReplace(parent, leaf, entry, "payload")
	}
	if err != nil {
		entry.Close()
		if cleanup := r.RemoveAll(id); cleanup != nil {
			return fmt.Errorf("%w (metadata cleanup: %v)", err, cleanup)
		}
	}
	return err
}
func readItem(entry *rooted.Root, id string) (Item, error) {
	var item Item
	f, err := entry.Open("metadata.json")
	if err != nil {
		return item, err
	}
	data, err := rooted.ReadBounded(f, 64<<10)
	f.Close()
	if err != nil {
		return item, err
	}
	if err := json.Unmarshal(data, &item); err != nil {
		return item, err
	}
	// Metadata is not authority for either storage paths or expiry duration.
	if item.DeletedAt.IsZero() {
		return item, fmt.Errorf("invalid deletion time")
	}
	if !filepath.IsLocal(item.OriginalPath) || strings.ContainsAny(item.OriginalPath, "\\:\x00") {
		return item, fmt.Errorf("invalid original path")
	}
	for _, p := range strings.Split(item.OriginalPath, "/") {
		if p == ".." || p == "." || p == "" {
			return item, fmt.Errorf("invalid original path")
		}
	}
	item.ID = id
	item.Name = filepath.Base(item.OriginalPath)
	item.ExpiresAt = item.DeletedAt.Add(Retention)
	return item, nil
}
func (s *Store) List(serverID string) ([]Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]Item, 0)
	r, err := s.server(serverID, false)
	if os.IsNotExist(err) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	defer r.Close()
	dirs, err := r.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, d := range dirs {
		if !validID(d.Name()) || !d.IsDir() {
			continue
		}
		e, err := pinDir(r, d.Name())
		if err != nil {
			continue
		}
		item, err := readItem(e, d.Name())
		_, payloadErr := e.Lstat("payload")
		e.Close()
		if err == nil && payloadErr == nil && s.now().Before(item.ExpiresAt) {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].DeletedAt.Equal(items[j].DeletedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].DeletedAt.After(items[j].DeletedAt)
	})
	return items, nil
}
func (s *Store) Restore(serverID, id string, tenant *rooted.Root) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !validID(id) {
		return os.ErrNotExist
	}
	r, err := s.server(serverID, false)
	if err != nil {
		return err
	}
	defer r.Close()
	e, err := pinDir(r, id)
	if err != nil {
		return err
	}
	defer e.Close()
	item, err := readItem(e, id)
	if err != nil {
		return err
	}
	if !s.now().Before(item.ExpiresAt) {
		return ErrExpired
	}
	// readItem validated every component lexically. Check ancestors via pinned
	// handles, and treat any existing leaf (including a symlink) as a collision.
	name := filepath.FromSlash(item.OriginalPath)
	parent, err := pinDir(tenant, filepath.Dir(name))
	if err != nil {
		return err
	}
	defer parent.Close()
	leaf := filepath.Base(name)
	if _, err := parent.Lstat(leaf); err == nil {
		return ErrCollision
	} else if !os.IsNotExist(err) {
		return err
	}
	st, err := e.Lstat("payload")
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 || (!st.IsDir() && !st.Mode().IsRegular()) {
		return fmt.Errorf("invalid trash payload")
	}
	if err := moveNoReplace(e, "payload", parent, leaf); err != nil {
		if os.IsExist(err) {
			return ErrCollision
		}
		return err
	}
	// Publication succeeded. Leftover metadata cannot be restored a second time
	// and will be removed by expiry cleanup if removal fails.
	e.Close()
	_ = r.RemoveAll(id)
	return nil
}
func (s *Store) Purge(serverID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !validID(id) {
		return os.ErrNotExist
	}
	r, err := s.server(serverID, false)
	if err != nil {
		return err
	}
	defer r.Close()
	if _, err := r.Lstat(id); err != nil {
		return err
	}
	return r.RemoveAll(id)
}
func (s *Store) DeleteServer(serverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.open()
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer r.Close()
	return r.RemoveAll(serverKey(serverID))
}
func (s *Store) closeScan() {
	if s.entries != nil {
		s.entries.Close()
		s.entries = nil
	}
	if s.scanServer != nil {
		s.scanServer.Close()
		s.scanServer = nil
	}
	if s.scan != nil {
		s.scan.Close()
		s.scan = nil
	}
	if s.scanRoot != nil {
		s.scanRoot.Close()
		s.scanRoot = nil
	}
}
func (s *Store) Close() { s.mu.Lock(); defer s.mu.Unlock(); s.closed = true; s.closeScan() }

// Sweep visits at most limit directory entries. Recursive deletion of one expired
// payload is not byte/time bounded. Corrupt metadata is retained for operator recovery.
func (s *Store) Sweep(limit int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	var err error
	if s.scan == nil {
		s.scanRoot, err = s.open()
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		s.scan, err = s.scanRoot.Root.Open(".")
		if err != nil {
			s.closeScan()
			return err
		}
	}
	for i := 0; i < limit; i++ {
		if s.entries == nil {
			names, e := s.scan.Readdirnames(1)
			if e == io.EOF {
				s.closeScan()
				return nil
			}
			if e != nil {
				s.closeScan()
				return e
			}
			key := names[0]
			if len(key) != 64 {
				continue
			}
			if _, e := hex.DecodeString(key); e != nil {
				continue
			}
			s.scanServer, e = pinDir(s.scanRoot, key)
			if e != nil {
				continue
			}
			s.entries, e = s.scanServer.Root.Open(".")
			if e != nil {
				s.scanServer.Close()
				s.scanServer = nil
				continue
			}
			continue
		}
		names, e := s.entries.Readdirnames(1)
		if e != nil {
			s.entries.Close()
			s.entries = nil
			s.scanServer.Close()
			s.scanServer = nil
			if e != io.EOF {
				return e
			}
			continue
		}
		id := names[0]
		if !validID(id) {
			continue
		}
		entry, e := pinDir(s.scanServer, id)
		if e != nil {
			continue
		}
		item, e := readItem(entry, id)
		entry.Close()
		if e == nil && !s.now().Before(item.ExpiresAt) {
			if e := s.scanServer.RemoveAll(id); e != nil {
				return e
			}
		}
	}
	return nil
}
