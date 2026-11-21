package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxServerResumeIDs = 10000

type serverResumeState struct {
	Version int      `json:"version"`
	Scope   string   `json:"scope"`
	IDs     []string `json:"ids"`
}

func serverResumeScope(baseDir string) (scope, key string, err error) {
	scope, err = filepath.Abs(baseDir)
	if err != nil {
		return "", "", fmt.Errorf("server resume directory: %w", err)
	}
	scope = filepath.Clean(scope)
	if runtime.GOOS == "windows" {
		scope = strings.ToLower(scope)
	}
	return scope, fmt.Sprintf("server_resume:%x", sha256.Sum256([]byte(scope))), nil
}

func canonicalServerResumeIDs(ids []string) ([]string, error) {
	if len(ids) > maxServerResumeIDs {
		return nil, fmt.Errorf("server resume state exceeds %d IDs", maxServerResumeIDs)
	}
	unique := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" || len(id) > 128 || id == "." || id == ".." ||
			!utf8.ValidString(id) || strings.ContainsAny(id, "/\\:") ||
			strings.ContainsFunc(id, unicode.IsControl) {
			return nil, fmt.Errorf("invalid server resume ID %q", id)
		}
		unique[id] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for id := range unique {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

// SaveServerResumeState replaces the one-shot local resume intent for baseDir.
// A nil or empty IDs slice stores an empty snapshot, clearing any older intent.
// Snapshots have no expiry and are scoped to a cleaned absolute directory, not
// a resolved symlink target. IDs are validated before any database mutation.
func (db *DB) SaveServerResumeState(baseDir string, ids []string) error {
	scope, key, err := serverResumeScope(baseDir)
	if err != nil {
		return err
	}
	ids, err = canonicalServerResumeIDs(ids)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(serverResumeState{Version: 1, Scope: scope, IDs: ids})
	if err != nil {
		return fmt.Errorf("encode server resume state: %w", err)
	}
	_, err = db.conn.Exec(`INSERT INTO panel_settings (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, string(payload), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save server resume state: %w", err)
	}
	return nil
}

// TakeServerResumeState validates and consumes a snapshot in one transaction.
// Missing state returns nil, nil. Corrupt state is preserved and returns no IDs.
// IDs are returned only after the delete commits; concurrent handles may receive
// SQLite lock errors, but cannot both successfully consume the same snapshot.
func (db *DB) TakeServerResumeState(baseDir string) ([]string, error) {
	scope, key, err := serverResumeScope(baseDir)
	if err != nil {
		return nil, err
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin server resume take: %w", err)
	}
	defer tx.Rollback()

	var payload string
	err = tx.QueryRow(`SELECT value FROM panel_settings WHERE key = ?`, key).Scan(&payload)
	if err == sql.ErrNoRows {
		return nil, tx.Commit()
	}
	if err != nil {
		return nil, fmt.Errorf("load server resume state: %w", err)
	}
	if !utf8.ValidString(payload) {
		return nil, fmt.Errorf("invalid UTF-8 in server resume state")
	}
	var state serverResumeState
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("decode server resume state: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("unexpected trailing data in server resume state")
	}
	if state.Version != 1 || state.Scope != scope || state.IDs == nil {
		return nil, fmt.Errorf("invalid server resume version, scope, or missing IDs")
	}
	ids, err := canonicalServerResumeIDs(state.IDs)
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(`DELETE FROM panel_settings WHERE key = ?`, key)
	if err != nil {
		return nil, fmt.Errorf("delete server resume state: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("count consumed server resume state: %w", err)
	}
	if count != 1 {
		return nil, fmt.Errorf("server resume state was not consumed")
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit server resume take: %w", err)
	}
	return ids, nil
}
