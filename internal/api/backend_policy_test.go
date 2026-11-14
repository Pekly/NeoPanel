package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func setBackendPolicy(t *testing.T, f *reliabilityFixture, allow bool) {
	t.Helper()
	node, err := f.db.GetNode("local")
	if err != nil {
		node = &database.ClusterNode{ID: "local", Name: "local"}
		if err := f.db.CreateNode(node); err != nil {
			t.Fatal(err)
		}
	}
	node.AllowUserRAM, node.AllowUserPorts, node.AllowUserJava = allow, allow, true
	if err := f.db.UpdateNode(node); err != nil {
		t.Fatal(err)
	}
	subs, err := f.db.ListServerSubusers("alpha")
	if err != nil || len(subs) != 1 {
		t.Fatal(subs, err)
	}
	if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, []string{"settings.edit"}); err != nil {
		t.Fatal(err)
	}
}

func TestLockedRAMRejectsJVMArgumentChanges(t *testing.T) {
	f := newReliabilityFixture(t)
	setBackendPolicy(t, f, false)
	before, err := f.db.GetServer("alpha")
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"owner", "subuser"} {
		for _, args := range []string{"-Xmx64G", "-XX:MaxHeapSize=64G", "@tenant.args", "-XX:VMOptionsFile=tenant.args", "-Dbenign=true", ""} {
			body, _ := json.Marshal(map[string]string{"java_args": args, "name": "must not change"})
			w := f.request("PUT", "/api/servers/alpha", string(body), user)
			if w.Code != 403 || !strings.Contains(w.Body.String(), "memory is locked") {
				t.Fatalf("%s %q: %d %s", user, args, w.Code, w.Body.String())
			}
			after, err := f.db.GetServer("alpha")
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("denial changed server", after, err)
			}
		}
		for _, body := range []string{`{"memory_max":8192}`, `{"memory_min":8192}`} {
			if w := f.request("PUT", "/api/servers/alpha", body, user); w.Code != 403 {
				t.Fatal(w.Code, w.Body.String())
			}
		}
		// Existing settings forms may echo unchanged JVM flags.
		if w := f.request("PUT", "/api/servers/alpha", `{"java_args":"-Dkeep=true"}`, user); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		before, _ = f.db.GetServer("alpha")
	}
	if w := f.request("PUT", "/api/servers/alpha", `{"java_args":"-Xmx64G"}`, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	setBackendPolicy(t, f, true)
	for _, user := range []string{"owner", "subuser"} {
		body := fmt.Sprintf(`{"java_args":"-Duser=%s"}`, user)
		if w := f.request("PUT", "/api/servers/alpha", body, user); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

func TestPortMutationNodePolicy(t *testing.T) {
	for _, operation := range []string{"add", "delete", "primary"} {
		for _, allow := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/allow=%v", operation, allow), func(t *testing.T) {
				f := newReliabilityFixture(t)
				setBackendPolicy(t, f, allow)
				if _, err := f.db.AddServerPort("alpha", 25565, "tcp/udp", "primary", true); err != nil {
					t.Fatal(err)
				}
				secondary, err := f.db.AddServerPort("alpha", 25566, "tcp", "secondary", false)
				if err != nil {
					t.Fatal(err)
				}
				propertyPath := filepath.Join(f.manager.GetWorkingDir("alpha"), "server.properties")
				properties := []byte("server-port=25565\nmotd=keep\n")
				if err := os.WriteFile(propertyPath, properties, 0600); err != nil {
					t.Fatal(err)
				}
				firewallCalls := 0
				f.router.servers.allowPort = func(int, string) { firewallCalls++ }
				f.router.servers.removePort = func(int) { firewallCalls++ }
				method, path, body, success := "POST", "/api/servers/alpha/ports", `{"port":25567,"protocol":"tcp","is_primary":true}`, 201
				if operation == "delete" {
					method, path, body, success = "DELETE", fmt.Sprintf("/api/servers/alpha/ports/%d", secondary.ID), "", 200
				} else if operation == "primary" {
					path, body, success = fmt.Sprintf("/api/servers/alpha/ports/%d/primary", secondary.ID), "", 200
				}
				beforePorts, _ := f.db.ListServerPorts("alpha")
				beforeServer, _ := f.db.GetServer("alpha")
				for _, user := range []string{"owner", "subuser"} {
					if allow {
						break // exercise a successful subuser mutation below
					}
					w := f.request(method, path, body, user)
					if w.Code != 403 || !strings.Contains(w.Body.String(), "Node policy") {
						t.Fatal(user, w.Code, w.Body.String())
					}
					afterPorts, _ := f.db.ListServerPorts("alpha")
					afterServer, _ := f.db.GetServer("alpha")
					afterProperties, err := os.ReadFile(propertyPath)
					if err != nil || !bytes.Equal(properties, afterProperties) || !reflect.DeepEqual(beforePorts, afterPorts) || !reflect.DeepEqual(beforeServer, afterServer) || firewallCalls != 0 {
						t.Fatal("denied port mutation changed DB, files or firewall", err, firewallCalls)
					}
				}
				user := "admin"
				if allow {
					user = "subuser"
				}
				if w := f.request(method, path, body, user); w.Code != success {
					t.Fatal(user, w.Code, w.Body.String())
				}
			})
		}
	}
}

