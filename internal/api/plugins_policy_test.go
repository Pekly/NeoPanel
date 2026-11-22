package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestPluginInstalledRoutesAuthorizationAndFeaturePolicy(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, route := range []struct{ method, suffix string }{{"GET", "catalog"}, {"POST", "identify"}, {"POST", "update"}} {
		path := "/api/servers/alpha/plugins/" + route.suffix
		for user, status := range map[string]int{"": 401, "stranger": 403, "subuser": 403} {
			if route.method == "GET" && user == "subuser" {
				status = 200
			}
			if w := f.request(route.method, path, "{}", user); w.Code != status {
				t.Fatal(route, user, w.Code, w.Body.String())
			}
		}
		permission := "files.edit"
		if route.method == "GET" {
			permission = "files.view"
		}
		if serverActionPermission(route.method, "plugins/"+route.suffix) != permission || featureForRoute(route.method, path) != "plugins" {
			t.Fatal("policy mapping", route)
		}
	}
	if err := f.db.SetFeaturePolicy("alpha", database.FeaturePolicy{"plugins": false}); err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct{ method, suffix string }{{"GET", "catalog"}, {"POST", "identify"}, {"POST", "update"}} {
		for _, user := range []string{"owner", "admin"} {
			if w := f.request(route.method, "/api/servers/alpha/plugins/"+route.suffix, "{}", user); w.Code != 403 {
				t.Fatal(route, user, w.Code, w.Body.String())
			}
		}
	}
}

func TestPluginInstalledWorkerRouting(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	// An empty catalog needs no upstream requests. The worker's mods folder proves
	// the request was routed rather than served from the master's Paper instance.
	srv, err := worker.db.GetServer("alpha")
	if err != nil {
		t.Fatal(err)
	}
	srv.Type = "fabric"
	if err := worker.db.UpdateServer(srv); err != nil {
		t.Fatal(err)
	}
	base := "/api/nodes/remote/proxy/api/servers/alpha/plugins/"
	w := master.request("GET", base+"catalog", "", "owner")
	var page installedCatalog
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || page.Folder != "mods" {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, suffix := range []string{"identify", "update"} {
		if w := master.request("POST", base+suffix, "{}", "subuser"); w.Code != 403 {
			t.Fatal("write permission bypass", suffix, w.Code)
		}
		if w := master.request("POST", base+suffix, "{}", "owner"); w.Code != 400 {
			t.Fatal("write not routed", suffix, w.Code, w.Body.String())
		}
	}
	if err := worker.db.SetFeaturePolicy("alpha", database.FeaturePolicy{"plugins": false}); err != nil {
		t.Fatal(err)
	}
	if w := master.request(http.MethodGet, base+"catalog", "", "owner"); w.Code != 403 || !strings.Contains(w.Body.String(), "disabled") {
		t.Fatal(w.Code, w.Body.String())
	}
}
