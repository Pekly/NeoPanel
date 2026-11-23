package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudflareClientVerifyAndListZones(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/tokens/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(cloudflareResponse{
				Success: false,
				Errors:  []cfError{{Code: 1000, Message: "Invalid token"}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(cloudflareResponse{
			Success: true,
			Result:  json.RawMessage(`{"id":"tok_1","status":"active"}`),
		})
	})

	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(cloudflareResponse{
			Success: true,
			Result: json.RawMessage(`[
				{"id":"zone_abc","name":"example.com","status":"active"},
				{"id":"zone_def","name":"mc.other.net","status":"active"}
			]`),
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewCloudflareClient("test-valid-token")
	client.BaseURL = srv.URL

	ctx := context.Background()
	status, err := client.VerifyToken(ctx)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if status != "active" {
		t.Errorf("Expected active, got %s", status)
	}

	zones, err := client.ListZones(ctx, "")
	if err != nil {
		t.Fatalf("ListZones failed: %v", err)
	}
	if len(zones) != 2 {
		t.Errorf("Expected 2 zones, got %d", len(zones))
	}

	matchedZone, err := client.FindZoneForFQDN(ctx, "play.mc.other.net")
	if err != nil {
		t.Fatalf("FindZoneForFQDN failed: %v", err)
	}
	if matchedZone.ID != "zone_def" {
		t.Errorf("Expected zone_def for play.mc.other.net, got %s", matchedZone.ID)
	}
}

func TestCloudflareClientSyncDNS(t *testing.T) {
	var createdRecords []CloudflareDNSRecord

	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(cloudflareResponse{
			Success: true,
			Result:  json.RawMessage(`[{"id":"zone_123","name":"myserver.com","status":"active"}]`),
		})
	})

	mux.HandleFunc("/zones/zone_123/dns_records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Return empty to trigger creates
			_ = json.NewEncoder(w).Encode(cloudflareResponse{
				Success: true,
				Result:  json.RawMessage(`[]`),
			})
			return
		}
		if r.Method == http.MethodPost {
			var rec CloudflareDNSRecord
			if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
				t.Fatalf("Decode record error: %v", err)
			}
			rec.ID = "rec_" + rec.Type
			createdRecords = append(createdRecords, rec)
			raw, _ := json.Marshal(rec)
			_ = json.NewEncoder(w).Encode(cloudflareResponse{
				Success: true,
				Result:  raw,
			})
			return
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewCloudflareClient("test-token")
	client.BaseURL = srv.URL

	ctx := context.Background()
	result, err := client.SyncServerDNS(ctx, "", "play.myserver.com", "192.168.1.100", 25565)
	if err != nil {
		t.Fatalf("SyncServerDNS failed: %v", err)
	}

	if result.ZoneID != "zone_123" {
		t.Errorf("Expected zone_123, got %s", result.ZoneID)
	}
	if len(createdRecords) != 2 {
		t.Fatalf("Expected 2 records created (A and SRV), got %d", len(createdRecords))
	}

	// Address record must be A, content 192.168.1.100, proxied FALSE!
	aRec := createdRecords[0]
	if aRec.Type != "A" || aRec.Content != "192.168.1.100" || aRec.Proxied {
		t.Errorf("Invalid A record: %+v", aRec)
	}

	// SRV record must be SRV, proxied FALSE
	srvRec := createdRecords[1]
	if srvRec.Type != "SRV" || !strings.Contains(srvRec.Name, "_minecraft._tcp") {
		t.Errorf("Invalid SRV record: %+v", srvRec)
	}
}
