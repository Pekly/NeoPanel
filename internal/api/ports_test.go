package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/proxyconfig"
	"github.com/Pekly/NeoPanel/internal/runner"
)

func portAPIFixture(t *testing.T) *reliabilityFixture {
	t.Helper()
	f := newReliabilityFixture(t)
	// Never invoke the host firewall, on any test platform.
	f.router.servers.allowPort = func(int, string) {}
	f.router.servers.removePort = func(int) {}
	s, _ := f.db.GetServer("alpha")
	s.Port = 25565
	if err := f.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "server.properties"), []byte("server-port=25565\nquery.port=25565\nmotd=keep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	setBackendPolicy(t, f, true)
	return f
}
func portDirectRequest(f *reliabilityFixture, handler http.HandlerFunc, method, body, user string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/servers/alpha/port-summary", strings.NewReader(body))
	req.SetPathValue("id", "alpha")
	u := f.users[user]
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{UserID: u.ID, Role: u.Role, Username: u.Username}))
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}
func TestPortAPIQuotaRandomAndSummary(t *testing.T) {
	f := portAPIFixture(t)
	if w := f.request("POST", "/api/servers/alpha/ports", `{"port":42000}`, "admin"); w.Code < 400 {
		t.Fatal("default quota bypass", w.Body.String())
	}
	h := f.router.servers
	for _, user := range []string{"owner", "subuser"} {
		if w := portDirectRequest(f, h.SetServerPortAllowance, "PUT", `{"allowed":3}`, user); w.Code != 403 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := portDirectRequest(f, h.SetServerPortAllowance, "PUT", `{"allowed":3}`, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := portDirectRequest(f, h.SetPortPool, "PUT", `{"start":30001,"end":30001}`, "owner"); w.Code != 403 {
		t.Fatal(w.Code)
	}
	if w := portDirectRequest(f, h.SetPortPool, "PUT", `{"start":30001,"end":30001}`, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w := f.request("POST", "/api/servers/alpha/ports", `{"port":42000,"protocol":"udp"}`, "owner")
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var got struct {
		Port               int    `json:"port"`
		ListenerConfigured bool   `json:"listener_configured"`
		Message            string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Port != 30001 || got.ListenerConfigured || !strings.Contains(got.Message, "Multiple ports can be used simultaneously") {
		t.Fatal(got)
	}
	if w = f.request("POST", "/api/servers/alpha/ports", `{"port":42000}`, "admin"); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = f.request("POST", "/api/servers/alpha/ports", `{"port":42001}`, "admin"); w.Code < 400 {
		t.Fatal("quota bypassed")
	}
	w = f.request("GET", "/api/servers/alpha/ports", "", "owner")
	if w.Code != 200 || !strings.HasPrefix(w.Body.String(), "[") || w.Header().Get("X-Ports-Used") != "3" || w.Header().Get("X-Ports-Allowed") != "3" {
		t.Fatal(w.Code, w.Header(), w.Body.String())
	}
	w = portDirectRequest(f, h.GetServerPortSummary, "GET", "", "owner")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"used":3`) {
		t.Fatal(w.Body.String())
	}
}
func TestPortAPIGenericUpdateCannotBypassAndOwnershipAdminOnly(t *testing.T) {
	f := portAPIFixture(t)
	for _, user := range []string{"owner", "subuser", "admin"} {
		for _, body := range []string{`{"port":30001}`, `{"ports_allowed":3}`} {
			w := f.request("PUT", "/api/servers/alpha", body, user)
			if w.Code < 400 {
				t.Fatal(user, body, w.Code)
			}
		}
	}
	original := f.users["owner"].ID
	target := f.users["stranger"].ID
	for _, user := range []string{"owner", "subuser"} {
		w := f.request("PUT", "/api/servers/alpha", fmt.Sprintf(`{"owner_id":%d}`, target), user)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		s, _ := f.db.GetServer("alpha")
		if s.OwnerID != original || s.Port != 25565 || s.PortsAllowed != 1 {
			t.Fatal("member changed protected settings", s)
		}
	}
	if w := f.request("PUT", "/api/servers/alpha", fmt.Sprintf(`{"owner_id":%d}`, target), "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ := f.db.GetServer("alpha")
	if s.OwnerID != target {
		t.Fatal("admin transfer failed")
	}
}
func TestPortAPIPolicyAndPermission(t *testing.T) {
	f := portAPIFixture(t)
	if err := f.db.SetServerPortAllowance("alpha", 3); err != nil {
		t.Fatal(err)
	}
	if err := f.db.SetPortPool(database.PortPool{Start: 30001, End: 30001}); err != nil {
		t.Fatal(err)
	}
	setBackendPolicy(t, f, false)
	for _, user := range []string{"owner", "subuser"} {
		w := f.request("POST", "/api/servers/alpha/ports", `{"port":40000}`, user)
		if w.Code != 403 || !strings.Contains(w.Body.String(), "Node policy") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	setBackendPolicy(t, f, true)
	subs, _ := f.db.ListServerSubusers("alpha")
	if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, []string{"settings.view"}); err != nil {
		t.Fatal(err)
	}
	if w := f.request("POST", "/api/servers/alpha/ports", `{}`, "subuser"); w.Code != 403 {
		t.Fatal(w.Code)
	}
	// A trusted forwarded policy still overrides the worker's permissive local row.
	req := httptest.NewRequest("POST", "/api/servers/alpha/ports", strings.NewReader(`{}`))
	req.SetPathValue("id", "alpha")
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{UserID: f.users["owner"].ID, Role: "user", ForwardedNodePolicy: &nodeUserPolicy{AllowUserPorts: false}}))
	w := httptest.NewRecorder()
	f.router.servers.AddServerPort(w, req)
	if w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestPortAPIPrimaryLifecycleAndFailureAtomicity(t *testing.T) {
	f := portAPIFixture(t)
	if err := f.db.SetServerPortAllowance("alpha", 3); err != nil {
		t.Fatal(err)
	}
	release, err := f.manager.BeginServerOperation("alpha")
	if err != nil {
		t.Fatal(err)
	}
	w := f.request("POST", "/api/servers/alpha/ports", `{"port":30001,"is_primary":true}`, "admin")
	release()
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ := f.db.GetServer("alpha")
	if s.Port != 25565 || s.PortsUsed != 1 {
		t.Fatal(s)
	}
	w = f.request("POST", "/api/servers/alpha/ports", `{"port":30001,"is_primary":true}`, "admin")
	if w.Code != 201 || !strings.Contains(w.Body.String(), "next manual start") {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ = f.db.GetServer("alpha")
	raw, _ := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "server.properties"))
	if s.Port != 30001 || !strings.Contains(string(raw), "server-port=30001") || f.manager.GetStatus("alpha") != "offline" {
		t.Fatal(s, string(raw))
	}
	// A malformed proxy config fails without committing the new allocation.
	s.Type = "velocity"
	if err = f.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.manager.GetWorkingDir("alpha"), "velocity.toml")
	bad := []byte("bind = [ broken")
	if err = os.WriteFile(path, bad, 0600); err != nil {
		t.Fatal(err)
	}
	w = f.request("POST", "/api/servers/alpha/ports", `{"port":30002,"is_primary":true}`, "admin")
	if w.Code < 400 {
		t.Fatal(w.Code)
	}
	s, _ = f.db.GetServer("alpha")
	raw, _ = os.ReadFile(path)
	if s.Port != 30001 || s.PortsUsed != 2 || string(raw) != string(bad) {
		t.Fatal(s, string(raw))
	}
}
func TestPortAPIExactConfigRollbackAndSteelSync(t *testing.T) {
	for _, kind := range []string{"paper", "steel", "velocity", "bungeecord", "waterfall"} {
		t.Run(kind, func(t *testing.T) {
			f := portAPIFixture(t)
			s, _ := f.db.GetServer("alpha")
			s.Type = kind
			if err := f.db.UpdateServer(s); err != nil {
				t.Fatal(err)
			}
			dir := f.manager.GetWorkingDir("alpha")
			name := "server.properties"
			if kind == "steel" {
				name = filepath.Join("config", "config.toml")
				if err := os.MkdirAll(filepath.Join(dir, "config"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name), []byte("[server]\nserver_port = 25565\nother = true\n"), 0600); err != nil {
					t.Fatal(err)
				}
			} else if proxyconfig.Supported(kind) {
				name = proxyconfig.FileName(kind)
				if _, err := proxyconfig.Ensure(dir, kind, 25565); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("commit failed")
			req := httptest.NewRequest("POST", "/", nil)
			err = f.router.servers.withPrimaryPortChange(req, "alpha", func(apply database.PortConfigApply) error {
				if err := f.manager.Start(s); !errors.Is(err, runner.ErrServerOperation) {
					t.Fatalf("power gate not held: %v", err)
				}
				if err := apply(30001); err != nil {
					return err
				}
				updated, _ := os.ReadFile(filepath.Join(dir, name))
				if !strings.Contains(string(updated), "30001") {
					t.Fatal("port not applied", string(updated))
				}
				return injected
			})
			if !errors.Is(err, injected) {
				t.Fatal(err)
			}
			after, _ := os.ReadFile(filepath.Join(dir, name))
			if string(before) != string(after) {
				t.Fatal("rollback did not restore exact bytes")
			}
			release, err := f.manager.BeginServerOperation("alpha")
			if err != nil {
				t.Fatal("gate not released", err)
			}
			release()
		})
	}
}
func TestPortAPICreationCollisionLeavesNoServer(t *testing.T) {
	f := portAPIFixture(t)
	w := f.request("POST", "/api/servers", `{"name":"collision","type":"custom","port":25565}`, "admin")
	if w.Code < 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	all, err := f.db.ListServers()
	if err != nil || len(all) != 2 {
		t.Fatal(len(all), err)
	}
	entries, err := os.ReadDir(f.cfg.Paths.ServersDir)
	if err != nil || len(entries) != 2 {
		t.Fatal("orphaned creation directory", len(entries), err)
	}
}

func TestPortAPISetPrimaryAndMissingVelocityRollback(t *testing.T) {
	f := portAPIFixture(t)
	if err := f.db.SetServerPortAllowance("alpha", 2); err != nil {
		t.Fatal(err)
	}
	p, err := f.db.AddServerPort("alpha", 30001, "tcp", "", false)
	if err != nil {
		t.Fatal(err)
	}
	w := f.request("POST", fmt.Sprintf("/api/servers/alpha/ports/%d/primary", p.ID), "", "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ := f.db.GetServer("alpha")
	if s.Port != 30001 || s.PortsUsed != 2 {
		t.Fatal(s)
	}
	w = f.request("DELETE", fmt.Sprintf("/api/servers/alpha/ports/%d", p.ID), "", "owner")
	if w.Code < 400 {
		t.Fatal("deleted primary")
	}
	s.Type = "velocity"
	if err := f.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("commit failed after initialize")
	err = f.router.servers.withPrimaryPortChange(httptest.NewRequest("POST", "/", nil), "alpha", func(apply database.PortConfigApply) error {
		if err := apply(30002); err != nil {
			return err
		}
		return injected
	})
	if !errors.Is(err, injected) {
		t.Fatal(err)
	}
	for _, name := range []string{"velocity.toml", "forwarding.secret"} {
		if _, err := os.Stat(filepath.Join(f.manager.GetWorkingDir("alpha"), name)); !os.IsNotExist(err) {
			t.Fatal("rollback left initialized file", name, err)
		}
	}
}

func TestPortAPIJavaAndBedrockAllocationsCoexist(t *testing.T) {
	f := portAPIFixture(t)
	if err := f.db.SetServerPortAllowance("alpha", 2); err != nil {
		t.Fatal(err)
	}
	properties := filepath.Join(f.manager.GetWorkingDir("alpha"), "server.properties")
	before, err := os.ReadFile(properties)
	if err != nil {
		t.Fatal(err)
	}
	var requested []database.ServerPort
	f.router.servers.allowPort = func(port int, protocol string) {
		requested = append(requested, database.ServerPort{Port: port, Protocol: protocol})
	}
	w := f.request("POST", "/api/servers/alpha/ports", `{"port":19132,"protocol":"udp","notes":"Bedrock / Geyser"}`, "admin")
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body.String())
	}
	if len(requested) != 1 || requested[0].Port != 19132 || requested[0].Protocol != "udp" {
		t.Fatalf("wrong host firewall request: %+v", requested)
	}
	srv, err := f.db.GetServer("alpha")
	if err != nil || srv.Port != 25565 || srv.PortsUsed != 2 {
		t.Fatalf("primary listener changed: %+v, %v", srv, err)
	}
	after, err := os.ReadFile(properties)
	if err != nil || string(after) != string(before) {
		t.Fatalf("extra UDP allocation changed Java configuration: %s, %v", after, err)
	}
	ports, err := f.db.ListServerPorts("alpha")
	if err != nil || len(ports) != 2 {
		t.Fatalf("ports: %+v, %v", ports, err)
	}
	var primary, bedrock bool
	for _, p := range ports {
		primary = primary || (p.Port == 25565 && p.IsPrimary)
		bedrock = bedrock || (p.Port == 19132 && p.Protocol == "udp" && !p.IsPrimary)
	}
	if !primary || !bedrock {
		t.Fatalf("both allocations must remain assigned: %+v", ports)
	}
	// A second server on the same node cannot claim the assigned extra port.
	w = f.request("POST", "/api/servers", `{"name":"port-collision","type":"custom","port":19132}`, "admin")
	if w.Code < 400 {
		t.Fatal("another server claimed the UDP allocation", w.Body.String())
	}
	if f.manager.GetStatus("alpha") != "offline" {
		t.Fatal("allocating an extra port started the server")
	}
}
