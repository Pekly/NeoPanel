#!/usr/bin/env bash
# Mocked installer regression tests. Never run package managers, systemd or bwrap.
# On Linux, run: bash tests/install-isolation.test.sh
set -eu
exec 3>&2
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# Load the real functions without truncating the installer log or invoking main.
source <(sed '/^# Entrypoint$/,$d; /^: > "\$LOG_FILE"$/d' "$ROOT/install.sh")
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/neopanel-installer-tests.XXXXXX")"
trap 'status=$?; if [ "$status" -ne 0 ]; then cat "${LOG_FILE:-/dev/null}" >&3 || true; fi; rm -rf -- "$TEST_ROOT"' EXIT
export TMPDIR="$TEST_ROOT/tmp"
mkdir -p "$TMPDIR"
LOG_FILE="$TEST_ROOT/install.log"
TRACE="$TEST_ROOT/trace"
ARGS="$TEST_ROOT/args"
CONFIG_DIR="$TEST_ROOT/config"
mkdir -p "$CONFIG_DIR"
printf 'turnstile: {}\nkeep: unchanged\n' > "$CONFIG_DIR/config.yaml"
COUNT=0

reset_mocks() {
  MOCK_PKG=apt HAVE_BWRAP=1 HAVE_TIMEOUT=1 HAVE_RUNUSER=1
  FAIL_PACKAGE=none NEVER_INSTALL=false VERSION_FAIL=false PROBE_MODE=ok
  SERVICE_USER='' SERVICE_GROUP='' SERVICE_DYNAMIC=no SERVICE_LOAD=loaded SHOW_FAIL=false
  : > "$TRACE"; : > "$ARGS"; : > "$LOG_FILE"
}
# Mask tool discovery so tests behave identically on Linux and Git Bash.
command() {
  if [ "${1:-}" = -v ]; then
    case "${2:-}" in
      bwrap) [ "$HAVE_BWRAP" = 1 ] && echo bwrap; return $? ;;
      timeout) [ "$HAVE_TIMEOUT" = 1 ] && echo timeout; return $? ;;
      runuser) [ "$HAVE_RUNUSER" = 1 ] && echo runuser; return $? ;;
      apt-get) [ "$MOCK_PKG" = apt ] && echo apt-get; return $? ;;
      dnf|yum|pacman|apk) [ "$MOCK_PKG" = "$2" ] && echo "$2"; return $? ;;
    esac
  fi
  builtin command "$@"
}
package_mock() {
  printf 'package %s\n' "$*" >> "$TRACE"
  if [ "$FAIL_PACKAGE" = all ] || [[ " $* " == *" $FAIL_PACKAGE "* ]]; then
    echo 'mock repository/package failure' >&2; return 1
  fi
  if [ "$NEVER_INSTALL" != true ]; then
    case " $* " in *' install '*|*' add '*|*' -S '*)
      [[ " $* " != *' bubblewrap '* ]] || HAVE_BWRAP=1
      [[ " $* " != *' coreutils '* ]] || HAVE_TIMEOUT=1
      ;;
    esac
  fi
}
apt-get() { package_mock apt "$@"; }
dnf() { package_mock dnf "$@"; }
yum() { package_mock yum "$@"; }
pacman() { package_mock pacman "$@"; }
apk() { package_mock apk "$@"; }
bwrap() {
  if [ "${1:-}" = --version ]; then
    [ "$VERSION_FAIL" != true ] || return 1
    echo 'bubblewrap mock'; return 0
  fi
  echo 'sandbox invoked' >> "$TRACE"
  printf '%s\n' "$@" > "$ARGS"
  if [ "$PROBE_MODE" = blocked ]; then echo 'bwrap: Operation not permitted (mock)'; return 1; fi
  if [ "$PROBE_MODE" = no-marker ]; then return 0; fi
  local work=''
  while [ "$#" -gt 0 ]; do
    if [ "$1" = --bind ]; then work="$2"; break; fi
    shift
  done
  [ -n "$work" ] || return 1
  printf 'sandbox-ok\n' > "$work/probe-ok"
}
timeout() {
  [ "$1 $2 $3" = '-k 5 20' ] || return 1
  shift 3
  if [ "$PROBE_MODE" = timeout ]; then echo 'probe timed out (mock)'; return 124; fi
  "$@"
}
runuser() {
  printf 'runuser %s\n' "$*" >> "$TRACE"
  while [ "$1" != -- ]; do shift; done
  shift; "$@"
}
chmod() { printf 'chmod %s\n' "$*" >> "$TRACE"; }
mkdir() {
  if [ "${1:-}" = -m ]; then shift 2; fi
  builtin command mkdir "$@"
}
chown() { printf 'chown %s\n' "$*" >> "$TRACE"; }
id() {
  case "${2:-}" in root|0) echo 0 ;; paneluser) echo 1001 ;; *) return 1 ;; esac
}
systemctl() {
  printf 'systemctl %s\n' "$*" >> "$TRACE"
  [ "$SHOW_FAIL" != true ] || return 1
  case "$*" in
    *--property=LoadState*) echo "$SERVICE_LOAD" ;;
    *--property=DynamicUser*) echo "$SERVICE_DYNAMIC" ;;
    *--property=User*) echo "$SERVICE_USER" ;;
    *--property=Group*) echo "$SERVICE_GROUP" ;;
    *) echo 'unexpected systemctl operation' >&2; return 99 ;;
  esac
}
assert_contains() { grep -Fq -- "$2" "$1" || { echo "Missing '$2' in $1"; cat "$1"; return 1; }; }
assert_not_contains() { if grep -Fq -- "$2" "$1"; then echo "Unexpected '$2' in $1"; cat "$1"; return 1; fi; }
assert_clean() { [ -z "$(find "$TMPDIR" -mindepth 1 -print -quit)" ] || { echo 'Probe temp directory leaked'; return 1; }; }
pass() { COUNT=$((COUNT+1)); echo "ok $COUNT - $*"; }

