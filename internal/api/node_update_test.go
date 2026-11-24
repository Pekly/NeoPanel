package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestNodeUpdateEndpoints(t *testing.T) {
	f := newReliabilityFixture(t)

	// Local node rejected
	w := f.request("POST", "/api/nodes/local/update", "", "admin")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for local node update, got %d", w.Code)
	}

	// Nonexistent node returns 404
	w = f.request("POST", "/api/nodes/nonexistent-node/update", "", "admin")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing node, got %d", w.Code)
	}

	// Non-admin forbidden
	w = f.request("POST", "/api/nodes/some-node/update", "", "owner")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", w.Code)
	}

	// Unauthenticated unauthorized
	w = f.request("POST", "/api/nodes/some-node/update", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated, got %d", w.Code)
	}

	// Mock worker node daemon
	var applyCalled int32
	var configCalled int32
	workerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/admin/updates/config") {
			atomic.AddInt32(&configCalled, 1)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/api/admin/updates/apply") {
			atomic.AddInt32(&applyCalled, 1)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "updating"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer workerServer.Close()

	remoteNode := &database.ClusterNode{
		ID:       "worker-01",
		Name:     "Worker One",
		BaseURL:  workerServer.URL,
		APIToken: "neo_node_mock_token_12345",
		FQDN:     "worker.example.com",
	}
	if err := f.db.CreateNode(remoteNode); err != nil {
		t.Fatalf("failed to create test remote node: %v", err)
	}

	// Test single node update endpoint
	w = f.request("POST", "/api/nodes/worker-01/update", `{"repo_url":"https://github.com/Pekly/NeoPanel.git","branch":"main"}`, "admin")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for node update, got %d: %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if res["status"] != "updating" {
		t.Fatalf("expected status=updating, got %v", res["status"])
	}

	// Test cluster-wide update endpoint
	w = f.request("POST", "/api/cluster/update", `{"branch":"main"}`, "admin")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for cluster update, got %d: %s", w.Code, w.Body.String())
	}
	var clusterRes map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &clusterRes); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if clusterRes["status"] != "updating" {
		t.Fatalf("expected status=updating, got %v", clusterRes["status"])
	}
	if clusterRes["dispatched"].(float64) < 1 {
		t.Fatalf("expected at least 1 dispatched node, got %v", clusterRes["dispatched"])
	}
}
