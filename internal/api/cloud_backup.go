package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	settingS3BackupEnabled   = "backup_s3_enabled"
	settingS3BackupEndpoint  = "backup_s3_endpoint"
	settingS3BackupBucket    = "backup_s3_bucket"
	settingS3BackupRegion    = "backup_s3_region"
	settingS3BackupAccessKey = "backup_s3_access_key"
	settingS3BackupSecretKey = "backup_s3_secret_key"
	settingS3BackupPrefix    = "backup_s3_prefix"
	settingS3BackupAuto      = "backup_s3_auto_upload"
	settingS3BackupRetention = "backup_s3_retention_days"
)

type CloudBackupConfig struct {
	Enabled       bool   `json:"enabled"`
	Endpoint      string `json:"endpoint"`
	Bucket        string `json:"bucket"`
	Region        string `json:"region"`
	AccessKey     string `json:"access_key"`
	SecretKey     string `json:"secret_key,omitempty"`
	Prefix        string `json:"prefix"`
	AutoUpload    bool   `json:"auto_upload"`
	RetentionDays int    `json:"retention_days"`
}

type CloudBackupItem struct {
	Key          string    `json:"key"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag,omitempty"`
}

type S3Client struct {
	Endpoint   string
	Bucket     string
	Region     string
	AccessKey  string
	SecretKey  string
	HTTPClient *http.Client
}

func NewS3Client(cfg CloudBackupConfig) *S3Client {
	endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	if endpoint == "" {
		endpoint = "https://s3.amazonaws.com"
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "us-east-1"
	}

	return &S3Client{
		Endpoint:  endpoint,
		Bucket:    strings.TrimSpace(cfg.Bucket),
		Region:    region,
		AccessKey: strings.TrimSpace(cfg.AccessKey),
		SecretKey: strings.TrimSpace(cfg.SecretKey),
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (s *S3Client) buildURL(key string, query url.Values) string {
	rawURL := s.Endpoint
	// If path-style vs virtual-hosted: For custom endpoints (R2, MinIO, Wasabi), use path-style: endpoint/bucket/key
	escapedKey := strings.TrimPrefix(key, "/")
	if s.Bucket != "" {
		rawURL += "/" + url.PathEscape(s.Bucket)
	}
	if escapedKey != "" {
		parts := strings.Split(escapedKey, "/")
		for i := range parts {
			parts[i] = url.PathEscape(parts[i])
		}
		rawURL += "/" + strings.Join(parts, "/")
	}
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}
	return rawURL
}

func (s *S3Client) signRequest(req *http.Request, payloadHash string, now time.Time) {
	dateISO := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")

	req.Header.Set("x-amz-date", dateISO)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	// Build Canonical Headers
	headersToSign := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	var headerEntries []string
	for _, h := range headersToSign {
		val := ""
		if h == "host" {
			val = req.Host
			if val == "" {
				val = req.URL.Host
			}
		} else {
			val = req.Header.Get(h)
		}
		headerEntries = append(headerEntries, fmt.Sprintf("%s:%s", h, strings.TrimSpace(val)))
	}
	canonicalHeaders := strings.Join(headerEntries, "\n") + "\n"
	signedHeaders := strings.Join(headersToSign, ";")

	// Canonical Query String
	var queryKeys []string
	for k := range req.URL.Query() {
		queryKeys = append(queryKeys, k)
	}
	sort.Strings(queryKeys)
	var queryParts []string
	for _, k := range queryKeys {
		for _, v := range req.URL.Query()[k] {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
		}
	}
	canonicalQuery := strings.Join(queryParts, "&")

	// Canonical URI
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Canonical Request
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.Region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		dateISO,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	)

	// Calculate Signature
	kSecret := []byte("AWS4" + s.SecretKey)
	kDate := hmacSHA256(kSecret, []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.Region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	signature := hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.AccessKey,
		credentialScope,
		signedHeaders,
		signature,
	)
	req.Header.Set("Authorization", authHeader)
}