for manager in apt dnf yum pacman apk; do
  reset_mocks; MOCK_PKG="$manager" HAVE_BWRAP=0 HAVE_TIMEOUT=0
  install_isolation_tools > "$LOG_FILE" 2>&1
  [ "$HAVE_BWRAP" = 1 ] && [ "$HAVE_TIMEOUT" = 1 ]
  assert_contains "$TRACE" 'bubblewrap coreutils'
  pass "$manager installs missing tools"
done
reset_mocks; install_isolation_tools
[ ! -s "$TRACE" ]; pass 'existing tools skip package operations'
reset_mocks; HAVE_BWRAP=0 MOCK_PKG=unknown
if install_isolation_tools > "$LOG_FILE" 2>&1; then exit 1; fi
assert_contains "$LOG_FILE" 'No supported package manager'; [ ! -s "$TRACE" ]; pass 'unsupported manager fails explicitly'
reset_mocks; HAVE_BWRAP=0 FAIL_PACKAGE=update
if install_isolation_tools > "$LOG_FILE" 2>&1; then exit 1; fi
assert_not_contains "$TRACE" 'install -y'; pass 'repository update failure stops installation'
reset_mocks; HAVE_BWRAP=0 FAIL_PACKAGE=install
if install_isolation_tools > "$LOG_FILE" 2>&1; then exit 1; fi
assert_contains "$LOG_FILE" 'Could not install'; pass 'package failure is propagated'
reset_mocks; HAVE_BWRAP=0 NEVER_INSTALL=true
if install_isolation_tools > "$LOG_FILE" 2>&1; then exit 1; fi
assert_contains "$LOG_FILE" 'still unavailable'; pass 'successful package command must actually provide tools'
reset_mocks; MOCK_PKG=unknown
if step_prerequisites > "$LOG_FILE" 2>&1; then exit 1; fi
[ ! -s "$TRACE" ]; pass 'base prerequisites no longer silently ignore unknown package manager'

reset_mocks; step_isolation fresh > "$LOG_FILE" 2>&1
assert_contains "$LOG_FILE" 'host probe passed as root'
for arg in --unshare-all --clearenv --cap-drop --die-with-parent '/etc/neopanel' 'NEOPANEL_ISOLATION_PROBE_SECRET'; do
  assert_contains "$ARGS" "$arg"
