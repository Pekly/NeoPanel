package database

// retirePanelRoles removes the retired custom panel-role schema atomically.
// Run after invitation lifecycle migration so its provenance backfill cannot
// later make formerly protected accounts eligible for automatic deletion.
func (db *DB) retirePanelRoles() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var assignedTable bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='panel_user_roles')`).Scan(&assignedTable); err != nil {
		return err
	}
	if assignedTable {
		// Assignments used to protect invitation-created accounts from cleanup.
		// Keep those accounts independent without changing any built-in role,
		// credential, ownership, or server-team permission.
		if _, err := tx.Exec(`UPDATE users SET invitation_created=0 WHERE invitation_created=1
 AND id IN (SELECT user_id FROM panel_user_roles)`); err != nil {
			return err
		}
	}
	for _, statement := range []string{
		`DROP TRIGGER IF EXISTS panel_role_delete_guard`,
		`DROP TRIGGER IF EXISTS panel_user_roles_insert_guard`,
		`DROP TRIGGER IF EXISTS panel_user_roles_update_guard`,
		`DROP TRIGGER IF EXISTS panel_user_roles_cleanup`,
		`DROP TABLE IF EXISTS panel_user_roles`,
		`DROP TABLE IF EXISTS panel_roles`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return tx.Commit()
}
