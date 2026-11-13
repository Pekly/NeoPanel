package database

import (
	"errors"
	"time"
)

const PanelPublicURLKey = "panel_public_url"

var ErrPublicURLPasskeys = errors.New("Remove all registered passkeys before changing the panel hostname; existing passkeys cannot move to another hostname")

// SetPanelPublicURL updates the URL and RP pin in one transaction. AddPasskey
// checks this same pin, so a registration racing a hostname change cannot commit
// a credential for the old hostname after the new address has been saved.
func (db *DB) SetPanelPublicURL(origin, rpID string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM passkeys WHERE rp_id <> ?`, rpID).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return ErrPublicURLPasskeys
	}
	if _, err := tx.Exec(`INSERT INTO webauthn_rp(singleton,rp_id) VALUES(1,?) ON CONFLICT(singleton) DO UPDATE SET rp_id=excluded.rp_id`, rpID); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO panel_settings(key,value,updated_at) VALUES(?,?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_at=excluded.updated_at`, PanelPublicURLKey, origin, time.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit()
}
