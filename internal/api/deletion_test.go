package api

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestNodeReassignmentCannotHideRunningServer(t *testing.T) {
	f := newReliabilityFixture(t)
	w := f.request("PUT", "/api/servers/alpha", `{"node_id":"other-node","name":"should-not-change"}`, "admin")
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	srv, err := f.db.GetServer("alpha")
	if err != nil || srv.NodeID != "local" || srv.Name != "alpha" {
		t.Fatal("metadata-only move accepted", srv, err)
	}
}

func TestMasterCannotDeleteRemoteTaggedMetadataLocally(t *testing.T) {
	f := newReliabilityFixture(t)
	srv, _ := f.db.GetServer("alpha")
	srv.NodeID = "remote"
	if err := f.db.UpdateServer(srv); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(f.manager.GetWorkingDir("alpha"), "keep")
	if err := os.WriteFile(marker, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	w := f.request("DELETE", "/api/servers/alpha", "", "admin")
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := f.db.GetServer("alpha"); err != nil {
		t.Fatal("deleted remote metadata", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("deleted files without worker confirmation", err)
	}
}

func TestRemoteDeleteRequiresWorkerCompletionAndKeepsSameIDLocalServer(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	worker.cfg.Server.Role = "node"
	node := httptest.NewServer(worker.router)
	defer node.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: node.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	release, err := worker.manager.BeginServerOperation("alpha")
	if err != nil {
		t.Fatal(err)
	}
	w := master.request("DELETE", "/api/nodes/remote/proxy/api/servers/alpha", "", "admin")
	release()
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, f := range []*reliabilityFixture{master, worker} {
		if _, err := f.db.GetServer("alpha"); err != nil {
			t.Fatal("worker refusal discarded metadata", err)
		}
	}
	w = master.request("DELETE", "/api/nodes/remote/proxy/api/servers/alpha", "", "admin")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var ack struct {
		Deleted bool   `json:"deleted"`
		ID      string `json:"server_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &ack); err != nil || !ack.Deleted || ack.ID != "alpha" {
		t.Fatal("missing completion acknowledgement", w.Body.String(), err)
	}
	if _, err := worker.db.GetServer("alpha"); err == nil {
		t.Fatal("worker row survived successful delete")
	}
	if _, err := master.db.GetServer("alpha"); err != nil {
		t.Fatal("deleted same-ID server on wrong host", err)
	}
	// An unreachable worker is never a reason to erase master-side metadata.
	node.Close()
	w = master.request("DELETE", "/api/nodes/remote/proxy/api/servers/beta", "", "admin")
	if w.Code != 502 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, f := range []*reliabilityFixture{master, worker} {
		if _, err := f.db.GetServer("beta"); err != nil {
			t.Fatal("unreachable worker discarded metadata", err)
		}
	}
}
