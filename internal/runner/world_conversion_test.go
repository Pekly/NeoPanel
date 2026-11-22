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
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/database"
)

func TestWorldConversionLeaseExcludesStartsAndJoinsShutdown(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	l, e := m.beginConversion(srv)
	if e != nil {
		t.Fatal(e)
	}
	if unlock, e := m.beginOperation(srv.ID); e == nil {
		unlock()
		l.close()
		t.Fatal("operation gate not held")
	} else if !errors.Is(e, ErrServerOperation) {
		t.Fatal(e)
	}
	if lock, e := acquireRunLock(m.GetWorkingDir(srv.ID)); e == nil {
		lock.Close()
		l.close()
		t.Fatal("run lock not held")
	}
	done := make(chan struct{})
	go func() { m.ShutdownAll(); close(done) }()
	select {
	case <-l.ctx.Done():
	case <-time.After(3 * time.Second):
		l.close()
		t.Fatal("shutdown did not cancel auxiliary work")
	}
	select {
	case <-done:
		l.close()
		t.Fatal("shutdown did not join auxiliary work")
	default:
	}
	if e = l.fence(func() error { t.Fatal("published during shutdown"); return nil }); e == nil {
		t.Fatal("shutdown fence bypass")
	}
	l.close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown blocked after cleanup")
	}
}
func TestWorldConversionCompatibility(t *testing.T) {
	live := t.TempDir()
	os.WriteFile(filepath.Join(live, "server.properties"), []byte("level-name=world\n"), 0600)
	f, _ := chunker.Lookup("JAVA_1_20_5")
	srv := &database.Server{Type: "vanilla", Version: "1.20.5"}
	if e := conversionCompatible(srv, "world", f, live); e != nil {
		t.Fatal(e)
	}
	for _, kind := range []string{"paper", "fabric", "bedrock", "custom", "velocity"} {
		srv.Type = kind
		if conversionCompatible(srv, "world", f, live) == nil {
			t.Fatal(kind)
		}
	}
	srv.Type = "vanilla"
	srv.Version = "1.20.4"
	if conversionCompatible(srv, "world", f, live) == nil {
		t.Fatal("version mismatch accepted")
	}
	srv.Version = "1.20.5"
	bedrock, _ := chunker.Lookup("BEDROCK_R20_80")
	if conversionCompatible(srv, "world", bedrock, live) == nil {
		t.Fatal("Bedrock replace accepted")
	}
	os.Mkdir(filepath.Join(live, "world_nether"), 0700)
	if conversionCompatible(srv, "world", f, live) == nil {
		t.Fatal("split dimensions accepted")
	}
}
func TestWorldConversionWorldBackupCompleteAndCancelled(t *testing.T) {
	src, dest := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(src, "level.dat"), []byte("original"), 0600)
	if e := conversionWorldBackup(context.Background(), src, dest, "original.tar.gz"); e != nil {
		t.Fatal(e)
	}
	f, e := os.Open(filepath.Join(dest, "original.tar.gz"))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	h, e := tr.Next()
	if e != nil || h.Name != "level.dat" {
		t.Fatal(h, e)
	}
	data, e := io.ReadAll(tr)
	if e != nil || string(data) != "original" {
		t.Fatal(string(data), e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e = conversionWorldBackup(ctx, src, dest, "cancelled.tar.gz"); e == nil {
		t.Fatal("cancelled backup succeeded")
	}
	if _, e = os.Stat(filepath.Join(dest, "cancelled.tar.gz")); !os.IsNotExist(e) {
		t.Fatal("partial backup published")
	}
}
func TestWorldConversionClosedGateAndUnsupportedHost(t *testing.T) {
	c := NewWorldConverter(nil, t.TempDir(), func(string) bool { return false })
	srv := &database.Server{ID: "x", MemoryMaxMB: 1024, CPULimitPercent: 100}
	if _, e := c.Start(srv, WorldConversionRequest{Input: "world", Format: "JAVA_1_20_5", AcknowledgeLoss: true}); e == nil {
		t.Fatal("disabled start accepted")
	}
	if e := c.Cancel("x", "missing"); !os.IsNotExist(e) {
		t.Fatal(e)
	}
	if runtime.GOOS != "linux" && WorldConversionSupport(srv) == "" {
		t.Fatal("unsupported host claims enforcement")
	}
}
func TestWorldConversionJavaExtractionBounds(t *testing.T) {
	for _, name := range []string{"jdk/bin/java", "jdk/../../escape"} {
		t.Run(name, func(t *testing.T) {
			var data bytes.Buffer
			gz := gzip.NewWriter(&data)
			tw := tar.NewWriter(gz)
			tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Typeflag: tar.TypeReg, Size: 4})
			tw.Write([]byte("java"))
			tw.Close()
			gz.Close()
			archive := filepath.Join(t.TempDir(), "runtime.tar.gz")
			os.WriteFile(archive, data.Bytes(), 0600)
			dest := t.TempDir()
			e := conversionExtractJava(context.Background(), archive, dest)
			if strings.Contains(name, "..") {
				if e == nil {
					t.Fatal("traversal accepted")
				}
			} else if e != nil {
				t.Fatal(e)
			}
		})
	}
}
