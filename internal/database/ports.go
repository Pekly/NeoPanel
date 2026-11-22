package database

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// migratePorts preserves every legacy assignment, including pre-existing collisions.
// Constraints apply to new allocations, not by deleting or reassigning old ports.
func (db *DB) migratePorts() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('servers') WHERE name='ports_allowed'`).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		if _, err = tx.Exec(`ALTER TABLE servers ADD COLUMN ports_allowed INTEGER NOT NULL DEFAULT 1`); err != nil {
			return err
		}
		if _, err = tx.Exec(`INSERT OR IGNORE INTO server_ports(server_id,port,protocol,notes,is_primary,created_at) SELECT id,port,'tcp/udp','Primary Game Port',1,created_at FROM servers WHERE port>0;
   UPDATE server_ports SET is_primary=(port=(SELECT port FROM servers WHERE id=server_id)) WHERE EXISTS(SELECT 1 FROM servers WHERE id=server_id);
   UPDATE servers SET ports_allowed=MAX(1,(SELECT COUNT(*) FROM server_ports WHERE server_id=servers.id));`); err != nil {
			return err
		}
	}
	_, err = tx.Exec(`
 CREATE TRIGGER IF NOT EXISTS ports_server_insert BEFORE INSERT ON servers BEGIN
  SELECT CASE WHEN NEW.ports_allowed < 1 OR NEW.ports_allowed > 64512 THEN RAISE(ABORT,'port allowance must be between 1 and 64512') END;
  SELECT CASE WHEN NEW.port>0 AND EXISTS(SELECT 1 FROM servers s WHERE COALESCE(NULLIF(s.node_id,''),'local')=COALESCE(NULLIF(NEW.node_id,''),'local') AND (s.port=NEW.port OR EXISTS(SELECT 1 FROM server_ports p WHERE p.server_id=s.id AND p.port=NEW.port))) THEN RAISE(ABORT,'port already allocated on this node') END;
 END;
 CREATE TRIGGER IF NOT EXISTS ports_server_created AFTER INSERT ON servers WHEN NEW.port>0 BEGIN
  INSERT INTO server_ports(server_id,port,protocol,notes,is_primary,created_at) VALUES(NEW.id,NEW.port,'tcp/udp','Primary Game Port',1,NEW.created_at);
 END;
 CREATE TRIGGER IF NOT EXISTS ports_quota_update BEFORE UPDATE OF ports_allowed ON servers BEGIN
  SELECT CASE WHEN NEW.ports_allowed < 1 OR NEW.ports_allowed > 64512 OR NEW.ports_allowed < (SELECT COUNT(*) FROM server_ports WHERE server_id=NEW.id) THEN RAISE(ABORT,'port allowance cannot be below current usage or outside 1..64512') END;
 END;
 CREATE TRIGGER IF NOT EXISTS ports_allocate BEFORE INSERT ON server_ports BEGIN
  SELECT CASE WHEN NOT EXISTS(SELECT 1 FROM servers WHERE id=NEW.server_id) THEN RAISE(ABORT,'server not found') END;
  SELECT CASE WHEN (SELECT COUNT(*) FROM server_ports WHERE server_id=NEW.server_id) >= (SELECT ports_allowed FROM servers WHERE id=NEW.server_id) THEN RAISE(ABORT,'port allowance reached') END;
  SELECT CASE WHEN EXISTS(SELECT 1 FROM servers s JOIN servers target ON target.id=NEW.server_id WHERE s.id!=NEW.server_id AND COALESCE(NULLIF(s.node_id,''),'local')=COALESCE(NULLIF(target.node_id,''),'local') AND (s.port=NEW.port OR EXISTS(SELECT 1 FROM server_ports p WHERE p.server_id=s.id AND p.port=NEW.port))) THEN RAISE(ABORT,'port already allocated on this node') END;
 END;
 CREATE TRIGGER IF NOT EXISTS ports_primary_update BEFORE UPDATE OF port,node_id ON servers WHEN NEW.port!=OLD.port OR COALESCE(NULLIF(NEW.node_id,''),'local')!=COALESCE(NULLIF(OLD.node_id,''),'local') BEGIN
  SELECT CASE WHEN NEW.port>0 AND NOT EXISTS(SELECT 1 FROM server_ports WHERE server_id=NEW.id AND port=NEW.port) THEN RAISE(ABORT,'primary port must be allocated first') END;
  SELECT CASE WHEN EXISTS(SELECT 1 FROM servers s WHERE s.id!=NEW.id AND COALESCE(NULLIF(s.node_id,''),'local')=COALESCE(NULLIF(NEW.node_id,''),'local') AND ((NEW.port>0 AND s.port=NEW.port) OR EXISTS(SELECT 1 FROM server_ports p WHERE p.server_id=s.id AND p.port=NEW.port) OR EXISTS(SELECT 1 FROM server_ports mine WHERE mine.server_id=NEW.id AND (mine.port=s.port OR EXISTS(SELECT 1 FROM server_ports other WHERE other.server_id=s.id AND other.port=mine.port))))) THEN RAISE(ABORT,'port already allocated on destination node') END;
 END;
 `)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// PortConfigApply runs under the allocation write transaction. It MUST NOT use DB
// methods (SQLite has one connection). The caller restores its file snapshot if
// this callback or the final commit fails, while retaining its lifecycle gate.
type PortConfigApply func(port int) error

func beginPortTx(db *DB, serverID string) (*sql.Tx, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	// Acquire the SQLite writer reservation before reads, even across DB handles.
	res, err := tx.Exec(`UPDATE servers SET id=id WHERE id=?`, serverID)
	if err == nil {
		var n int64
		n, err = res.RowsAffected()
		if n == 0 && err == nil {
			err = sql.ErrNoRows
		}
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}

func (db *DB) SetServerPortAllowance(serverID string, allowed int) error {
	if allowed < 1 || allowed > 64512 {
		return fmt.Errorf("port allowance must be between 1 and 64512")
	}
	res, err := db.conn.Exec(`UPDATE servers SET ports_allowed=?,updated_at=? WHERE id=?`, allowed, time.Now().UTC(), serverID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if n == 0 && err == nil {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) AllocateServerPort(serverID string, port int, protocol, notes string, primary, random bool, apply PortConfigApply) (*ServerPort, error) {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	if protocol == "" {
		protocol = "tcp/udp"
	}
	if protocol != "tcp/udp" && protocol != "tcp" && protocol != "udp" {
		return nil, fmt.Errorf("protocol must be tcp, udp or tcp/udp")
	}
	if !random && (port < 1024 || port > 65535) {
		return nil, fmt.Errorf("port must be between 1024 and 65535")
	}
	tx, err := beginPortTx(db, serverID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if random {
		pool, err := readPortPool(tx)
		if err != nil {
			return nil, err
		}
		rows, err := tx.Query(`SELECT p.port FROM server_ports p JOIN servers s ON s.id=p.server_id WHERE COALESCE(NULLIF(s.node_id,''),'local')=(SELECT COALESCE(NULLIF(node_id,''),'local') FROM servers WHERE id=?) UNION SELECT port FROM servers WHERE COALESCE(NULLIF(node_id,''),'local')=(SELECT COALESCE(NULLIF(node_id,''),'local') FROM servers WHERE id=?)`, serverID, serverID)
		if err != nil {
			return nil, err
		}
		occupied := map[int]bool{}
		for rows.Next() {
			var n int
			if err = rows.Scan(&n); err != nil {
				rows.Close()
				return nil, err
			}
			occupied[n] = true
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		free := make([]int, 0, pool.End-pool.Start+1)
		for n := pool.Start; n <= pool.End; n++ {
			if !occupied[n] {
				free = append(free, n)
			}
		}
		if len(free) == 0 {
			return nil, fmt.Errorf("no free ports in the configured pool")
		}
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(free))))
		if err != nil {
			return nil, err
		}
		port = free[index.Int64()]
	}
	// Backward-compatible idempotent recording of an already-created primary.
	if primary && !random {
		p := &ServerPort{}
		err = tx.QueryRow(`SELECT id,server_id,port,protocol,notes,is_primary,created_at FROM server_ports WHERE server_id=? AND port=? AND is_primary=1`, serverID, port).Scan(&p.ID, &p.ServerID, &p.Port, &p.Protocol, &p.Notes, &p.IsPrimary, &p.CreatedAt)
		if err == nil {
			if apply != nil {
				if err = apply(port); err != nil {
					return nil, err
				}
			}
			return p, tx.Commit()
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if notes == "" {
		notes = "Additional Allocation"
		if primary {
			notes = "Primary Game Port"
		}
	}
	p := &ServerPort{ServerID: serverID, Port: port, Protocol: protocol, Notes: notes, IsPrimary: primary, CreatedAt: time.Now().UTC()}
	res, err := tx.Exec(`INSERT INTO server_ports(server_id,port,protocol,notes,is_primary,created_at) VALUES(?,?,?,?,0,?)`, serverID, port, protocol, notes, p.CreatedAt)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	p.ID = int(id)
	if primary {
		if err = setPrimaryPortTx(tx, serverID, p.ID, port); err != nil {
			return nil, err
		}
		if apply != nil {
			if err = apply(port); err != nil {
				return nil, err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}

func setPrimaryPortTx(tx *sql.Tx, serverID string, portID, port int) error {
	if _, err := tx.Exec(`UPDATE servers SET port=?,updated_at=? WHERE id=?`, port, time.Now().UTC(), serverID); err != nil {
		return err
	}
	_, err := tx.Exec(`UPDATE server_ports SET is_primary=(id=?) WHERE server_id=?`, portID, serverID)
	return err
}

func (db *DB) ChangePrimaryServerPort(serverID string, portID int, apply PortConfigApply) (int, error) {
	tx, err := beginPortTx(db, serverID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var port int
	if err = tx.QueryRow(`SELECT port FROM server_ports WHERE id=? AND server_id=?`, portID, serverID).Scan(&port); err != nil {
		return 0, fmt.Errorf("port allocation not found: %w", err)
	}
	if err = setPrimaryPortTx(tx, serverID, portID, port); err != nil {
		return 0, err
	}
	if apply != nil {
		if err = apply(port); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return port, nil
}
