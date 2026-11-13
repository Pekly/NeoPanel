package api

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestSFTPDetailsRequireExplicitPermission(t *testing.T) {
	f := newReliabilityFixture(t)
	if w := f.request("GET", "/api/servers/alpha/sftp", "", "subuser"); w.Code != 403 {
		t.Fatal("files.view implied SFTP access", w.Code, w.Body.String())
	}
	for _, user := range []string{"owner", "admin"} {
		if w := f.request("GET", "/api/servers/alpha/sftp", "", user); w.Code != 200 {
			t.Fatal(user, w.Code, w.Body.String())
		}
	}
	members, err := f.db.ListServerSubusers("alpha")
	if err != nil || len(members) != 1 {
		t.Fatal(members, err)
	}
	if err := f.db.UpdateServerSubuser("alpha", members[0].ID, []string{"sftp.access", "files.view"}); err != nil {
		t.Fatal(err)
	}
	if w := f.request("GET", "/api/servers/alpha/sftp", "", "subuser"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := f.db.UpdateServerSubuser("alpha", members[0].ID, []string{"files.view"}); err != nil {
		t.Fatal(err)
	}
	if w := f.request("GET", "/api/servers/alpha/sftp", "", "subuser"); w.Code != 403 {
		t.Fatal("revocation did not apply", w.Code)
	}
	if permission := serverActionPermission("GET", "sftp"); permission != "sftp.access" {
		t.Fatal(permission)
	}
}
func TestTeamManagersCannotGrantUnheldPermissions(t *testing.T) {
	f := newReliabilityFixture(t)
	members, err := f.db.ListServerSubusers("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateServerSubuser("alpha", members[0].ID, []string{"console.view", "subusers.manage"}); err != nil {
		t.Fatal(err)
	}
	if w := f.request("POST", "/api/servers/alpha/subusers", `{"username":"stranger","permissions":["sftp.access"]}`, "subuser"); w.Code != 403 {
		t.Fatal("manager escalated another member", w.Code, w.Body.String())
	}
	path := fmt.Sprintf("/api/servers/alpha/subusers/%d", members[0].ID)
	if w := f.request("PUT", path, `{"permissions":["console.view","subusers.manage","sftp.access"]}`, "subuser"); w.Code != 403 {
		t.Fatal("manager escalated self", w.Code, w.Body.String())
	}
	if w := f.request("POST", "/api/servers/alpha/subusers", `{"username":"stranger","permissions":["console.view"]}`, "subuser"); w.Code != 201 {
		t.Fatal("valid delegated grant rejected", w.Code, w.Body.String())
	}
	if w := f.request("PUT", path, `{"permissions":["made.up"]}`, "owner"); w.Code != 400 {
		t.Fatal("unknown permission accepted", w.Code, w.Body.String())
	}
	if w := f.request("PUT", path, `{"permissions":["sftp.access","files.view"]}`, "owner"); w.Code != 200 {
		t.Fatal("owner cannot grant SFTP", w.Code, w.Body.String())
	}
}
func TestInvitationGrantRechecksInviterPermissions(t *testing.T) {
	f := newReliabilityFixture(t)
	members, _ := f.db.ListServerSubusers("alpha")
	if err := f.db.UpdateServerSubuser("alpha", members[0].ID, []string{"subusers.manage", "console.view"}); err != nil {
		t.Fatal(err)
	}
	user := f.users["subuser"]
	claims := &JWTClaims{UserID: user.ID, Username: user.Username, Role: user.Role}
	request := httptest.NewRequest("GET", "http://panel.example", nil)
	if _, err := f.router.invitations.authorize(request, "alpha", nil, claims, []string{"sftp.access"}); err == nil {
		t.Fatal("invitation bypassed grant limits")
	}
	if _, err := f.router.invitations.authorize(request, "alpha", nil, claims, []string{"console.view"}); err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateServerSubuser("alpha", members[0].ID, []string{"subusers.manage"}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.router.invitations.authorize(request, "alpha", nil, claims, []string{"console.view"}); err == nil {
		t.Fatal("revoked inviter permission still grantable")
	}
	if err := validateInvitationPermissions([]string{"sftp.access"}); err != nil {
		t.Fatal(err)
	}
}
