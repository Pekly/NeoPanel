//go:build linux

package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorldConversionAtomicExchangeRetainsOriginal(t *testing.T) {
	base := t.TempDir()
	live := filepath.Join(base, "server")
	stage := filepath.Join(base, "private")
	os.MkdirAll(filepath.Join(live, "world"), 0700)
	os.MkdirAll(filepath.Join(stage, "world"), 0700)
	os.WriteFile(filepath.Join(live, "world", "level.dat"), []byte("original"), 0600)
	os.WriteFile(filepath.Join(stage, "world", "level.dat"), []byte("converted"), 0600)
	if err := conversionExchange(live, "world", filepath.Join(stage, "world"), "world"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(live, "world", "level.dat"))
	old, _ := os.ReadFile(filepath.Join(stage, "world", "level.dat"))
	if string(got) != "converted" || string(old) != "original" {
		t.Fatal("exchange did not retain both trees")
	}
	// The same atomic primitive performs administrator rollback without merging.
	if err := conversionExchange(live, "world", filepath.Join(stage, "world"), "world"); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(filepath.Join(live, "world", "level.dat"))
	if string(got) != "original" {
		t.Fatal("rollback lost original")
	}
	if err := conversionExchange(live, "world", filepath.Join(stage, "missing"), "world"); err == nil {
		t.Fatal("missing candidate accepted")
	}
	got, _ = os.ReadFile(filepath.Join(live, "world", "level.dat"))
	if string(got) != "original" {
		t.Fatal("failed swap changed original")
	}
}
