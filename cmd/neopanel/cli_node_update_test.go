package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCLINodeUpdate(t *testing.T) {
	// Missing target or --all should fail
	if code := cliNode("", []string{"update"}); code != 1 {
		t.Fatalf("expected exit code 1 for empty update flags, got %d", code)
	}

	// Mock Master Panel
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/api/cluster/update" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     "updating",
				"message":    "Dispatched updates to 2 node(s)",
				"dispatched": 2,
			})
			return
		}
		if r.URL.Path == "/api/nodes/node-1/update" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "updating",
				"message": "Dispatched update to node-1",
				"node_id": "node-1",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Update via --master for single node
	code := cliNode("", []string{"update", "node-1", "--master", server.URL, "--api-key", "test-api-key"})
	if code != 0 {
		t.Fatalf("expected exit code 0 for single node update, got %d", code)
	}

	// Update via --master with --all
	code = cliNode("", []string{"update", "--all", "--master", server.URL, "--api-key", "test-api-key"})
	if code != 0 {
		t.Fatalf("expected exit code 0 for cluster update, got %d", code)
	}
}