done
assert_clean; pass 'fresh-install probe uses sandbox flags and cleans temporary files'
for mode in blocked timeout no-marker; do
  reset_mocks; PROBE_MODE="$mode"
  if step_isolation fresh > "$LOG_FILE" 2>&1; then echo 'Failed probe was accepted'; exit 1; fi
  assert_not_contains "$LOG_FILE" 'host probe passed'; assert_clean
  pass "$mode probe fails safely and cleans temporary files"
done
reset_mocks; VERSION_FAIL=true
if step_isolation fresh > "$LOG_FILE" 2>&1; then exit 1; fi
assert_not_contains "$TRACE" 'sandbox invoked'; assert_clean; pass 'broken executable fails before probe'
reset_mocks; SERVICE_USER=paneluser SERVICE_GROUP=panelgroup
step_isolation installed > "$LOG_FILE" 2>&1
assert_contains "$TRACE" 'runuser -u paneluser -g panelgroup --'
assert_contains "$LOG_FILE" 'host probe passed as paneluser'; assert_clean; pass 'update tests configured service user and group'
reset_mocks; SERVICE_USER=paneluser HAVE_RUNUSER=0
if step_isolation installed > "$LOG_FILE" 2>&1; then exit 1; fi
assert_contains "$LOG_FILE" 'Install runuser'; assert_not_contains "$TRACE" 'sandbox invoked'; assert_clean; pass 'missing identity launcher cannot fall back to root'
for failure in dynamic missing unreadable unknown-user; do
  reset_mocks
  case "$failure" in dynamic) SERVICE_DYNAMIC=yes ;; missing) SERVICE_LOAD=not-found ;; unreadable) SHOW_FAIL=true ;; unknown-user) SERVICE_USER=missing-user ;; esac
  if step_isolation installed > "$LOG_FILE" 2>&1; then exit 1; fi
  assert_not_contains "$TRACE" 'sandbox invoked'; assert_clean; pass "$failure service identity fails without modifying service"
done

# Exercise the real update orchestration and real run_step error handling.
print_banner() { :; }
check_root() { :; }
detect_installed_role() { INSTALL_ROLE="${TEST_ROLE:-panel}"; }
draw_progress() { :; }
step_update_ensure_java() { echo 'java' >> "$TRACE"; }
step_update_fetch() { echo 'fetch' >> "$TRACE"; }
step_update_compile() { echo 'compile' >> "$TRACE"; }
step_update_swap() { echo 'swap' >> "$TRACE"; }
for role in panel node; do
  reset_mocks; TEST_ROLE="$role" HAVE_BWRAP=0
  (update_neopanel) > "$TEST_ROOT/update-output" 2>&1
  assert_contains "$TRACE" 'bubblewrap'
  assert_contains "$TRACE" 'sandbox invoked'; assert_contains "$TRACE" 'swap'
  [ "$(grep -n 'sandbox invoked' "$TRACE" | cut -d: -f1)" -lt "$(grep -n '^compile$' "$TRACE" | cut -d: -f1)" ]
  assert_clean; pass "$role update installs/probes before compilation and swap"
done
reset_mocks; PROBE_MODE=blocked
if (update_neopanel) > "$TEST_ROOT/update-output" 2>&1; then exit 1; fi
assert_not_contains "$TRACE" 'fetch'; assert_not_contains "$TRACE" 'compile'; assert_not_contains "$TRACE" 'swap'
assert_contains "$TEST_ROOT/update-output" 'Operation not permitted'
assert_contains "$TEST_ROOT/update-output" 'Full detailed installation log'
assert_contains "$CONFIG_DIR/config.yaml" 'keep: unchanged'
assert_clean; pass 'failed update probe stops real progress runner and preserves config/binary stages'

# Both fresh role flows must run the probe before building/configuring services.
for fn in run_full_installation run_node_installation; do
  body="$(declare -f "$fn")"
  [[ "$body" == *'step_isolation fresh'* ]] || exit 1
  before="${body%%step_isolation fresh*}"
  [[ "$before" != *'step_build_neopanel'* && "$before" != *'step_config_service '* ]] || exit 1
  pass "$fn wires isolation before build/service changes"
done
printf '\n%s installer isolation checks passed (mocked host tools).\n' "$COUNT"
