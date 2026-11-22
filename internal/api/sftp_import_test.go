package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/sftpimport"
	"github.com/golang-jwt/jwt/v5"
)

const importURL = "/api/servers/alpha/files/import-sftp"

func importTestBody(t *testing.T) string {
	t.Helper()
	body, err := json.Marshal(sftpimport.Options{Host: "sftp.example.com", Port: 22, Username: "transfer-user", Password: "private-test-password", Fingerprint: "SHA256:" + strings.Repeat("A", 43), RemotePath: ".", MaxBytes: 1024, TimeoutMinutes: 1})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func waitImportJob(t *testing.T, f *reliabilityFixture, state string) sftpImportJob {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		f.router.sftpImports.mu.Lock()
		rec := f.router.sftpImports.jobs["alpha"]
		var job sftpImportJob
		if rec != nil {
			job = rec.job
		}
		f.router.sftpImports.mu.Unlock()
		if job.State == state {
			return job
		}
		if job.State != "" && !importActive(job.State) {
			t.Fatalf("job = %+v, want %s", job, state)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for import state", state)
	return sftpImportJob{}
}

func TestSFTPImportPermissions(t *testing.T) {
	f := newReliabilityFixture(t)
	routes := []struct{ method, path string }{{"GET", importURL}, {"POST", importURL}, {"POST", importURL + "/host-key"}, {"POST", importURL + "/test"}, {"DELETE", importURL + "/missing"}}
	for _, route := range routes {
		for _, user := range []string{"", "stranger", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request(route.method, route.path, "{}", user); w.Code != want {
				t.Fatalf("%s %s %s: %d %s", user, route.method, route.path, w.Code, w.Body.String())
			}
		}
	}
	subs, _ := f.db.ListServerSubusers("alpha")
	for _, perms := range [][]string{{"files.view", "files.edit"}, {"files.view", "sftp.access"}, {"files.edit", "sftp.access"}, {"files.view", "files.edit", "sftp.access"}} {
		if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, perms); err != nil {
			t.Fatal(err)
		}
		want := 403
		if len(perms) == 3 {
			want = 200
		}
		if w := f.request("GET", importURL, "", "subuser"); w.Code != want {
			t.Fatal(perms, w.Code, w.Body.String())
		}
	}
	for _, user := range []string{"owner", "admin"} {
		if w := f.request("GET", importURL, "", user); w.Code != 200 {
			t.Fatal(user, w.Code, w.Body.String())
		}
	}
}

func TestSFTPImportEligibilityAndInputBounds(t *testing.T) {
	f := newReliabilityFixture(t)
	w := f.request("GET", importURL, "", "owner")
	if !strings.Contains(w.Body.String(), `"eligible":true`) {
		t.Fatal(w.Body.String())
	}
	for _, body := range []string{"null", "{}{}", `{"overwrite":true}`, strings.Repeat(" ", sftpImportBodyLimit+1)} {
		w = f.request("POST", importURL, body, "owner")
		if w.Code != 400 && w.Code != 413 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	body := strings.Replace(importTestBody(t), `"max_bytes":1024`, `"max_bytes":5368709120`, 1)
	if w = f.request("POST", importURL, body, "owner"); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := os.WriteFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "eula.txt"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if w = f.request("POST", importURL, importTestBody(t), "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = f.request("GET", importURL, "", "owner"); !strings.Contains(w.Body.String(), `"eligible":false`) {
		t.Fatal(w.Body.String())
	}
	if entries, _ := os.ReadDir(f.manager.GetWorkingDir("alpha")); len(entries) != 1 {
		t.Fatal("nonempty target changed")
	}
}

func TestSFTPImportPublishesAndDoesNotRetainSecrets(t *testing.T) {
	f := newReliabilityFixture(t)
	h := f.router.sftpImports
	h.transfer = func(ctx context.Context, opts sftpimport.Options, allowed []string, stage string, progress func(sftpimport.Progress)) error {
		if opts.Password != "private-test-password" {
			return fmt.Errorf("credentials not passed to transport")
		}
		if entries, _ := os.ReadDir(f.manager.GetWorkingDir("alpha")); len(entries) != 0 {
			return fmt.Errorf("live tree changed before publication")
		}
		progress(sftpimport.Progress{State: "transferring", FilesTotal: 1, FilesDone: 1, BytesTotal: 5, BytesDone: 5, CurrentPath: "world.dat"})
		return os.WriteFile(filepath.Join(stage, "world.dat"), []byte("world"), 0600)
	}
	w := f.request("POST", importURL, importTestBody(t), "owner")
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	job := waitImportJob(t, f, "completed")
	if job.SafetyCopy == "" || job.FilesDone != 1 || f.manager.GetStatus("alpha") != "offline" {
		t.Fatal(job)
	}
	if data, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "world.dat")); err != nil || string(data) != "world" {
		t.Fatal(string(data), err)
	}
	w = f.request("GET", importURL, "", "owner")
	for _, secret := range []string{"private-test-password", "transfer-user", "sftp.example.com", "private_key", "passphrase"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatal("job leaked", secret)
		}
	}
	if w = f.request("DELETE", importURL+"/"+job.ID, "", "owner"); w.Code != 200 || !strings.Contains(w.Body.String(), `"state":"completed"`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestSFTPImportCancellationAndDuplicateStart(t *testing.T) {
	f := newReliabilityFixture(t)
	entered := make(chan struct{})
	f.router.sftpImports.transfer = func(ctx context.Context, _ sftpimport.Options, _ []string, stage string, progress func(sftpimport.Progress)) error {
		if err := os.WriteFile(filepath.Join(stage, "partial"), []byte("partial"), 0600); err != nil {
			return err
		}
		progress(sftpimport.Progress{State: "transferring"})
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}
	w := f.request("POST", importURL, importTestBody(t), "owner")
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	var job sftpImportJob
	if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("transfer not started")
	}
	if w = f.request("POST", importURL, importTestBody(t), "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = f.request("DELETE", importURL+"/old-job-id", "", "owner"); w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = f.request("DELETE", importURL+"/"+job.ID, "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	waitImportJob(t, f, "cancelled")
	if entries, _ := os.ReadDir(f.manager.GetWorkingDir("alpha")); len(entries) != 0 {
		t.Fatal("partial transfer published")
	}
}

func TestSFTPImportRevokedSessionCancelsTransfer(t *testing.T) {
	f := newReliabilityFixture(t)
	entered := make(chan struct{})
	f.router.sftpImports.transfer = func(ctx context.Context, _ sftpimport.Options, _ []string, _ string, _ func(sftpimport.Progress)) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}
	if w := f.request("POST", importURL, importTestBody(t), "owner"); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("transfer not started")
	}
	if err := f.db.UpdateUserPassword(f.users["owner"].ID, "new-hash"); err != nil {
		t.Fatal(err)
	}
	waitImportJob(t, f, "cancelled")
	if entries, _ := os.ReadDir(f.manager.GetWorkingDir("alpha")); len(entries) != 0 {
		t.Fatal("revoked transfer published")
	}
}

