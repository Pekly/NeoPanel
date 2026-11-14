//go:build !linux && !windows

package sftp

import (
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPortableFIFORejectedWithoutBlocking(t *testing.T) {
	h, dir := portableTestHandler(t)
	must(t, unix.Mkfifo(filepath.Join(dir, "fifo"), 0600))
	done := make(chan error, 1)
	go func() {
		f, err := h.fsOpen("/fifo", os.O_RDONLY)
		if f != nil {
			f.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		denied(t, err)
	case <-time.After(time.Second):
		t.Fatal("FIFO open blocked")
	}
	_, err := h.fsList("/")
	denied(t, err)
}
