package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/trash"
)

const trashURL = "/api/servers/alpha/files/trash"

func trashItems(t *testing.T, f *reliabilityFixture, path string) []trash.Item {
	t.Helper()
	w := f.request("GET", path, "", "owner")
	var result struct {
		Items []trash.Item `json:"items"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Items == nil {
		t.Fatal(w.Code, w.Body.String())
	}
	return result.Items
}
func TestTrashDefaultAndPermanentSingleBulk(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	for _, name := range []string{"single", "permanent", "bulk", "bulk-permanent"} {
		bulkDeleteFile(t, root, name)
	}
	for _, tc := range []struct{ method, path, body string }{
		{"DELETE", "/api/servers/alpha/files?path=single", ""},
		{"DELETE", "/api/servers/alpha/files?path=permanent&permanent=true", ""},
		{"POST", bulkDeleteURL, `{"paths":["bulk","bulk","missing"]}`},
		{"POST", bulkDeleteURL, `{"paths":["bulk-permanent","missing"],"permanent":true}`},
	} {
		w := f.request(tc.method, tc.path, tc.body, "owner")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if tc.method == "POST" {
			var req struct{ Paths []string }
			json.Unmarshal([]byte(tc.body), &req)
			requireBulkDeleted(t, w.Body.Bytes(), req.Paths)
		}
	}
	items := trashItems(t, f, trashURL)
	if len(items) != 2 || items[0].OriginalPath != "bulk" || items[1].OriginalPath != "single" {
		t.Fatal(items)
	}
	for _, name := range []string{"single", "permanent", "bulk", "bulk-permanent"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatal(name, err)
		}
	}
	if dirs, err := os.ReadDir(root); err != nil || len(dirs) != 0 {
		t.Fatal("trash exposed in tenant root", dirs, err)
	}
	for _, item := range items {
		w := f.request("POST", trashURL+"/"+item.ID+"/restore", `{}`, "owner")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if data, err := os.ReadFile(filepath.Join(root, item.OriginalPath)); err != nil || string(data) != "keep" {
			t.Fatal(string(data), err)
		}
	}
	if len(trashItems(t, f, trashURL)) != 0 {
		t.Fatal("restore did not remove trash entries")
	}
}
func TestTrashRestoreCollisionPurgeAndCrossServer(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	bulkDeleteFile(t, root, "dir/file")
	if w := f.request("DELETE", "/api/servers/alpha/files?path=dir", "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	items := trashItems(t, f, trashURL)
	if len(items) != 1 {
		t.Fatal(items)
	}
	id := items[0].ID
	bulkDeleteFile(t, root, "dir/other")
	if w := f.request("POST", trashURL+"/"+id+"/restore", `{}`, "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, tc := range []struct{ method, suffix string }{{"POST", "/restore"}, {"DELETE", ""}} {
		if w := f.request(tc.method, "/api/servers/beta/files/trash/"+id+tc.suffix, `{}`, "owner"); w.Code != 404 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := f.request("DELETE", trashURL+"/"+id, "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := f.request("POST", trashURL+"/"+id+"/restore", `{}`, "owner"); w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "dir/other")); err != nil {
		t.Fatal("collision overwritten", err)
	}
}
func TestTrashPermissions(t *testing.T) {
	f := newReliabilityFixture(t)
	id := strings.Repeat("a", 32)
	routes := []struct{ method, path string }{{"GET", trashURL}, {"POST", trashURL + "/" + id + "/restore"}, {"DELETE", trashURL + "/" + id}}
	for _, route := range routes {
		for _, user := range []string{"", "stranger", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request(route.method, route.path, `{}`, user); w.Code != want {
				t.Fatal(user, route, w.Code, w.Body.String())
			}
		}
	}
	subs, err := f.db.ListServerSubusers("alpha")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		perms  []string
		status []int
	}{
		{[]string{"files.delete"}, []int{403, 403, 404}},
		{[]string{"files.view", "files.delete"}, []int{200, 403, 404}},
		{[]string{"files.view", "files.edit"}, []int{403, 403, 403}},
		{[]string{"files.view", "files.edit", "files.delete"}, []int{200, 404, 404}},
	} {
		if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, tc.perms); err != nil {
			t.Fatal(err)
		}
		for i, route := range routes {
			if w := f.request(route.method, route.path, `{}`, "subuser"); w.Code != tc.status[i] {
				t.Fatal(tc.perms, route, w.Code, w.Body.String())
			}
		}
	}
	for _, user := range []string{"owner", "admin"} {
		for _, route := range routes {
			w := f.request(route.method, route.path, `{}`, user)
			if w.Code == 401 || w.Code == 403 {
				t.Fatal(user, w.Code)
			}
		}
	}
}
func TestTrashProxyContract(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	base := "/api/nodes/remote/proxy/api/servers/alpha/files"
	bulkDeleteFile(t, worker.manager.GetWorkingDir("alpha"), "remote-file")
	if w := master.request("DELETE", base+"?path=remote-file", "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	items := trashItems(t, master, base+"/trash")
	if len(items) != 1 {
		t.Fatal(items)
	}
	if len(trashItems(t, master, trashURL)) != 0 {
		t.Fatal("proxy used master filesystem")
	}
	if w := master.request("GET", base+"/trash", "", "subuser"); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := master.request("POST", fmt.Sprintf("%s/trash/%s/restore", base, items[0].ID), `{}`, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := master.request("POST", base+"/delete-bulk", `{"paths":["remote-file"]}`, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	items = trashItems(t, master, base+"/trash")
	if len(items) != 1 {
		t.Fatal(items)
	}
	if w := master.request("DELETE", base+"/trash/"+items[0].ID, "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if len(trashItems(t, master, base+"/trash")) != 0 {
		t.Fatal("proxy purge failed")
	}
}

func TestTrashRestoreWaitsForServerOperationGate(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	bulkDeleteFile(t, root, "config/file")
	if w := f.request("DELETE", "/api/servers/alpha/files?path=config/file", "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	item := trashItems(t, f, trashURL)[0]
	release, err := f.manager.BeginServerOperation("alpha")
	if err != nil {
		t.Fatal(err)
	}
	released := false
	defer func() {
		if !released {
			release()
		}
	}()
	url := trashURL + "/" + item.ID + "/restore"
	if w := f.request("POST", url, `{}`, "owner"); w.Code != 409 {
		t.Fatal("restore entered a busy server", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "config", "file")); !os.IsNotExist(err) {
		t.Fatal("busy restore wrote destination", err)
	}
	if items := trashItems(t, f, trashURL); len(items) != 1 || items[0].ID != item.ID {
		t.Fatal("busy restore consumed recoverable payload", items)
	}
	release()
	released = true
	if w := f.request("POST", url, `{}`, "owner"); w.Code != 200 {
		t.Fatal("restore failed after maintenance released", w.Code, w.Body.String())
	}
}
