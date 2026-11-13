package runner

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestBackupCompleteArchiveAndConcurrentCollision(t *testing.T) {
	source := t.TempDir()
	dest := t.TempDir()
	content := strings.Repeat("important world contents", 1000)
	if err := os.WriteFile(filepath.Join(source, "world.dat"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := CreateBackup(source, dest, "world.tar.gz"); results <- err }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, os.ErrExist) {
			t.Error(err)
		}
	}
	if successes != 1 {
		t.Fatalf("%d publications succeeded", successes)
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 1 || entries[0].Name() != "world.tar.gz" {
		t.Fatalf("temporary archive leaked: %v", entries)
	}
	f, err := os.Open(filepath.Join(dest, "world.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	header, err := tr.Next()
	if err != nil || header.Name != "world.dat" {
		t.Fatal(header, err)
	}
	got, err := io.ReadAll(tr)
	if err != nil || string(got) != content {
		t.Fatal("invalid archive contents", err)
	}
	if _, err := tr.Next(); err != io.EOF {
		t.Fatal("tar not finalized", err)
	}
	if _, err := io.Copy(io.Discard, gz); err != nil {
		t.Fatal("gzip footer corrupt", err)
	}
}

type failBackupWriter struct{ remaining int }

func (w *failBackupWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		return 0, io.ErrClosedPipe
	}
	w.remaining -= len(p)
	return len(p), nil
}

func TestBackupErrorsPropagateAndDoNotPublish(t *testing.T) {
	source := t.TempDir()
	dest := t.TempDir()
	os.WriteFile(filepath.Join(source, "data"), []byte("data"), 0600)
	// The first ten bytes allow the gzip header; failure then occurs on final flush.
	if err := writeBackupArchive(source, &failBackupWriter{remaining: 10}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("lost writer/close error: %v", err)
	}
	if err := writeBackupArchive(filepath.Join(source, "missing"), io.Discard); err == nil {
		t.Fatal("lost walk error")
	}
	if _, err := CreateBackup(filepath.Join(source, "missing"), dest, "bad.tar.gz"); err == nil {
		t.Fatal("missing source succeeded")
	}
	if _, err := CreateBackup(source, filepath.Join(source, "backups"), "recursive.tar.gz"); err == nil {
		t.Fatal("recursive backup allowed")
	}
	file := filepath.Join(dest, "not-directory")
	os.WriteFile(file, []byte("keep"), 0600)
	if _, err := CreateBackup(source, file, "bad.tar.gz"); err == nil {
		t.Fatal("invalid destination succeeded")
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 1 {
		t.Fatal("failed archive published", entries)
	}
}

func TestScheduledActionsPersistActualCompletion(t *testing.T) {
	root := t.TempDir()
	db, err := database.InitDB(filepath.Join(root, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	srv := &database.Server{ID: "server", Name: "test", Type: "paper", Version: "1.21", JarFile: "server.jar"}
	if err := db.CreateServer(srv); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(filepath.Join(root, "servers"), db)
	defer m.ShutdownAll()
	// Intentionally non-default location: manual and scheduled backups must agree.
	dest := filepath.Join(root, "custom-data", "backups")
	m.SetBackupDirectory(dest)
	os.MkdirAll(m.GetWorkingDir(srv.ID), 0755)
	os.WriteFile(filepath.Join(m.GetWorkingDir(srv.ID), "world"), []byte("world"), 0600)
	wait := func(id string) *database.ServerSchedule {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			s, err := db.GetServerSchedule(srv.ID, id)
			if err != nil {
				t.Fatal(err)
			}
			if s.LastStatus == "succeeded" || s.LastStatus == "failed" {
				return s
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("schedule did not complete")
		return nil
	}
	s := &database.ServerSchedule{ID: "backup", ServerID: srv.ID, Name: "backup", CronExpr: "* * * * *", Action: "backup", IsActive: true}
	if err := db.CreateServerSchedule(s); err != nil {
		t.Fatal(err)
	}
	if err := m.StartSchedule(s); err != nil {
		t.Fatal(err)
	}
	result := wait(s.ID)
	if result.LastStatus != "succeeded" || result.LastError != "" || result.LastRunAt == nil || result.LastFinishedAt == nil {
		t.Fatal(result)
	}
	entries, err := os.ReadDir(filepath.Join(dest, srv.ID))
	if err != nil || len(entries) != 1 {
		t.Fatal("scheduled archive not published", entries, err)
	}
	os.RemoveAll(m.GetWorkingDir(srv.ID))
	if err := m.StartSchedule(s); err != nil {
		t.Fatal(err)
	}
	result = wait(s.ID)
	if result.LastStatus != "failed" || !strings.Contains(result.LastError, "source") {
		t.Fatalf("backup error not persisted: %+v", result)
	}
	cmd := &database.ServerSchedule{ID: "command", ServerID: srv.ID, Name: "command", CronExpr: "* * * * *", Action: "command", Payload: "say hi", IsActive: true}
	db.CreateServerSchedule(cmd)
	if err := m.StartSchedule(cmd); err != nil {
		t.Fatal(err)
	}
	if result := wait(cmd.ID); result.LastStatus != "failed" || result.LastError == "" {
		t.Fatal("offline command reported success", result)
	}
}
