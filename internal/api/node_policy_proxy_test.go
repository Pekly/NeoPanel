package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestRemoteNodePolicyEnforcedWithoutWorkerPolicyRow(t *testing.T) {
	for _, operation := range []string{"ram", "args", "java", "port", "add", "delete", "primary"} {
		for _, caller := range []string{"owner", "subuser", "admin", "permitted-subuser"} {
			t.Run(operation+"/"+caller, func(t *testing.T) {
				master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
				if node, _ := worker.db.GetNode("local"); node != nil {
					t.Fatal("worker unexpectedly has a local policy")
				}
				subs, _ := worker.db.ListServerSubusers("alpha")
				if err := worker.db.UpdateServerSubuser("alpha", subs[0].ID, []string{"settings.edit"}); err != nil {
					t.Fatal(err)
				}
				if _, err := worker.db.AddServerPort("alpha", 25565, "tcp", "primary", true); err != nil {
					t.Fatal(err)
				}
				secondary, err := worker.db.AddServerPort("alpha", 25566, "tcp", "secondary", false)
				if err != nil {
					t.Fatal(err)
				}
				properties := filepath.Join(worker.manager.GetWorkingDir("alpha"), "server.properties")
				if err := os.WriteFile(properties, []byte("server-port=25565\n"), 0600); err != nil {
					t.Fatal(err)
				}
				firewallCalls := 0
				worker.router.servers.allowPort = func(int, string) { firewallCalls++ }
				worker.router.servers.removePort = func(int) { firewallCalls++ }
				allow := caller == "permitted-subuser"
				user := caller
				if allow {
					user = "subuser"
				}
				// Inspect only the actual mutation, not the existing identity probe.
				seenPolicy := make(chan string, 1)
				remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/api/auth/me" {
						seenPolicy <- r.Header.Get(forwardedNodePolicyHeader)
					}
					worker.router.ServeHTTP(w, r)
				}))
				defer remote.Close()
				node := &database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken, AllowUserRAM: allow, AllowUserPorts: allow, AllowUserJava: allow}
				if err := master.db.CreateNode(node); err != nil {
					t.Fatal(err)
				}
				method, path, body, success := "PUT", "/api/servers/alpha", `{"memory_max":4096}`, 200
				switch operation {
				case "args":
					body = `{"java_args":"-Xmx8G"}`
				case "java":
					body = `{"java_path":"java-new"}`
				case "port":
					body = `{"port":25567}`
				case "add":
					method, path, body, success = "POST", path+"/ports", `{"port":25567,"protocol":"tcp","is_primary":true}`, 201
				case "delete":
					method, path, body = "DELETE", fmt.Sprintf("%s/ports/%d", path, secondary.ID), ""
				case "primary":
					method, path, body = "POST", fmt.Sprintf("%s/ports/%d/primary", path, secondary.ID), ""
				}
				beforeServer, _ := worker.db.GetServer("alpha")
				beforePorts, _ := worker.db.ListServerPorts("alpha")
				w := master.request(method, "/api/nodes/remote/proxy"+path, body, user)
				want := 403
				if user == "admin" || allow {
					want = success
				}
				if w.Code != want {
					t.Fatal(w.Code, w.Body.String())
				}
				select {
				case raw := <-seenPolicy:
					if user == "admin" {
						if raw != "" {
							t.Fatal("admin unexpectedly carries tenant policy", raw)
						}
					} else {
						var policy nodeUserPolicy
						if err := json.Unmarshal([]byte(raw), &policy); err != nil || !reflect.DeepEqual(&policy, userPolicyForNode(node)) {
							t.Fatal("wrong forwarded policy", raw, err)
						}
					}
				default:
					t.Fatal("request never reached worker")
				}
				if want == 403 {
					if !strings.Contains(w.Body.String(), "Node policy") {
						t.Fatal("denied for wrong reason", w.Body.String())
					}
					afterServer, _ := worker.db.GetServer("alpha")
					afterPorts, _ := worker.db.ListServerPorts("alpha")
					data, err := os.ReadFile(properties)
					if err != nil || string(data) != "server-port=25565\n" || !reflect.DeepEqual(beforeServer, afterServer) || !reflect.DeepEqual(beforePorts, afterPorts) || firewallCalls != 0 {
						t.Fatal("denied mutation changed state", err, firewallCalls)
					}
				}
			})
		}
	}
}