func TestSFTPImportWorkerProxyKeepsUserIdentity(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	prefix := "/api/nodes/remote/proxy"
	if w := master.request("GET", prefix+importURL, "", "subuser"); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	worker.router.sftpImports.transfer = func(ctx context.Context, _ sftpimport.Options, _ []string, stage string, _ func(sftpimport.Progress)) error {
		return os.WriteFile(filepath.Join(stage, "imported"), []byte("worker"), 0600)
	}
	if w := master.request("POST", prefix+importURL, importTestBody(t), "admin"); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	waitImportJob(t, worker, "completed")
	worker.router.sftpImports.mu.Lock()
	id := worker.router.sftpImports.jobs["alpha"].userID
	worker.router.sftpImports.mu.Unlock()
	if id != master.users["admin"].ID {
		t.Fatal("admin became a machine job", id)
	}
	if entries, _ := os.ReadDir(master.manager.GetWorkingDir("alpha")); len(entries) != 0 {
		t.Fatal("proxy wrote on master")
	}
}

func TestSFTPImportNoPrivateConnectionByDefault(t *testing.T) {
	f := newReliabilityFixture(t)
	w := f.request(http.MethodPost, importURL+"/host-key", `{"host":"127.0.0.1","port":22}`, "owner")
	if w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestSFTPImportHistoryAndProbeConcurrencyAreBounded(t *testing.T) {
	f := newReliabilityFixture(t)
	h := f.router.sftpImports
	for i := 0; i < cap(h.probes); i++ {
		h.probes <- struct{}{}
	}
	w := f.request("POST", importURL+"/host-key", `{"host":"127.0.0.1","port":22}`, "owner")
	if w.Code != 429 {
		t.Fatal(w.Code, w.Body.String())
	}
	for i := 0; i < cap(h.probes); i++ {
		<-h.probes
	}
	for i := 0; i < sftpImportHistoryLimit; i++ {
		id := fmt.Sprintf("finished-%d", i)
		h.jobs[id] = &sftpImportRecord{job: sftpImportJob{Progress: sftpimport.Progress{State: "completed"}, UpdatedAt: time.Now()}}
	}
	h.transfer = func(context.Context, sftpimport.Options, []string, string, func(sftpimport.Progress)) error {
		return nil
	}
	if w = f.request("POST", importURL, importTestBody(t), "owner"); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	waitImportJob(t, f, "completed")
	h.mu.Lock()
	count := len(h.jobs)
	h.mu.Unlock()
	if count != sftpImportHistoryLimit {
		t.Fatal("unbounded job history", count)
	}
}

func TestSFTPImportAttemptLimitsArePerAccount(t *testing.T) {
	f := newReliabilityFixture(t)
	h := f.router.sftpImports
	request := func(user string) *http.Request {
		r := httptest.NewRequest("POST", importURL, nil)
		claims := &JWTClaims{UserID: f.users[user].ID, Role: f.users[user].Role}
		return r.WithContext(context.WithValue(r.Context(), contextUserKey, claims))
	}
	for i := 0; i < 6; i++ {
		if !h.allowAttempt(request("owner")) {
			t.Fatal("normal connection workflow exhausted burst", i)
		}
	}
	if h.allowAttempt(request("owner")) {
		t.Fatal("connection attempt throttle not enforced")
	}
	if !h.allowAttempt(request("admin")) {
		t.Fatal("one account exhausted another account's proxy quota")
	}
}

func TestSFTPImportWorkerCancelsWhenOriginSessionExpires(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	worker.router.sftpImports.transfer = func(ctx context.Context, _ sftpimport.Options, _ []string, _ string, _ func(sftpimport.Progress)) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}
	user := master.users["owner"]
	claims := JWTClaims{UserID: user.ID, Username: user.Username, Role: user.Role, AuthVersion: user.AuthVersion, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Second))}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(master.router.auth.jwtSecret)
	if err != nil {
		t.Fatal(err)
	}
	master.tokens["owner"] = token
	w := master.request("POST", "/api/nodes/remote/proxy"+importURL, importTestBody(t), "owner")
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("worker transfer not started")
	}
	waitImportJob(t, worker, "cancelled")
	if entries, _ := os.ReadDir(worker.manager.GetWorkingDir("alpha")); len(entries) != 0 {
		t.Fatal("expired session published files")
	}
}

func TestSFTPImportUntrustedExpiryHeaderIsIgnored(t *testing.T) {
	f := newReliabilityFixture(t)
	r := httptest.NewRequest("POST", importURL, nil)
	r.Header.Set(sftpImportExpiryHeader, "1")
	r.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
	claims, err := f.router.auth.ValidateAuthToken(f.tokens["owner"])
	if err != nil {
		t.Fatal(err)
	}
	r = r.WithContext(context.WithValue(r.Context(), contextUserKey, claims))
	authority, err := f.router.sftpImports.captureAuthority(r)
	if err != nil || !authority.expiresAt.Equal(claims.ExpiresAt.Time) {
		t.Fatal("untrusted header changed expiry", authority.expiresAt, err)
	}
}
