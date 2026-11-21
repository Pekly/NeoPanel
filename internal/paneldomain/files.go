package paneldomain

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type pathKind int

const (
	regularPath pathKind = iota
	executablePath
	directoryPath
)

type fileRecord struct {
	data     []byte
	info     os.FileInfo
	uid, gid int
}

func (m *CaddyManager) snapshot(path string) (fileRecord, error) {
	var result fileRecord
	if err := m.safe(path, regularPath); err != nil {
		return result, err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return result, err
	}
	f, err := openNoFollow(path)
	if err != nil {
		return result, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return result, err
	}
	if !info.Mode().IsRegular() || !os.SameFile(before, info) || info.Size() > maxConfigBytes {
		return result, errors.New("unsafe or oversized config file")
	}
	data, err := readBounded(f, maxConfigBytes)
	if err != nil {
		return result, err
	}
	after, err := f.Stat()
	if err != nil {
		return result, err
	}
	if !sameMetadata(info, after) {
		return result, errors.New("file changed while reading")
	}
	uid, gid, _ := ownership(info)
	return fileRecord{data: data, info: info, uid: uid, gid: gid}, nil
}

func sameMetadata(a, b os.FileInfo) bool {
	au, ag, al := ownership(a)
	bu, bg, bl := ownership(b)
	return os.SameFile(a, b) && a.Mode() == b.Mode() && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime()) && au == bu && ag == bg && al == bl
}

func (m *CaddyManager) unchanged(want fileRecord) error {
	got, err := m.read(m.path)
	if err != nil {
		return err
	}
	if !sameMetadata(want.info, got.info) || !bytes.Equal(want.data, got.data) {
		return errors.New("Caddyfile changed concurrently; refusing to overwrite")
	}
	return nil
}

// All staging and backups use O_EXCL-generated names in the verified config directory.
func (m *CaddyManager) stage(pattern string, data []byte, metadata *fileRecord) (string, error) {
	dir := filepath.Dir(m.path)
	if err := m.safe(dir, directoryPath); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	name := f.Name()
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(name)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return "", err
	}
	if metadata != nil {
		if err = preserveMetadata(f, *metadata); err != nil {
			return "", err
		}
	}
	if err = f.Sync(); err != nil {
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = m.safe(name, regularPath); err != nil {
		return "", err
	}
	if err = m.syncDir(dir); err != nil {
		return "", err
	}
	ok = true
	return name, nil
}

func (m *CaddyManager) publish(staged string, expected fileRecord) error {
	if err := m.safe(filepath.Dir(m.path), directoryPath); err != nil {
		return err
	}
	if err := m.safe(staged, regularPath); err != nil {
		return err
	}
	if err := m.unchanged(expected); err != nil {
		return err
	}
	if err := os.Rename(staged, m.path); err != nil {
		return err
	}
	if err := m.syncDir(filepath.Dir(m.path)); err != nil {
		return fmt.Errorf("config published but directory sync failed: %w", err)
	}
	return nil
}

func (m *CaddyManager) pendingAbsent() error {
	_, err := os.Lstat(m.pendingPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("unfinished Caddy provisioning recovery state at %s; administrator must reconcile disk/running configuration and retained backup before removing it", m.pendingPath())
}

func (m *CaddyManager) writePending(data []byte) error {
	if err := m.safe(filepath.Dir(m.path), directoryPath); err != nil {
		return err
	}
	f, err := os.OpenFile(m.pendingPath(), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return m.syncDir(filepath.Dir(m.path))
}

func (m *CaddyManager) clearPending(expected []byte) error {
	state, err := m.read(m.pendingPath())
	if err != nil {
		return err
	}
	if !bytes.Equal(state.data, expected) {
		return errors.New("recovery state changed concurrently; refusing to remove it")
	}
	if err := os.Remove(m.pendingPath()); err != nil {
		return err
	}
	return m.syncDir(filepath.Dir(m.path))
}
