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

