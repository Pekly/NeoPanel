package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/runner"
)

func TestBackupRestoreAPIConfirmationAuthorizationAndReplacement(t *testing.T) {
	f := newReliabilityFixture(t)
	live := f.manager.GetWorkingDir("alpha")
	os.WriteFile(filepath.Join(live, "world"), []byte("saved"), 0600)
	if w := f.request("POST", "/api/servers/alpha/backups", `{"name":"snapshot"}`, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	os.WriteFile(filepath.Join(live, "world"), []byte("changed"), 0600)
	os.WriteFile(filepath.Join(live, "extra"), []byte("remove"), 0600)
	path := "/api/servers/alpha/backups/restore"
	for _, body := range []string{`{}`, `{"name":"snapshot.tar.gz"}`, `{"name":"snapshot.tar.gz","confirm":false}`, `{"name":"../snapshot.tar.gz","confirm":true}`, `{"name":"snapshot.tar.gz","confirm":"true"}`} {
		if w := f.request("POST", path, body, "owner"); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
		got, _ := os.ReadFile(filepath.Join(live, "world"))
		if string(got) != "changed" {
			t.Fatal("unconfirmed restore changed data")
		}
	}
	body := `{"name":"snapshot.tar.gz","confirm":true}`
	for _, user := range []string{"stranger", "subuser"} {
		if w := f.request("POST", path, body, user); w.Code != 403 {
			t.Fatal(user, w.Code, w.Body.String())
		}
	}
	if w := f.request("POST", path, `{"name":"missing.tar.gz","confirm":true}`, "owner"); w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
	w := f.request("POST", path, body, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result runner.RestoreResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.SafetyCopy != "" || !strings.Contains(w.Body.String(), `"previous_files_preserved":true`) {
		t.Fatal(result, err)
	}
	got, _ := os.ReadFile(filepath.Join(live, "world"))
	if string(got) != "saved" {
		t.Fatal("archive not restored")
	}
	if _, err := os.Stat(filepath.Join(live, "extra")); !os.IsNotExist(err) {
		t.Fatal("extra file survived replacement")
	}
	w = f.request("POST", path, body, "admin")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.SafetyCopy == "" {
		t.Fatal("administrator lost recovery path", result, err)
	}
}

func TestBackupRestoreErrorsKeepPathsForAdministrators(t *testing.T) {
	for _, cause := range []error{runner.ErrRestoreRecovery, runner.ErrRestoreRetention, runner.ErrServerNotStopped, runner.ErrServerOperation, runner.ErrInvalidArchive, os.ErrNotExist} {
		err := fmt.Errorf("host /private/data/.restore-hidden: %w", cause)
		for _, role := range []string{"user", "admin"} {
			r := httptest.NewRequest("POST", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), contextUserKey, &JWTClaims{Role: role}))
			message := restoreErrorForViewer(r, err)
			if strings.Contains(message, "/private/data") != (role == "admin") {
				t.Fatalf("%s: %s", role, message)
			}
			if role == "user" && cause == runner.ErrRestoreRecovery && !strings.Contains(message, "Keep the server stopped") {
				t.Fatal("recovery warning removed", message)
			}
		}
	}
	f := newReliabilityFixture(t)
	w := f.request("POST", "/api/servers/alpha/backups/restore", `{"name":"missing.tar.gz","confirm":true}`, "owner")
	if w.Code != 404 || !strings.Contains(w.Body.String(), "Refresh the backup list") || strings.Contains(w.Body.String(), f.cfg.Paths.DataDir) {
		t.Fatal(w.Code, w.Body.String())
	}
}
