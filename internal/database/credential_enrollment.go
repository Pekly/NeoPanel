package database

import (
	"database/sql"
	"time"
)

// CredentialEnrollmentAuth holds the credentials captured when the request was
// authenticated. Use the immutable user ID and exactly one of AuthVersion or
// APIKeyHash; enrollment checks that those credentials are still valid.
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

// AddSSHKeyForAuth checks revocation and inserts the key in one database write.
// It returns sql.ErrNoRows for revoked or non-user credentials.
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

// AddAPIKeyForAuth checks revocation and inserts the key hash in one database
// write. It returns sql.ErrNoRows for revoked or non-user credentials. The caller
// must not return the plaintext key if enrollment fails.
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

// EnableUserTwoFactorForAuth checks revocation and the un-enrolled state in the
// same write. Keep the failure classification in that write transaction so a
// revoked credential returns ErrNoRows, not an existing-factor conflict.
func (db *DB) EnableUserTwoFactorForAuth(auth CredentialEnrollmentAuth, secret, recoveryCodes string) error {
	predicate, args, err := auth.predicate()
	if err != nil {
		return err
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	writeArgs := append([]interface{}{secret, recoveryCodes}, args...)
	res, err := tx.Exec(`UPDATE users SET two_factor_enabled = 1, two_factor_secret = ?, recovery_codes = ? WHERE two_factor_enabled = 0 AND `+predicate, writeArgs...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		var id int
		if err := tx.QueryRow(`SELECT id FROM users WHERE `+predicate, args...).Scan(&id); err != nil {
			return err
		}
		return ErrFactorAlreadyEnabled
	}
	return tx.Commit()
}
