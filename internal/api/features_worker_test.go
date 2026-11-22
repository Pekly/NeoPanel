package api

import (
	"encoding/json"
	"github.com/Pekly/NeoPanel/internal/database"
	"net/http/httptest"
	"testing"
)

func TestFeaturesWorkerDeliveryAndRetry(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	node := &database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}
	if err := master.db.CreateNode(node); err != nil {
		t.Fatal(err)
	}
	w := master.request("PUT", "/api/admin/features", `{"plugins":false,"schedules":false}`, "admin")
	var view featureView
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil || !view.Nodes["remote"] {
		t.Fatal(w.Code, w.Body.String())
	}
	if worker.db.FeatureEnabled("alpha", "schedules") || worker.db.FeatureEnabled("alpha", "plugins") {
		t.Fatal("worker policy not saved")
	}
	if w := worker.request("GET", "/api/servers/alpha/plugins", "", "owner"); w.Code != 403 {
		t.Fatal("direct worker bypass", w.Code)
	}
	if w := master.request("GET", "/api/nodes/remote/proxy/api/servers/alpha/plugins", "", "owner"); w.Code != 403 {
		t.Fatal("proxy bypass", w.Code, w.Body.String())
	}
	// An operator changed the worker independently: identical Save must redeliver.
	if err := worker.db.SetFeaturePolicy("", database.FeaturePolicy{}); err != nil {
		t.Fatal(err)
	}
	w = master.request("PUT", "/api/admin/features", `{"plugins":false,"schedules":false}`, "admin")
	if w.Code != 200 || worker.db.FeatureEnabled("alpha", "plugins") {
		t.Fatal("retry did not deliver", w.Code, w.Body.String())
	}
	remote.Close()
	w = master.request("PUT", "/api/admin/features", `{"plugins":true}`, "admin")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil || view.Nodes["remote"] {
		t.Fatal("false acknowledgement", w.Code, w.Body.String())
	}
	// Changing back after failed delivery must not revive an earlier ack.
	w = master.request("PUT", "/api/admin/features", `{"plugins":false,"schedules":false}`, "admin")
	if json.Unmarshal(w.Body.Bytes(), &view) != nil || view.Nodes["remote"] {
		t.Fatal("stale ack", w.Body.String())
	}
}
