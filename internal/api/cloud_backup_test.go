package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestS3ClientOperations(t *testing.T) {
	storage := make(map[string][]byte)

	s3Mux := http.NewServeMux()
	s3Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Missing or invalid SigV4 authorization"))
			return
		}

		key := strings.TrimPrefix(r.URL.Path, "/test-bucket/")

		switch r.Method {
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			storage[key] = data
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			if r.URL.Query().Get("list-type") == "2" {
				prefix := r.URL.Query().Get("prefix")
				var contentsXML []string
				for k, v := range storage {
					if prefix == "" || strings.HasPrefix(k, prefix) {
						contentsXML = append(contentsXML, fmt.Sprintf("<Contents><Key>%s</Key><LastModified>2026-11-23T12:00:00Z</LastModified><Size>%d</Size><ETag>\"testetag\"</ETag></Contents>", k, len(v)))
					}
				}
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` + strings.Join(contentsXML, "") + `</ListBucketResult>`))
				return
			}
			data, ok := storage[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		case http.MethodDelete:
			delete(storage, key)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mockS3 := httptest.NewServer(s3Mux)
	defer mockS3.Close()

	cfg := CloudBackupConfig{
		Enabled:   true,
		Endpoint:  mockS3.URL,
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Prefix:    "backups",
	}

	client := NewS3Client(cfg)
	ctx := context.Background()

	// 1. Test PutObject
	testData := []byte("compressed-backup-bytes-tar-gz")
	err := client.PutObject(ctx, "backups/server1/backup1.tar.gz", testData, "application/gzip")
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// 2. Test ListObjects
	items, err := client.ListObjects(ctx, "backups/server1/")
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}
	if len(items) != 1 || items[0].Name != "backup1.tar.gz" || items[0].Size != int64(len(testData)) {
		t.Fatalf("Unexpected ListObjects result: %+v", items)
	}

	// 3. Test GetObject
	rc, size, err := client.GetObject(ctx, "backups/server1/backup1.tar.gz")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	downloaded, _ := io.ReadAll(rc)
	rc.Close()
	if string(downloaded) != string(testData) {
		t.Fatalf("Downloaded data mismatch: got %s, want %s", string(downloaded), string(testData))
	}
	_ = size

	// 4. Test DeleteObject
	err = client.DeleteObject(ctx, "backups/server1/backup1.tar.gz")
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}

	itemsAfter, _ := client.ListObjects(ctx, "backups/server1/")
	if len(itemsAfter) != 0 {
		t.Fatalf("Expected 0 items after delete, got %d", len(itemsAfter))
	}

	// 5. Test TestConnection
	if err := client.TestConnection(ctx); err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}

func TestCloudBackupAPIEndpoints(t *testing.T) {
	f := newReliabilityFixture(t)

	// 1. Admin config GET initially disabled
	w := f.request("GET", "/api/admin/backups/cloud", "", "admin")
	if w.Code != 200 {
		t.Fatalf("GET /api/admin/backups/cloud failed: %d %s", w.Code, w.Body.String())
	}
	var cfg CloudBackupConfig
	if err := json.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("Unmarshal config failed: %v", err)
	}
	if cfg.Enabled {
		t.Errorf("Expected cloud backup initially disabled")
	}

	// 2. Admin config POST update
	saveBody := `{
		"enabled": true,
		"endpoint": "https://r2.example.com",
		"bucket": "my-backups",
		"region": "auto",
		"access_key": "AKIA12345",
		"secret_key": "SECRETKEY987",
		"prefix": "cluster-backups",
		"auto_upload": true,
		"retention_days": 14
	}`
	w = f.request("POST", "/api/admin/backups/cloud", saveBody, "admin")
	if w.Code != 200 {
		t.Fatalf("POST config failed: %d %s", w.Code, w.Body.String())
	}

	// 3. Verify saved and secret key masked
	w = f.request("GET", "/api/admin/backups/cloud", "", "admin")
	_ = json.Unmarshal(w.Body.Bytes(), &cfg)
	if !cfg.Enabled || cfg.Bucket != "my-backups" || !cfg.AutoUpload || cfg.RetentionDays != 14 {
		t.Errorf("Unexpected saved config: %+v", cfg)
	}
	if !strings.Contains(cfg.SecretKey, "•") {
		t.Errorf("Expected masked secret key, got %s", cfg.SecretKey)
	}

	// 4. Server cloud backups list (returns empty/disabled if unreachable endpoint)
	w = f.request("GET", "/api/servers/alpha/backups/cloud", "", "owner")
	if w.Code != 200 && w.Code != 502 {
		t.Errorf("Expected 200 or 502 for unresolvable endpoint, got %d %s", w.Code, w.Body.String())
	}

	// 5. Create local backup on alpha and test cloud upload validation
	bDir := filepath.Join(f.cfg.Paths.DataDir, "backups", "alpha")
	_ = os.MkdirAll(bDir, 0755)
	_ = os.WriteFile(filepath.Join(bDir, "test.tar.gz"), []byte("backup-content"), 0644)

	// Invalid name check
	w = f.request("POST", "/api/servers/alpha/backups/invalid-name/cloud/upload", "", "owner")
	if w.Code != 400 {
		t.Errorf("Expected 400 for invalid backup name, got %d", w.Code)
	}
}
