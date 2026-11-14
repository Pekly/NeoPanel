package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func lifecycleAPIInvitation(t *testing.T, f *reliabilityFixture, name, node string) *database.Invitation {
	t.Helper()
	now := time.Now().UTC()
	i := &database.Invitation{ID: name, ServerID: "alpha", NodeID: node, Username: name, Email: name + "@example.test", Permissions: []string{"console.view"}, InviterID: f.users["owner"].ID, TokenHash: name + "hash", Status: "pending", DeliveryStatus: "sent", CreatedAt: now, LastSentAt: now, ExpiresAt: now.Add(time.Hour)}
	if err := f.db.CreateInvitation(i); err != nil {
		t.Fatal(err)
	}
	return i
}
func TestInvitationLifecycleAPICredentialsAndFailedRuntimeDeletion(t *testing.T) {
	for _, kind := range []string{"subuser", "server"} {
		t.Run(kind, func(t *testing.T) {
			f := newReliabilityFixture(t)
			i := lifecycleAPIInvitation(t, f, "created", "local")
			hash, err := f.router.auth.HashPassword("long-test-password")
			if err != nil {
				t.Fatal(err)
			}
			u, err := f.db.AcceptInvitation(i.ID, i.TokenHash, 0, hash, nil)
			if err != nil {
				t.Fatal(err)
			}
			token, err := f.router.auth.GenerateToken(u)
			if err != nil {
				t.Fatal(err)
			}
			f.tokens["created"] = token
			if _, err = f.router.auth.ValidateToken(token); err != nil {
				t.Fatal(err)
			}
			apiKey := "neo_lifecycle_test_key"
			sum := sha256.Sum256([]byte(apiKey))
			if _, err = f.db.AddAPIKey(u.ID, "test", hex.EncodeToString(sum[:]), "neo_life"); err != nil {
				t.Fatal(err)
			}
			if _, err = f.router.auth.ValidateAuthToken(apiKey); err != nil {
				t.Fatal(err)
			}
			path := "/api/servers/alpha"
			if kind == "server" {
				release, err := f.manager.BeginServerOperation("alpha")
				if err != nil {
					t.Fatal(err)
				}
				w := f.request("DELETE", path, "", "admin")
				release()
				if w.Code != 409 {
					t.Fatal(w.Code, w.Body.String())
				}
				if _, err = f.db.GetUserByID(u.ID); err != nil {
					t.Fatal("runtime failure deleted account", err)
				}
				if _, err = f.db.GetUserServerPermissions("alpha", u.ID); err != nil {
					t.Fatal("runtime failure removed membership", err)
				}
			} else {
				members, err := f.db.ListServerSubusers("alpha")
				if err != nil {
					t.Fatal(err)
				}
				for _, m := range members {
					if m.UserID == u.ID {
						path = fmt.Sprintf("%s/subusers/%d", path, m.ID)
					}
				}
			}
			w := f.request("DELETE", path, "", "admin")
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			if _, err = f.db.GetUserByID(u.ID); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("account survived", err)
			}
			if _, err = f.router.auth.ValidateToken(token); err == nil {
				t.Fatal("deleted user's JWT accepted")
			}
			if _, err = f.router.auth.ValidateAuthToken(apiKey); err == nil {
				t.Fatal("deleted user's API key accepted")
			}
			if _, err = f.router.auth.revalidateConsoleClaims(httptest.NewRequest("GET", "/api/servers/alpha/ws", nil), token, &JWTClaims{UserID: u.ID}); err == nil {
				t.Fatal("console revalidation accepted deleted user")
			}
			if w = f.request("GET", "/api/servers", "", "created"); w.Code != 401 {
				t.Fatal("deleted credentials accepted", w.Code, w.Body.String())
			}
			if _, err = f.db.GetUserByID(f.users["subuser"].ID); err != nil {
				t.Fatal("independent account deleted", err)
			}
		})
	}
}

func TestInvitationLifecycleProxyReconciliation(t *testing.T) {
	for _, kind := range []string{"subuser", "server"} {
		t.Run(kind, func(t *testing.T) {
			master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
			worker.cfg.Server.Role = "node"
			nodeHTTP := httptest.NewServer(worker.router)
			defer nodeHTTP.Close()
			node := &database.ClusterNode{ID: "remote", Name: "worker", BaseURL: nodeHTTP.URL, APIToken: worker.cfg.Security.NodeToken}
			if err := master.db.CreateNode(node); err != nil {
				t.Fatal(err)
			}
			i := lifecycleAPIInvitation(t, master, "remotecreated", "remote")
			if err := master.db.ReserveInvitationAccount(i.ID, i.TokenHash, 0, "hash"); err != nil {
				t.Fatal(err)
			}
			u, err := master.db.AcceptInvitation(i.ID, i.TokenHash, 0, "", func(i *database.Invitation, u, inviter *database.User) error {
				if err := worker.db.ProvisionInvitationUser(u); err != nil {
					return err
				}
				return worker.db.InvitationMembership(i.ID, i.ServerID, u.ID, worker.users["owner"].ID, "user", i.Permissions, i.ExpiresAt, false)
			})
			if err != nil {
				t.Fatal(err)
			}
			path := "/api/nodes/remote/proxy/api/servers/alpha"
			if kind == "subuser" {
				members, err := worker.db.ListServerSubusers("alpha")
				if err != nil {
					t.Fatal(err)
				}
				for _, m := range members {
					if m.UserID == u.ID {
						path = fmt.Sprintf("%s/subusers/%d", path, m.ID)
					}
				}
			}
			w := master.request("DELETE", path, "", "admin")
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			if w.Header().Get("X-NeoPanel-Invitation-Cleanup") != "reconciled" {
				t.Fatal(w.Header())
			}
			saved, err := master.db.GetInvitation(i.ID)
			if err != nil || saved.Status != "revoked" {
				t.Fatal(saved, err)
			}
			for _, f := range []*reliabilityFixture{master, worker} {
				if _, err = f.db.GetUserByID(u.ID); err != nil {
					t.Fatal("remote identity deleted without global proof", err)
				}
			}
			nodeHTTP.Close()
			w = master.request("POST", "/api/nodes/remote/invitation-removals/reconcile?server_id=alpha", "", "admin")
			if w.Code != 502 {
				t.Fatal(w.Code, w.Body.String())
			}
			if _, err = master.db.GetUserByID(u.ID); err != nil {
				t.Fatal("unreachable node caused deletion", err)
			}
		})
	}
}
