package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSandboxOSReleaseMountIsSinglePublicFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "os-release")
	if err := os.WriteFile(file, []byte("NAME=TestLinux\n"), 0644); err != nil {
		t.Fatal(err)
	}
	want := []string{"--ro-bind-try", file, "/etc/os-release"}
	if got := sandboxOSReleaseArgs(file); !reflect.DeepEqual(got, want) {
		t.Fatalf("mount: %v", got)
	}
	if got := sandboxOSReleaseArgs(filepath.Join(dir, "missing"), file); !reflect.DeepEqual(got, want) {
		t.Fatalf("fallback: %v", got)
	}
	if got := sandboxOSReleaseArgs(dir, filepath.Join(dir, "missing")); got != nil {
		t.Fatalf("must not mount directories: %v", got)
	}
}

func TestSandboxOSReleaseResolvesSymlinkWithoutParentMount(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "public-release")
	link := filepath.Join(dir, "os-release")
	if err := os.WriteFile(target, []byte("ID=test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if got, want := sandboxOSReleaseArgs(link), []string{"--ro-bind-try", target, "/etc/os-release"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
