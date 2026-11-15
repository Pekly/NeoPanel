package database

import (
	"database/sql"
	"errors"
	"time"
)

// PrepareDesktopGuest is called only by the native desktop host, before HTTP
// admission. A marker, not a username, identifies its passwordless local owner.
// Existing databases are never adopted. Revocation is permanent, even if a
// password/security method is subsequently removed or the user is deleted.
func (db *DB) PrepareDesktopGuest() (*User, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
 CREATE TABLE IF NOT EXISTS desktop_guest (
  singleton INTEGER PRIMARY KEY CHECK(singleton=1), user_id INTEGER NOT NULL, disabled INTEGER NOT NULL DEFAULT 0
 );
 CREATE TRIGGER IF NOT EXISTS desktop_guest_security AFTER UPDATE OF password_hash, auth_version, two_factor_enabled, two_factor_secret, role ON users
 WHEN OLD.password_hash != NEW.password_hash OR OLD.auth_version != NEW.auth_version OR OLD.two_factor_enabled != NEW.two_factor_enabled OR OLD.two_factor_secret != NEW.two_factor_secret OR OLD.role != NEW.role
 BEGIN UPDATE desktop_guest SET disabled=1 WHERE user_id=OLD.id; END;
 CREATE TRIGGER IF NOT EXISTS desktop_guest_deleted AFTER DELETE ON users
 BEGIN UPDATE desktop_guest SET disabled=1 WHERE user_id=OLD.id; END;
 CREATE TRIGGER IF NOT EXISTS desktop_guest_passkey AFTER INSERT ON passkeys
 BEGIN UPDATE desktop_guest SET disabled=1 WHERE user_id=NEW.user_id; END;
 `)
	if err != nil {
		return nil, err
	}
	var count, marked int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return nil, err
	}
	if err = tx.QueryRow(`SELECT COUNT(*) FROM desktop_guest`).Scan(&marked); err != nil {
		return nil, err
	}
	if count > 0 && marked == 0 {
		if _, err = tx.Exec(`INSERT INTO desktop_guest(singleton,user_id,disabled) VALUES(1,0,1)`); err != nil {
			return nil, err
		}
	}
	if count == 0 && marked == 0 {
		res, e := tx.Exec(`INSERT INTO users (username,password_hash,role,created_at) VALUES ('Guest','!desktop-guest-no-password','admin',?)`, time.Now().UTC())
		if e != nil {
			return nil, e
		}
		id, e := res.LastInsertId()
		if e != nil {
			return nil, e
		}
		if _, err = tx.Exec(`INSERT INTO desktop_guest(singleton,user_id) VALUES(1,?)`, id); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return db.DesktopGuest()
}

// DesktopGuest returns only an unchanged, explicitly marked local guest.
// There is deliberately no HTTP endpoint for acquiring this identity.
func (db *DB) DesktopGuest() (*User, error) {
	var id int
	err := db.conn.QueryRow(`SELECT u.id FROM users u JOIN desktop_guest g ON g.user_id=u.id
 WHERE g.disabled=0 AND u.role='admin' AND u.password_hash='!desktop-guest-no-password'
 AND u.auth_version=0 AND COALESCE(u.two_factor_enabled,0)=0
 AND NOT EXISTS(SELECT 1 FROM passkeys p WHERE p.user_id=u.id)`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return db.GetUserByID(id)
}
