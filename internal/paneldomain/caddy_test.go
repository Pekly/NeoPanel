package paneldomain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type fixture struct {
	m                          *CaddyManager
	t                          *testing.T
	original                   []byte
	active                     []byte
	validateErr                error
	reloadErr                  error
	reloadMode                 string
	onValidate                 func()
	onReload                   func()
	reloads, validates, adapts int
}

func fakeJSON(b []byte) []byte { return []byte(fmt.Sprintf(`{"config":"%x"}`, sha256.Sum256(b))) }

func newFixture(t *testing.T) *fixture {
	t.Helper()
	m := NewCaddyManager("127.0.0.1", 8080, true, true)
	m.path = filepath.Join(t.TempDir(), "Caddyfile")
	f := &fixture{m: m, t: t, original: []byte(legacyBlock("old.example.com", "old@example.com", m.upstream)), reloadMode: "candidate"}
	if err := os.WriteFile(m.path, f.original, 0644); err != nil {
		t.Fatal(err)
	}
	f.active = fakeJSON(f.original)
	m.platform = func() error { return nil }
	m.service = func(context.Context) error { return nil }
	m.syncDir = func(string) error { return nil }
	m.safe = func(path string, kind pathKind) error {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		// Real Lstat, fake trusted ownership so tests do not need root or /etc.
		mode := info.Mode()
		if runtime.GOOS == "windows" {
			mode = (mode &^ os.ModePerm) | 0644
			if info.IsDir() {
				mode |= 0111
			}
		}
		_, _, links := ownership(info)
		return checkSafety(mode, 0, links, kind)
	}
	m.running = func(context.Context) ([]byte, error) { return f.active, nil }
	m.run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if name == caddyBinary && len(args) == 5 && args[1] == "--config" && args[3] == "--adapter" && args[4] == "caddyfile" {
			if args[0] == "adapt" {
				f.adapts++
				b, err := os.ReadFile(args[2])
				return fakeJSON(b), err
			}
			if args[0] == "validate" {
				f.validates++
				if f.onValidate != nil {
					f.onValidate()
				}
				return nil, f.validateErr
			}
		}
		if name == systemctlBinary && reflect.DeepEqual(args, []string{"reload", "caddy.service"}) {
			f.reloads++
			switch f.reloadMode {
			case "candidate":
				b, err := os.ReadFile(m.path)
				if err != nil {
					return nil, err
				}
				f.active = fakeJSON(b)
			case "original":
				f.active = fakeJSON(f.original)
			case "unknown":
				f.active = []byte(`{"different":true}`)
			default:
				t.Fatalf("unexpected reload mode %q", f.reloadMode)
			}
			if f.onReload != nil {
				f.onReload()
			}
			return nil, f.reloadErr
		}
		t.Fatalf("unexpected command: %q %q", name, args)
		return nil, errors.New("unexpected command")
	}
	return f
}

func (f *fixture) disk() []byte {
	f.t.Helper()
	b, err := os.ReadFile(f.m.path)
	if err != nil {
		f.t.Fatal(err)
	}
	return b
}
func (f *fixture) provision() (ProvisionResult, error) {
	return f.m.Provision(context.Background(), "new.example.com", "admin@example.com")
}
func (f *fixture) backups() []string {
	f.t.Helper()
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(f.m.path), ".neopanel-backup-*"))
	if err != nil {
		f.t.Fatal(err)
	}
	return paths
}

