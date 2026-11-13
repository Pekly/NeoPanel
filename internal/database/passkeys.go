package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

const MaxPasskeysPerUser = 10

var ErrPasskeyUnavailable = errors.New("passkey unavailable")

// Passkey contains only public credential material. CredentialJSON is the library's
// complete credential record (COSE public key, counter, transports and backup flags).
// Never expose credential IDs, handles or public key material in account list APIs.
type Passkey struct {
	ID             string  `json:"id"`
	Label          string  `json:"label"`
	CreatedAt      string  `json:"created_at"`
	LastUsedAt     *string `json:"last_used_at"`
	UserID         int     `json:"-"`
	CredentialID   []byte  `json:"-"`
	CredentialJSON string  `json:"-"`
	RPID           string  `json:"-"`
}

func (db *DB) MigratePasskeys() error {
	_, err := db.conn.Exec(`
 CREATE TABLE IF NOT EXISTS webauthn_users (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  handle BLOB NOT NULL UNIQUE
 );
 CREATE TABLE IF NOT EXISTS passkeys (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  credential_id BLOB NOT NULL UNIQUE,
  credential_json TEXT NOT NULL,
  rp_id TEXT NOT NULL,
  label TEXT NOT NULL,
  created_at TEXT NOT NULL,
  last_used_at TEXT
 );
 CREATE INDEX IF NOT EXISTS idx_passkeys_user ON passkeys(user_id);
 CREATE TABLE IF NOT EXISTS webauthn_rp (singleton INTEGER PRIMARY KEY CHECK(singleton=1), rp_id TEXT NOT NULL);
 -- Existing SQLite installations do not enable foreign_keys on every connection.
 CREATE TRIGGER IF NOT EXISTS passkeys_user_deleted AFTER DELETE ON users BEGIN
  DELETE FROM passkeys WHERE user_id=OLD.id;
  DELETE FROM webauthn_users WHERE user_id=OLD.id;
 END;
 `)
	return err
}

// PasskeyRPID is pinned only by a successful, password-authorized registration,
// never by an unauthenticated request's Host header.
func (db *DB) PasskeyRPID() (string, error) {
	var rp string
	err := db.conn.QueryRow(`SELECT rp_id FROM webauthn_rp WHERE singleton=1`).Scan(&rp)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return rp, err
}

func (db *DB) GetOrCreateWebAuthnHandle(userID int) ([]byte, error) {
	handle := make([]byte, 32)
	if _, err := rand.Read(handle); err != nil {
		return nil, err
	}
	_, err := db.conn.Exec(`INSERT INTO webauthn_users(user_id,handle) SELECT id,? FROM users WHERE id=? ON CONFLICT(user_id) DO NOTHING`, handle, userID)
	if err != nil {
		return nil, err
	}
	err = db.conn.QueryRow(`SELECT handle FROM webauthn_users JOIN users ON users.id=webauthn_users.user_id WHERE user_id=?`, userID).Scan(&handle)
	return handle, err
}

func (db *DB) PasskeyUserByHandle(handle []byte) (int, error) {
	var id int
	err := db.conn.QueryRow(`SELECT user_id FROM webauthn_users JOIN users ON users.id=webauthn_users.user_id WHERE handle=?`, handle).Scan(&id)
	return id, err
}

func (db *DB) ListPasskeys(userID int) ([]Passkey, error) {
	rows, err := db.conn.Query(`SELECT p.id,p.user_id,p.credential_id,p.credential_json,p.rp_id,p.label,p.created_at,p.last_used_at FROM passkeys p JOIN users u ON u.id=p.user_id WHERE p.user_id=? ORDER BY p.created_at,p.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := make([]Passkey, 0)
	for rows.Next() {
		var p Passkey
		if err := rows.Scan(&p.ID, &p.UserID, &p.CredentialID, &p.CredentialJSON, &p.RPID, &p.Label, &p.CreatedAt, &p.LastUsedAt); err != nil {
			return nil, err
		}
		keys = append(keys, p)
	}
	return keys, rows.Err()
}

// AddPasskey atomically checks ownership proof, the per-account limit and RP pin.
// Global credential_id uniqueness prevents reassigning a credential to another user.
func (db *DB) AddPasskey(userID int, passwordHash, rpID, label string, credentialID []byte, credentialJSON string) (*Passkey, error) {
	if len(credentialID) == 0 || len(credentialID) > 1024 || rpID == "" || len(label) > 80 {
		return nil, ErrPasskeyUnavailable
	}
	id := make([]byte, 24)
	if _, err := rand.Read(id); err != nil {
		return nil, err
	}
	p := &Passkey{ID: base64.RawURLEncoding.EncodeToString(id), UserID: userID, CredentialID: credentialID, CredentialJSON: credentialJSON, RPID: rpID, Label: label, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM users WHERE id=? AND password_hash=?`, userID, passwordHash).Scan(&count); err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, ErrPasskeyUnavailable
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM passkeys WHERE user_id=?`, userID).Scan(&count); err != nil {
		return nil, err
	}
	if count >= MaxPasskeysPerUser {
		return nil, ErrPasskeyUnavailable
	}
	if _, err := tx.Exec(`INSERT INTO webauthn_rp(singleton,rp_id) VALUES(1,?) ON CONFLICT(singleton) DO NOTHING`, rpID); err != nil {
		return nil, err
	}
	var storedRP string
	if err := tx.QueryRow(`SELECT rp_id FROM webauthn_rp WHERE singleton=1`).Scan(&storedRP); err != nil {
		return nil, err
	}
	if storedRP != rpID {
		return nil, ErrPasskeyUnavailable
	}
	if _, err := tx.Exec(`INSERT INTO passkeys(id,user_id,credential_id,credential_json,rp_id,label,created_at) VALUES(?,?,?,?,?,?,?)`, p.ID, userID, credentialID, credentialJSON, rpID, label, p.CreatedAt); err != nil {
		return nil, ErrPasskeyUnavailable
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}

// UpdatePasskey uses compare-and-swap so concurrent assertions cannot roll back
// the signature counter or resurrect a credential removed during verification.
func (db *DB) UpdatePasskey(p Passkey, credentialJSON string) error {
	result, err := db.conn.Exec(`UPDATE passkeys SET credential_json=?,last_used_at=? WHERE id=? AND user_id=? AND credential_json=? AND EXISTS(SELECT 1 FROM users WHERE id=?)`, credentialJSON, time.Now().UTC().Format(time.RFC3339Nano), p.ID, p.UserID, p.CredentialJSON, p.UserID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrPasskeyUnavailable
	}
	return nil
}

func (db *DB) DeletePasskey(userID int, id, passwordHash string) error {
	result, err := db.conn.Exec(`DELETE FROM passkeys WHERE id=? AND user_id=? AND EXISTS(SELECT 1 FROM users WHERE id=? AND password_hash=?)`, id, userID, userID, passwordHash)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrPasskeyUnavailable
	}
	return nil
}