func TestDestructiveFileWritesRequireEditAndDelete(t *testing.T) {
	for _, action := range []string{"rename", "move", "unzip", "extract"} {
		for _, grant := range []string{"delete", "edit", "both", "owner", "admin"} {
			t.Run(action+"/"+grant, func(t *testing.T) {
				f := newReliabilityFixture(t)
				permissions := []string{"files.delete"}
				if grant == "edit" {
					permissions = []string{"files.edit"}
				}
				if grant == "both" {
					permissions = []string{"files.edit", "files.delete"}
				}
				subs, _ := f.db.ListServerSubusers("alpha")
				if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, permissions); err != nil {
					t.Fatal(err)
				}
				root := f.manager.GetWorkingDir("alpha")
				if err := os.WriteFile(filepath.Join(root, "source.txt"), []byte("source"), 0600); err != nil {
					t.Fatal(err)
				}
				var archive bytes.Buffer
				zw := zip.NewWriter(&archive)
				entry, err := zw.Create("result.txt")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := entry.Write([]byte("source")); err != nil {
					t.Fatal(err)
				}
				if err := zw.Close(); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "source.zip"), archive.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
				body := `{"old_path":"source.txt","new_path":"result.txt"}`
				if action == "move" {
					body = `{"from":"source.txt","to":"result.txt"}`
				}
				if action == "unzip" || action == "extract" {
					body = `{"path":"source.zip"}`
				}
				user, want := "subuser", 200
				if grant == "owner" || grant == "admin" {
					user = grant
				}
				if grant == "delete" || grant == "edit" {
					want = 403
				}
				w := f.request("POST", "/api/servers/alpha/files/"+action, body, user)
				if w.Code != want {
					t.Fatal(w.Code, w.Body.String())
				}
				result, err := os.ReadFile(filepath.Join(root, "result.txt"))
				if want == 403 {
					if !os.IsNotExist(err) {
						t.Fatal("denied operation wrote destination", err)
					}
					source, err := os.ReadFile(filepath.Join(root, "source.txt"))
					if err != nil || string(source) != "source" {
						t.Fatal("denied operation changed source", err)
					}
					logs, err := f.db.ListActivity(0, true, 100)
					if err != nil || len(logs) != 0 {
						t.Fatal("denied operation emitted audit write", logs, err)
					}
				} else if err != nil || string(result) != "source" {
					t.Fatal("authorized operation did not write", err)
				}
			})
		}
	}
}
