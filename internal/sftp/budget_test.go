package sftp

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	sftpproto "github.com/pkg/sftp"
)

func assignDisk(t *testing.T, f *fixture, mb int64) {
	t.Helper()
	srv, err := f.db.GetServer("srv-alpha")
	must(t, err)
	srv.DiskLimitMB = mb
	must(t, f.db.UpdateServer(srv))
}

func TestSFTPLogicalSizeLimitAndExistingHandles(t *testing.T) {
	f := newFixture(t)
	_, client := f.connect(t, "editor#srv-alpha", "password")
	file, err := client.Create("large")
	must(t, err)
	defer file.Close()
	assignDisk(t, f, 1) // handle predates the assignment
	for _, offset := range []int64{16 << 20, 1 << 20, math.MaxInt64} {
		_, err := file.WriteAt([]byte("x"), offset)
		denied(t, err)
	}
	denied(t, file.Truncate(32<<20))
	denied(t, client.Truncate("large", 32<<20))
	st, err := file.Stat()
	must(t, err)
	if st.Size() != 0 {
		t.Fatalf("rejected growth mutated file: %d", st.Size())
	}
	// Exact boundary is allowed; both ordinary uploads and overwrites work.
	_, err = file.WriteAt([]byte("x"), (1<<20)-1)
	must(t, err)
	_, err = file.WriteAt([]byte("ok"), 0)
	must(t, err)
	must(t, file.Truncate(1<<20))
	must(t, file.Truncate(4))
	_, err = file.WriteAt([]byte("ok"), 0)
	must(t, err)
}

func TestSFTPFiniteDefaultAndOverflowLimits(t *testing.T) {
	f := newFixture(t)
	_, client := f.connect(t, "owner", "key")
	file, err := client.Create("bounded")
	must(t, err)
	defer file.Close()
	for _, assignment := range []int64{0, -1, math.MaxInt64} {
		assignDisk(t, f, assignment)
		_, err := file.WriteAt([]byte("x"), maxSFTPFileSize)
		denied(t, err)
		denied(t, file.Truncate(maxSFTPFileSize+1))
	}
	for _, offset := range []int64{-1, math.MaxInt64, maxSFTPFileSize} {
		_, err := checkedWriteEnd(offset, 1, maxSFTPFileSize)
		denied(t, err)
	}
	// Exercise unsigned protocol SETSTAT without making a huge file.
	h := &chrootHandler{rootDir: filepath.Join(f.root, "srv-alpha"), server: f.server, userID: f.users["owner"].ID, serverID: "srv-alpha"}
	attrs := make([]byte, 8)
	binary.BigEndian.PutUint64(attrs, math.MaxUint64)
	denied(t, h.Filecmd(&sftpproto.Request{Method: "Setstat", Filepath: "bounded", Flags: 1, Attrs: attrs}))
}

func TestSFTPOpenOversizedAndTruncateRecovery(t *testing.T) {
	f := newFixture(t)
	assignDisk(t, f, 1)
	_, client := f.connect(t, "owner", "key")
	old, err := client.Create("external")
	must(t, err)
	defer old.Close()
	// Model a small preexisting/external oversized file; no production tree.
	must(t, os.Truncate(filepath.Join(f.root, "srv-alpha", "external"), 2<<20))
	_, err = old.WriteAt([]byte("x"), 0)
	denied(t, err)
	opened, err := client.OpenFile("external", os.O_WRONLY)
	if opened != nil {
		opened.Close()
	}
	denied(t, err)
	readable, err := client.Open("external")
	must(t, err)
	must(t, readable.Close())
	// OPEN/TRUNC shrinks to zero, so it must remain usable for recovery.
	replacement, err := client.Create("external")
	must(t, err)
	defer replacement.Close()
	_, err = replacement.Write([]byte("small"))
	must(t, err)
}

func TestSFTPRepeatedReplacementDoesNotExhaustLifetimeBudget(t *testing.T) {
	f := newFixture(t)
	assignDisk(t, f, 1)
	connection, client := f.connect(t, "owner", "key")
	for i := 0; i < 3; i++ {
		file, err := client.Create("replacement")
		must(t, err)
		_, err = file.WriteAt([]byte("x"), (1<<20)-1)
		must(t, err)
		must(t, file.Truncate(0))
		must(t, file.Truncate(1<<20))
		must(t, file.Close())
		must(t, client.Remove("replacement"))
	}
	must(t, connection.Close())
	_, reconnected := f.connect(t, "owner", "key")
	file, err := reconnected.Create("replacement")
	must(t, err)
	defer file.Close()
	must(t, file.Truncate(1<<20))
	denied(t, file.Truncate((1<<20)+1))
}