func TestProvisionPreservesOldBytesMetadataAndDuplicate(t *testing.T) {
	f := newFixture(t)
	before, _ := os.Stat(f.m.path)
	r, err := f.provision()
	if err != nil || !r.Changed || !strings.Contains(r.Message, "issuance is pending") {
		t.Fatalf("result=%+v err=%v", r, err)
	}
	after, _ := os.Stat(f.m.path)
	bu, bg, _ := ownership(before)
	au, ag, _ := ownership(after)
	if before.Mode() != after.Mode() || bu != au || bg != ag {
		t.Fatal("metadata changed")
	}
	want := append(append([]byte(nil), f.original...), []byte(aliasBlock("new.example.com", "admin@example.com", f.m.upstream))...)
	if !bytes.Equal(f.disk(), want) {
		t.Fatal("original bytes not preserved or incorrect additive block")
	}
	if _, err := parseConfig(f.disk(), f.m.upstream); err != nil {
		t.Fatal(err)
	}
	if f.reloads != 1 || f.validates != 1 {
		t.Fatal("unexpected mutation calls")
	}
	if len(f.backups()) != 1 {
		t.Fatal("missing retained backup")
	}
	saved, err := os.ReadFile(f.backups()[0])
	if err != nil || !bytes.Equal(saved, f.original) {
		t.Fatal("bad backup")
	}
	if err := f.m.pendingAbsent(); err != nil {
		t.Fatal(err)
	}
	r, err = f.provision()
	duplicateInfo, _ := os.Stat(f.m.path)
	if err != nil || r.Changed || f.reloads != 1 || f.validates != 1 || !os.SameFile(after, duplicateInfo) || !after.ModTime().Equal(duplicateInfo.ModTime()) {
		t.Fatalf("duplicate not idempotent: %+v %v", r, err)
	}
	r, err = f.m.Provision(context.Background(), "old.example.com", "other@example.com")
	if err != nil || r.Changed || f.reloads != 1 {
		t.Fatal("original site should also be idempotent")
	}
}

func TestCapabilityReadOnly(t *testing.T) {
	f := newFixture(t)
	before, _ := os.Stat(f.m.path)
	cap := f.m.Capability(context.Background())
	entries, _ := os.ReadDir(filepath.Dir(f.m.path))
	after, _ := os.Stat(f.m.path)
	if !cap.Available || cap.Upstream != "127.0.0.1:8080" || !reflect.DeepEqual(cap.Hosts, []string{"old.example.com"}) || len(entries) != 1 || !sameMetadata(before, after) || f.reloads != 0 || f.validates != 0 {
		t.Fatalf("unexpected capability: %+v", cap)
	}
}

func TestInputValidationBeforeCommands(t *testing.T) {
	for _, host := range []string{"", "https://new.example.com", "new.example.com:443", "*.example.com", "example.com\nimport evil", "foo.{env}.com", "127.0.0.1", "localhost", "a..com", "-a.example.com", "a.example.com.", "UPPER.example.com", "example.com#", "example.com;", "пример.рф"} {
		t.Run(host, func(t *testing.T) {
			f := newFixture(t)
			if _, err := f.m.Provision(context.Background(), host, "admin@example.com"); err == nil || f.adapts != 0 {
				t.Fatal("accepted invalid hostname")
			}
		})
	}
	for _, email := range []string{"", "a@example.com\nimport evil", "{env.X}@example.com", "a@example.com#", "A <a@example.com>", "a@@example.com", "a@localhost", "a..b@example.com", "a;@example.com", "\"a\"@example.com"} {
		t.Run(email, func(t *testing.T) {
			f := newFixture(t)
			if _, err := f.m.Provision(context.Background(), "new.example.com", email); err == nil || f.adapts != 0 {
				t.Fatal("accepted invalid email")
			}
		})
	}
	if !validEmail("First.last+tag@example.com") || !validHostname("xn--bcher-kva.example") {
		t.Fatal("safe input rejected")
	}
}

