package runner_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/api"
	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type maintenanceTransport func(*http.Request) (*http.Response, error)

func (f maintenanceTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDestructiveAPIsShareRestoreGate(t *testing.T) {
	root := t.TempDir()
	db, err := database.InitDB(filepath.Join(root, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cfg := &config.Config{}
	cfg.Paths.DataDir = root
	cfg.Paths.ServersDir = filepath.Join(root, "servers")
	m := runner.NewServerManager(cfg.Paths.ServersDir, db)
	defer m.ShutdownAll()
	srv := &database.Server{ID: "alpha", Name: "alpha", Type: "paper", Version: "1.21", JarFile: "server.jar"}
	if err := db.CreateServer(srv); err != nil {
		t.Fatal(err)
	}
	live := m.GetWorkingDir(srv.ID)
	os.MkdirAll(live, 0755)
	os.WriteFile(filepath.Join(live, "original"), []byte("keep"), 0600)
	if _, err := m.CreateServerBackup(srv.ID, "saved.tar.gz"); err != nil {
		t.Fatal(err)
	}
	software := api.NewSoftwareHandler(db, m)
	servers := api.NewServerHandler(cfg, db, m)
	request := func(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.SetPathValue("id", srv.ID)
		w := httptest.NewRecorder()
		handler(w, r)
		return w
	}
	// Hold the exact gate acquired by RestoreBackup across its extraction/swap.
	release, err := m.BeginServerOperation(srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, handler := range []http.HandlerFunc{software.ReinstallServer, software.ChangeSoftware, servers.DeleteServer} {
		w := request(handler, `{"software":"paper","version":"1.21"}`)
		if w.Code != 409 {
			t.Fatal(w.Code, w.Body.String())
		}
		got, _ := os.ReadFile(filepath.Join(live, "original"))
		if string(got) != "keep" {
			t.Fatal("conflicting operation changed original")
		}
		if _, err := db.GetServer(srv.ID); err != nil {
			t.Fatal("conflicting delete removed metadata", err)
		}
	}
	release()
	// A real reinstall is paused in its download, after cleanup. Restore/delete
	// must remain excluded until the download AND metadata update finish.
	entered, proceed := make(chan struct{}), make(chan struct{})
	oldTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = oldTransport }()
	http.DefaultTransport = maintenanceTransport(func(r *http.Request) (*http.Response, error) {
		close(entered)
		<-proceed
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("fresh jar")), Request: r}, nil
	})
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- request(software.ReinstallServer, `{"software":"paper","version":"1.21"}`) }()
	<-entered
	if _, err := m.RestoreBackup(context.Background(), srv, "saved.tar.gz"); !errors.Is(err, runner.ErrServerOperation) {
		t.Fatal("restore bypassed reinstall", err)
	}
	if w := request(servers.DeleteServer, ""); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	close(proceed)
	if w := <-done; w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	got, _ := os.ReadFile(filepath.Join(live, "server.jar"))
	if string(got) != "fresh jar" {
		t.Fatal("reinstall failed")
	}
	result, err := m.RestoreBackup(context.Background(), srv, "saved.tar.gz")
	if err != nil {
		t.Fatal("reinstall gate leaked", err)
	}
	got, _ = os.ReadFile(filepath.Join(live, "original"))
	if string(got) != "keep" {
		t.Fatal("restore failed")
	}
	if w := request(servers.DeleteServer, ""); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := os.Stat(live); !os.IsNotExist(err) {
		t.Fatal("delete left live files", err)
	}
	if _, err := db.GetServer(srv.ID); err == nil {
		t.Fatal("delete left metadata")
	}
	if _, err := os.Stat(filepath.Join(cfg.Paths.ServersDir, result.SafetyCopy)); err != nil {
		t.Fatal("delete removed restore safety copy", err)
	}
	release, err = m.BeginServerOperation(srv.ID)
	if err != nil {
		t.Fatal("delete gate leaked", err)
	}
	release()
}
