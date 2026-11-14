//go:build !windows

package rooted

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestFIFOOpenDoesNotBlock(t *testing.T) {
	dir, r := testRoot(t)
	if e := syscall.Mkfifo(filepath.Join(dir, "fifo"), 0600); e != nil {
		t.Fatal(e)
	}
	done := make(chan error, 1)
	go func() {
		f, e := r.Open("fifo")
		if f != nil {
			f.Close()
		}
		done <- e
	}()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("FIFO accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("FIFO open blocked")
	}
}
