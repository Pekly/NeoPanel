package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func waitURLJob(t *testing.T, f *reliabilityFixture, jobID string) URLDownloadJob {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		w := f.request("GET", "/api/servers/alpha/files/upload-url", "", "owner")
		if w.Code != http.StatusOK {
			t.Fatalf("GET upload-url failed: %d %s", w.Code, w.Body.String())
		}
		var res struct {
			Jobs []URLDownloadJob `json:"jobs"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("unmarshal jobs: %v", err)
		}
		for _, j := range res.Jobs {
			if j.ID == jobID && (j.State == "completed" || j.State == "failed" || j.State == "cancelled") {
				return j
			}
		}
		time.Sleep(15 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for URL download job %s", jobID)
	return URLDownloadJob{}
}

func TestFilesUploadFromURL(t *testing.T) {
	f := newReliabilityFixture(t)

	// Mock remote file server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/plugin.jar":
			w.Header().Set("Content-Type", "application/java-archive")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mock-jar-content-12345"))
		case "/with-header":
			w.Header().Set("Content-Disposition", `attachment; filename="custom-plugin.jar"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("header-jar-content"))
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("default-content"))
		}
	}))
	defer mockServer.Close()

	// 1. Invalid URLs rejected
	for _, invalidURL := range []string{"", "ftp://example.com/file.txt", "javascript:alert(1)", "not-a-url"} {
		body, _ := json.Marshal(map[string]interface{}{"url": invalidURL})
		w := f.request("POST", "/api/servers/alpha/files/upload-url", string(body), "owner")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid url %q, got %d: %s", invalidURL, w.Code, w.Body.String())
		}
	}

	// 2. Traversal path rejected
	body, _ := json.Marshal(map[string]interface{}{
		"url":  mockServer.URL + "/plugin.jar",
		"path": "../../../etc",
	})
	w := f.request("POST", "/api/servers/alpha/files/upload-url", string(body), "owner")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for directory traversal, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Subuser without files.edit forbidden
	body, _ = json.Marshal(map[string]interface{}{"url": mockServer.URL + "/plugin.jar"})
	w = f.request("POST", "/api/servers/alpha/files/upload-url", string(body), "subuser")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unauthorized subuser, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Remote 404 results in failed job
	body, _ = json.Marshal(map[string]interface{}{"url": mockServer.URL + "/not-found"})
	w = f.request("POST", "/api/servers/alpha/files/upload-url", string(body), "owner")
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for async download start, got %d: %s", w.Code, w.Body.String())
	}
	var start404 struct {
		JobID string `json:"job_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &start404)
	if job404 := waitURLJob(t, f, start404.JobID); job404.State != "failed" {
		t.Fatalf("expected failed state for 404 remote, got %+v", job404)
	}

	// 5. Successful download using URL path
	_ = os.MkdirAll(filepath.Join(f.manager.GetWorkingDir("alpha"), "plugins"), 0755)
	body, _ = json.Marshal(map[string]interface{}{
		"url":  mockServer.URL + "/plugin.jar",
		"path": "plugins",
	})
	w = f.request("POST", "/api/servers/alpha/files/upload-url", string(body), "owner")
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for download, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	doneJob := waitURLJob(t, f, res.JobID)
	if doneJob.State != "completed" || doneJob.Filename != "plugin.jar" || doneJob.Percent != 100 {
		t.Fatalf("expected completed plugin.jar, got %+v", doneJob)
	}

	// Verify file was written to server working dir
	targetFile := filepath.Join(f.manager.GetWorkingDir("alpha"), "plugins", "plugin.jar")
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read saved file %s: %v", targetFile, err)
	}
	if string(content) != "mock-jar-content-12345" {
		t.Fatalf("unexpected content: %q", string(content))
	}

	// 6. Conflict when duplicate without overwrite
	w = f.request("POST", "/api/servers/alpha/files/upload-url", string(body), "owner")
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for job, got %d", w.Code)
	}
	var dupRes struct {
		JobID string `json:"job_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &dupRes)
	if dupJob := waitURLJob(t, f, dupRes.JobID); dupJob.State != "failed" {
		t.Fatalf("expected failed conflict when file exists, got %+v", dupJob)
	}

	// 7. Success with overwrite=true
	bodyOverwrite, _ := json.Marshal(map[string]interface{}{
		"url":       mockServer.URL + "/plugin.jar",
		"path":      "plugins",
		"overwrite": true,
	})
	w = f.request("POST", "/api/servers/alpha/files/upload-url", string(bodyOverwrite), "owner")
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 with overwrite, got %d: %s", w.Code, w.Body.String())
	}
	var owRes struct {
		JobID string `json:"job_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &owRes)
	if owJob := waitURLJob(t, f, owRes.JobID); owJob.State != "completed" {
		t.Fatalf("expected completed with overwrite, got %+v", owJob)
	}

	// 8. Download with filename from Content-Disposition header
	bodyHeader, _ := json.Marshal(map[string]interface{}{
		"url": mockServer.URL + "/with-header",
	})
	w = f.request("POST", "/api/servers/alpha/files/upload-url", string(bodyHeader), "owner")
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 with header, got %d: %s", w.Code, w.Body.String())
	}
	var resHeader struct {
		JobID string `json:"job_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resHeader)
	if headerJob := waitURLJob(t, f, resHeader.JobID); headerJob.State != "completed" || headerJob.Filename != "custom-plugin.jar" {
		t.Fatalf("expected filename=custom-plugin.jar, got %+v", headerJob)
	}
}
