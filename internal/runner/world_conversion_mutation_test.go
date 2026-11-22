package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestMaintenanceLeasesExcludeManagedMutations(t *testing.T) {
	for _, kind := range []string{"conversion", "proxy"} {
		t.Run(kind, func(t *testing.T) {
			m, srv, _ := restoreFixture(t)
			dir := m.GetWorkingDir(srv.ID)
			begin := func() (func(), error) {
				if kind == "proxy" {
					return m.BeginProxyConfigOperation(context.Background(), srv.ID)
				}
				l, err := m.beginConversion(srv)
				if err != nil {
					return nil, err
				}
				return l.close, nil
			}
			mutation, err := rooted.TryMutation(dir)
			if err != nil {
				t.Fatal(err)
			}
			if release, err := begin(); !errors.Is(err, ErrServerOperation) {
				if release != nil {
					release()
				}
				t.Fatalf("maintenance raced an active mutation: %v", err)
			}
			mutation()
			release, err := begin()
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "managed.txt")
			if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
				release()
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				lease, err := rooted.TryMutation(dir)
				if err == nil {
					defer lease()
					err = os.WriteFile(path, []byte("changed"), 0600)
				}
				done <- err
			}()
			if err := <-done; !errors.Is(err, rooted.ErrMutationBusy) {
				release()
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != "original" {
				release()
				t.Fatal(string(got), err)
			}
			release()
			mutation, err = rooted.TryMutation(dir)
			if err != nil {
				t.Fatal(err)
			}
			mutation()
		})
	}
}

func TestWorldConversionPropertiesFailClosed(t *testing.T) {
	format, _ := chunker.Lookup("JAVA_1_20_5")
	srv := &database.Server{Type: "vanilla", Version: "1.20.5"}
	for _, tc := range []struct {
		name, text, input string
		allowed           bool
	}{
		{"equals", "level-name=custom\n", "custom", true},
		{"default", "motd=hello\n", "world", true},
		{"comments", " # level-name:custom\\\n!comment\nlevel-name=world\n", "world", true},
		{"java-whitespace", "\tlevel-name\t=\f custom\r\n", "custom", true},
		{"last-wins", "level-name=world\rlevel-name=custom\r", "custom", true},
		{"colon", "level-name:custom\n", "world", false},
		{"space", "level-name custom\n", "world", false},
		{"tab", "level-name\tcustom\n", "world", false},
		{"colon-equals", "level-name:custom=foo\n", "world", false},
		{"escaped-key", "level\\-name=custom\n", "world", false},
		{"unicode-key", "level\\u002dname=custom\n", "world", false},
		{"escaped-value", "level-name=wo\\rld\n", "world", false},
		{"continued-key", "level-\\\n name=custom\n", "world", false},
		{"continued-value", "level-name=wo\\\n rld\n", "world", false},
		{"trailing-space", "level-name=world \n", "world", false},
		{"unsupported-other", "motd:hi\nlevel-name=world\n", "world", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			err := conversionCompatible(srv, tc.input, format, dir)
			if (err == nil) != tc.allowed {
				t.Fatalf("allowed=%v: %v", tc.allowed, err)
			}
		})
	}
}
