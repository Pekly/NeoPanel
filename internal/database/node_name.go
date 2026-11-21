package database

import (
	"database/sql"
	"time"
)

// RenameNode changes only display metadata, without rewriting connection details
// or policy from a potentially stale node snapshot. A missing ID is not created.
func (db *DB) RenameNode(id, name string) error {
	result, err := db.conn.Exec(`UPDATE nodes SET name = ?, updated_at = ? WHERE id = ?`, name, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
