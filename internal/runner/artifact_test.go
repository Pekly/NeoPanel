package runner

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/serverjar"
)

func artifactJar(t *testing.T, path, manifest string) {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, err := z.Create("META-INF/MANIFEST.MF")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(w, manifest); err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

type artifactTransport func(*http.Request) (*http.Response, error)

func (f artifactTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestArtifactExactRequirementAndHashInvalidation(t *testing.T) {
	db, base, srv, _ := restartFixture(t)
	m := NewServerManager(base, db)
	t.Cleanup(m.ShutdownAll)
	srv.JavaPath = "auto"
	srv.JarFile = "server.jar"
	path := filepath.Join(m.GetWorkingDir(srv.ID), srv.JarFile)
	artifactJar(t, path, "Manifest-Version: 1.0\r\nMain-Class: example.Main\r\n\r\n")
	local, err := serverjar.Inspect(context.Background(), filepath.Dir(path), srv.JarFile)
	if err != nil {
		t.Fatal(err)
	}
	local.Source, local.Status, local.Software, local.MinecraftVersion = "mcjars", "identified", "paper", "1.21.4"
	local.RequiredJava, local.JavaSource = 21, "mcjars-version"
	if err := m.RecordArtifactInfo(srv, local); err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = artifactTransport(func(r *http.Request) (*http.Response, error) {
		t.Error("launch resolution made a network request")
		return nil, errors.New("no network")
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	major, err := m.resolveAutoJava(srv)
	if err != nil || major != "21" {
		t.Fatalf("major=%q err=%v", major, err)
	}
	if got := m.ArtifactSummary(srv); got == nil || got.MinecraftVersion != "1.21.4" {
		t.Fatalf("missing summary: %+v", got)
	}
	// Deliberately preserve length and mtime: cached stat is NOT launch evidence.
	fi, _ := os.Stat(path)
	// Rebuild ZIP to retain valid CRCs, then preserve timestamp.
	artifactJar(t, path, "Manifest-Version: 1.0\r\nMain-Class: changed.Main\r\n\r\n")
	_ = os.Chtimes(path, fi.ModTime(), fi.ModTime())
	if _, err = m.resolveAutoJava(srv); err == nil {
		t.Fatal("changed bytes reused old exact requirement")
	}
	got, _ := m.db.GetServer(srv.ID)
	if got.Type != srv.Type || got.Version != srv.Version {
		t.Fatal("inspection changed configured identity")
	}
}

func TestArtifactSummaryRejectsReplacedFileAndFilename(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	srv.JarFile = "server.jar"
	path := filepath.Join(m.GetWorkingDir(srv.ID), srv.JarFile)
	artifactJar(t, path, "Manifest-Version: 1.0\r\n\r\n")
	info, err := serverjar.Inspect(context.Background(), filepath.Dir(path), srv.JarFile)
	if err != nil {
		t.Fatal(err)
	}
	m.cacheArtifact(srv, info)
	srv.JarFile = "other.jar"
	if m.ArtifactSummary(srv) != nil {
		t.Fatal("filename retained identity")
	}
	srv.JarFile = "server.jar"
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	artifactJar(t, path, "Manifest-Version: 1.0\r\nMain-Class: Other\r\n\r\n")
	if m.ArtifactSummary(srv) != nil {
		t.Fatal("replacement retained identity")
	}
}

func TestArtifactAutoRequiresExactEvidence(t *testing.T) {
	for _, major := range []int{7, 8, 11, 16, 17, 21, 25} {
		got, err := autoJavaMajor(serverjar.Info{RequiredJava: major})
		if err != nil || got == "" {
			t.Fatal(major, err)
		}
		if major == 7 && got != "8" {
			t.Fatal("legacy policy changed")
		}
	}
	for _, info := range []serverjar.Info{{MinimumJava: 21}, {}, {RequiredJava: 99}} {
		if _, err := autoJavaMajor(info); err == nil {
			t.Fatalf("accepted uncertain/unsupported evidence: %+v", info)
		}
	}
}

func TestArtifactInspectionProviderFailureAndBusy(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	srv.JarFile = "server.jar"
	path := filepath.Join(m.GetWorkingDir(srv.ID), srv.JarFile)
	artifactJar(t, path, "Manifest-Version: 1.0\r\n\r\n")
	release, err := rooted.TryMutation(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.InspectArtifact(context.Background(), srv); !errors.Is(err, ErrArtifactBusy) {
		t.Fatal(err)
	}
	release()
	original := http.DefaultTransport
	calls := 0
	http.DefaultTransport = artifactTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: r}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	info, err := m.InspectArtifact(context.Background(), srv)
	if err != nil || info.RequiredJava != 0 || info.Software != "" || !strings.Contains(info.Message, "unavailable") {
		t.Fatalf("%+v %v", info, err)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
	_, _ = m.InspectArtifact(context.Background(), srv)
	if calls != 1 {
		t.Fatal("cache caused extra provider work")
	}
	_, _ = m.InspectArtifact(context.Background(), srv, true)
	if calls != 2 {
		t.Fatal("provider failure prevented retry")
	}
}

func TestArtifactAutoStartExcludesManagedWritersAndReleasesOnFailure(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	srv.JavaPath = "auto"
	inst := m.getOrCreateInstance(srv)
	if inst.resolveAutoJava == nil {
		t.Fatal("manager did not wire Auto resolution")
	}
	want := errors.New("stop before executing any process")
	called := false
	inst.resolveAutoJava = func(*database.Server) (string, error) {
		called = true
		if release, err := rooted.TryMutation(inst.workingDir); !errors.Is(err, rooted.ErrMutationBusy) {
			if release != nil {
				release()
			}
			t.Fatalf("Auto inspection allowed writer: %v", err)
		}
		return "", want
	}
	release, err := rooted.TryMutation(inst.workingDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := inst.Start(); err == nil || called {
		release()
		t.Fatal("Auto bypassed active writer", err)
	}
	release()
	if err := inst.Start(); !errors.Is(err, want) || !called {
		t.Fatal(err, called)
	}
	release, err = rooted.TryMutation(inst.workingDir)
	if err != nil {
		t.Fatal("failed launch leaked lease", err)
	}
	release()
	srv.JavaPath = filepath.Join(t.TempDir(), "missing-manual-java")
	called = false
	if err := inst.Start(); err == nil || called {
		t.Fatal("manual runtime went through Auto", err)
	}
}
