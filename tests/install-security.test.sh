#!/usr/bin/env bash
# Mocked security regression checks: never install compilers or modify host services.
set -eu
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source <(sed '/^# Entrypoint$/,$d' "$ROOT/install.sh")
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/neopanel-security-tests.XXXXXX")"
trap 'command rm -rf -- "$TEST_ROOT"' EXIT
export TMPDIR="$TEST_ROOT/tmp"
mkdir -p "$TMPDIR"
TRACE="$TEST_ROOT/trace"
COUNT=0
pass() { COUNT=$((COUNT+1)); printf 'ok %s - %s\n' "$COUNT" "$*"; }

[[ "$LOG_FILE" == /var/lib/neopanel/installer/* && "$BUILD_DIR" == /var/lib/neopanel/installer/* ]]
pass 'installer state is not in predictable shared tmp paths'
MOCK_OWNER=0
stat() {
  if [ "${1:-}" = -c ] && [ "${2:-}" = %u ]; then echo "$MOCK_OWNER"; else command stat "$@"; fi
}
INSTALL_STATE_DIR="$TEST_ROOT/state"
LOG_FILE="$INSTALL_STATE_DIR/install.log"
INSTALL_WORKSPACE_READY=false
initialize_installer_workspace
[ -d "$INSTALL_STATE_DIR" ] && [ -f "$LOG_FILE" ]
printf 'keep-log\n' > "$LOG_FILE"
INSTALL_WORKSPACE_READY=false
initialize_installer_workspace
grep -q keep-log "$LOG_FILE"
pass 'private installer workspace preserves existing log contents'

INSTALL_WORKSPACE_READY=false
MOCK_OWNER=1001
if initialize_installer_workspace > "$TEST_ROOT/error" 2>&1; then exit 1; fi
MOCK_OWNER=0
pass 'untrusted state-directory ownership is rejected'

command rm -f -- "$LOG_FILE"
printf 'outside-sentinel\n' > "$TEST_ROOT/outside"
ln "$TEST_ROOT/outside" "$LOG_FILE"
if initialize_installer_workspace > "$TEST_ROOT/error" 2>&1; then exit 1; fi
grep -q outside-sentinel "$TEST_ROOT/outside"
command rm -f -- "$LOG_FILE"
pass 'hardlinked update logs are rejected without modifying outside data'

ln -s "$TEST_ROOT/outside" "$LOG_FILE"
if [ -L "$LOG_FILE" ]; then
  if initialize_installer_workspace > "$TEST_ROOT/error" 2>&1; then exit 1; fi
  grep -q outside-sentinel "$TEST_ROOT/outside"
  pass 'symlinked update logs are rejected'
else
  printf 'skip - platform did not create a native symlink\n'
fi
command rm -f -- "$LOG_FILE"

MOCK_GO=go1.26.0
go() { printf '%s\n' "$MOCK_GO"; }
for MOCK_GO in go1.22.4 go1.24.9 go1.25.9 invalid; do
  if go_compiler_supported; then exit 1; fi
done
for MOCK_GO in go1.26.0 go1.27.1; do go_compiler_supported; done
pass 'compiler minimum is Go 1.26'

detect_os_pkg_manager() { PKG_MGR=unknown; }
uname() { echo x86_64; }
curl() {
  echo curl >> "$TRACE"
  if [[ "$*" == *'?mode=json'* ]]; then echo '{}'; return; fi
  while [ "$#" -gt 0 ]; do
    if [ "$1" = -o ]; then printf 'untrusted-download' > "$2"; return; fi
    shift
  done
  return 1
}
jq() {
  if [[ "$*" == *'.filename'* ]]; then echo go1.26.9.linux-amd64.tar.gz; else printf '%064d\n' 0; fi
}
sha256sum() { cat > /dev/null; return 1; }
tar() { echo 'UNSAFE EXTRACTION' >> "$TRACE"; return 1; }
: > "$TRACE"
MOCK_GO=go1.26.0
step_go_compiler
[ ! -s "$TRACE" ]
pass 'supported compiler avoids package/download operations'
MOCK_GO=go1.25.9
if step_go_compiler > "$TEST_ROOT/error" 2>&1; then exit 1; fi
grep -q 'checksum mismatch' "$TEST_ROOT/error"
if grep -q 'UNSAFE EXTRACTION' "$TRACE"; then exit 1; fi
[ -z "$(find "$TMPDIR" -mindepth 1 -print -quit)" ]
pass 'bad compiler checksum fails before extraction and cleans temporary files'

[[ "$(declare -f step_update_compile)" == *step_go_compiler* ]]
[[ "$(declare -f step_config_service)" == *'trusted_proxies:'* ]]
pass 'updates enforce compiler version and panel installs declare trusted loopback proxies'
printf '\n%s installer security checks passed (mocked host tools).\n' "$COUNT"
