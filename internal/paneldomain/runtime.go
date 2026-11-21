package paneldomain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

const (
	caddyBinary     = "/usr/bin/caddy"
	systemctlBinary = "/usr/bin/systemctl"
	configPath      = "/etc/caddy/Caddyfile"
	maxConfigBytes  = 64 << 10
	maxOutputBytes  = 1 << 20
	commandTimeout  = 20 * time.Second
)

var errOutputLimit = errors.New("command output exceeded safety limit")

type boundedBuffer struct {
	buf      bytes.Buffer
	limit    int
	exceeded bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.buf.Len() {
		b.exceeded = true
		return 0, errOutputLimit
	}
	return b.buf.Write(p)
}

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = time.Second
	// Do not inherit proxy settings, Caddy substitutions, plugin settings or caller environment.
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=/var/lib/caddy", "USER=caddy", "LOGNAME=caddy", "LANG=C"}
	if name == caddyBinary {
		if err := caddyIdentity(cmd); err != nil {
			return nil, err
		}
	}
	out := &boundedBuffer{limit: maxOutputBytes}
	stderr := &boundedBuffer{limit: 64 << 10}
	cmd.Stdout, cmd.Stderr = out, stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if out.exceeded || stderr.exceeded {
		return nil, errOutputLimit
	}
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w", name, err)
	}
	return out.buf.Bytes(), nil
}

func runningConfig(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// No environment proxy and no redirects: this must remain the local Caddy admin API.
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("admin API redirect rejected") }}
	return fetchRunningConfig(ctx, client)
}

func fetchRunningConfig(ctx context.Context, client *http.Client) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:2019/config/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect local Caddy admin API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("local Caddy admin API did not return HTTP 200")
	}
	return readBounded(resp.Body, maxOutputBytes)
}

func readBounded(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errOutputLimit
	}
	return b, nil
}

func sameJSON(a, b []byte) bool {
	decode := func(data []byte) (any, error) {
		if len(data) > maxOutputBytes {
			return nil, errOutputLimit
		}
		d := json.NewDecoder(bytes.NewReader(data))
		d.UseNumber()
		var v any
		if err := d.Decode(&v); err != nil {
			return nil, err
		}
		if _, ok := v.(map[string]any); !ok {
			return nil, errors.New("config is not an object")
		}
		var extra any
		if d.Decode(&extra) != io.EOF {
			return nil, errors.New("trailing JSON")
		}
		return v, nil
	}
	av, ae := decode(a)
	bv, be := decode(b)
	return ae == nil && be == nil && reflect.DeepEqual(av, bv)
}

const serviceProperties = "LoadState,ActiveState,SubState,FragmentPath,DropInPaths,Environment,EnvironmentFiles,User,Group,Type,ExecStart,ExecReload,NeedDaemonReload"

func (m *CaddyManager) inspectService(ctx context.Context) error {
	for _, p := range []string{caddyBinary, systemctlBinary} {
		if err := m.safe(p, executablePath); err != nil {
			return err
		}
	}
	out, err := m.run(ctx, systemctlBinary, "show", "caddy.service", "--no-pager", "--property="+serviceProperties)
	if err != nil {
		return err
	}
	props := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return errors.New("unsupported systemd property output")
		}
		if _, exists := props[k]; exists {
			return errors.New("duplicate systemd property")
		}
		props[k] = v
	}
	for _, key := range strings.Split(serviceProperties, ",") {
		if _, ok := props[key]; !ok {
			return fmt.Errorf("missing systemd property %s", key)
		}
	}
	for k, want := range map[string]string{"LoadState": "loaded", "ActiveState": "active", "SubState": "running", "DropInPaths": "", "Environment": "", "EnvironmentFiles": "", "User": "caddy", "Group": "caddy", "Type": "notify", "NeedDaemonReload": "no"} {
		if props[k] != want {
			return fmt.Errorf("unsupported caddy.service %s (requires standard file-based service without overrides)", k)
		}
	}
	fragment, err := canonicalServiceUnit(props["FragmentPath"])
	if err != nil {
		return err
	}
	unit, err := m.read(fragment)
	if err != nil {
		return err
	}
	if err = checkUnit(unit.data); err != nil {
		return err
	}
	for key, argv := range map[string]string{
		"ExecStart":  "/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile",
		"ExecReload": "/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force",
	} {
		// systemctl show appends execution timestamps/status. Only the one exact command is allowed.
		prefix := "{ path=/usr/bin/caddy ; argv[]=" + argv + " ; ignore_errors=no ;"
		if !strings.HasPrefix(props[key], prefix) || strings.Count(props[key], "argv[]=") != 1 || strings.Count(props[key], "{ path=") != 1 || !strings.HasSuffix(props[key], " }") {
			return fmt.Errorf("unsupported caddy.service %s", key)
		}
	}
	return nil
}

func checkUnit(data []byte) error {
	allowed := map[string][]string{
		"Unit.Description": {"Caddy"}, "Unit.Documentation": {"https://caddyserver.com/docs/"},
		"Unit.After": {"network.target network-online.target"}, "Unit.Requires": {"network-online.target"},
		"Service.Type": {"notify"}, "Service.User": {"caddy"}, "Service.Group": {"caddy"},
		"Service.ExecStart":      {"/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile"},
		"Service.ExecReload":     {"/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force"},
		"Service.TimeoutStopSec": {"5s"}, "Service.LimitNOFILE": {"1048576"}, "Service.LimitNPROC": {"512"},
		"Service.PrivateTmp": {"true"}, "Service.ProtectSystem": {"full"},
		"Service.AmbientCapabilities": {"CAP_NET_ADMIN CAP_NET_BIND_SERVICE", "CAP_NET_BIND_SERVICE"},
		"Service.Restart":             {"on-abnormal", "on-failure", "always"}, "Service.RestartSec": {"5s"},
		"Service.RestartPreventExitStatus": {"1"}, "Service.NoNewPrivileges": {"true"}, "Service.ProtectHome": {"true"},
		"Install.WantedBy": {"multi-user.target"},
	}
	section := ""
	seen := map[string]bool{}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if line == "[Unit]" || line == "[Service]" || line == "[Install]" {
			section = line[1 : len(line)-1]
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		key := section + "." + k
		good := false
		for _, candidate := range allowed[key] {
			if candidate == v {
				good = true
			}
		}
		if !ok || !good || seen[key] {
			return fmt.Errorf("unsupported caddy.service unit directive %s", key)
		}
		seen[key] = true
	}
	for _, k := range []string{"Service.Type", "Service.User", "Service.Group", "Service.ExecStart", "Service.ExecReload"} {
		if !seen[k] {
			return fmt.Errorf("missing caddy.service directive %s", k)
		}
	}
	return nil
}
