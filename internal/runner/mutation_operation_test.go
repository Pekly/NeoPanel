package runner

import (
	"archive/tar"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestDirectoryReplacementExcludesManagedWrites(t *testing.T) {
	m, srv, backups := restoreFixture(t)
	live := m.GetWorkingDir(srv.ID)
	if err := os.WriteFile(filepath.Join(backups, "safe.tar.gz"), restoreArchive(t, tar.Header{Name: "new"}), 0600); err != nil {
		t.Fatal(err)
	}
	release, err := rooted.TryMutation(live)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := m.RestoreBackup(context.Background(), srv, "safe.tar.gz"); !errors.Is(err, ErrServerOperation) {
		t.Fatal("restore bypassed active writer", err)
	}
	called := false
	if _, err := m.ImportServerFiles(context.Background(), srv, func(context.Context, string) error { called = true; return nil }); !errors.Is(err, ErrServerOperation) || called {
		t.Fatal("import bypassed writer", err)
	}
	if finish, err := m.BeginServerMaintenance(context.Background(), srv.ID); !errors.Is(err, ErrServerOperation) {
		if finish != nil {
			finish()
		}
		t.Fatal("maintenance bypassed writer", err)
	}
	if _, err := os.Stat(filepath.Join(live, "original")); err != nil {
		t.Fatal("original moved despite writer", err)
	}
	release()
	if _, err := m.RestoreBackup(context.Background(), srv, "safe.tar.gz"); err != nil {
		t.Fatal("restore did not resume", err)
	}
	writeRelease, err := rooted.TryMutation(live)
	if err != nil {
		t.Fatal("restore leaked exclusive lease", err)
	}
	writeRelease()
	finish, err := m.BeginServerMaintenance(context.Background(), srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if unexpected, err := rooted.TryMutation(live); !errors.Is(err, rooted.ErrMutationBusy) {
		if unexpected != nil {
			unexpected()
		}
		t.Fatal("writer bypassed maintenance", err)
	}
	finish()
	writeRelease, err = rooted.TryMutation(live)
	if err != nil {
		t.Fatal("maintenance leaked exclusive lease", err)
	}
	writeRelease()
}

func TestImportHoldsManagedWriteLeaseThroughTransfer(t *testing.T) {
	m, srv, live := importFixture(t)
	_, err := m.ImportServerFiles(context.Background(), srv, func(ctx context.Context, stage string) error {
		if release, err := rooted.TryMutation(live); !errors.Is(err, rooted.ErrMutationBusy) {
			if release != nil {
				release()
			}
			t.Error("write allowed during import", err)
		}
		return os.WriteFile(filepath.Join(stage, "imported"), []byte("safe"), 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
	release, err := rooted.TryMutation(live)
	if err != nil {
		t.Fatal("import leaked exclusive lease", err)
	}
	release()
}