func (s *S3Client) PutObject(ctx context.Context, key string, data []byte, contentType string) error {
	reqURL := s.buildURL(key, nil)
	payloadHash := sha256Hex(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	if contentType == "" {
		contentType = "application/gzip"
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))

	s.signRequest(req, payloadHash, time.Now())

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("S3 PutObject error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("S3 PutObject failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *S3Client) GetObject(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	reqURL := s.buildURL(key, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}

	emptyHash := sha256Hex([]byte{})
	s.signRequest(req, emptyHash, time.Now())

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("S3 GetObject error: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, 0, os.ErrNotExist
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, 0, fmt.Errorf("S3 GetObject failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return resp.Body, resp.ContentLength, nil
}

func (s *S3Client) DeleteObject(ctx context.Context, key string) error {
	reqURL := s.buildURL(key, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	emptyHash := sha256Hex([]byte{})
	s.signRequest(req, emptyHash, time.Now())

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("S3 DeleteObject error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("S3 DeleteObject failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

type s3ListBucketResult struct {
	XMLName  xml.Name `xml:"ListBucketResult"`
	Contents []struct {
		Key          string    `xml:"Key"`
		LastModified time.Time `xml:"LastModified"`
		ETag         string    `xml:"ETag"`
		Size         int64     `xml:"Size"`
	} `xml:"Contents"`
}

func (s *S3Client) ListObjects(ctx context.Context, prefix string) ([]CloudBackupItem, error) {
	q := url.Values{}
	q.Set("list-type", "2")
	if prefix != "" {
		q.Set("prefix", prefix)
	}

	reqURL := s.buildURL("", q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	emptyHash := sha256Hex([]byte{})
	s.signRequest(req, emptyHash, time.Now())

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("S3 ListObjects error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("S3 ListObjects failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var xmlRes s3ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&xmlRes); err != nil {
		return nil, fmt.Errorf("parsing S3 XML list response: %w", err)
	}

	var items []CloudBackupItem
	for _, c := range xmlRes.Contents {
		name := filepath.Base(c.Key)
		items = append(items, CloudBackupItem{
			Key:          c.Key,
			Name:         name,
			Size:         c.Size,
			LastModified: c.LastModified,
			ETag:         strings.Trim(c.ETag, "\""),
		})
	}

	return items, nil
}

func (s *S3Client) TestConnection(ctx context.Context) error {
	// Simple test: Put a small test object then immediately delete it
	testKey := fmt.Sprintf(".neopanel_test_%d.txt", time.Now().UnixNano())
	testData := []byte("NeoPanel Cloud Storage Connection Test - OK")

	if err := s.PutObject(ctx, testKey, testData, "text/plain"); err != nil {
		return err
	}

	_ = s.DeleteObject(ctx, testKey)
	return nil
}

// BackupHandler helper methods for Cloud Backup

func (h *BackupHandler) loadCloudConfig() CloudBackupConfig {
	cfg := CloudBackupConfig{
		Enabled:       false,
		Endpoint:      "https://s3.amazonaws.com",
		Region:        "us-east-1",
		Prefix:        "backups",
		AutoUpload:    false,
		RetentionDays: 30,
	}

	if h.db != nil {
		if v, err := h.db.GetPanelSetting(settingS3BackupEnabled); err == nil {
			cfg.Enabled = v == "true"
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupEndpoint); err == nil && v != "" {
			cfg.Endpoint = v
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupBucket); err == nil {
			cfg.Bucket = v
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupRegion); err == nil && v != "" {
			cfg.Region = v
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupAccessKey); err == nil {
			cfg.AccessKey = v
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupSecretKey); err == nil {
			cfg.SecretKey = v
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupPrefix); err == nil && v != "" {
			cfg.Prefix = v
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupAuto); err == nil {
			cfg.AutoUpload = v == "true"
		}
		if v, err := h.db.GetPanelSetting(settingS3BackupRetention); err == nil && v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.RetentionDays = n
			}
		}
	}

	return cfg
}

func (h *BackupHandler) GetAdminCloudBackupConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.loadCloudConfig()
	maskedSecret := ""
	if len(cfg.SecretKey) > 8 {
		maskedSecret = cfg.SecretKey[:4] + strings.Repeat("•", len(cfg.SecretKey)-8) + cfg.SecretKey[len(cfg.SecretKey)-4:]
	} else if cfg.SecretKey != "" {
		maskedSecret = "••••••••"
	}

	cfg.SecretKey = maskedSecret
	writeJSON(w, http.StatusOK, cfg)
}

func (h *BackupHandler) SaveAdminCloudBackupConfig(w http.ResponseWriter, r *http.Request) {
	var req CloudBackupConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if h.db != nil {
		_ = h.db.SetPanelSetting(settingS3BackupEnabled, strconv.FormatBool(req.Enabled))
		_ = h.db.SetPanelSetting(settingS3BackupEndpoint, strings.TrimSpace(req.Endpoint))
		_ = h.db.SetPanelSetting(settingS3BackupBucket, strings.TrimSpace(req.Bucket))
		_ = h.db.SetPanelSetting(settingS3BackupRegion, strings.TrimSpace(req.Region))
		_ = h.db.SetPanelSetting(settingS3BackupAccessKey, strings.TrimSpace(req.AccessKey))
		if req.SecretKey != "" && !strings.Contains(req.SecretKey, "•") {
			_ = h.db.SetPanelSetting(settingS3BackupSecretKey, strings.TrimSpace(req.SecretKey))
		}
		prefix := strings.Trim(strings.TrimSpace(req.Prefix), "/")
		if prefix == "" {
			prefix = "backups"
		}
		_ = h.db.SetPanelSetting(settingS3BackupPrefix, prefix)
		_ = h.db.SetPanelSetting(settingS3BackupAuto, strconv.FormatBool(req.AutoUpload))
		_ = h.db.SetPanelSetting(settingS3BackupRetention, strconv.Itoa(req.RetentionDays))
	}

	logAudit(h.db, r, "admin_cloud_backup_config", fmt.Sprintf("Updated S3 cloud backup configuration (bucket: %s, enabled: %v)", req.Bucket, req.Enabled))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cloud backup configuration saved successfully"})
}

func (h *BackupHandler) TestAdminCloudBackup(w http.ResponseWriter, r *http.Request) {
	var req CloudBackupConfig
	_ = json.NewDecoder(r.Body).Decode(&req)

	saved := h.loadCloudConfig()
	if req.Endpoint == "" {
		req.Endpoint = saved.Endpoint
	}
	if req.Bucket == "" {
		req.Bucket = saved.Bucket
	}
	if req.Region == "" {
		req.Region = saved.Region
	}
	if req.AccessKey == "" {
		req.AccessKey = saved.AccessKey
	}
	if req.SecretKey == "" || strings.Contains(req.SecretKey, "•") {
		req.SecretKey = saved.SecretKey
	}

	if req.Bucket == "" || req.AccessKey == "" || req.SecretKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Bucket, Access Key ID, and Secret Key are required to test connection"})
		return
	}

	client := NewS3Client(req)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cloud storage connection test failed: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Successfully connected to S3 bucket %q and verified write/delete permissions!", req.Bucket),
	})
}

func (h *BackupHandler) ListCloudBackups(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if _, err := h.db.GetServer(serverID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	cfg := h.loadCloudConfig()
	if !cfg.Enabled || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
			"backups": []CloudBackupItem{},
		})
		return
	}

	client := NewS3Client(cfg)
	prefix := fmt.Sprintf("%s/%s/", cfg.Prefix, serverID)

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	items, err := client.ListObjects(ctx, prefix)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to list cloud backups: " + err.Error()})
		return
	}
	if items == nil {
		items = []CloudBackupItem{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"bucket":  cfg.Bucket,
		"prefix":  prefix,
		"backups": items,
	})
}

