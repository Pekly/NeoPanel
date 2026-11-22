package rooted

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestMutationLeases(t *testing.T) {
	dir := t.TempDir()
	r1, err := TryMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r1()
	r2, err := TryMutation(filepath.Join(dir, "."))
	if err != nil {
		t.Fatal(err)
	}
	if release, err := TryExclusiveMutation(dir); !errors.Is(err, ErrMutationBusy) {
		if release != nil {
			release()
		}
		t.Fatalf("exclusive acquired during mutation: %v", err)
	}
	r1()
	r2()
	exclusive, err := TryExclusiveMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer exclusive()
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if release, err := TryMutation(dir); !errors.Is(err, ErrMutationBusy) {
				if release != nil {
					release()
				}
				t.Errorf("mutation acquired during exclusive job: %v", err)
			}
		}()
	}
	wg.Wait()
	other, err := TryMutation(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other()
	exclusive()
	after, err := TryMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	after()
}

func TestMutationCanonicalRoot(t *testing.T) {
	dir := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(dir, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	release, err := TryExclusiveMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if r, err := TryMutation(alias); !errors.Is(err, ErrMutationBusy) {
		if r != nil {
			r()
		}
		t.Fatalf("alias bypassed lease: %v", err)
	}
}

func TestMutationMissingRootKeepsKeyAfterCreation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "server")
	release, err := TryExclusiveMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if next, err := TryMutation(dir); !errors.Is(err, ErrMutationBusy) {
		if next != nil {
			next()
		}
		t.Fatal("newly created root bypassed maintenance", err)
	}
	release()
	next, err := TryMutation(dir)
	if err != nil {
		t.Fatal(err)
	}
	next()
}
