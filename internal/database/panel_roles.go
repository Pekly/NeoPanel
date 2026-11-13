package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const PermissionServersCreate = "servers.create"

var (
	ErrInvalidPanelRole  = errors.New("invalid panel role")
	ErrPanelRoleAssigned = errors.New("panel role is assigned to users")
	ErrPanelRoleConflict = errors.New("panel role name already exists")
)

type PanelRole struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Color       string   `json:"color,omitempty"`
	Permissions []string `json:"permissions"`
	Builtin     bool     `json:"builtin,omitempty"`
}

func (db *DB) migratePanelRoles() error {
	// Triggers enforce referential integrity even on legacy connections where
	// SQLite foreign_keys is disabled. Deletion and assignment are atomic.
	_, err := db.conn.Exec(`
 CREATE TABLE IF NOT EXISTS panel_roles (
  id TEXT PRIMARY KEY, name TEXT NOT NULL COLLATE NOCASE UNIQUE,
  description TEXT NOT NULL DEFAULT '', color TEXT NOT NULL DEFAULT '',
  permissions TEXT NOT NULL DEFAULT '[]'
 );
 CREATE TABLE IF NOT EXISTS panel_user_roles (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  role_id TEXT NOT NULL REFERENCES panel_roles(id) ON DELETE RESTRICT
 );
 CREATE INDEX IF NOT EXISTS idx_panel_user_roles_role ON panel_user_roles(role_id);
 CREATE TRIGGER IF NOT EXISTS panel_role_delete_guard BEFORE DELETE ON panel_roles
 WHEN EXISTS (SELECT 1 FROM panel_user_roles WHERE role_id = OLD.id)
 BEGIN SELECT RAISE(ABORT, 'panel role assigned'); END;
 CREATE TRIGGER IF NOT EXISTS panel_user_roles_insert_guard BEFORE INSERT ON panel_user_roles
 WHEN NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id)
 OR NOT EXISTS (SELECT 1 FROM panel_roles WHERE id = NEW.role_id)
 BEGIN SELECT RAISE(ABORT, 'invalid panel role assignment'); END;
 CREATE TRIGGER IF NOT EXISTS panel_user_roles_update_guard BEFORE UPDATE ON panel_user_roles
 WHEN NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id)
 OR NOT EXISTS (SELECT 1 FROM panel_roles WHERE id = NEW.role_id)
 BEGIN SELECT RAISE(ABORT, 'invalid panel role assignment'); END;
 CREATE TRIGGER IF NOT EXISTS panel_user_roles_cleanup AFTER DELETE ON users
 BEGIN DELETE FROM panel_user_roles WHERE user_id = OLD.id; END;
 `)
	return err
}

var panelRoleColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func ValidatePanelRole(role *PanelRole) error {
	role.Name = strings.TrimSpace(role.Name)
	role.Description = strings.TrimSpace(role.Description)
	if !utf8.ValidString(role.Name) || utf8.RuneCountInString(role.Name) < 1 || utf8.RuneCountInString(role.Name) > 64 {
		return fmt.Errorf("%w: name must contain 1-64 characters", ErrInvalidPanelRole)
	}
	if strings.EqualFold(role.Name, "admin") || strings.EqualFold(role.Name, "user") {
		return fmt.Errorf("%w: built-in names are reserved", ErrInvalidPanelRole)
	}
	for _, c := range role.Name {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != ' ' && c != '-' && c != '_' {
			return fmt.Errorf("%w: name may only contain letters, numbers, spaces, hyphens and underscores", ErrInvalidPanelRole)
		}
	}
	if !utf8.ValidString(role.Description) || utf8.RuneCountInString(role.Description) > 500 {
		return fmt.Errorf("%w: description exceeds 500 characters", ErrInvalidPanelRole)
	}
	for _, c := range role.Description {
		if unicode.IsControl(c) && c != '\n' && c != '\t' {
			return fmt.Errorf("%w: invalid description", ErrInvalidPanelRole)
		}
	}
	if role.Color != "" && !panelRoleColor.MatchString(role.Color) {
		return fmt.Errorf("%w: color must be #RRGGBB", ErrInvalidPanelRole)
	}
	if err := validatePanelPermissions(role.Permissions); err != nil {
		return err
	}
	if role.Permissions == nil {
		role.Permissions = []string{}
	}
	return nil
}

func validatePanelPermissions(permissions []string) error {
	seen := map[string]bool{}
	for _, p := range permissions {
		if p != PermissionServersCreate || seen[p] {
			return fmt.Errorf("%w: unsupported or duplicate permission %q", ErrInvalidPanelRole, p)
		}
		seen[p] = true
	}
	return nil
}

