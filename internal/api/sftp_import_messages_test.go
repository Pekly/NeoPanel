package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/sftpimport"
)

func diagnosticImportFixture(f *reliabilityFixture) {
	f.router.sftpImports.jobs["alpha"] = &sftpImportRecord{job: sftpImportJob{
		ID: "diagnostic-job", ServerID: "alpha", UpdatedAt: time.Now(),
		Progress:   sftpimport.Progress{State: "failed"},
		Error:      "open C:/private-host/data/.restore-alpha-hidden: access denied",
		Message:    "internal staging failure on private-host",
		SafetyCopy: ".restore-alpha-hidden/original",
	}}
}

func TestSFTPImportMessagesRespectViewerRole(t *testing.T) {
	f := newReliabilityFixture(t)
	diagnosticImportFixture(f)
	for _, method := range []string{"GET", "DELETE"} {
		url := importURL
		if method == "DELETE" {
			url += "/diagnostic-job"
		}
		for _, user := range []string{"owner", "admin"} {
			w := f.request(method, url, "", user)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			var job sftpImportJob
			if method == "GET" {
				var response struct {
					Job sftpImportJob `json:"job"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				job = response.Job
			} else if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
				t.Fatal(err)
			}
			if job.UserMessage == "" || job.State != "failed" {
				t.Fatal(job)
			}
			if user == "admin" {
				if job.Error == "" || job.SafetyCopy == "" {
					t.Fatal("admin diagnostics removed", job)
				}
			} else {
				if job.Error != "" || job.SafetyCopy != "" || strings.Contains(w.Body.String(), "private-host") || strings.Contains(w.Body.String(), ".restore-alpha") {
					t.Fatal("host diagnostics exposed", w.Body.String())
				}
				if !strings.Contains(job.UserMessage, "contact support") {
					t.Fatal("missing next step", job)
				}
			}
		}
	}
}

func TestSFTPImportWorkerMessagesRespectForwardedRole(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	diagnosticImportFixture(worker)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"owner", "admin"} {
		w := master.request("GET", "/api/nodes/remote/proxy"+importURL, "", user)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "private-host") != (user == "admin") {
			t.Fatal("wrong worker diagnostic audience", user, w.Body.String())
		}
	}
}

func TestSFTPImportConnectionErrorsGiveSafeNextSteps(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, user := range []string{"owner", "admin"} {
		w := f.request("POST", importURL+"/host-key", `{"host":"127.0.0.1","port":22}`, user)
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
		if user == "owner" {
			if !strings.Contains(w.Body.String(), "Contact support") || strings.Contains(w.Body.String(), "outbound address") {
				t.Fatal(w.Body.String())
			}
		} else if !strings.Contains(w.Body.String(), "outbound address denied") {
			t.Fatal(w.Body.String())
		}
	}
	if got := importUserError("sftp import: host key mismatch"); !strings.Contains(got, "Verify it with your provider") {
		t.Fatal(got)
	}
	if got := importUserError("import transfer failed: remote file changed"); !strings.Contains(got, "Stop the source server") {
		t.Fatal(got)
	}
}
