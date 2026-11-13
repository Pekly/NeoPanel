package runner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func restoreArchive(t *testing.T, headers ...tar.Header) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, h := range headers {
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if h.Size > 0 {
			if _, err := tw.Write(bytes.Repeat([]byte("x"), int(h.Size))); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func restoreFixture(t *testing.T) (*ServerManager, *database.Server, string) {
	t.Helper()
	root := t.TempDir()
	m := NewServerManager(filepath.Join(root, "servers"), nil)
	t.Cleanup(m.ShutdownAll)
	srv := &database.Server{ID: "alpha", DiskLimitMB: 8}
	live := m.GetWorkingDir(srv.ID)
	if err := os.MkdirAll(live, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(live, "original"), []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	backupDir := filepath.Join(root, "backups", srv.ID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatal(err)
	}
	return m, srv, backupDir
}

func TestRestoreFullReplacementAndSafetyCopy(t *testing.T) {
	m, srv, dir := restoreFixture(t)
	source := t.TempDir()
	os.MkdirAll(filepath.Join(source, "world"), 0755)
	os.WriteFile(filepath.Join(source, "world", "level.dat"), []byte("restored world"), 0600)
	if _, err := CreateBackup(source, dir, "valid.tar.gz"); err != nil {
		t.Fatal(err)
	}
	result, err := m.RestoreBackup(context.Background(), srv, "valid.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if result.SafetyCopy == "" || m.GetStatus(srv.ID) != "offline" {
		t.Fatal(result)
	}
	live := m.GetWorkingDir(srv.ID)
	got, err := os.ReadFile(filepath.Join(live, "world", "level.dat"))
	if err != nil || string(got) != "restored world" {
		t.Fatal(string(got), err)
	}
	if _, err := os.Stat(filepath.Join(live, "original")); !os.IsNotExist(err) {
		t.Fatal("restore merged instead of replaced")
	}
	got, err = os.ReadFile(filepath.Join(m.baseDir, result.SafetyCopy, "original"))
	if err != nil || string(got) != "keep me" {
		t.Fatal("safety copy missing", err)
	}
}

func TestRestoreRejectsArchiveWithoutChangingLiveData(t *testing.T) {
	good := restoreArchive(t, tar.Header{Name: "new", Typeflag: tar.TypeReg, Size: 4, Mode: 0600})
	cases := map[string][]byte{
		"truncated-gzip":  good[:len(good)-5],
		"not-gzip":        []byte("broken"),
		"duplicate":       restoreArchive(t, tar.Header{Name: "x"}, tar.Header{Name: "x"}),
		"file-parent":     restoreArchive(t, tar.Header{Name: "x"}, tar.Header{Name: "x/y"}),
		"implicit-parent": restoreArchive(t, tar.Header{Name: "x/y"}, tar.Header{Name: "x"}),
		"case-alias":      restoreArchive(t, tar.Header{Name: "x"}, tar.Header{Name: "X"}),
		"symlink":         restoreArchive(t, tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "../../outside"}),
		"hardlink":        restoreArchive(t, tar.Header{Name: "link", Typeflag: tar.TypeLink, Linkname: "../outside"}),
		"device":          restoreArchive(t, tar.Header{Name: "device", Typeflag: tar.TypeChar}),
	}
	for _, name := range []string{"../outside", "a/../outside", "/absolute", "a\\outside", "C:/outside", "./x", "a//x", "x. ", "a/.."} {
		cases[name] = restoreArchive(t, tar.Header{Name: name})
	}
	// Valid gzip containing a truncated tar (including complete last file data)
	// must not be accepted just because tar.Reader tolerates absent end blocks.
	gz, _ := gzip.NewReader(bytes.NewReader(good))
	raw, _ := io.ReadAll(gz)
	gz.Close()
	for _, cut := range []int{1024, 1536} {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		w.Write(raw[:len(raw)-cut])
		w.Close()
		cases["truncated-tar-"+strconv.Itoa(cut)] = b.Bytes()
	}
	corrupt := append([]byte(nil), good...)
	corrupt[len(corrupt)-8] ^= 1
	cases["bad-crc"] = corrupt
	for label, archive := range cases {
		t.Run(label, func(t *testing.T) {
			m, srv, dir := restoreFixture(t)
			os.WriteFile(filepath.Join(dir, "bad.tar.gz"), archive, 0600)
			if _, err := m.RestoreBackup(context.Background(), srv, "bad.tar.gz"); !errors.Is(err, ErrInvalidArchive) {
				t.Fatalf("expected invalid archive, got %v", err)
			}
			got, err := os.ReadFile(filepath.Join(m.GetWorkingDir(srv.ID), "original"))
			if err != nil || string(got) != "keep me" {
				t.Fatal("live data changed", err)
			}
			entries, _ := os.ReadDir(m.GetWorkingDir(srv.ID))
			if len(entries) != 1 {
				t.Fatal("partial extraction in live tree")
			}
			stages, _ := filepath.Glob(filepath.Join(m.baseDir, ".restore-*"))
			if len(stages) != 0 {
				t.Fatal("failed extraction leaked staging", stages)
			}
		})
	}
}

func TestRestoreBudgetsAndCancellation(t *testing.T) {
	archive := restoreArchive(t, tar.Header{Name: "a/b", Size: 10}, tar.Header{Name: "c"})
	for _, limits := range []struct {
		bytes   int64
		entries int
	}{{9, 100}, {100, 1}, {100, 2}} {
		if err := extractRestoreArchive(context.Background(), bytes.NewReader(archive), t.TempDir(), limits.bytes, limits.entries); !errors.Is(err, ErrInvalidArchive) {
			t.Fatal("budget not enforced", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := extractRestoreArchive(ctx, bytes.NewReader(archive), t.TempDir(), 100, 100); err == nil {
		t.Fatal("cancel ignored")
	}
}

func TestRestoreStoppedAndConcurrentPowerGate(t *testing.T) {
	m, srv, dir := restoreFixture(t)
	os.WriteFile(filepath.Join(dir, "ok.tar.gz"), restoreArchive(t, tar.Header{Name: "new"}), 0600)
	inst := NewProcessInstance(srv, m.GetWorkingDir(srv.ID), DefaultStartupTimeout, nil, nil)
	m.instances[srv.ID] = inst
	for _, status := range []string{"online", "starting", "stopping"} {
		inst.mu.Lock()
		inst.status = status
		inst.mu.Unlock()
		if _, err := m.RestoreBackup(context.Background(), srv, "ok.tar.gz"); !errors.Is(err, ErrServerNotStopped) {
			t.Fatalf("%s: %v", status, err)
		}
	}
	inst.mu.Lock()
	inst.status = "offline"
	inst.mu.Unlock()
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	results := make(chan error, 4)
	go func() { results <- m.Start(srv) }()
	go func() { results <- m.Restart(srv) }()
	go func() { results <- m.Stop(srv.ID, true) }()
	go func() { _, err := m.RestoreBackup(context.Background(), srv, "ok.tar.gz"); results <- err }()
	for i := 0; i < 4; i++ {
		if err := <-results; !errors.Is(err, ErrServerOperation) {
			t.Fatal("operation bypassed lock", err)
		}
	}
	if _, err := m.CreateServerBackup(srv.ID, "another.tar.gz"); !errors.Is(err, ErrServerOperation) {
		t.Fatal(err)
	}
}

func TestRestorePublicationRollback(t *testing.T) {
	for _, failRollback := range []bool{false, true} {
		root := t.TempDir()
		live, fresh, old := filepath.Join(root, "live"), filepath.Join(root, "fresh"), filepath.Join(root, "old")
		os.Mkdir(live, 0755)
		os.Mkdir(fresh, 0755)
		os.WriteFile(filepath.Join(live, "original"), []byte("keep"), 0600)
		calls := 0
		err := publishRestore(live, fresh, old, func(from, to string) error {
			calls++
			if calls == 2 || (failRollback && calls == 3) {
				return errors.New("injected rename failure")
			}
			return os.Rename(from, to)
		})
		if err == nil {
			t.Fatal("failed swap reported success")
		}
		expected := live
		if failRollback {
			expected = old
			if !errors.Is(err, ErrRestoreRecovery) || !strings.Contains(err.Error(), old) {
				t.Fatal("missing recovery information", err)
			}
		}
		got, e := os.ReadFile(filepath.Join(expected, "original"))
		if e != nil || string(got) != "keep" {
			t.Fatal("original lost", e)
		}
	}
}

func TestRestoreRejectsSymlinkRootsAndPreservesExistingLinkTargets(t *testing.T) {
	m, srv, dir := restoreFixture(t)
	outside := t.TempDir()
	target := filepath.Join(outside, "valuable")
	os.WriteFile(target, []byte("outside"), 0600)
	if err := os.Symlink(target, filepath.Join(m.GetWorkingDir(srv.ID), "link")); err != nil {
		t.Skip("symlink creation unavailable: ", err)
	}
	os.WriteFile(filepath.Join(dir, "ok.tar.gz"), restoreArchive(t, tar.Header{Name: "new"}), 0600)
	if _, err := m.RestoreBackup(context.Background(), srv, "ok.tar.gz"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "outside" {
		t.Fatal("symlink target modified")
	}
	if err := os.Symlink(filepath.Join(dir, "ok.tar.gz"), filepath.Join(dir, "link.tar.gz")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.RestoreBackup(context.Background(), srv, "link.tar.gz"); !errors.Is(err, ErrInvalidArchive) {
		t.Fatal("archive symlink allowed", err)
	}
	os.RemoveAll(m.GetWorkingDir(srv.ID))
	if err := os.Symlink(outside, m.GetWorkingDir(srv.ID)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.RestoreBackup(context.Background(), srv, "ok.tar.gz"); !errors.Is(err, ErrInvalidArchive) {
		t.Fatal("server root symlink allowed", err)
	}
}

func TestRestoreSuppressesLateAutoRestart(t *testing.T) {
	m, srv, dir := restoreFixture(t)
	db, err := database.InitDB(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	srv.Name, srv.Type, srv.Version, srv.JarFile, srv.AutoRestart = "alpha", "paper", "1.21", "server.jar", true
	if err := db.CreateServer(srv); err != nil {
		t.Fatal(err)
	}
	m.db = db
	m.getOrCreateInstance(srv)
	retry := make(chan struct{})
	m.retryCancels[srv.ID] = retry
	os.WriteFile(filepath.Join(dir, "ok.tar.gz"), restoreArchive(t, tar.Header{Name: "new"}), 0600)
	if _, err := m.RestoreBackup(context.Background(), srv, "ok.tar.gz"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-retry:
	default:
		t.Fatal("pending restart not cancelled")
	}
	m.handleProcessExit(srv.ID, true, true, false, time.Hour)
	if len(m.retryCancels) != 0 {
		t.Fatal("late crash callback queued auto-start after restore")
	}
}