func TestUnsupportedConfigAndDrift(t *testing.T) {
	for _, transform := range []func(string) string{
		func(s string) string { return "{ debug }\n" + s },
		func(s string) string { return s + "import /tmp/evil\n" },
		func(s string) string { return strings.Replace(s, "encode gzip zstd", "encode gzip", 1) },
		func(s string) string { return strings.Replace(s, "127.0.0.1:8080", "127.0.0.1:8081", 1) },
		func(s string) string { return s + legacyBlock("other.example.com", "a@example.com", "127.0.0.1:8080") },
		func(s string) string { return s + aliasBlock("old.example.com", "a@example.com", "127.0.0.1:8080") },
	} {
		f := newFixture(t)
		data := []byte(transform(string(f.original)))
		if err := os.WriteFile(f.m.path, data, 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := f.provision(); err == nil || f.reloads != 0 || !bytes.Equal(f.disk(), data) {
			t.Fatal("custom configuration accepted or modified")
		}
	}
	f := newFixture(t)
	f.active = []byte(`{"drift":true}`)
	if cap := f.m.Capability(context.Background()); cap.Available || !strings.Contains(cap.Reason, "differs") {
		t.Fatalf("drift accepted: %+v", cap)
	}
	if _, err := f.provision(); err == nil || f.validates != 0 {
		t.Fatal("drift mutation")
	}
}

func TestValidationFailureAndConcurrentSourceChange(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		f := newFixture(t)
		f.validateErr = errors.New("invalid candidate")
		if _, err := f.provision(); err == nil || f.reloads != 0 || !bytes.Equal(f.disk(), f.original) || len(f.backups()) != 0 {
			t.Fatal("validation failure mutated original")
		}
		entries, _ := os.ReadDir(filepath.Dir(f.m.path))
		if len(entries) != 1 {
			t.Fatal("candidate not cleaned")
		}
	})
	t.Run("source", func(t *testing.T) {
		f := newFixture(t)
		changed := []byte("administrator replacement\n")
		f.onValidate = func() {
			if err := os.WriteFile(f.m.path, changed, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := f.provision(); err == nil || !strings.Contains(err.Error(), "changed concurrently") || f.reloads != 0 || !bytes.Equal(f.disk(), changed) {
			t.Fatalf("concurrent edit lost: %v", err)
		}
	})
	t.Run("live", func(t *testing.T) {
		f := newFixture(t)
		f.onValidate = func() { f.active = []byte(`{"drift":true}`) }
		if _, err := f.provision(); err == nil || f.reloads != 0 || !bytes.Equal(f.disk(), f.original) {
			t.Fatal("live concurrent edit lost")
		}
	})
}

func TestReloadFailureAndTimeoutReconciliation(t *testing.T) {
	for _, test := range []struct {
		name, mode       string
		err              error
		changed, pending bool
	}{
		{"definite failure", "original", &exec.ExitError{}, false, false},
		{"failed client but applied", "candidate", &exec.ExitError{}, true, false},
		{"failed client uncertain", "unknown", &exec.ExitError{}, false, true},
		{"timeout applied", "candidate", context.DeadlineExceeded, true, false},
		{"timeout unknown", "unknown", context.DeadlineExceeded, false, true},
		{"cancellation applied", "candidate", context.Canceled, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newFixture(t)
			f.reloadMode, f.reloadErr = test.mode, test.err
			r, err := f.provision()
			if r.Changed != test.changed || (err == nil) != test.changed {
				t.Fatalf("result=%+v err=%v", r, err)
			}
			if (f.m.pendingAbsent() != nil) != test.pending || len(f.backups()) != 1 {
				t.Fatal("incorrect recovery state")
			}
			if test.pending {
				if !strings.Contains(err.Error(), "uncertain") || !bytes.HasPrefix(f.disk(), f.original) || bytes.Equal(f.disk(), f.original) {
					t.Fatal("uncertain candidate not retained")
				}
				if cap := f.m.Capability(context.Background()); cap.Available || !strings.Contains(cap.Reason, "recovery") {
					t.Fatal("pending transaction not blocked")
				}
			} else if !test.changed && !bytes.Equal(f.disk(), f.original) {
				t.Fatal("rollback failed")
			}
		})
	}
}

