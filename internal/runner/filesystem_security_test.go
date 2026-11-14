package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestCustomJavaDoesNotExposeAncestorMounts(t *testing.T) {
	base := t.TempDir()
	a := filepath.Join(base, "A")
	b := filepath.Join(base, "B")
	os.Mkdir(a, 0755)
	os.Mkdir(b, 0755)
	os.WriteFile(filepath.Join(b, "secret"), []byte("secret"), 0600)
	binary := filepath.Join(a, "custom-java")
	os.WriteFile(binary, []byte("custom"), 0755)
	if args := javaRuntimeMountArgs(binary, a); len(args) != 0 {
		t.Fatal("custom runtime mounts", args)
	}
	bin := t.TempDir()
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	os.WriteFile(filepath.Join(bin, "bwrap"+suffix), []byte("stub"), 0755)
	t.Setenv("PATH", bin)
	p := NewProcessInstance(&database.Server{ID: "a", Type: "paper", JavaPath: binary}, a, time.Second, nil, nil)
	p.SetIsolationConfig("bwrap", 512)
	cmd, _, e := p.buildCommand(binary, nil)
	if e != nil {
		t.Fatal(e)
	}
	for i, arg := range cmd.Args {
		if (arg == "--ro-bind" || arg == "--ro-bind-try" || arg == "--bind") && i+1 < len(cmd.Args) {
			src := cmd.Args[i+1]
			if src == base || src == b || src == filepath.Dir(base) {
				t.Fatal("sibling exposed", cmd.Args)
			}
		}
	}
	// Symlinked arbitrary conf is not used to derive mounts, either.
	if e := os.Symlink(b, filepath.Join(a, "conf")); e == nil {
		if args := javaRuntimeMountArgs(binary, a); len(args) != 0 {
			t.Fatal(args)
		}
	}
}
func TestStartupConfigConfinement(t *testing.T) {
	for _, tc := range []struct{ kind, name, content string }{{"paper", "server.properties", "server-port=1\n"}, {"velocity", "velocity.toml", "bind = \"0.0.0.0:1\"\n"}, {"bungeecord", "config.yml", "  host: 0.0.0.0:25577\n"}} {
		t.Run(tc.kind, func(t *testing.T) {
			dir := t.TempDir()
			outside := filepath.Join(t.TempDir(), "victim")
			os.WriteFile(outside, []byte(tc.content), 0600)
			name := filepath.Join(dir, tc.name)
			if e := os.Link(outside, name); e != nil {
				t.Fatal(e)
			}
			if e := syncServerPort(dir, tc.kind, 25570); e != nil {
				t.Fatal(e)
			}
			b, _ := os.ReadFile(outside)
			if string(b) != tc.content {
				t.Fatal("hardlink modified")
			}
			os.Remove(name)
			if e := os.Symlink(outside, name); e != nil {
				t.Skip(e)
			}
			if e := syncServerPort(dir, tc.kind, 25570); e == nil {
				t.Fatal("startup accepted symlink")
			}
			b, _ = os.ReadFile(outside)
			if string(b) != tc.content {
				t.Fatal("outside config modified")
			}
		})
	}
}

type backupSwapWriter struct {
	bytes.Buffer
	swap func()
}

func (w *backupSwapWriter) Write(p []byte) (int, error) {
	if w.swap != nil {
		swap := w.swap
		w.swap = nil
		swap()
	}
	return w.Buffer.Write(p)
}
func TestBackupParentSwapCannotReadOutside(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	os.Mkdir(d, 0755)
	os.WriteFile(filepath.Join(d, "inside"), []byte("inside"), 0600)
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "secret"), []byte("outside-secret"), 0600)
	probe := filepath.Join(dir, "probe")
	if e := os.Symlink(outside, probe); e != nil {
		t.Skip(e)
	}
	os.Remove(probe)
	out := &backupSwapWriter{swap: func() {
		if e := os.Rename(d, d+"-old"); e != nil {
			t.Fatal(e)
		}
		if e := os.Symlink(outside, d); e != nil {
			t.Fatal(e)
		}
	}}
	if e := writeBackupArchive(dir, out); e == nil {
		t.Fatal("backup accepted parent swap")
	}
}
