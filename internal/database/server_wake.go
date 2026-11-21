package database

import (
	"database/sql"
	"fmt"
)

// ServerWakeState is host-local authority, never reconstructed from logs or status.
type ServerWakeState struct {
	JoinToStartEnabled  bool
	StoppedByIdlePolicy bool
}

func (db *DB) migrateServerWake() error {
	_, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS server_wake_state (
 server_id TEXT PRIMARY KEY,
 join_to_start_enabled INTEGER NOT NULL DEFAULT 1 CHECK(join_to_start_enabled IN (0,1)),
 stopped_by_idle_policy INTEGER NOT NULL DEFAULT 0 CHECK(stopped_by_idle_policy IN (0,1))
 );
 CREATE TRIGGER IF NOT EXISTS delete_server_wake_state AFTER DELETE ON servers
 BEGIN DELETE FROM server_wake_state WHERE server_id = OLD.id; END;
 CREATE TRIGGER IF NOT EXISTS relocate_server_wake_state AFTER UPDATE OF node_id ON servers
 WHEN COALESCE(NEW.node_id,'') NOT IN ('','local')
 BEGIN DELETE FROM server_wake_state WHERE server_id = OLD.id; END;`)
	return err
}

func (db *DB) GetServerWakeState(id string) (ServerWakeState, error) {
	var enabled, idle int
	err := db.conn.QueryRow(`SELECT COALESCE(w.join_to_start_enabled,1), COALESCE(w.stopped_by_idle_policy,0)
 FROM servers s LEFT JOIN server_wake_state w ON w.server_id=s.id
 WHERE s.id=? AND COALESCE(s.node_id,'') IN ('','local')`, id).Scan(&enabled, &idle)
	if err != nil {
		return ServerWakeState{}, err
	}
	if (enabled != 0 && enabled != 1) || (idle != 0 && idle != 1) {
		return ServerWakeState{}, fmt.Errorf("invalid server wake state")
	}
	return ServerWakeState{enabled == 1, idle == 1}, nil
}

func (db *DB) setServerWakeField(id, field string, value bool) error {
	result, err := db.conn.Exec(`INSERT INTO server_wake_state(server_id,`+field+`)
 SELECT id,? FROM servers WHERE id=? AND COALESCE(node_id,'') IN ('','local')
 ON CONFLICT(server_id) DO UPDATE SET `+field+`=excluded.`+field, value, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) SetServerWakePreference(id string, enabled bool) error {
	return db.setServerWakeField(id, "join_to_start_enabled", enabled)
}

// SetServerIdleStopped is trusted runtime-only state, not a user-editable setting.
func (db *DB) SetServerIdleStopped(id string, stopped bool) error {
	return db.setServerWakeField(id, "stopped_by_idle_policy", stopped)
}
