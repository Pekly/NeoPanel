package database

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestPanelRolesPersistenceRevocationAndCleanup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roles.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	admin, err := db.CreateUser("admin", "unused", "admin")
	if err != nil {
		t.Fatal(err)
	}
	member, err := db.CreateUser("member", "unused", "user")
	if err != nil {
		t.Fatal(err)
	}
	check := func(id int, want bool) {
		t.Helper()
		got, err := db.UserCanCreateServer(id)
		if err != nil || got != want {
			t.Fatalf("create(%d) = %v, %v; want %v", id, got, err, want)
		}
	}
	check(admin.ID, true)
	check(member.ID, false)
	role := &PanelRole{Name: "Creators", Description: "Can create", Color: "#abcdef", Permissions: []string{PermissionServersCreate}}
	if err := db.CreatePanelRole(role); err != nil {
		t.Fatal(err)
	}
	if err := db.AssignPanelRole(member.ID, role.ID); err != nil {
		t.Fatal(err)
	}
	check(member.ID, true)
	if err := db.DeletePanelRole(role.ID); !errors.Is(err, ErrPanelRoleAssigned) {
		t.Fatalf("delete assigned: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	check(member.ID, true)
	loaded, err := db.GetUserByID(member.ID)
	if err != nil || loaded.Role != "user" || loaded.PanelRoleID != role.ID || loaded.PanelRoleName != role.Name {
		t.Fatalf("loaded: %+v %v", loaded, err)
	}
	roles, err := db.ListPanelRoles()
	if err != nil || len(roles) != 1 || roles[0].Color != role.Color {
		t.Fatalf("roles: %+v %v", roles, err)
	}
	role.Permissions = []string{}
	if err := db.UpdatePanelRole(role); err != nil {
		t.Fatal(err)
	}
	check(member.ID, false)
	role.Permissions = []string{PermissionServersCreate}
	if err := db.UpdatePanelRole(role); err != nil {
		t.Fatal(err)
	}
	check(member.ID, true)
	if err := db.AssignPanelRole(member.ID, ""); err != nil {
		t.Fatal(err)
	}
	check(member.ID, false)
	if err := db.AssignPanelRole(member.ID, role.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.DeleteUser(member.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM panel_user_roles").Scan(&count); err != nil || count != 0 {
		t.Fatalf("orphan assignments: %d %v", count, err)
	}
	if err := db.DeletePanelRole(role.ID); err != nil {
		t.Fatal(err)
	}
	if allowed, err := db.UserCanCreateServer(member.ID); allowed || err == nil {
		t.Fatalf("deleted account allowed: %v %v", allowed, err)
	}
}

func TestPanelRoleValidationAndFailClosed(t *testing.T) {
	db, err := InitDB(filepath.Join(t.TempDir(), "roles.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, role := range []PanelRole{
		{Name: ""}, {Name: "AdMiN"}, {Name: "user"}, {Name: "<script>"}, {Name: strings.Repeat("a", 65)},
		{Name: "Good", Color: "red"}, {Name: "Good", Description: strings.Repeat("a", 501)},
		{Name: "Good", Permissions: []string{"admin"}}, {Name: "Good", Permissions: []string{"files.view"}},
		{Name: "Good", Permissions: []string{PermissionServersCreate, PermissionServersCreate}},
	} {
		if err := db.CreatePanelRole(&role); !errors.Is(err, ErrInvalidPanelRole) {
			t.Errorf("accepted %+v: %v", role, err)
		}
	}
	role := &PanelRole{Name: "Good", Permissions: []string{PermissionServersCreate}}
	if err := db.CreatePanelRole(role); err != nil {
		t.Fatal(err)
	}
	if err := db.CreatePanelRole(&PanelRole{Name: "good"}); !errors.Is(err, ErrPanelRoleConflict) {
		t.Fatalf("duplicate: %v", err)
	}
	user, err := db.CreateUser("member", "unused", "user")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AssignPanelRole(999, role.ID); err == nil {
		t.Fatal("assigned missing user")
	}
	if err := db.AssignPanelRole(user.ID, "admin"); err == nil {
		t.Fatal("assigned built-in as custom")
	}
	if err := db.AssignPanelRole(user.ID, role.ID); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`broken`, `["admin"]`, `["servers.create","files.view"]`} {
		if _, err := db.conn.Exec("UPDATE panel_roles SET permissions = ? WHERE id = ?", raw, role.ID); err != nil {
			t.Fatal(err)
		}
		if allowed, err := db.UserCanCreateServer(user.ID); allowed || err == nil {
			t.Fatalf("corrupt permissions allowed: %v %v", allowed, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if allowed, err := db.UserCanCreateServer(user.ID); allowed || err == nil {
		t.Fatalf("storage error allowed: %v %v", allowed, err)
	}
}