func (h *BackupHandler) UploadBackupToCloud(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	name := filepath.Base(getURLParam(r, "name"))

	if name == "" || (!strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".zip")) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid backup filename required"})
		return
	}

	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	cfg := h.loadCloudConfig()
	if !cfg.Enabled || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cloud backup is not configured or disabled in Admin settings"})
		return
	}

	localBackupPath := filepath.Join(h.getBackupsDir(serverID), name)
	data, err := os.ReadFile(localBackupPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Local backup file not found: " + name})
		return
	}

	client := NewS3Client(cfg)
	cloudKey := fmt.Sprintf("%s/%s/%s", cfg.Prefix, serverID, name)

	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()

	if err := client.PutObject(ctx, cloudKey, data, "application/gzip"); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to upload to cloud: " + err.Error()})
		return
	}

	logAudit(h.db, r, "backup_cloud_upload", fmt.Sprintf("Uploaded backup '%s' (%d bytes) to cloud (%s) on %s", name, len(data), cfg.Bucket, auditServerLabel(h.db, srv.ID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Backup %s successfully uploaded to cloud storage!", name),
		"cloud_key": cloudKey,
		"bucket":    cfg.Bucket,
		"size":      len(data),
	})
}

func (h *BackupHandler) DeleteCloudBackup(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	name := filepath.Base(getURLParam(r, "name"))

	cfg := h.loadCloudConfig()
	if !cfg.Enabled || cfg.Bucket == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cloud backup is not configured"})
		return
	}

	client := NewS3Client(cfg)
	cloudKey := fmt.Sprintf("%s/%s/%s", cfg.Prefix, serverID, name)

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	if err := client.DeleteObject(ctx, cloudKey); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to delete cloud backup: " + err.Error()})
		return
	}

	logAudit(h.db, r, "backup_cloud_delete", fmt.Sprintf("Deleted cloud backup '%s' from %s", name, cfg.Bucket))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Cloud backup %s deleted successfully", name),
	})
}

func (h *BackupHandler) RestoreCloudBackup(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	name := filepath.Base(getURLParam(r, "name"))

	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	cfg := h.loadCloudConfig()
	if !cfg.Enabled || cfg.Bucket == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cloud backup is not configured"})
		return
	}

	client := NewS3Client(cfg)
	cloudKey := fmt.Sprintf("%s/%s/%s", cfg.Prefix, serverID, name)

	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()

	rc, _, err := client.GetObject(ctx, cloudKey)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Cloud backup not found: " + err.Error()})
		return
	}
	defer rc.Close()

	// Save to local backup directory
	localDir := h.getBackupsDir(serverID)
	_ = os.MkdirAll(localDir, 0755)
	localPath := filepath.Join(localDir, name)

	f, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Saving downloaded backup: " + err.Error()})
		return
	}
	_, copyErr := io.Copy(f, rc)
	f.Close()
	if copyErr != nil {
		_ = os.Remove(localPath)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Writing downloaded backup: " + copyErr.Error()})
		return
	}

	// Restore using server manager
	_, restoreErr := h.manager.RestoreBackup(r.Context(), srv, name)
	if restoreErr != nil {
		msg := restoreErrorForViewer(r, restoreErr)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
		return
	}

	logAudit(h.db, r, "backup_cloud_restore", fmt.Sprintf("Restored server %s from cloud backup '%s'", auditServerLabel(h.db, srv.ID), name))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Server successfully restored from cloud backup %s", name),
	})
}
