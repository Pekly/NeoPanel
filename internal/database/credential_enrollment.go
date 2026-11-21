package database

import (
	"database/sql"
	"time"
)

// CredentialEnrollmentAuth identifies the authority that may enroll a human
// credential. Supply exactly one of AuthVersion (JWT) or APIKeyHash (API key).
// API keys are independent of password changes, but must still exist at insert.
// UserID is the authenticated immutable ID, never resolved from a username.
type CredentialEnrollmentAuth struct {
	UserID      int
	AuthVersion *int64
	APIKeyHash  string
}

func (a CredentialEnrollmentAuth) predicate() (string, []interface{}, error) {
	if a.UserID <= 0 {
		return "", nil, sql.ErrNoRows
	}
	if a.AuthVersion != nil && a.APIKeyHash == "" {
		return `id = ? AND auth_version = ?`, []interface{}{a.UserID, *a.AuthVersion}, nil
	}
	if a.AuthVersion == nil && a.APIKeyHash != "" {
		return `id = ? AND EXISTS (SELECT 1 FROM api_keys WHERE user_id = users.id AND key_hash = ?)`, []interface{}{a.UserID, a.APIKeyHash}, nil
	}
	return "", nil, sql.ErrNoRows
}

// AddSSHKeyForAuth checks revocation and inserts in a single database write.
// sql.ErrNoRows means authority was revoked or is not a human credential.
// AddSSHKey remains available for trusted internal callers.
func (db *DB) AddSSHKeyForAuth(auth CredentialEnrollmentAuth, name, fingerprint, publicKey string) (*SSHKey, error) {
	predicate, args, err := auth.predicate()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	args = append([]interface{}{name, fingerprint, publicKey, now}, args...)
	res, err := db.conn.Exec(`INSERT INTO ssh_keys (user_id, name, fingerprint, public_key, created_at)
 SELECT id, ?, ?, ?, ? FROM users WHERE `+predicate, args...)
	if err := affectedRow(res, err); err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &SSHKey{ID: int(id), UserID: auth.UserID, Name: name, Fingerprint: fingerprint, PublicKey: publicKey, CreatedAt: now}, nil
}

// AddAPIKeyForAuth has the same atomic revocation contract as AddSSHKeyForAuth.
// Only the hash is stored; the caller must not return a secret on any error.
// AddAPIKey remains available for trusted internal callers.
func (db *DB) AddAPIKeyForAuth(auth CredentialEnrollmentAuth, name, keyHash, keyPrefix string) (*APIKey, error) {
	predicate, args, err := auth.predicate()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	args = append([]interface{}{name, keyHash, keyPrefix, now}, args...)
	res, err := db.conn.Exec(`INSERT INTO api_keys (user_id, name, key_hash, key_prefix, created_at)
 SELECT id, ?, ?, ?, ? FROM users WHERE `+predicate, args...)
	if err := affectedRow(res, err); err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &APIKey{ID: int(id), UserID: auth.UserID, Name: name, KeyPrefix: keyPrefix, CreatedAt: now}, nil
}