func TestRollbackWillNotOverwriteConcurrentChanges(t *testing.T) {
	f := newFixture(t)
	f.reloadMode, f.reloadErr = "original", &exec.ExitError{}
	changed := []byte("concurrent admin change")
	f.onReload = func() {
		if err := os.WriteFile(f.m.path, changed, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.provision(); err == nil || !strings.Contains(err.Error(), "uncertain") || !bytes.Equal(f.disk(), changed) || f.m.pendingAbsent() == nil {
		t.Fatalf("unsafe rollback: %v", err)
	}
}

func TestPendingNeverOverwritten(t *testing.T) {
	f := newFixture(t)
	state := []byte("previous recovery")
	if err := os.WriteFile(f.m.pendingPath(), state, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := f.provision(); err == nil || f.adapts != 0 {
		t.Fatal("pending transaction accepted")
	}
	got, _ := os.ReadFile(f.m.pendingPath())
	if !bytes.Equal(got, state) {
		t.Fatal("pending overwritten")
	}
	if err := f.m.writePending([]byte("new")); !errors.Is(err, os.ErrExist) {
		t.Fatalf("reservation was not exclusive: %v", err)
	}
}

func TestAliasLimitAndSerialization(t *testing.T) {
	f := newFixture(t)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := f.provision(); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if f.reloads != 1 {
		t.Fatal("duplicate concurrent reloads")
	}
	for i := 1; i < maxAliases; i++ {
		if _, err := f.m.Provision(context.Background(), fmt.Sprintf("alias%d.example.com", i), "admin@example.com"); err != nil {
			t.Fatal(err)
		}
	}
	if len(f.backups()) != maxAliases {
		t.Fatal("previous backup overwritten")
	}
	if _, err := f.m.Provision(context.Background(), "overflow.example.com", "admin@example.com"); err == nil {
		t.Fatal("alias limit ignored")
	}
	if _, err := f.provision(); err != nil {
		t.Fatal("duplicate rejected at limit")
	}
}

func TestListenerAndPlatformRestrictions(t *testing.T) {
	for _, args := range []struct {
		host         string
		port         int
		panel, proxy bool
	}{
		{"0.0.0.0", 8080, true, true}, {"localhost", 8080, true, true}, {"::1", 8080, true, true},
		{"127.0.0.1", 0, true, true}, {"127.0.0.1", 65536, true, true}, {"127.0.0.1", 8080, false, true}, {"127.0.0.1", 8080, true, false},
	} {
		m := NewCaddyManager(args.host, args.port, args.panel, args.proxy)
		m.platform = func() error { t.Fatal("invalid setup reached platform probe"); return nil }
		if cap := m.Capability(context.Background()); cap.Available || cap.Reason == "" {
			t.Fatal("invalid listener accepted")
		}
	}
	if runtime.GOOS != "linux" {
		if cap := NewCaddyManager("127.0.0.1", 8080, true, true).Capability(context.Background()); cap.Available || !strings.Contains(cap.Reason, "Linux") {
			t.Fatal("unsupported platform accepted")
		}
	}
}

func TestFinalPublicationGuardAndSyncFailure(t *testing.T) {
	t.Run("source change after backup", func(t *testing.T) {
		f := newFixture(t)
		changed := []byte("changed after backup")
		f.m.syncDir = func(string) error {
			if _, err := os.Lstat(f.m.pendingPath()); err == nil {
				return os.WriteFile(f.m.path, changed, 0644)
			}
			return nil
		}
		if _, err := f.provision(); err == nil || f.reloads != 0 || !bytes.Equal(f.disk(), changed) || len(f.backups()) != 1 {
			t.Fatalf("final publication guard failed: %v", err)
		}
	})
	t.Run("published but sync failed", func(t *testing.T) {
		f := newFixture(t)
		f.m.syncDir = func(string) error {
			data, err := os.ReadFile(f.m.path)
			if err != nil {
				return err
			}
			if !bytes.Equal(data, f.original) {
				return errors.New("injected directory sync failure")
			}
			return nil
		}
		if _, err := f.provision(); err == nil || !strings.Contains(err.Error(), "uncertain") || f.reloads != 0 || bytes.Equal(f.disk(), f.original) || f.m.pendingAbsent() == nil || len(f.backups()) != 1 {
			t.Fatalf("recovery state lost after publication: %v", err)
		}
	})
}

func TestUnavailableReconciliationRetainsRecovery(t *testing.T) {
	f := newFixture(t)
	f.reloadErr = context.DeadlineExceeded
	f.onReload = func() {
		f.m.running = func(ctx context.Context) ([]byte, error) {
			if ctx.Err() != nil {
				t.Fatal("reconciliation reused a canceled request context")
			}
			return nil, errors.New("admin API unavailable")
		}
	}
	if _, err := f.provision(); err == nil || !strings.Contains(err.Error(), "uncertain") || f.m.pendingAbsent() == nil || len(f.backups()) != 1 {
		t.Fatalf("missing uncertainty recovery: %v", err)
	}
}

func TestConcurrentRecoveryStateNotRemoved(t *testing.T) {
	f := newFixture(t)
	replacement := []byte("administrator recovery note")
	f.onReload = func() {
		if err := os.WriteFile(f.m.pendingPath(), replacement, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.provision(); err == nil || !strings.Contains(err.Error(), "uncertain") {
		t.Fatalf("changed recovery state ignored: %v", err)
	}
	got, err := os.ReadFile(f.m.pendingPath())
	if err != nil || !bytes.Equal(got, replacement) {
		t.Fatal("changed recovery state removed")
	}
}

func TestAmbiguousReloadDelayedCompletionRetainsRecovery(t *testing.T) {
	for _, reloadErr := range []error{context.DeadlineExceeded, context.Canceled} {
		t.Run(reloadErr.Error(), func(t *testing.T) {
			f := newFixture(t)
			f.reloadMode, f.reloadErr = "original", reloadErr
			var finishReload func()
			f.onReload = func() {
				// The service has read the candidate, but has not activated it yet.
				loaded := append([]byte(nil), f.disk()...)
				finishReload = func() { f.active = fakeJSON(loaded) }
			}
			r, err := f.provision()
			if err == nil || !strings.Contains(err.Error(), "uncertain") || r.Changed {
				t.Fatalf("ambiguous reload reported complete: %+v %v", r, err)
			}
			candidate := f.disk()
			if bytes.Equal(candidate, f.original) || !bytes.HasPrefix(candidate, f.original) || !sameJSON(f.active, fakeJSON(f.original)) {
				t.Fatal("candidate must remain on disk while original is still active")
			}
			pending, err := os.ReadFile(f.m.pendingPath())
			if err != nil || len(f.backups()) != 1 {
				t.Fatalf("recovery missing: %v", err)
			}
			for _, phase := range []string{"before completion", "after completion"} {
				if phase == "after completion" {
					finishReload()
				}
				if _, err := f.provision(); err == nil || !strings.Contains(err.Error(), "recovery") {
					t.Fatalf("retry not blocked %s: %v", phase, err)
				}
				got, err := os.ReadFile(f.m.pendingPath())
				if err != nil || !bytes.Equal(got, pending) || !bytes.Equal(f.disk(), candidate) || f.reloads != 1 {
					t.Fatalf("recovery or candidate changed %s: %v", phase, err)
				}
			}
			if !sameJSON(f.active, fakeJSON(candidate)) {
				t.Fatal("delayed activation diverged from disk")
			}
		})
	}
}

func TestOperationWaitingHonorsContext(t *testing.T) {
	owner := newFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	ownerDone := make(chan error, 1)
	owner.onValidate = func() { close(entered); <-release }
	go func() { _, err := owner.provision(); ownerDone <- err }()
	defer func() {
		close(release)
		select {
		case err := <-ownerDone:
			if err != nil {
				t.Errorf("holding provision failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("holding provision did not finish")
		}
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("provision did not acquire global operation slot")
	}
	for _, operation := range []string{"capability", "provision"} {
		for _, expiry := range []string{"cancel", "deadline"} {
			t.Run(operation+"/"+expiry, func(t *testing.T) {
				waiter := newFixture(t) // Separate manager still shares the global slot.
				ctx, cancel := context.WithCancel(context.Background())
				want := error(context.Canceled)
				if expiry == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
					want = context.DeadlineExceeded
				}
				defer cancel()
				started, done := make(chan struct{}), make(chan error, 1)
				go func() {
					close(started)
					if operation == "capability" {
						cap := waiter.m.Capability(ctx)
						if cap.Available || cap.Reason != want.Error() || len(cap.Hosts) != 0 {
							done <- fmt.Errorf("unexpected waiting capability: %+v", cap)
						} else {
							done <- nil
						}
						return
					}
					r, err := waiter.m.Provision(ctx, "new.example.com", "admin@example.com")
					if !errors.Is(err, want) || r.Changed {
						done <- fmt.Errorf("unexpected waiting provision: %+v %v", r, err)
					} else {
						done <- nil
					}
				}()
				<-started
				if expiry == "cancel" {
					cancel()
				}
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("canceled operation remained blocked behind provisioning")
				}
				if waiter.adapts != 0 || waiter.validates != 0 || waiter.reloads != 0 {
					t.Fatal("waiting operation reached commands")
				}
			})
		}
	}
}

func TestAcquireOperationAlreadyCanceledDoesNotLeakSlot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := acquireOperation(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled acquisition: %v", err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := acquireOperation(ctx); err != nil {
		t.Fatalf("operation slot leaked: %v", err)
	}
	releaseOperation()
}
