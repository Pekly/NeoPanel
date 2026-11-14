package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestJavaVMArgsManagedHeapLast(t *testing.T) {
	for _, custom := range []string{
		"", "-Xmx64G -Xms32G", "-XX:MaxHeapSize=64G -XX:InitialHeapSize=32G",
		"-XX:MaxRAMPercentage=99 -XX:InitialRAMPercentage=99",
		"-XX:+UseG1GC -Dterminal.ansi=true --add-opens=java.base/java.lang=ALL-UNNAMED",
	} {
		srv := database.Server{MemoryMinMB: 512, MemoryMaxMB: 2048, JavaArgs: custom}
		args, err := javaVMArgs(srv)
		if err != nil {
			t.Fatal(custom, err)
		}
		want := append(strings.Fields(custom), "-Xms512M", "-Xmx2048M")
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("%q: got %v, want %v", custom, args, want)
		}
	}
}

func TestJavaVMArgsRejectIndirectAndLauncherArguments(t *testing.T) {
	for _, custom := range []string{
		"@tenant.args", "@@tenant.args", "-XX:Flags=tenant.flags", "-XX:VMOptionsFile=tenant.flags",
		"-jar other.jar", "--module=other/main", "-m other/main", "other.Main", "-cp", "--class-path=.", "--",
	} {
		if args, err := javaVMArgs(database.Server{JavaArgs: custom, MemoryMaxMB: 2048}); err == nil {
			t.Fatalf("accepted launcher/indirect options %q: %v", custom, args)
		}
	}
}

func TestServerMemoryCeiling(t *testing.T) {
	for _, tc := range []struct {
		software string
		heap     int
		want     int64
	}{
		{"paper", 0, 0}, {"paper", -1, 0}, {"paper", 256, 768}, {"paper", 1024, 1536},
		{"paper", 2048, 3072}, {"paper", 8192, 12288}, {"steel", 2048, 2048},
	} {
		srv := database.Server{Type: tc.software, MemoryMaxMB: tc.heap}
		if got := serverMemoryCeilingMB(srv); got != tc.want {
			t.Fatalf("%+v: %d", tc, got)
		}
	}
	maxInt := int(^uint(0) >> 1)
	if got := serverMemoryCeilingMB(database.Server{Type: "paper", MemoryMaxMB: maxInt}); got < int64(maxInt) {
		t.Fatal("memory ceiling overflowed", got)
	}
}

// Exercise both command builders without starting any host isolation command.
func TestSystemdCommandsIncludeJavaMemoryCeiling(t *testing.T) {
	bin := t.TempDir()
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	for _, name := range []string{"systemd-run", "bwrap"} {
		if err := os.WriteFile(filepath.Join(bin, name+ext), []byte("not executed"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	for _, engine := range []string{"systemd", "bwrap"} {
		for _, software := range []string{"paper", "steel"} {
			srv := &database.Server{ID: "memory-test", Type: software, MemoryMaxMB: 2048}
			p := NewProcessInstance(srv, t.TempDir(), time.Second, nil, nil)
			p.isolationEngine = engine
			cmd, label, err := p.buildCommand("java", []string{"-Xmx2048M"})
			if err != nil {
				t.Fatal(engine, software, err)
			}
			want := fmt.Sprintf("--property=MemoryMax=%dM", serverMemoryCeilingMB(*srv))
			count := 0
			for _, arg := range cmd.Args {
				if arg == want {
					count++
				}
			}
			if count != 1 || !strings.Contains(label, "MemoryMax=") {
				t.Fatal(engine, software, cmd.Args, label)
			}
		}
	}
}
