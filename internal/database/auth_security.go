package database

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrSetupCompleted = errors.New("setup has already been completed")
var ErrFactorAlreadyEnabled = errors.New("two-factor authentication is already enabled")

// CreateInitialAdmin performs the empty-database check and insert as one write.
// Password hashing and request decoding must happen before calling this method.
func (db *DB) CreateInitialAdmin(username, email, passwordHash string) (*User, error) {
	res, err := db.conn.Exec(`INSERT INTO users (username, email, password_hash, role, created_at)
 SELECT ?, ?, ?, 'admin', ? WHERE NOT EXISTS (SELECT 1 FROM users)`, strings.TrimSpace(username), strings.TrimSpace(email), passwordHash, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, ErrSetupCompleted
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return db.GetUserByID(int(id))
}

// EnableUserTwoFactor never replaces an enrolled factor, even with a stale user snapshot.
func (db *DB) EnableUserTwoFactor(id int, secret, recoveryCodes string) error {
	res, err := db.conn.Exec(`UPDATE users SET two_factor_enabled = 1, two_factor_secret = ?, recovery_codes = ? WHERE id = ? AND two_factor_enabled = 0`, secret, recoveryCodes, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrFactorAlreadyEnabled
	}
	return nil
}

// ResetUserPassword consumes the recovery credential and revokes sessions atomically.
func (db *DB) ResetUserPassword(token, newHash string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int
	err = tx.QueryRow(`UPDATE password_resets SET used = 1 WHERE token = ? AND used = 0 AND expires_at > ? RETURNING user_id`, strings.TrimSpace(token), time.Now().UTC()).Scan(&userID)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE users SET password_hash = ?, auth_version = auth_version + 1 WHERE id = ?`, newHash, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return sql.ErrNoRows
	}
	if _, err = tx.Exec(`UPDATE password_resets SET used = 1 WHERE user_id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateUserProfileForSession checks revocation at the write, including the optional
// avatar mutation. A nil expectedVersion is reserved for API-key authentication,
// whose credentials are not revoked by password changes.
func (db *DB) UpdateUserProfileForSession(id int, expectedVersion *int64, username, firstName, lastName, email string, avatarURL *string) error {
	query := `UPDATE users SET username = ?, first_name = ?, last_name = ?, email = ?`
	args := []interface{}{strings.TrimSpace(username), strings.TrimSpace(firstName), strings.TrimSpace(lastName), strings.TrimSpace(email)}
	if avatarURL != nil {
		query += `, avatar_url = ?`
		args = append(args, strings.TrimSpace(*avatarURL))
	}
	query += ` WHERE id = ?`
	args = append(args, id)
	if expectedVersion != nil {
		query += ` AND auth_version = ?`
		args = append(args, *expectedVersion)
	}
	res, err := db.conn.Exec(query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

// ConsumeUserRecoveryCode commits a previously verified recovery-code removal
// only if the complete credential snapshot still matches. Never retry a stale
// snapshot: doing so could restore a code consumed by another request.
func (db *DB) ConsumeUserRecoveryCode(id int, authVersion int64, secret, originalCodes, remainingCodes string) error {
	res, err := db.conn.Exec(`UPDATE users SET recovery_codes = ? WHERE id = ? AND auth_version = ? AND two_factor_enabled = 1 AND two_factor_secret = ? AND recovery_codes = ?`, remainingCodes, id, authVersion, secret, originalCodes)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}