func TestForwardedNodePolicyTrustBoundary(t *testing.T) {
	f := newReliabilityFixture(t)
	setBackendPolicy(t, f, false)
	permissive := `{"allow_user_ram":true,"allow_user_ports":true,"allow_user_java":true}`
	for _, path := range []string{"/api/servers/alpha", "/api/nodes/remote/proxy/api/servers/alpha"} {
		for _, identity := range []bool{false, true} {
			for _, caller := range []string{"owner", "subuser"} {
				req := httptest.NewRequest("PUT", path, strings.NewReader(`{"java_args":"-Xmx8G"}`))
				req.Header.Set("Authorization", "Bearer "+f.tokens[caller])
				req.Header.Set(forwardedNodePolicyHeader, permissive)
				if identity {
					req.Header.Set("X-NeoPanel-User-ID", fmt.Sprint(f.users[caller].ID))
					req.Header.Set("X-NeoPanel-User-Role", "admin")
				}
				w := httptest.NewRecorder()
				f.router.ServeHTTP(w, req)
				if w.Code != 403 {
					t.Fatal("spoofed policy accepted", path, caller, identity, w.Code, w.Body.String())
				}
			}
		}
	}
	server, _ := f.db.GetServer("alpha")
	if server.JavaArgs != "-Dkeep=true" {
		t.Fatal("spoof changed args")
	}
	for _, role := range []string{"user", "admin"} {
		headers := make(http.Header)
		headers.Set(forwardedNodePolicyHeader, permissive)
		setProxyIdentity(headers, &JWTClaims{UserID: f.users["owner"].ID, Role: role}, "node-token")
		if headers.Get(forwardedNodePolicyHeader) != "" {
			t.Fatal("proxy retained client policy")
		}
	}
	// Neither serialization nor parsing JWT claims may persist the trusted field.
	claims := JWTClaims{ForwardedNodePolicy: &nodeUserPolicy{true, true, true}}
	encoded, err := json.Marshal(claims)
	if err != nil || strings.Contains(string(encoded), "allow_user") {
		t.Fatal(string(encoded), err)
	}
	var parsed JWTClaims
	if err := json.Unmarshal([]byte(`{"ForwardedNodePolicy":`+permissive+`,"forwarded_node_policy":`+permissive+`}`), &parsed); err != nil || parsed.ForwardedNodePolicy != nil {
		t.Fatal("policy parsed from JWT", err)
	}
}

func TestLegacyWorkerPolicyFallbackAndMalformedForwarding(t *testing.T) {
	f := newReliabilityFixture(t)
	setBackendPolicy(t, f, false)
	request := func(policy string, withPolicy bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest("PUT", "/api/servers/alpha", strings.NewReader(`{"java_args":"-Xmx8G"}`))
		req.Header.Set("Authorization", "Bearer "+f.cfg.Security.NodeToken)
		req.Header.Set("X-NeoPanel-User-ID", fmt.Sprint(f.users["owner"].ID))
		req.Header.Set("X-NeoPanel-User-Role", "user")
		if withPolicy {
			req.Header.Set(forwardedNodePolicyHeader, policy)
		}
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, req)
		return w
	}
	// Older master carries identity only: use worker's local node restrictions.
	if w := request("", false); w.Code != 403 || !strings.Contains(w.Body.String(), "Node policy") {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, bad := range []string{"{", "null", "[]", `{"allow_user_ram":"true"}`, strings.Repeat("x", 257)} {
		if w := request(bad, true); w.Code != 403 || !strings.Contains(w.Body.String(), "invalid forwarded node policy") {
			t.Fatal(bad, w.Code, w.Body.String())
		}
	}
	if err := f.db.DeleteNode("local"); err != nil {
		t.Fatal(err)
	}
	// Preserve direct/legacy behavior when the worker has no configured policy.
	if w := request("", false); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := f.request("PUT", "/api/servers/alpha", `{"java_args":"-Ddirect=true"}`, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	// A trusted explicit permissive master policy overrides a restrictive worker row.
	setBackendPolicy(t, f, false)
	if w := request(`{"allow_user_ram":true,"allow_user_ports":true,"allow_user_java":true}`, true); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
