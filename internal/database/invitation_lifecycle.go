package database

import (
	"database/sql"
	"fmt"
)

// migrateInvitationLifecycle deliberately does not infer creation from membership.
func (db *DB) migrateInvitationLifecycle() error {
	for _, column := range []string{"invitation_created", "identity_exported"} {
		_ = db.addColumnIfNotExists("users", column, "BOOLEAN NOT NULL DEFAULT 0")
		if _, err := db.conn.Exec("SELECT " + column + " FROM users LIMIT 0"); err != nil {
			return err
		}
	}
	_, err := db.conn.Exec(`
 CREATE TABLE IF NOT EXISTS deleted_invitation_users (id INTEGER PRIMARY KEY);
 CREATE TABLE IF NOT EXISTS invitation_lifecycle_migrations (id INTEGER PRIMARY KEY);
 UPDATE users SET identity_exported=1 WHERE EXISTS (SELECT 1 FROM nodes WHERE id<>'local')
 AND NOT EXISTS (SELECT 1 FROM invitation_lifecycle_migrations WHERE id=1);
 -- account_id was written ONLY by the account INSERT, never by existing-user acceptance.
 -- Require the original identity and creation window as well; ambiguous legacy users survive.
 UPDATE users SET invitation_created=1 WHERE NOT EXISTS (SELECT 1 FROM invitation_lifecycle_migrations WHERE id=1)
 AND EXISTS (SELECT 1 FROM team_invitations i WHERE i.account_id=users.id
 AND i.username=users.username AND lower(i.email)=lower(users.email)
 AND julianday(substr(users.created_at,1,19))>=julianday(substr(i.created_at,1,19))
 AND julianday(substr(users.created_at,1,19))<=julianday(substr(i.expires_at,1,19)));
 UPDATE users SET identity_exported=1 WHERE EXISTS
 (SELECT 1 FROM team_invitations i WHERE i.node_id<>'local' AND
 (i.account_id=users.id OR (i.username=users.username AND lower(i.email)=lower(users.email)))
 AND i.status IN ('granting','accepted','revoked'));
 INSERT OR IGNORE INTO invitation_lifecycle_migrations(id) VALUES(1);
 CREATE TRIGGER IF NOT EXISTS invitation_user_no_resurrection BEFORE INSERT ON users
 WHEN EXISTS (SELECT 1 FROM deleted_invitation_users WHERE id=NEW.id)
 BEGIN SELECT RAISE(ABORT, 'deleted invitation identity'); END;
 CREATE TRIGGER IF NOT EXISTS invitation_member_insert_guard BEFORE INSERT ON server_subusers
 WHEN NOT EXISTS (SELECT 1 FROM users WHERE id=NEW.user_id)
 OR NOT EXISTS (SELECT 1 FROM servers WHERE id=NEW.server_id)
 BEGIN SELECT RAISE(ABORT, 'missing membership identity or server'); END;
 CREATE TRIGGER IF NOT EXISTS invitation_member_update_guard BEFORE UPDATE OF user_id,server_id ON server_subusers
 WHEN NOT EXISTS (SELECT 1 FROM users WHERE id=NEW.user_id)
 OR NOT EXISTS (SELECT 1 FROM servers WHERE id=NEW.server_id)
 BEGIN SELECT RAISE(ABORT, 'missing membership identity or server'); END;
 CREATE TRIGGER IF NOT EXISTS invitation_owner_insert_guard BEFORE INSERT ON servers
 WHEN EXISTS (SELECT 1 FROM deleted_invitation_users WHERE id=NEW.owner_id)
 BEGIN SELECT RAISE(ABORT, 'deleted invitation owner'); END;
 CREATE TRIGGER IF NOT EXISTS invitation_owner_update_guard BEFORE UPDATE OF owner_id ON servers
 WHEN EXISTS (SELECT 1 FROM deleted_invitation_users WHERE id=NEW.owner_id)
 BEGIN SELECT RAISE(ABORT, 'deleted invitation owner'); END;
 `)
	return err
}

