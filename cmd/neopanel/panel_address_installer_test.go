package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func recoveryTestBash(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		if path, err := exec.LookPath("bash"); err == nil {
			return path
		}
	} else {
		// Do not accidentally launch Windows' WSL bash shim.
		for _, path := range []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "bash.exe"),
			filepath.Join(os.Getenv("USERPROFILE"), "scoop", "apps", "Git", "current", "bin", "bash.exe"),
		} {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	t.Skip("Bash/Git Bash unavailable")
	return ""
}

func TestPanelAddressInstaller(t *testing.T) {
	bash := recoveryTestBash(t)
	installer, err := os.ReadFile(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	installerText := strings.ReplaceAll(string(installer), "\r\n", "\n")
	cmd := exec.Command(bash, "-n")
	cmd.Stdin = strings.NewReader(installerText)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bash syntax: %v\n%s", err, output)
	}
	definitions, _, found := strings.Cut(installerText, "# Entrypoint\n")
	if !found {
		t.Fatal("installer entrypoint marker missing")
	}
	dir := t.TempDir()
	mock := filepath.Join(dir, "mock-neopanel")
	mockBody := `#!/usr/bin/env bash
if [ "${4:-}" = --supported ]; then
  if [ "${MOCK_OLD:-}" = true ]; then echo 'Unknown subcommand' >&2; exit 1; fi
  echo neopanel-panel-address-v1
  exit 0
fi
printf '%s\n' "$@" > "$MOCK_ARGS"
`
	if err := os.WriteFile(mock, []byte(mockBody), 0700); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfg, []byte("# mock config; never parsed or changed\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// Source definitions only; all operations are mocked and use temporary files.
	checks := `
fail() { echo "FAIL: $*" >&2; exit 1; }
require_panel_address_admin() { return 0; }
for forbidden in run_step step_ssl_caddy initialize_installer_workspace run_full_installation run_node_installation update_neopanel systemctl curl; do
  # Static function names only; no user input enters this declaration.
  eval "$forbidden() { fail 'unexpected side effect: $forbidden'; }"
done
for forbidden in --update --install --node --yes --ssl --turnstile; do
  if parse_cli_args --change-fqdn "$forbidden" >/dev/null; then fail "accepted $forbidden"; fi
  if parse_cli_args "$forbidden" --change-fqdn >/dev/null; then fail "accepted leading $forbidden"; fi
done
if parse_cli_args --change-fqdn --panel-url >/dev/null; then fail 'missing URL accepted'; fi
parse_cli_args --repo https://example.invalid/repo --branch main --change-fqdn --panel-url 'https://panel.example.com' --confirm-address
[ "$ACTION" = change-fqdn ] || fail 'incorrect action'
[ "$RECOVERY_CONFIG" = "$NEOPANEL_CONFIG" ] || fail 'config environment ignored'
[ "$RECOVERY_BINARY" = "$NEOPANEL_BINARY" ] || fail 'binary environment ignored'
# Old binaries must fail safely with an update instruction and no save attempt.
export MOCK_OLD=true
if output=$(change_panel_address); then fail 'old binary accepted'; fi
[[ "$output" == *'Update NeoPanel first'* ]] || fail 'no update guidance'
[ ! -e "$MOCK_ARGS" ] || fail 'old binary reached save'
unset MOCK_OLD
# Even hostile URL text must remain one literal argument, never shell evaluation.
RECOVERY_URL='https://$(touch SHOULD_NOT_EXIST).example.com; echo bad'
change_panel_address >/dev/null
mapfile -t saved < "$MOCK_ARGS"
[ "${#saved[@]}" = 6 ] || fail 'wrong argument count'
[ "${saved[0]}" = -config ] && [ "${saved[1]}" = "$RECOVERY_CONFIG" ] || fail 'config argument changed'
[ "${saved[2]}" = panel-address ] && [ "${saved[3]}" = --url ] || fail 'wrong CLI command'
[ "${saved[4]}" = "$RECOVERY_URL" ] && [ "${saved[5]}" = --confirm ] || fail 'URL or confirmation changed'
[ ! -e SHOULD_NOT_EXIST ] || fail 'URL evaluated'
# Missing URL must not turn explicit unattended confirmation into a prompt/default.
RECOVERY_URL=''
if output=$(change_panel_address); then fail 'empty confirmed URL accepted'; fi
[[ "$output" == *'requires an explicit --panel-url'* ]] || fail 'missing URL guidance'
echo 'installer recovery checks passed'
`
	cmd = exec.Command(bash, "--noprofile", "--norc")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NEOPANEL_CONFIG="+filepath.ToSlash(cfg), "NEOPANEL_BINARY="+filepath.ToSlash(mock), "MOCK_ARGS="+filepath.ToSlash(filepath.Join(dir, "args.txt")))
	cmd.Stdin = strings.NewReader(definitions + checks)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("installer recovery: %v\n%s", err, output)
	}
}