func (db *DB) ListPanelRoles() ([]*PanelRole, error) {
	rows, err := db.conn.Query("SELECT id, name, description, color, permissions FROM panel_roles ORDER BY name COLLATE NOCASE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []*PanelRole{}
	for rows.Next() {
		role, err := scanPanelRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func scanPanelRole(row interface{ Scan(...interface{}) error }) (*PanelRole, error) {
	role := &PanelRole{}
	var raw string
	if err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Color, &raw); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(raw), &role.Permissions); err != nil {
		return nil, err
	}
	if err := validatePanelPermissions(role.Permissions); err != nil {
		return nil, err
	}
	if role.Permissions == nil {
		role.Permissions = []string{}
	}
	return role, nil
}

func (db *DB) GetPanelRole(id string) (*PanelRole, error) {
	return scanPanelRole(db.conn.QueryRow("SELECT id, name, description, color, permissions FROM panel_roles WHERE id = ?", id))
}

func panelRoleWriteError(err error) error {
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: panel_roles.name") {
		return ErrPanelRoleConflict
	}
	return err
}

func (db *DB) CreatePanelRole(role *PanelRole) error {
	if err := ValidatePanelRole(role); err != nil {
		return err
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return err
	}
	role.ID = "role-" + hex.EncodeToString(id)
	role.Builtin = false
	permissions, _ := json.Marshal(role.Permissions)
	_, err := db.conn.Exec("INSERT INTO panel_roles (id, name, description, color, permissions) VALUES (?, ?, ?, ?, ?)", role.ID, role.Name, role.Description, role.Color, string(permissions))
	return panelRoleWriteError(err)
}

func (db *DB) UpdatePanelRole(role *PanelRole) error {
	if err := ValidatePanelRole(role); err != nil {
		return err
	}
	permissions, _ := json.Marshal(role.Permissions)
	result, err := db.conn.Exec("UPDATE panel_roles SET name = ?, description = ?, color = ?, permissions = ? WHERE id = ?", role.Name, role.Description, role.Color, string(permissions), role.ID)
	if err != nil {
		return panelRoleWriteError(err)
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	role.Builtin = false
	return err
}

func (db *DB) DeletePanelRole(id string) error {
	result, err := db.conn.Exec("DELETE FROM panel_roles WHERE id = ?", id)
	if err != nil {
		if strings.Contains(err.Error(), "panel role assigned") {
			return ErrPanelRoleAssigned
		}
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) AssignPanelRole(userID int, roleID string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int
	if err := tx.QueryRow("SELECT id FROM users WHERE id = ?", userID).Scan(&id); err != nil {
		return err
	}
	if roleID == "" {
		_, err = tx.Exec("DELETE FROM panel_user_roles WHERE user_id = ?", userID)
	} else {
		var role string
		if err := tx.QueryRow("SELECT id FROM panel_roles WHERE id = ?", roleID).Scan(&role); err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO panel_user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT(user_id) DO UPDATE SET role_id = excluded.role_id", userID, roleID)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// loadUserPanelRole always resolves current effective capabilities, never token data.
func (db *DB) loadUserPanelRole(user *User) (*User, error) {
	var raw string
	err := db.conn.QueryRow(`SELECT u.role, COALESCE(pr.id, ''), COALESCE(pr.name, ''), COALESCE(pr.permissions, '[]')
 FROM users u LEFT JOIN panel_user_roles pur ON pur.user_id = u.id
 LEFT JOIN panel_roles pr ON pr.id = pur.role_id WHERE u.id = ?`, user.ID).
		Scan(&user.Role, &user.PanelRoleID, &user.PanelRoleName, &raw)
	if err != nil {
		return nil, err
	}
	user.PanelPermissions = []string{}
	if err := json.Unmarshal([]byte(raw), &user.PanelPermissions); err != nil {
		return nil, err
	}
	if err := validatePanelPermissions(user.PanelPermissions); err != nil {
		return nil, err
	}
	if user.PanelPermissions == nil {
		user.PanelPermissions = []string{}
	}
	if user.Role == "admin" {
		user.PanelPermissions = []string{PermissionServersCreate}
	}
	return user, nil
}

func (db *DB) UserCanCreateServer(userID int) (bool, error) {
	user, err := db.loadUserPanelRole(&User{ID: userID})
	if err != nil {
		return false, err
	}
	for _, p := range user.PanelPermissions {
		if p == PermissionServersCreate {
			return true, nil
		}
	}
	return false, nil
}
