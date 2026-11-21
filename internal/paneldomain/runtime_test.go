package paneldomain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const standardUnit = `[Unit]
Description=Caddy
Documentation=https://caddyserver.com/docs/
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=notify
User=caddy
Group=caddy
ExecStart=/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile
ExecReload=/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force
TimeoutStopSec=5s
LimitNOFILE=1048576
LimitNPROC=512
PrivateTmp=true
ProtectSystem=full
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
`

func TestServiceInspection(t *testing.T) {
	standard := map[string]string{
		"LoadState": "loaded", "ActiveState": "active", "SubState": "running", "FragmentPath": "/usr/lib/systemd/system/caddy.service",
		"DropInPaths": "", "Environment": "", "EnvironmentFiles": "", "User": "caddy", "Group": "caddy", "Type": "notify", "NeedDaemonReload": "no",
		"ExecStart":  "{ path=/usr/bin/caddy ; argv[]=/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile ; ignore_errors=no ; start_time=[time] ; stop_time=[n/a] ; pid=1 ; code=(null) ; status=0/0 }",
		"ExecReload": "{ path=/usr/bin/caddy ; argv[]=/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force ; ignore_errors=no ; start_time=[time] ; stop_time=[n/a] ; pid=2 ; code=exited ; status=0 }",
	}
	for _, change := range []struct{ key, value string }{
		{}, {"ActiveState", "inactive"}, {"SubState", "dead"}, {"DropInPaths", "/etc/systemd/system/caddy.service.d/override.conf"},
		{"Environment", "CADDY_ADMIN=elsewhere"}, {"EnvironmentFiles", "/etc/caddy/env"}, {"User", "root"}, {"Type", "simple"},
		{"NeedDaemonReload", "yes"}, {"FragmentPath", "/etc/systemd/system/caddy.service"},
		{"ExecStart", strings.Replace(standard["ExecStart"], "--environ", "--resume", 1)},
		{"ExecStart", strings.Replace(standard["ExecStart"], "/etc/caddy/Caddyfile", "/tmp/custom", 1)},
		{"ExecReload", standard["ExecReload"] + " ; " + standard["ExecReload"]},
	} {
		t.Run(change.key+change.value, func(t *testing.T) {
			m := NewCaddyManager("127.0.0.1", 8080, true, true)
			m.safe = func(p string, k pathKind) error {
				if k != executablePath || (p != caddyBinary && p != systemctlBinary) {
					t.Fatal("unexpected executable probe")
				}
				return nil
			}
			m.read = func(p string) (fileRecord, error) {
				if p != "/usr/lib/systemd/system/caddy.service" {
					t.Fatalf("unexpected unit path %q", p)
				}
				return fileRecord{data: []byte(standardUnit)}, nil
			}
			m.run = func(_ context.Context, name string, args ...string) ([]byte, error) {
				if name != systemctlBinary || !reflect.DeepEqual(args, []string{"show", "caddy.service", "--no-pager", "--property=" + serviceProperties}) {
					t.Fatal("unexpected service command")
				}
				var b strings.Builder
				for _, key := range strings.Split(serviceProperties, ",") {
					value := standard[key]
					if key == change.key {
						value = change.value
					}
					fmt.Fprintf(&b, "%s=%s\n", key, value)
				}
				return []byte(b.String()), nil
			}
			err := m.inspectService(context.Background())
			if (err == nil) != (change.key == "") {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestUnitRejectsCustomDirectives(t *testing.T) {
	if err := checkUnit([]byte(standardUnit)); err != nil {
		t.Fatal(err)
	}
	for _, directive := range []string{"Environment=FOO=bar", "EnvironmentFile=/etc/custom", "ExecStartPre=/bin/true", "ExecStart=/usr/bin/caddy run --resume", "RootDirectory=/other", "WorkingDirectory=/tmp", "PassEnvironment=FOO", "ExecReload=/bin/sh -c evil"} {
		unit := strings.Replace(standardUnit, "[Service]", "[Service]\n"+directive, 1)
		if err := checkUnit([]byte(unit)); err == nil {
			t.Fatalf("accepted %q", directive)
		}
	}
}

func TestFileSafetyPolicy(t *testing.T) {
	for _, test := range []struct {
		name  string
		mode  os.FileMode
		uid   int
		links uint64
		kind  pathKind
		ok    bool
	}{
		{"regular", 0644, 0, 1, regularPath, true}, {"private", 0600, 0, 1, regularPath, true}, {"directory", os.ModeDir | 0755, 0, 2, directoryPath, true},
		{"executable", 0755, 0, 1, executablePath, true}, {"nonexecutable", 0644, 0, 1, executablePath, false},
		{"symlink", os.ModeSymlink | 0777, 0, 1, regularPath, false}, {"hardlink", 0644, 0, 2, regularPath, false},
		{"group writable", 0664, 0, 1, regularPath, false}, {"world writable", 0666, 0, 1, regularPath, false},
		{"wrong owner", 0644, 1000, 1, regularPath, false}, {"fifo", os.ModeNamedPipe | 0600, 0, 1, regularPath, false},
		{"setuid", os.ModeSetuid | 0755, 0, 1, executablePath, false}, {"directory writable", os.ModeDir | 0775, 0, 2, directoryPath, false},
		{"directory symlink", os.ModeSymlink | 0755, 0, 1, directoryPath, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := checkSafety(test.mode, test.uid, test.links, test.kind); (err == nil) != test.ok {
				t.Fatalf("unexpected safety result %v", err)
			}
		})
	}
}

func TestSnapshotRejectsSymlinkAndOversize(t *testing.T) {
	t.Run("symlink", func(t *testing.T) {
		f := newFixture(t)
		link := filepath.Join(filepath.Dir(f.m.path), "link")
		if err := os.Symlink(f.m.path, link); err != nil {
			t.Skipf("host cannot create unprivileged symlinks: %v", err)
		}
		if _, err := f.m.snapshot(link); err == nil {
			t.Fatal("symlink accepted")
		}
	})
	t.Run("oversize", func(t *testing.T) {
		f := newFixture(t)
		if err := os.WriteFile(f.m.path, bytes.Repeat([]byte("x"), maxConfigBytes+1), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := f.m.snapshot(f.m.path); err == nil {
			t.Fatal("oversize accepted")
		}
	})
}

func TestBoundedOutputAndJSON(t *testing.T) {
	b := &boundedBuffer{limit: 4}
	if _, err := io.Copy(b, strings.NewReader("12345")); !errors.Is(err, errOutputLimit) || b.buf.Len() > 4 || !b.exceeded {
		t.Fatalf("unbounded writer: %v", err)
	}
	if _, err := readBounded(strings.NewReader("12345"), 4); !errors.Is(err, errOutputLimit) {
		t.Fatal("unbounded reader")
	}
	if !sameJSON([]byte(`{"b":2,"a":[1]}`), []byte(`{ "a": [1], "b": 2 }`)) {
		t.Fatal("object order should be irrelevant")
	}
	for _, bad := range []string{"null", "[]", "{}{}", "not-json", `{"a":1}`} {
		if sameJSON([]byte(`{}`), []byte(bad)) {
			t.Fatalf("accepted bad/different config %s", bad)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBoundedAdminHTTP(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		ok     bool
	}{
		{"valid", 200, `{"apps":{}}`, true}, {"oversize", 200, strings.Repeat("x", maxOutputBytes+1), false}, {"error", 500, "no", false}, {"redirect", 302, "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodGet || r.URL.String() != "http://127.0.0.1:2019/config/" {
					t.Fatal("unexpected admin request")
				}
				return &http.Response{StatusCode: test.status, Body: io.NopCloser(strings.NewReader(test.body)), Header: make(http.Header), Request: r}, nil
			})}
			out, err := fetchRunningConfig(context.Background(), client)
			if (err == nil) != test.ok || len(out) > maxOutputBytes {
				t.Fatalf("unexpected result %v", err)
			}
		})
	}
}

// Execute only this test binary as a deterministic unprivileged command helper.
func TestCommandHelper(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "--paneldomain-helper" {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "stdout":
		_, _ = os.Stdout.Write(bytes.Repeat([]byte("x"), maxOutputBytes+100))
		os.Exit(0)
	case "stderr":
		_, _ = os.Stderr.Write(bytes.Repeat([]byte("x"), maxOutputBytes+100))
		os.Exit(0)
	case "sleep":
		time.Sleep(5 * time.Second)
		os.Exit(0)
	case "ok":
		fmt.Print("safe")
		os.Exit(0)
	}
	os.Exit(2)
}

func TestCommandRunnerBounds(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"stdout", "stderr"} {
		if _, err := runCommand(context.Background(), executable, "-test.run=TestCommandHelper", "--", "--paneldomain-helper", mode); !errors.Is(err, errOutputLimit) {
			t.Fatalf("%s not bounded: %v", mode, err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := runCommand(ctx, executable, "-test.run=TestCommandHelper", "--", "--paneldomain-helper", "sleep"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout not honored: %v", err)
	}
	out, err := runCommand(context.Background(), executable, "-test.run=TestCommandHelper", "--", "--paneldomain-helper", "ok")
	if err != nil || string(out) != "safe" {
		t.Fatalf("helper failed: %s %v", out, err)
	}
}
