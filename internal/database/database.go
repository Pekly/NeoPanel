package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func InitDB(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	// Connect to SQLite with WAL mode enabled for high concurrency & minimal lock contention
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	conn.SetMaxOpenConns(1) // SQLite works best with single writer or connection pooling handled properly

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS servers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		type TEXT NOT NULL DEFAULT 'paper',
		version TEXT NOT NULL DEFAULT '1.20.4',
		jar_file TEXT NOT NULL DEFAULT 'server.jar',
		port INTEGER NOT NULL DEFAULT 25565,
		memory_min INTEGER NOT NULL DEFAULT 1024,
		memory_max INTEGER NOT NULL DEFAULT 2048,
		java_args TEXT DEFAULT '',
		java_path TEXT DEFAULT '',
		auto_restart BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_servers_port ON servers(port);

	CREATE TABLE IF NOT EXISTS ssh_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		fingerprint TEXT NOT NULL,
		public_key TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		key_prefix TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		action TEXT NOT NULL,
		details TEXT NOT NULL,
		ip_address TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_activity_created ON activity_logs(created_at DESC);

	CREATE TABLE IF NOT EXISTS server_ports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		port INTEGER NOT NULL,
		protocol TEXT NOT NULL DEFAULT 'tcp/udp',
		notes TEXT NOT NULL DEFAULT '',
		is_primary BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		UNIQUE(server_id, port)
	);

	CREATE INDEX IF NOT EXISTS idx_server_ports_srv ON server_ports(server_id);
	CREATE INDEX IF NOT EXISTS idx_server_ports_port ON server_ports(port);

	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		base_url TEXT NOT NULL,
		api_token TEXT NOT NULL DEFAULT '',
		fqdn TEXT NOT NULL DEFAULT '',
		location TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS server_databases (
		id TEXT PRIMARY KEY,
		server_id TEXT NOT NULL,
		name TEXT NOT NULL,
		engine TEXT NOT NULL DEFAULT 'sqlite',
		username TEXT NOT NULL DEFAULT '',
		password TEXT NOT NULL DEFAULT '',
		host TEXT NOT NULL DEFAULT '127.0.0.1',
		port INTEGER NOT NULL DEFAULT 0,
		file_path TEXT NOT NULL DEFAULT '',
		max_connections INTEGER NOT NULL DEFAULT 25,
		created_at DATETIME NOT NULL,
		UNIQUE(server_id, name)
	);

	CREATE INDEX IF NOT EXISTS idx_server_databases_srv ON server_databases(server_id);

	CREATE TABLE IF NOT EXISTS panel_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) GetUserCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	row := db.conn.QueryRow("SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?", username)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) GetUserByID(id int) (*User, error) {
	row := db.conn.QueryRow("SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?", id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) CreateUser(username, passwordHash, role string) (*User, error) {
	now := time.Now().UTC()
	res, err := db.conn.Exec("INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
		username, passwordHash, role, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           int(id),
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
	}, nil
}

func (db *DB) ListUsers() ([]*User, error) {
	rows, err := db.conn.Query("SELECT id, username, role, created_at FROM users ORDER BY created_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []*User{}
	}
	return users, nil
}

func (db *DB) DeleteUser(id int) error {
	_, err := db.conn.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (db *DB) ListServers() ([]*Server, error) {
	rows, err := db.conn.Query(`SELECT id, name, description, type, version, jar_file, port, memory_min, memory_max, java_args, java_path, auto_restart, created_at, updated_at FROM servers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*Server
	for rows.Next() {
		s := &Server{}
		var autoRestart int
		err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.Type, &s.Version, &s.JarFile,
			&s.Port, &s.MemoryMinMB, &s.MemoryMaxMB, &s.JavaArgs, &s.JavaPath,
			&autoRestart, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		s.AutoRestart = autoRestart == 1
		servers = append(servers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if servers == nil {
		servers = []*Server{}
	}
	return servers, nil
}

func (db *DB) GetServer(id string) (*Server, error) {
	row := db.conn.QueryRow(`SELECT id, name, description, type, version, jar_file, port, memory_min, memory_max, java_args, java_path, auto_restart, created_at, updated_at FROM servers WHERE id = ?`, id)
	s := &Server{}
	var autoRestart int
	err := row.Scan(
		&s.ID, &s.Name, &s.Description, &s.Type, &s.Version, &s.JarFile,
		&s.Port, &s.MemoryMinMB, &s.MemoryMaxMB, &s.JavaArgs, &s.JavaPath,
		&autoRestart, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.AutoRestart = autoRestart == 1
	return s, nil
}

func (db *DB) CreateServer(s *Server) error {
	autoRestart := 0
	if s.AutoRestart {
		autoRestart = 1
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	_, err := db.conn.Exec(`INSERT INTO servers (id, name, description, type, version, jar_file, port, memory_min, memory_max, java_args, java_path, auto_restart, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.Description, s.Type, s.Version, s.JarFile, s.Port, s.MemoryMinMB, s.MemoryMaxMB, s.JavaArgs, s.JavaPath, autoRestart, s.CreatedAt, s.UpdatedAt)
	return err
}

func (db *DB) UpdateServer(s *Server) error {
	autoRestart := 0
	if s.AutoRestart {
		autoRestart = 1
	}
	s.UpdatedAt = time.Now().UTC()

	_, err := db.conn.Exec(`UPDATE servers SET name = ?, description = ?, type = ?, version = ?, jar_file = ?, port = ?, memory_min = ?, memory_max = ?, java_args = ?, java_path = ?, auto_restart = ?, updated_at = ? WHERE id = ?`,
		s.Name, s.Description, s.Type, s.Version, s.JarFile, s.Port, s.MemoryMinMB, s.MemoryMaxMB, s.JavaArgs, s.JavaPath, autoRestart, s.UpdatedAt, s.ID)
	return err
}

func (db *DB) DeleteServer(id string) error {
	_, err := db.conn.Exec("DELETE FROM servers WHERE id = ?", id)
	return err
}

func (db *DB) UpdateUserPassword(userID int, newHash string) error {
	_, err := db.conn.Exec("UPDATE users SET password_hash = ? WHERE id = ?", newHash, userID)
	return err
}

// --- SSH Keys ---

func (db *DB) AddSSHKey(userID int, name, fingerprint, publicKey string) (*SSHKey, error) {
	now := time.Now().UTC()
	res, err := db.conn.Exec("INSERT INTO ssh_keys (user_id, name, fingerprint, public_key, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, name, fingerprint, publicKey, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &SSHKey{
		ID:          int(id),
		UserID:      userID,
		Name:        name,
		Fingerprint: fingerprint,
		PublicKey:   publicKey,
		CreatedAt:   now,
	}, nil
}

func (db *DB) ListSSHKeys(userID int) ([]*SSHKey, error) {
	rows, err := db.conn.Query("SELECT id, user_id, name, fingerprint, public_key, created_at FROM ssh_keys WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*SSHKey
	for rows.Next() {
		k := &SSHKey{}
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Fingerprint, &k.PublicKey, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (db *DB) DeleteSSHKey(userID, keyID int) error {
	_, err := db.conn.Exec("DELETE FROM ssh_keys WHERE id = ? AND user_id = ?", keyID, userID)
	return err
}

func (db *DB) GetAllSSHKeysForUser(username string) ([]*SSHKey, error) {
	user, err := db.GetUserByUsername(username)
	if err != nil || user == nil {
		return nil, err
	}
	return db.ListSSHKeys(user.ID)
}

// --- API Keys ---

func (db *DB) AddAPIKey(userID int, name, keyHash, keyPrefix string) (*APIKey, error) {
	now := time.Now().UTC()
	res, err := db.conn.Exec("INSERT INTO api_keys (user_id, name, key_hash, key_prefix, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, name, keyHash, keyPrefix, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &APIKey{
		ID:        int(id),
		UserID:    userID,
		Name:      name,
		KeyPrefix: keyPrefix,
		CreatedAt: now,
	}, nil
}

func (db *DB) ListAPIKeys(userID int) ([]*APIKey, error) {
	rows, err := db.conn.Query("SELECT id, user_id, name, key_prefix, created_at FROM api_keys WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		k := &APIKey{}
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyPrefix, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (db *DB) DeleteAPIKey(userID, keyID int) error {
	_, err := db.conn.Exec("DELETE FROM api_keys WHERE id = ? AND user_id = ?", keyID, userID)
	return err
}

// --- Activity Logs ---

func (db *DB) LogActivity(userID int, action, details, ip string) error {
	now := time.Now().UTC()
	_, err := db.conn.Exec("INSERT INTO activity_logs (user_id, action, details, ip_address, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, action, details, ip, now)
	return err
}

func (db *DB) ListActivity(userID int, includeAll bool, limit int) ([]*ActivityLog, error) {
	if limit <= 0 {
		limit = 200
	} else if limit > 500 {
		limit = 500
	}

	var rows *sql.Rows
	var err error
	if includeAll {
		rows, err = db.conn.Query(`
			SELECT a.id, a.user_id, COALESCE(u.username, CASE WHEN a.user_id = 0 THEN 'system' ELSE 'deleted-user' END),
			       a.action, a.details, a.ip_address, a.created_at
			FROM activity_logs a
			LEFT JOIN users u ON u.id = a.user_id
			ORDER BY a.created_at DESC, a.id DESC
			LIMIT ?`, limit)
	} else {
		rows, err = db.conn.Query(`
			SELECT a.id, a.user_id, COALESCE(u.username, CASE WHEN a.user_id = 0 THEN 'system' ELSE 'deleted-user' END),
			       a.action, a.details, a.ip_address, a.created_at
			FROM activity_logs a
			LEFT JOIN users u ON u.id = a.user_id
			WHERE a.user_id = ?
			ORDER BY a.created_at DESC, a.id DESC
			LIMIT ?`, userID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []*ActivityLog{}
	for rows.Next() {
		l := &ActivityLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Details, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// --- Server Port Allocations ---

func (db *DB) ListServerPorts(serverID string) ([]*ServerPort, error) {
	rows, err := db.conn.Query("SELECT id, server_id, port, protocol, notes, is_primary, created_at FROM server_ports WHERE server_id = ? ORDER BY is_primary DESC, port ASC", serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []*ServerPort
	for rows.Next() {
		p := &ServerPort{}
		var isPrimary int
		if err := rows.Scan(&p.ID, &p.ServerID, &p.Port, &p.Protocol, &p.Notes, &isPrimary, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.IsPrimary = isPrimary == 1
		ports = append(ports, p)
	}

	// Backwards compatibility: if no ports in server_ports, populate from servers table
	if len(ports) == 0 {
		srv, err := db.GetServer(serverID)
		if err == nil && srv.Port > 0 {
			now := time.Now().UTC()
			res, err := db.conn.Exec("INSERT OR IGNORE INTO server_ports (server_id, port, protocol, notes, is_primary, created_at) VALUES (?, ?, 'tcp/udp', 'Primary Game Port', 1, ?)",
				serverID, srv.Port, now)
			if err == nil {
				id, _ := res.LastInsertId()
				ports = append(ports, &ServerPort{
					ID:        int(id),
					ServerID:  serverID,
					Port:      srv.Port,
					Protocol:  "tcp/udp",
					Notes:     "Primary Game Port",
					IsPrimary: true,
					CreatedAt: now,
				})
			}
		}
	}

	if ports == nil {
		ports = []*ServerPort{}
	}
	return ports, nil
}

func (db *DB) AddServerPort(serverID string, port int, protocol, notes string, isPrimary bool) (*ServerPort, error) {
	if port < 1024 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1024 and 65535")
	}
	if protocol == "" {
		protocol = "tcp/udp"
	}
	if notes == "" {
		if isPrimary {
			notes = "Primary Game Port"
		} else {
			notes = "Additional Allocation"
		}
	}

	// Check if port is already allocated on another server
	var otherServerID string
	err := db.conn.QueryRow("SELECT server_id FROM server_ports WHERE port = ? AND server_id != ?", port, serverID).Scan(&otherServerID)
	if err == nil && otherServerID != "" {
		return nil, fmt.Errorf("port %d is already allocated to another server (%s)", port, otherServerID)
	}

	// Also check servers table for legacy entries
	var srvID string
	err = db.conn.QueryRow("SELECT id FROM servers WHERE port = ? AND id != ?", port, serverID).Scan(&srvID)
	if err == nil && srvID != "" {
		return nil, fmt.Errorf("port %d is already used as the primary port of server %s", port, srvID)
	}

	if isPrimary {
		// Set all other ports for this server as non-primary
		_, _ = db.conn.Exec("UPDATE server_ports SET is_primary = 0 WHERE server_id = ?", serverID)
		// Update servers table
		_, _ = db.conn.Exec("UPDATE servers SET port = ? WHERE id = ?", port, serverID)
	}

	now := time.Now().UTC()
	primaryInt := 0
	if isPrimary {
		primaryInt = 1
	}

	res, err := db.conn.Exec("INSERT INTO server_ports (server_id, port, protocol, notes, is_primary, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		serverID, port, protocol, notes, primaryInt, now)
	if err != nil {
		return nil, fmt.Errorf("port %d is already assigned to this server", port)
	}

	id, _ := res.LastInsertId()
	return &ServerPort{
		ID:        int(id),
		ServerID:  serverID,
		Port:      port,
		Protocol:  protocol,
		Notes:     notes,
		IsPrimary: isPrimary,
		CreatedAt: now,
	}, nil
}

func (db *DB) DeleteServerPort(serverID string, portID int) error {
	var isPrimary int
	err := db.conn.QueryRow("SELECT is_primary FROM server_ports WHERE id = ? AND server_id = ?", portID, serverID).Scan(&isPrimary)
	if err != nil {
		return fmt.Errorf("port allocation not found")
	}
	if isPrimary == 1 {
		return fmt.Errorf("cannot delete the primary game port. Assign another port as primary first.")
	}

	_, err = db.conn.Exec("DELETE FROM server_ports WHERE id = ? AND server_id = ?", portID, serverID)
	return err
}

func (db *DB) SetPrimaryServerPort(serverID string, portID int) (int, error) {
	var port int
	err := db.conn.QueryRow("SELECT port FROM server_ports WHERE id = ? AND server_id = ?", portID, serverID).Scan(&port)
	if err != nil {
		return 0, fmt.Errorf("port allocation not found")
	}

	_, _ = db.conn.Exec("UPDATE server_ports SET is_primary = 0 WHERE server_id = ?", serverID)
	_, err = db.conn.Exec("UPDATE server_ports SET is_primary = 1 WHERE id = ? AND server_id = ?", portID, serverID)
	if err != nil {
		return 0, err
	}
	_, _ = db.conn.Exec("UPDATE servers SET port = ? WHERE id = ?", port, serverID)
	return port, nil
}

func (db *DB) DeleteAllServerPorts(serverID string) error {
	_, err := db.conn.Exec("DELETE FROM server_ports WHERE server_id = ?", serverID)
	return err
}

// --- API Key Lookup ---

func (db *DB) GetUserByAPIKeyHash(keyHash string) (*User, error) {
	row := db.conn.QueryRow(`
		SELECT u.id, u.username, u.password_hash, u.role, u.created_at
		FROM api_keys ak
		JOIN users u ON u.id = ak.user_id
		WHERE ak.key_hash = ?
		LIMIT 1`, keyHash)
	user := &User{}
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt); err != nil {
		return nil, err
	}
	return user, nil
}

// --- Cluster Nodes ---

func (db *DB) ListNodes() ([]*ClusterNode, error) {
	rows, err := db.conn.Query(`SELECT id, name, base_url, api_token, fqdn, location, created_at, updated_at FROM nodes ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*ClusterNode
	for rows.Next() {
		n := &ClusterNode{}
		if err := rows.Scan(&n.ID, &n.Name, &n.BaseURL, &n.APIToken, &n.FQDN, &n.Location, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if nodes == nil {
		nodes = []*ClusterNode{}
	}
	return nodes, nil
}

func (db *DB) GetNode(id string) (*ClusterNode, error) {
	row := db.conn.QueryRow(`SELECT id, name, base_url, api_token, fqdn, location, created_at, updated_at FROM nodes WHERE id = ?`, id)
	n := &ClusterNode{}
	if err := row.Scan(&n.ID, &n.Name, &n.BaseURL, &n.APIToken, &n.FQDN, &n.Location, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, err
	}
	return n, nil
}

func (db *DB) CreateNode(n *ClusterNode) error {
	now := time.Now().UTC()
	n.CreatedAt = now
	n.UpdatedAt = now
	_, err := db.conn.Exec(`INSERT INTO nodes (id, name, base_url, api_token, fqdn, location, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.Name, n.BaseURL, n.APIToken, n.FQDN, n.Location, n.CreatedAt, n.UpdatedAt)
	return err
}

func (db *DB) UpdateNode(n *ClusterNode) error {
	n.UpdatedAt = time.Now().UTC()
	_, err := db.conn.Exec(`UPDATE nodes SET name = ?, base_url = ?, api_token = ?, fqdn = ?, location = ?, updated_at = ? WHERE id = ?`,
		n.Name, n.BaseURL, n.APIToken, n.FQDN, n.Location, n.UpdatedAt, n.ID)
	return err
}

func (db *DB) DeleteNode(id string) error {
	_, err := db.conn.Exec(`DELETE FROM nodes WHERE id = ?`, id)
	return err
}

// --- Server Databases ---

func (db *DB) ListServerDatabases(serverID string) ([]*ServerDatabase, error) {
	rows, err := db.conn.Query(`SELECT id, server_id, name, engine, username, password, host, port, file_path, max_connections, created_at FROM server_databases WHERE server_id = ? ORDER BY created_at DESC`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []*ServerDatabase
	for rows.Next() {
		d := &ServerDatabase{}
		if err := rows.Scan(&d.ID, &d.ServerID, &d.Name, &d.Engine, &d.Username, &d.Password, &d.Host, &d.Port, &d.FilePath, &d.MaxConnections, &d.CreatedAt); err != nil {
			return nil, err
		}
		dbs = append(dbs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if dbs == nil {
		dbs = []*ServerDatabase{}
	}
	return dbs, nil
}

func (db *DB) GetServerDatabase(serverID, dbID string) (*ServerDatabase, error) {
	row := db.conn.QueryRow(`SELECT id, server_id, name, engine, username, password, host, port, file_path, max_connections, created_at FROM server_databases WHERE server_id = ? AND id = ?`, serverID, dbID)
	d := &ServerDatabase{}
	if err := row.Scan(&d.ID, &d.ServerID, &d.Name, &d.Engine, &d.Username, &d.Password, &d.Host, &d.Port, &d.FilePath, &d.MaxConnections, &d.CreatedAt); err != nil {
		return nil, err
	}
	return d, nil
}

func (db *DB) CreateServerDatabase(d *ServerDatabase) error {
	now := time.Now().UTC()
	d.CreatedAt = now
	_, err := db.conn.Exec(`INSERT INTO server_databases (id, server_id, name, engine, username, password, host, port, file_path, max_connections, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.ServerID, d.Name, d.Engine, d.Username, d.Password, d.Host, d.Port, d.FilePath, d.MaxConnections, d.CreatedAt)
	return err
}

func (db *DB) UpdateServerDatabasePassword(serverID, dbID, newPassword string) error {
	_, err := db.conn.Exec(`UPDATE server_databases SET password = ? WHERE server_id = ? AND id = ?`, newPassword, serverID, dbID)
	return err
}

func (db *DB) DeleteServerDatabase(serverID, dbID string) error {
	_, err := db.conn.Exec(`DELETE FROM server_databases WHERE server_id = ? AND id = ?`, serverID, dbID)
	return err
}

func (db *DB) DeleteAllServerDatabases(serverID string) error {
	_, err := db.conn.Exec(`DELETE FROM server_databases WHERE server_id = ?`, serverID)
	return err
}

// --- Panel Settings ---

func (db *DB) GetPanelSetting(key string) (string, error) {
	var val string
	err := db.conn.QueryRow(`SELECT value FROM panel_settings WHERE key = ?`, key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (db *DB) SetPanelSetting(key, value string) error {
	now := time.Now().UTC()
	_, err := db.conn.Exec(`INSERT INTO panel_settings (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`, key, value, now)
	return err
}

func (db *DB) DeletePanelSetting(key string) error {
	_, err := db.conn.Exec(`DELETE FROM panel_settings WHERE key = ?`, key)
	return err
}
