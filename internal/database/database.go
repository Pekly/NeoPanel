package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	if err := db.MigrateInvitations(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("invitation migration failed: %w", err)
	}
	if err := db.MigrateScheduleTasks(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("schedule task migration failed: %w", err)
	}
	if err := db.MigratePasskeys(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("passkey migration failed: %w", err)
	}

	if err := db.migratePanelRoles(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("panel role migration failed: %w", err)
	}
	if err := db.migrateInvitationLifecycle(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("invitation lifecycle migration failed: %w", err)
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

	CREATE TABLE IF NOT EXISTS server_subusers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		permissions TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME NOT NULL,
		UNIQUE(server_id, user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_subusers_srv ON server_subusers(server_id);
	CREATE INDEX IF NOT EXISTS idx_subusers_usr ON server_subusers(user_id);

	CREATE TABLE IF NOT EXISTS server_schedules (
		id TEXT PRIMARY KEY,
		server_id TEXT NOT NULL,
		name TEXT NOT NULL,
		cron_expr TEXT NOT NULL,
		action TEXT NOT NULL,
		payload TEXT NOT NULL DEFAULT '',
		is_active BOOLEAN NOT NULL DEFAULT 1,
		last_run_at DATETIME,
		next_run_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_schedules_srv ON server_schedules(server_id);

	CREATE TABLE IF NOT EXISTS password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		used BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
	CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets(user_id);
	`
	if _, err := db.conn.Exec(schema); err != nil {
		return err
	}

	// Safe automatic schema column migrations
	_ = db.addColumnIfNotExists("server_schedules", "last_status", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("server_schedules", "last_error", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("server_schedules", "last_finished_at", "DATETIME")
	// A daemon restart cannot resume an in-flight action.
	_, _ = db.conn.Exec(`UPDATE server_schedules SET last_status = 'failed', last_error = 'Execution interrupted by daemon restart', last_finished_at = ? WHERE last_status IN ('started', 'running')`, time.Now().UTC())
	_ = db.addColumnIfNotExists("users", "email", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("users", "first_name", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("users", "last_name", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("users", "avatar_url", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("users", "two_factor_enabled", "BOOLEAN NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("users", "two_factor_secret", "TEXT NOT NULL DEFAULT ''")
	_ = db.addColumnIfNotExists("users", "recovery_codes", "TEXT NOT NULL DEFAULT ''")
	// Existing databases start at version zero; fail startup if migration fails.
	_ = db.addColumnIfNotExists("users", "auth_version", "INTEGER NOT NULL DEFAULT 0")
	if _, err := db.conn.Exec("SELECT auth_version FROM users LIMIT 0"); err != nil {
		return fmt.Errorf("auth version migration: %w", err)
	}
	_ = db.addColumnIfNotExists("servers", "icon_url", "TEXT DEFAULT ''")
	_ = db.addColumnIfNotExists("servers", "owner_id", "INTEGER DEFAULT 1")
	_ = db.addColumnIfNotExists("servers", "node_id", "TEXT DEFAULT 'local'")
	_ = db.addColumnIfNotExists("servers", "cpu_limit_percent", "INTEGER NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("servers", "disk_limit_mb", "INTEGER NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("nodes", "is_free", "BOOLEAN NOT NULL DEFAULT 0")

	// Migrate legacy "Powered by NeoPanel Daemon" footer in database
	_, _ = db.conn.Exec(`
		UPDATE panel_settings 
		SET value = REPLACE(value, '"footer_text":"Powered by NeoPanel Daemon"', '"footer_text":"[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026"')
		WHERE key = 'branding_json' AND value LIKE '%"footer_text":"Powered by NeoPanel Daemon"%';
	`)
	_, _ = db.conn.Exec(`
		UPDATE panel_settings 
		SET value = '[NeoPanel](https://github.com/Pekly/NeoPanel/)© 2026'
		WHERE key = 'footer_text' AND (value = 'Powered by NeoPanel Daemon' OR value = '');
	`)
	_ = db.addColumnIfNotExists("nodes", "idle_timeout_min", "INTEGER NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("nodes", "allow_user_ram", "BOOLEAN NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("nodes", "allow_user_ports", "BOOLEAN NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("nodes", "allow_user_java", "BOOLEAN NOT NULL DEFAULT 1")
	_ = db.addColumnIfNotExists("nodes", "isolation_engine", "TEXT NOT NULL DEFAULT 'auto'")
	_ = db.addColumnIfNotExists("nodes", "cpu_limit_percent", "INTEGER NOT NULL DEFAULT 0")
	_ = db.addColumnIfNotExists("nodes", "tasks_max", "INTEGER NOT NULL DEFAULT 512")

	return nil
}

func (db *DB) addColumnIfNotExists(table, column, colDef string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, colDef)
	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) GetUserCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	row := db.conn.QueryRow("SELECT id, username, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(email, ''), password_hash, role, COALESCE(avatar_url, ''), COALESCE(two_factor_enabled, 0), COALESCE(two_factor_secret, ''), COALESCE(recovery_codes, ''), created_at, auth_version, invitation_created FROM users WHERE username = ?", username)
	user := &User{}
	var tfa int
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Role, &user.AvatarURL, &tfa, &user.TwoFactorSecret, &user.RecoveryCodes, &user.CreatedAt, &user.AuthVersion, &user.InvitationCreated)
	if err != nil {
		return nil, err
	}
	user.TwoFactorEnabled = tfa == 1
	return db.loadUserPanelRole(user)
}

func (db *DB) GetUserByID(id int) (*User, error) {
	row := db.conn.QueryRow("SELECT id, username, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(email, ''), password_hash, role, COALESCE(avatar_url, ''), COALESCE(two_factor_enabled, 0), COALESCE(two_factor_secret, ''), COALESCE(recovery_codes, ''), created_at, auth_version, invitation_created FROM users WHERE id = ?", id)
	user := &User{}
	var tfa int
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Role, &user.AvatarURL, &tfa, &user.TwoFactorSecret, &user.RecoveryCodes, &user.CreatedAt, &user.AuthVersion, &user.InvitationCreated)
	if err != nil {
		return nil, err
	}
	user.TwoFactorEnabled = tfa == 1
	return db.loadUserPanelRole(user)
}

func (db *DB) GetUserByEmailOrUsername(lookup string) (*User, error) {
	lookup = strings.TrimSpace(lookup)
	row := db.conn.QueryRow("SELECT id, username, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(email, ''), password_hash, role, COALESCE(avatar_url, ''), COALESCE(two_factor_enabled, 0), COALESCE(two_factor_secret, ''), COALESCE(recovery_codes, ''), created_at, auth_version, invitation_created FROM users WHERE username = ? OR (email != '' AND LOWER(email) = LOWER(?))", lookup, lookup)
	user := &User{}
	var tfa int
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Role, &user.AvatarURL, &tfa, &user.TwoFactorSecret, &user.RecoveryCodes, &user.CreatedAt, &user.AuthVersion, &user.InvitationCreated)
	if err != nil {
		return nil, err
	}
	user.TwoFactorEnabled = tfa == 1
	return db.loadUserPanelRole(user)
}

func (db *DB) CreateUser(username, passwordHash, role string) (*User, error) {
	return db.CreateUserWithDetails(username, "", passwordHash, role, "")
}

func (db *DB) CreateUserWithDetails(username, email, passwordHash, role, avatarURL string) (*User, error) {
	return db.CreateUserWithFullDetails(username, "", "", email, passwordHash, role, avatarURL)
}

func (db *DB) CreateUserWithFullDetails(username, firstName, lastName, email, passwordHash, role, avatarURL string) (*User, error) {
	now := time.Now().UTC()
	username = strings.TrimSpace(username)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	email = strings.TrimSpace(email)
	if role != "admin" {
		role = "user"
	}
	res, err := db.conn.Exec("INSERT INTO users (username, first_name, last_name, email, password_hash, role, avatar_url, two_factor_enabled, two_factor_secret, recovery_codes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, 0, '', '', ?)",
		username, firstName, lastName, email, passwordHash, role, avatarURL, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return db.GetUserByID(int(id))
}

func (db *DB) UpdateUserProfile(id int, username, firstName, lastName, email string) error {
	username = strings.TrimSpace(username)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	email = strings.TrimSpace(email)

	if username != "" {
		_, err := db.conn.Exec("UPDATE users SET username = ?, first_name = ?, last_name = ?, email = ? WHERE id = ?",
			username, firstName, lastName, email, id)
		return err
	}
	_, err := db.conn.Exec("UPDATE users SET first_name = ?, last_name = ?, email = ? WHERE id = ?",
		firstName, lastName, email, id)
	return err
}

func (db *DB) AdminUpdateUser(id int, username, firstName, lastName, email, role, newPasswordHash string) error {
	username = strings.TrimSpace(username)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	email = strings.TrimSpace(email)
	if role != "admin" {
		role = "user"
	}

	if newPasswordHash != "" {
		_, err := db.conn.Exec("UPDATE users SET username = ?, first_name = ?, last_name = ?, email = ?, role = ?, password_hash = ?, auth_version = auth_version + 1 WHERE id = ?",
			username, firstName, lastName, email, role, newPasswordHash, id)
		return err
	}
	_, err := db.conn.Exec("UPDATE users SET username = ?, first_name = ?, last_name = ?, email = ?, role = ? WHERE id = ?",
		username, firstName, lastName, email, role, id)
	return err
}

func (db *DB) UpdateUserEmail(id int, email string) error {
	_, err := db.conn.Exec("UPDATE users SET email = ? WHERE id = ?", strings.TrimSpace(email), id)
	return err
}

func (db *DB) UpdateUserAvatar(id int, avatarURL string) error {
	_, err := db.conn.Exec("UPDATE users SET avatar_url = ? WHERE id = ?", strings.TrimSpace(avatarURL), id)
	return err
}

func (db *DB) UpdateUserTwoFactor(id int, enabled bool, secret, recoveryCodes string) error {
	val := 0
	if enabled {
		val = 1
	}
	_, err := db.conn.Exec("UPDATE users SET two_factor_enabled = ?, two_factor_secret = ?, recovery_codes = ? WHERE id = ?",
		val, secret, recoveryCodes, id)
	return err
}

func (db *DB) UpdateUserRole(id int, role string) error {
	if role != "admin" {
		role = "user"
	}
	_, err := db.conn.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	return err
}

func (db *DB) ListUsers() ([]*User, error) {
	rows, err := db.conn.Query("SELECT id, username, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(email, ''), role, COALESCE(avatar_url, ''), COALESCE(two_factor_enabled, 0), created_at, auth_version, invitation_created FROM users ORDER BY created_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		var tfa int
		if err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Role, &u.AvatarURL, &tfa, &u.CreatedAt, &u.AuthVersion, &u.InvitationCreated); err != nil {
			return nil, err
		}
		u.TwoFactorEnabled = tfa == 1
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for _, user := range users {
		if _, err := db.loadUserPanelRole(user); err != nil {
			return nil, err
		}
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

func (db *DB) ListServersForUser(userID int, role string) ([]*Server, error) {
	if role == "admin" || userID == 0 {
		return db.ListServers()
	}
	query := `SELECT s.id, s.name, s.description, s.type, s.version, s.jar_file, s.port, s.memory_min, s.memory_max, s.java_args, s.java_path, s.auto_restart, COALESCE(s.icon_url, ''), COALESCE(s.owner_id, 1), COALESCE(s.node_id, 'local'), COALESCE(s.cpu_limit_percent, 0), COALESCE(s.disk_limit_mb, 0), s.created_at, s.updated_at 
	FROM servers s 
	LEFT JOIN server_subusers su ON s.id = su.server_id AND su.user_id = ?
	WHERE s.owner_id = ? OR su.user_id IS NOT NULL
	GROUP BY s.id
	ORDER BY s.created_at DESC`
	rows, err := db.conn.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*Server
	for rows.Next() {
		s := &Server{}
		var autoRestart int
		var iconURL, nodeID string
		var ownerID int
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.Type, &s.Version, &s.JarFile,
			&s.Port, &s.MemoryMinMB, &s.MemoryMaxMB, &s.JavaArgs, &s.JavaPath,
			&autoRestart, &iconURL, &ownerID, &nodeID, &s.CPULimitPercent, &s.DiskLimitMB, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.AutoRestart = autoRestart == 1
		s.IconURL = iconURL
		s.OwnerID = ownerID
		s.NodeID = nodeID
		servers = append(servers, s)
	}
	if servers == nil {
		servers = []*Server{}
	}
	return servers, nil
}

func (db *DB) ListServers() ([]*Server, error) {
	rows, err := db.conn.Query(`SELECT id, name, description, type, version, jar_file, port, memory_min, memory_max, java_args, java_path, auto_restart, COALESCE(icon_url, ''), COALESCE(owner_id, 1), COALESCE(node_id, 'local'), COALESCE(cpu_limit_percent, 0), COALESCE(disk_limit_mb, 0), created_at, updated_at FROM servers ORDER BY created_at DESC`)
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
			&autoRestart, &s.IconURL, &s.OwnerID, &s.NodeID, &s.CPULimitPercent, &s.DiskLimitMB, &s.CreatedAt, &s.UpdatedAt,
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
	row := db.conn.QueryRow(`SELECT id, name, description, type, version, jar_file, port, memory_min, memory_max, java_args, java_path, auto_restart, COALESCE(icon_url, ''), COALESCE(owner_id, 1), COALESCE(node_id, 'local'), COALESCE(cpu_limit_percent, 0), COALESCE(disk_limit_mb, 0), created_at, updated_at FROM servers WHERE id = ?`, id)
	s := &Server{}
	var autoRestart int
	err := row.Scan(
		&s.ID, &s.Name, &s.Description, &s.Type, &s.Version, &s.JarFile,
		&s.Port, &s.MemoryMinMB, &s.MemoryMaxMB, &s.JavaArgs, &s.JavaPath,
		&autoRestart, &s.IconURL, &s.OwnerID, &s.NodeID, &s.CPULimitPercent, &s.DiskLimitMB, &s.CreatedAt, &s.UpdatedAt,
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
	if s.NodeID == "" {
		s.NodeID = "local"
	}
	if s.OwnerID <= 0 {
		s.OwnerID = 1
	}

	_, err := db.conn.Exec(`INSERT INTO servers (id, name, description, type, version, jar_file, port, memory_min, memory_max, java_args, java_path, auto_restart, icon_url, owner_id, node_id, cpu_limit_percent, disk_limit_mb, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.Description, s.Type, s.Version, s.JarFile, s.Port, s.MemoryMinMB, s.MemoryMaxMB, s.JavaArgs, s.JavaPath, autoRestart, s.IconURL, s.OwnerID, s.NodeID, s.CPULimitPercent, s.DiskLimitMB, s.CreatedAt, s.UpdatedAt)
	return err
}

func (db *DB) UpdateServer(s *Server) error {
	autoRestart := 0
	if s.AutoRestart {
		autoRestart = 1
	}
	s.UpdatedAt = time.Now().UTC()
	if s.NodeID == "" {
		s.NodeID = "local"
	}
	if s.OwnerID <= 0 {
		s.OwnerID = 1
	}

	_, err := db.conn.Exec(`UPDATE servers SET name = ?, description = ?, type = ?, version = ?, jar_file = ?, port = ?, memory_min = ?, memory_max = ?, java_args = ?, java_path = ?, auto_restart = ?, icon_url = ?, owner_id = ?, node_id = ?, cpu_limit_percent = ?, disk_limit_mb = ?, updated_at = ? WHERE id = ?`,
		s.Name, s.Description, s.Type, s.Version, s.JarFile, s.Port, s.MemoryMinMB, s.MemoryMaxMB, s.JavaArgs, s.JavaPath, autoRestart, s.IconURL, s.OwnerID, s.NodeID, s.CPULimitPercent, s.DiskLimitMB, s.UpdatedAt, s.ID)
	return err
}

func (db *DB) UpdateServerOwner(serverID string, ownerID int) error {
	_, err := db.conn.Exec(`UPDATE servers SET owner_id = ?, updated_at = ? WHERE id = ?`, ownerID, time.Now().UTC(), serverID)
	return err
}

func (db *DB) UpdateServerIcon(id string, iconURL string) error {
	_, err := db.conn.Exec(`UPDATE servers SET icon_url = ?, updated_at = ? WHERE id = ?`, iconURL, time.Now().UTC(), id)
	return err
}

func (db *DB) DeleteServer(id string) error {
	return db.deleteInvitationServer(id)
}

func (db *DB) UpdateUserPassword(userID int, newHash string) error {
	_, err := db.conn.Exec("UPDATE users SET password_hash = ?, auth_version = auth_version + 1 WHERE id = ?", newHash, userID)
	return err
}

func (db *DB) CreatePasswordResetToken(userID int, token string, expiresAt time.Time) error {
	now := time.Now().UTC()
	_, _ = db.conn.Exec("UPDATE password_resets SET used = 1 WHERE user_id = ? AND used = 0", userID)
	_, err := db.conn.Exec(`
		INSERT INTO password_resets (user_id, token, expires_at, used, created_at)
		VALUES (?, ?, ?, 0, ?)
	`, userID, token, expiresAt.UTC(), now)
	return err
}

func (db *DB) GetPasswordResetToken(token string) (*PasswordReset, error) {
	row := db.conn.QueryRow(`
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_resets
		WHERE token = ?
	`, strings.TrimSpace(token))
	pr := &PasswordReset{}
	var used int
	err := row.Scan(&pr.ID, &pr.UserID, &pr.Token, &pr.ExpiresAt, &used, &pr.CreatedAt)
	if err != nil {
		return nil, err
	}
	pr.Used = used == 1
	return pr, nil
}

func (db *DB) MarkPasswordResetUsed(token string) error {
	_, err := db.conn.Exec("UPDATE password_resets SET used = 1 WHERE token = ?", strings.TrimSpace(token))
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
	user, err := db.GetUserByEmailOrUsername(username)
	if err != nil || user == nil {
		return nil, err
	}
	return db.ListSSHKeys(user.ID)
}

func (db *DB) GetAllSSHKeys() ([]*SSHKey, error) {
	rows, err := db.conn.Query("SELECT id, user_id, name, fingerprint, public_key, created_at FROM ssh_keys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []*SSHKey
	for rows.Next() {
		k := &SSHKey{}
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Fingerprint, &k.PublicKey, &k.CreatedAt); err == nil {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (db *DB) UpsertSyncedSSHKey(id, userID int, name, publicKey string) error {
	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		return nil
	}
	now := time.Now().UTC()
	var existingID int
	err := db.conn.QueryRow("SELECT id FROM ssh_keys WHERE id = ?", id).Scan(&existingID)
	if err == nil {
		_, err = db.conn.Exec("UPDATE ssh_keys SET user_id=?, name=?, public_key=? WHERE id=?", userID, name, publicKey, id)
		return err
	}
	_, err = db.conn.Exec("INSERT INTO ssh_keys (id, user_id, name, fingerprint, public_key, created_at) VALUES (?, ?, ?, '', ?, ?)", id, userID, name, publicKey, now)
	return err
}

func (db *DB) UpsertSyncedUser(id int, username, email, passwordHash, role string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("empty username")
	}
	now := time.Now().UTC()
	var existingID int
	err := db.conn.QueryRow("SELECT id FROM users WHERE username = ? OR id = ?", username, id).Scan(&existingID)
	if err == nil {
		res, updateErr := db.conn.Exec("UPDATE users SET identity_exported=1, username=?, email=?, auth_version=auth_version + CASE WHEN password_hash != ? THEN 1 ELSE 0 END, password_hash=?, role=? WHERE id=? AND NOT EXISTS (SELECT 1 FROM deleted_invitation_users WHERE id=?)", username, email, passwordHash, passwordHash, role, existingID, id)
		return affectedRow(res, updateErr)
	}
	_, err = db.conn.Exec("INSERT INTO users (id, username, email, password_hash, role, created_at, identity_exported) VALUES (?, ?, ?, ?, ?, ?, 1)", id, username, email, passwordHash, role, now)
	return err
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

func (db *DB) CreateAuditLog(username, action, details, ip string) error {
	userID := 0
	if username != "" && username != "system" {
		if u, _ := db.GetUserByUsername(username); u != nil {
			userID = u.ID
		}
	}
	return db.LogActivity(userID, action, details, ip)
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
		SELECT u.id, u.username, u.password_hash, u.role, u.created_at, u.auth_version
		FROM api_keys ak
		JOIN users u ON u.id = ak.user_id
		WHERE ak.key_hash = ?
		LIMIT 1`, keyHash)
	user := &User{}
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.AuthVersion); err != nil {
		return nil, err
	}
	return db.loadUserPanelRole(user)
}

// --- Cluster Nodes ---

func (db *DB) ListNodes() ([]*ClusterNode, error) {
	rows, err := db.conn.Query(`SELECT id, name, base_url, api_token, fqdn, location, COALESCE(is_free, 0), COALESCE(idle_timeout_min, 0), COALESCE(allow_user_ram, 0), COALESCE(allow_user_ports, 0), COALESCE(allow_user_java, 1), COALESCE(isolation_engine, 'auto'), COALESCE(cpu_limit_percent, 0), COALESCE(tasks_max, 512), created_at, updated_at FROM nodes ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*ClusterNode
	for rows.Next() {
		n := &ClusterNode{}
		var isFree, allowRam, allowPorts, allowJava int
		if err := rows.Scan(
			&n.ID, &n.Name, &n.BaseURL, &n.APIToken, &n.FQDN, &n.Location,
			&isFree, &n.IdleTimeoutMin, &allowRam, &allowPorts, &allowJava,
			&n.IsolationEngine, &n.CPULimitPercent, &n.TasksMax,
			&n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		n.IsFree = isFree == 1
		n.AllowUserRAM = allowRam == 1
		n.AllowUserPorts = allowPorts == 1
		n.AllowUserJava = allowJava == 1
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
	row := db.conn.QueryRow(`SELECT id, name, base_url, api_token, fqdn, location, COALESCE(is_free, 0), COALESCE(idle_timeout_min, 0), COALESCE(allow_user_ram, 0), COALESCE(allow_user_ports, 0), COALESCE(allow_user_java, 1), COALESCE(isolation_engine, 'auto'), COALESCE(cpu_limit_percent, 0), COALESCE(tasks_max, 512), created_at, updated_at FROM nodes WHERE id = ?`, id)
	n := &ClusterNode{}
	var isFree, allowRam, allowPorts, allowJava int
	if err := row.Scan(
		&n.ID, &n.Name, &n.BaseURL, &n.APIToken, &n.FQDN, &n.Location,
		&isFree, &n.IdleTimeoutMin, &allowRam, &allowPorts, &allowJava,
		&n.IsolationEngine, &n.CPULimitPercent, &n.TasksMax,
		&n.CreatedAt, &n.UpdatedAt,
	); err != nil {
		return nil, err
	}
	n.IsFree = isFree == 1
	n.AllowUserRAM = allowRam == 1
	n.AllowUserPorts = allowPorts == 1
	n.AllowUserJava = allowJava == 1
	return n, nil
}

func (db *DB) CreateNode(n *ClusterNode) error {
	now := time.Now().UTC()
	n.CreatedAt = now
	n.UpdatedAt = now
	if n.IsolationEngine == "" {
		n.IsolationEngine = "auto"
	}
	if n.TasksMax <= 0 {
		n.TasksMax = 512
	}
	isFree := 0
	if n.IsFree {
		isFree = 1
	}
	allowRam := 0
	if n.AllowUserRAM {
		allowRam = 1
	}
	allowPorts := 0
	if n.AllowUserPorts {
		allowPorts = 1
	}
	allowJava := 1
	if !n.AllowUserJava {
		allowJava = 0
	}
	_, err := db.conn.Exec(`INSERT INTO nodes (id, name, base_url, api_token, fqdn, location, is_free, idle_timeout_min, allow_user_ram, allow_user_ports, allow_user_java, isolation_engine, cpu_limit_percent, tasks_max, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.Name, n.BaseURL, n.APIToken, n.FQDN, n.Location, isFree, n.IdleTimeoutMin, allowRam, allowPorts, allowJava, n.IsolationEngine, n.CPULimitPercent, n.TasksMax, n.CreatedAt, n.UpdatedAt)
	return err
}

func (db *DB) UpdateNode(n *ClusterNode) error {
	n.UpdatedAt = time.Now().UTC()
	if n.IsolationEngine == "" {
		n.IsolationEngine = "auto"
	}
	if n.TasksMax <= 0 {
		n.TasksMax = 512
	}
	isFree := 0
	if n.IsFree {
		isFree = 1
	}
	allowRam := 0
	if n.AllowUserRAM {
		allowRam = 1
	}
	allowPorts := 0
	if n.AllowUserPorts {
		allowPorts = 1
	}
	allowJava := 1
	if !n.AllowUserJava {
		allowJava = 0
	}
	_, err := db.conn.Exec(`UPDATE nodes SET name = ?, base_url = ?, api_token = ?, fqdn = ?, location = ?, is_free = ?, idle_timeout_min = ?, allow_user_ram = ?, allow_user_ports = ?, allow_user_java = ?, isolation_engine = ?, cpu_limit_percent = ?, tasks_max = ?, updated_at = ? WHERE id = ?`,
		n.Name, n.BaseURL, n.APIToken, n.FQDN, n.Location, isFree, n.IdleTimeoutMin, allowRam, allowPorts, allowJava, n.IsolationEngine, n.CPULimitPercent, n.TasksMax, n.UpdatedAt, n.ID)
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

// InitializeServerDatabase reserves the unique name until initialization commits.
// The callback must not use the metadata DB (its writer is held by this transaction).
func (db *DB) InitializeServerDatabase(d *ServerDatabase, initialize func() error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	d.CreatedAt = time.Now().UTC()
	_, err = tx.Exec(`INSERT INTO server_databases (id, server_id, name, engine, username, password, host, port, file_path, max_connections, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, d.ID, d.ServerID, d.Name, d.Engine, d.Username, d.Password, d.Host, d.Port, d.FilePath, d.MaxConnections, d.CreatedAt)
	if err != nil {
		return err
	}
	if err = initialize(); err != nil {
		return err
	}
	return tx.Commit()
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

// --- Server Subusers & Permission Management ---

func (db *DB) ListServerSubusers(serverID string) ([]*ServerSubuser, error) {
	rows, err := db.conn.Query(`
		SELECT su.id, su.server_id, su.user_id, u.username, COALESCE(u.email, ''), su.permissions, su.created_at
		FROM server_subusers su
		JOIN users u ON su.user_id = u.id
		WHERE su.server_id = ?
		ORDER BY su.created_at ASC
	`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ServerSubuser
	for rows.Next() {
		su := &ServerSubuser{}
		var permJSON string
		if err := rows.Scan(&su.ID, &su.ServerID, &su.UserID, &su.Username, &su.Email, &permJSON, &su.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(permJSON), &su.Permissions)
		if su.Permissions == nil {
			su.Permissions = []string{}
		}
		list = append(list, su)
	}
	if list == nil {
		list = []*ServerSubuser{}
	}
	return list, nil
}

func (db *DB) AddServerSubuser(serverID string, userID int, permissions []string) (*ServerSubuser, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	now := time.Now().UTC()

	res, err := db.conn.Exec(`INSERT INTO server_subusers (server_id, user_id, permissions, created_at) VALUES (?, ?, ?, ?)`,
		serverID, userID, string(permJSON), now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	var username string
	_ = db.conn.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&username)

	return &ServerSubuser{
		ID:          int(id),
		ServerID:    serverID,
		UserID:      userID,
		Username:    username,
		Permissions: permissions,
		CreatedAt:   now,
	}, nil
}

func (db *DB) UpdateServerSubuser(serverID string, id int, permissions []string) error {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	res, err := db.conn.Exec(`UPDATE server_subusers SET permissions = ? WHERE server_id = ? AND id = ?`, string(permJSON), serverID, id)
	return affectedRow(res, err)
}

func (db *DB) DeleteServerSubuser(serverID string, id int) error {
	return db.deleteInvitationSubuser(serverID, id)
}

func (db *DB) GetUserServerPermissions(serverID string, userID int) ([]string, error) {
	var permJSON string
	err := db.conn.QueryRow(`SELECT permissions FROM server_subusers WHERE server_id = ? AND user_id = ?`, serverID, userID).Scan(&permJSON)
	if err != nil {
		return nil, err
	}
	var perms []string
	_ = json.Unmarshal([]byte(permJSON), &perms)
	if perms == nil {
		perms = []string{}
	}
	return perms, nil
}

func (db *DB) IsServerSubuser(serverID string, userID int) (bool, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM server_subusers WHERE server_id = ? AND user_id = ?`, serverID, userID).Scan(&count)
	return count > 0, err
}

// --- Server Schedules ---

const scheduleColumns = `id, server_id, name, cron_expr, action, payload, is_active, last_run_at, next_run_at, created_at, updated_at, last_status, last_error, last_finished_at`

func scanSchedule(row interface{ Scan(...any) error }) (*ServerSchedule, error) {
	s := &ServerSchedule{}
	var active int
	var last, next, finished sql.NullTime
	err := row.Scan(&s.ID, &s.ServerID, &s.Name, &s.CronExpr, &s.Action, &s.Payload, &active, &last, &next, &s.CreatedAt, &s.UpdatedAt, &s.LastStatus, &s.LastError, &finished)
	if err != nil {
		return nil, err
	}
	s.IsActive = active == 1
	if last.Valid {
		s.LastRunAt = &last.Time
	}
	if next.Valid {
		s.NextRunAt = &next.Time
	}
	if finished.Valid {
		s.LastFinishedAt = &finished.Time
	}
	return s, nil
}

func (db *DB) listSchedules(where string, args ...any) ([]*ServerSchedule, error) {
	rows, err := db.conn.Query(`SELECT `+scheduleColumns+` FROM server_schedules WHERE `+where+` ORDER BY created_at ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*ServerSchedule{}
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (db *DB) ListServerSchedules(serverID string) ([]*ServerSchedule, error) {
	return db.listSchedules(`server_id = ?`, serverID)
}

func (db *DB) ListAllActiveSchedules() ([]*ServerSchedule, error) {
	return db.listSchedules(`is_active = 1`)
}

func (db *DB) GetServerSchedule(serverID, id string) (*ServerSchedule, error) {
	return scanSchedule(db.conn.QueryRow(`SELECT `+scheduleColumns+` FROM server_schedules WHERE server_id = ? AND id = ?`, serverID, id))
}

func (db *DB) GetServerScheduleByID(id string) (*ServerSchedule, error) {
	return scanSchedule(db.conn.QueryRow(`SELECT `+scheduleColumns+` FROM server_schedules WHERE id = ?`, id))
}

func (db *DB) CreateServerSchedule(s *ServerSchedule) error {
	if _, err := ParseCron(s.CronExpr); err != nil {
		return err
	}
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt
	_, err := db.conn.Exec(`INSERT INTO server_schedules (id, server_id, name, cron_expr, action, payload, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, s.ID, s.ServerID, s.Name, s.CronExpr, s.Action, s.Payload, s.IsActive, s.CreatedAt, s.UpdatedAt)
	return err
}

func (db *DB) UpdateServerSchedule(s *ServerSchedule) error {
	if _, err := ParseCron(s.CronExpr); err != nil {
		return err
	}
	s.UpdatedAt = time.Now().UTC()
	res, err := db.conn.Exec(`UPDATE server_schedules SET name = ?, cron_expr = ?, action = ?, payload = ?, is_active = ?, updated_at = ? WHERE server_id = ? AND id = ?`, s.Name, s.CronExpr, s.Action, s.Payload, s.IsActive, s.UpdatedAt, s.ServerID, s.ID)
	return affectedRow(res, err)
}

func affectedRow(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) DeleteServerSchedule(serverID, id string) error {
	res, err := db.conn.Exec(`DELETE FROM server_schedules WHERE server_id = ? AND id = ? AND last_status NOT IN ('started', 'running')`, serverID, id)
	return affectedRow(res, err)
}

func (db *DB) ToggleServerSchedule(serverID, id string, active bool) error {
	res, err := db.conn.Exec(`UPDATE server_schedules SET is_active = ?, updated_at = ? WHERE server_id = ? AND id = ?`, active, time.Now().UTC(), serverID, id)
	return affectedRow(res, err)
}

// ClaimScheduleRun is atomic across manual/timer executions and daemon instances.
func (db *DB) ClaimScheduleRun(id string, now time.Time, scheduled bool) (bool, error) {
	query := `UPDATE server_schedules SET last_status = 'started', last_error = '', last_run_at = ?, last_finished_at = NULL WHERE id = ? AND last_status NOT IN ('started', 'running')`
	args := []any{now.UTC(), id}
	if scheduled {
		query += ` AND is_active = 1 AND (last_run_at IS NULL OR last_run_at < ?)`
		args = append(args, now.UTC().Truncate(time.Minute))
	}
	res, err := db.conn.Exec(query, args...)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (db *DB) SetScheduleRunStatus(id, status, message string) error {
	var finished *time.Time
	if status == "succeeded" || status == "failed" {
		now := time.Now().UTC()
		finished = &now
	}
	_, err := db.conn.Exec(`UPDATE server_schedules SET last_status = ?, last_error = ?, last_finished_at = ? WHERE id = ?`, status, message, finished, id)
	return err
}