// cleanupInvitationUser runs only while the association-removal write transaction
// is held. Nodes are not a global association ledger: exported identities and any
// installation with configured nodes are intentionally retained.
func cleanupInvitationUser(tx *sql.Tx, id int) error {
	var eligible bool
	err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users u WHERE id=? AND id<>1
 AND invitation_created=1 AND identity_exported=0 AND role='user'
 AND NOT EXISTS (SELECT 1 FROM nodes WHERE id<>'local')
 AND NOT EXISTS (SELECT 1 FROM servers WHERE owner_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM server_subusers WHERE user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM panel_user_roles WHERE user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM invitation_grants WHERE user_id=u.id AND status<>'revoked')
 AND NOT EXISTS (SELECT 1 FROM team_invitations i WHERE
 (i.account_id=u.id OR (lower(i.username)=lower(u.username) OR lower(i.email)=lower(u.email)))
 AND (i.status='granting' OR (i.node_id<>'local' AND i.status='accepted'))))`, id).Scan(&eligible)
	if err != nil || !eligible {
		return err
	}
	if _, err = tx.Exec(`INSERT OR IGNORE INTO deleted_invitation_users(id) VALUES(?)`, id); err != nil {
		return err
	}
	// Do not rely on PRAGMA foreign_keys: older installations leave it disabled.
	for _, table := range []string{"ssh_keys", "api_keys", "password_resets", "passkeys", "webauthn_users"} {
		if _, err = tx.Exec("DELETE FROM "+table+" WHERE user_id=?", id); err != nil {
			return err
		}
	}
	_, err = tx.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (db *DB) deleteInvitationSubuser(server string, membership int) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Acquire the writer before the read, including with multiple DB handles.
	if _, err = tx.Exec(`UPDATE server_subusers SET id=id WHERE id=? AND server_id=?`, membership, server); err != nil {
		return err
	}
	var user int
	if err = tx.QueryRow(`SELECT user_id FROM server_subusers WHERE id=? AND server_id=?`, membership, server).Scan(&user); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM server_subusers WHERE id=? AND server_id=?`, membership, server); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE invitation_grants SET status='revoked' WHERE membership_id=? AND server_id=?`, membership, server); err != nil {
		return err
	}
	if err = cleanupInvitationUser(tx, user); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) deleteInvitationServer(server string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE servers SET id=id WHERE id=?`, server); err != nil {
		return err
	}
	var exists int
	if err = tx.QueryRow(`SELECT 1 FROM servers WHERE id=?`, server).Scan(&exists); err != nil {
		return err
	}
	rows, err := tx.Query(`SELECT user_id FROM server_subusers WHERE server_id=? UNION SELECT account_id FROM team_invitations WHERE server_id=? AND node_id='local' AND account_id<>0 UNION SELECT owner_id FROM servers WHERE id=?`, server, server, server)
	if err != nil {
		return err
	}
	var users []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		users = append(users, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE invitation_grants SET status='revoked' WHERE server_id=?`, server); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE team_invitations SET status='revoked' WHERE server_id=? AND node_id='local' AND status IN ('pending','granting')`, server); err != nil {
		return err
	}
	for _, table := range []string{"server_subusers", "server_schedules", "server_ports", "server_databases"} {
		if _, err = tx.Exec("DELETE FROM "+table+" WHERE server_id=?", server); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(`DELETE FROM servers WHERE id=?`, server); err != nil {
		return err
	}
	for _, id := range users {
		if err = cleanupInvitationUser(tx, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// PinExportedIdentity is durable BEFORE sending credentials over the network.
// A failed/uncertain delivery must never be interpreted as proof of absence.
func (db *DB) PinExportedIdentity(id int) error {
	res, err := db.conn.Exec(`UPDATE users SET identity_exported=1 WHERE id=?`, id)
	return affectedRow(res, err)
}

func (db *DB) RevokedInvitationGrants(server string) ([]string, error) {
	rows, err := db.conn.Query(`SELECT id FROM invitation_grants WHERE server_id=? AND status='revoked'`, server)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Reconcile only durable worker tombstones, never a missing/unreachable identity.
// Keep account_id as provenance; revoked remote users remain pinned for manual
// cluster-wide deprovisioning until there is a distributed association protocol.
func (db *DB) ReconcileInvitationRemovals(node, server string, ids []string) error {
	if node == "" || node == "local" {
		return fmt.Errorf("remote node required")
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range ids {
		if _, err = tx.Exec(`UPDATE team_invitations SET status='revoked' WHERE id=? AND node_id=? AND server_id=? AND status IN ('accepted','granting')`, id, node, server); err != nil {
			return err
		}
	}
	return tx.Commit()
}
