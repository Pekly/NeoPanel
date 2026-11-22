package proxyconfig

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

const maxConfig = 1 << 20
const maxSecret = 4096

var operations sync.Mutex
var errIO = errors.New("Cannot access safe regular proxy configuration file")

// Pin the filesystem root first, so even creation of a missing server directory
// follows the same no-symlink policy as config and secret access.
func openDir(dir string, create bool) (*rooted.Root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, errIO
	}
	volume := filepath.VolumeName(abs)
	base := volume + string(filepath.Separator)
	parent, err := rooted.Open(base)
	if err != nil {
		return nil, errIO
	}
	defer parent.Close()
	rel, err := filepath.Rel(base, abs)
	if err != nil {
		return nil, errIO
	}
	if err = parent.Check(rel); err != nil {
		return nil, errIO
	}
	if create {
		if err = parent.MkdirAll(rel, 0755); err != nil {
			return nil, errIO
		}
		if err = parent.Check(rel); err != nil {
			return nil, errIO
		}
	}
	root, err := parent.Sub(rel)
	if os.IsNotExist(err) {
		return nil, ErrMissing
	}
	if err != nil {
		return nil, errIO
	}
	return root, nil
}
func readFile(root *rooted.Root, name string, limit int) ([]byte, os.FileInfo, error) {
	if err := root.Check(name); err != nil {
		return nil, nil, errIO
	}
	before, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return nil, nil, ErrMissing
	}
	if err != nil || !before.Mode().IsRegular() || before.Size() > int64(limit) {
		return nil, nil, errIO
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, nil, errIO
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !os.SameFile(before, st) || st.Size() > int64(limit) {
		return nil, nil, errIO
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil || len(b) > limit {
		return nil, nil, errIO
	}
	if err = root.Check(name); err != nil {
		return nil, nil, errIO
	}
	after, err := root.Lstat(name)
	if err != nil || !os.SameFile(st, after) {
		return nil, nil, ErrConflict
	}
	return b, st, nil
}
func revision(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
func arguments(kind string, port int) error {
	if !Supported(kind) {
		return errors.New("Unsupported proxy type")
	}
	if !validPort(port) {
		return errors.New("Invalid assigned port")
	}
	return nil
}
func snapshot(kind string, raw []byte, m object) (Snapshot, error) {
	s, err := settings(kind, m)
	if err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{Type: kind, File: FileName(kind), Exists: true, Revision: revision(raw), Config: s}
	if err := Validate(kind, s); err != nil {
		result.ValidationError = err.Error()
	}
	return result, nil
}
func Read(dir, kind string, port int) (Snapshot, error) {
	operations.Lock()
	defer operations.Unlock()
	kind = strings.ToLower(kind)
	if err := arguments(kind, port); err != nil {
		return Snapshot{}, err
	}
	missing := Snapshot{Type: kind, File: FileName(kind), Revision: "missing", Config: defaults(kind)}
	root, err := openDir(dir, false)
	if errors.Is(err, ErrMissing) {
		return missing, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	defer root.Close()
	raw, _, err := readFile(root, FileName(kind), maxConfig)
	if errors.Is(err, ErrMissing) {
		return missing, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	m, err := parse(kind, raw)
	if err != nil {
		return Snapshot{}, err
	}
	return snapshot(kind, raw, m)
}

// Ensure never repairs or rewrites an existing config, even if malformed. For
// existing Velocity configs it validates an existing configured secret without
// creating one. Save creates a missing secret when forwarding requires it.
func Ensure(dir, kind string, port int) (bool, error) {
	operations.Lock()
	defer operations.Unlock()
	kind = strings.ToLower(kind)
	if err := arguments(kind, port); err != nil {
		return false, err
	}
	root, err := openDir(dir, true)
	if err != nil {
		return false, err
	}
	defer root.Close()
	raw, _, err := readFile(root, FileName(kind), maxConfig)
	if err == nil {
		if kind == "velocity" {
			if m, e := parse(kind, raw); e == nil {
				name, e := secretPath(m)
				if e != nil {
					return false, e
				}
				if _, e = readSecret(root, name); e != nil && !errors.Is(e, ErrMissing) {
					return false, e
				}
			}
		}
		return false, nil
	}
	if !errors.Is(err, ErrMissing) {
		return false, err
	}
	m := starter(kind, port)
	b, err := encode(kind, m)
	if err != nil {
		return false, err
	}
	if kind == "velocity" {
		if err = ensureSecret(root, m); err != nil {
			return false, err
		}
	}
	if err = createExclusive(root, FileName(kind), b, 0644); err != nil {
		return false, err
	}
	return true, nil
}
func createExclusive(root *rooted.Root, name string, b []byte, mode os.FileMode) error {
	if err := root.Check(name); err != nil {
		return errIO
	}
	f, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if os.IsExist(err) {
		return ErrConflict
	}
	if err != nil {
		return errIO
	}
	_, err = f.Write(b)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return errIO
	}
	return nil
}
func publish(root *rooted.Root, name string, old []byte, identity os.FileInfo, b []byte) error {
	// Publish creates a private fresh inode: hardlinks can be read, never truncated.
	// Recheck identity and revision immediately before rename, including after the
	// temporary file has been written. External writers must still coordinate.
	err := root.Publish(name, identity.Mode().Perm(), func(f *os.File) error {
		if _, err := f.Write(b); err != nil {
			return errIO
		}
		if err := f.Sync(); err != nil {
			return errIO
		}
		current, st, err := readFile(root, name, maxConfig)
		if err != nil {
			return err
		}
		if !os.SameFile(identity, st) || revision(current) != revision(old) {
			return ErrConflict
		}
		return nil
	})
	if errors.Is(err, ErrConflict) || errors.Is(err, ErrMissing) {
		return err
	}
	if err != nil {
		return errIO
	}
	return nil
}
func Save(dir, kind string, port int, expected string, s Settings) (Snapshot, error) {
	operations.Lock()
	defer operations.Unlock()
	kind = strings.ToLower(kind)
	if err := arguments(kind, port); err != nil {
		return Snapshot{}, err
	}
	if expected == "" {
		return Snapshot{}, errors.New("Expected revision is required")
	}
	if err := Validate(kind, s); err != nil {
		return Snapshot{}, err
	}
	root, err := openDir(dir, false)
	if err != nil {
		return Snapshot{}, err
	}
	defer root.Close()
	raw, st, err := readFile(root, FileName(kind), maxConfig)
	if err != nil {
		return Snapshot{}, err
	}
	if revision(raw) != expected {
		return Snapshot{}, ErrConflict
	}
	m, err := parse(kind, raw)
	if err != nil {
		return Snapshot{}, err
	}
	if _, err = settings(kind, m); err != nil {
		return Snapshot{}, err
	}
	apply(kind, m, s)
	if err = setPort(kind, m, port); err != nil {
		return Snapshot{}, err
	}
	b, err := encode(kind, m)
	if err != nil {
		return Snapshot{}, err
	}
	// Check the resulting DTO before any publication or secret creation.
	result, err := snapshot(kind, b, m)
	if err != nil {
		return Snapshot{}, err
	}
	if kind == "velocity" {
		name, e := secretPath(m)
		if e != nil {
			return Snapshot{}, e
		}
		if _, e = readSecret(root, name); e != nil && !errors.Is(e, ErrMissing) {
			return Snapshot{}, e
		}
		if s.ForwardingMode == "MODERN" || s.ForwardingMode == "BUNGEEGUARD" {
			if err = ensureSecret(root, m); err != nil {
				return Snapshot{}, err
			}
		}
	}
	if err = publish(root, FileName(kind), raw, st, b); err != nil {
		return Snapshot{}, err
	}
	return result, nil
}
func SyncPort(dir, kind string, port int) error {
	operations.Lock()
	defer operations.Unlock()
	kind = strings.ToLower(kind)
	if err := arguments(kind, port); err != nil {
		return err
	}
	root, err := openDir(dir, false)
	if err != nil {
		return err
	}
	defer root.Close()
	raw, st, err := readFile(root, FileName(kind), maxConfig)
	if err != nil {
		return err
	}
	m, err := parse(kind, raw)
	if err != nil {
		return err
	}
	before, err := encode(kind, m)
	if err != nil {
		return err
	}
	if err = setPort(kind, m, port); err != nil {
		return err
	}
	b, err := encode(kind, m)
	if err != nil {
		return err
	}
	if bytes.Equal(before, b) {
		return nil
	}
	return publish(root, FileName(kind), raw, st, b)
}
func secretPath(m object) (string, error) {
	name := "forwarding.secret"
	if v, ok := m["forwarding-secret-file"]; ok {
		var valid bool
		name, valid = v.(string)
		if !valid {
			return "", errSecret
		}
	}
	// Accept portable confined relative paths, not URLs, drive names, ADS, or
	// traversal. Reject backslashes rather than interpreting them differently by OS.
	if name == "" || len(name) > 1024 || !utf8.ValidString(name) || !filepath.IsLocal(name) || strings.ContainsAny(name, "\\:") {
		return "", errSecret
	}
	for _, c := range name {
		if unicode.IsControl(c) {
			return "", errSecret
		}
	}
	for _, p := range strings.Split(name, "/") {
		if p == "" || p == "." || p == ".." || strings.HasSuffix(p, ".") || strings.HasSuffix(p, " ") {
			return "", errSecret
		}
	}
	return filepath.FromSlash(name), nil
}
func readSecret(root *rooted.Root, name string) (string, error) {
	b, _, err := readFile(root, name, maxSecret)
	if errors.Is(err, ErrMissing) {
		return "", ErrMissing
	}
	if err != nil {
		return "", errSecret
	}
	if !utf8.Valid(b) {
		return "", errSecret
	}
	s := strings.TrimRight(string(b), "\r\n")
	if strings.TrimSpace(s) == "" {
		return "", errSecret
	}
	for _, c := range s {
		if unicode.IsControl(c) {
			return "", errSecret
		}
	}
	return s, nil
}
func ensureSecret(root *rooted.Root, m object) error {
	name, err := secretPath(m)
	if err != nil {
		return err
	}
	if _, err = readSecret(root, name); err == nil {
		return nil
	}
	if !errors.Is(err, ErrMissing) {
		return err
	}
	if err = root.Check(name); err != nil {
		return errSecret
	}
	parent := filepath.Dir(name)
	if err = root.MkdirAll(parent, 0700); err != nil {
		return errSecret
	}
	if err = root.Check(name); err != nil {
		return errSecret
	}
	var random [32]byte
	if _, err = rand.Read(random[:]); err != nil {
		return errSecret
	}
	if err = createExclusive(root, name, []byte(hex.EncodeToString(random[:])), 0600); err != nil {
		return errSecret
	}
	return nil
}

// ReadSecret is an explicit reveal operation. Permission checks belong to the
// caller; ordinary Read/Snapshot never read or include secret contents.
func ReadSecret(dir, kind string) (string, error) {
	operations.Lock()
	defer operations.Unlock()
	kind = strings.ToLower(kind)
	if kind != "velocity" {
		return "", errors.New("Secret reveal is only supported for Velocity")
	}
	root, err := openDir(dir, false)
	if err != nil {
		return "", err
	}
	defer root.Close()
	raw, _, err := readFile(root, FileName(kind), maxConfig)
	if err != nil {
		return "", err
	}
	m, err := parse(kind, raw)
	if err != nil {
		return "", err
	}
	name, err := secretPath(m)
	if err != nil {
		return "", err
	}
	return readSecret(root, name)
}
