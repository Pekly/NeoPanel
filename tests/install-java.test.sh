#!/usr/bin/env bash
# Offline Java setup regression tests. Never run a real backend or host services.
set -eu
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source <(sed '/^# Entrypoint$/,$d' "$ROOT/install.sh")
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/neopanel-java-tests.XXXXXX")"
trap 'command rm -rf -- "$TEST_ROOT"' EXIT
INSTALL_DIR="$TEST_ROOT/hosting binary"
BUILD_DIR="$TEST_ROOT/new build"
mkdir -p "$INSTALL_DIR" "$BUILD_DIR"
export JAVA_TEST_TRACE="$TEST_ROOT/trace"
export JAVA_TEST_EXIT=0
cat > "$INSTALL_DIR/$BINARY_NAME" <<'MOCK'
#!/usr/bin/env bash
printf '%s|%s\n' "$0" "$*" >> "$JAVA_TEST_TRACE"
[ "$#" -eq 2 ] && [ "$1" = java ] && [ "$2" = install-all ] || exit 91
exit "$JAVA_TEST_EXIT"
MOCK
chmod +x "$INSTALL_DIR/$BINARY_NAME"
cp "$INSTALL_DIR/$BINARY_NAME" "$BUILD_DIR/neopanel_new"
step_java
step_java
[ "$(wc -l < "$JAVA_TEST_TRACE")" -eq 2 ] || fail 'repeat setup did not delegate verification'
grep -Fq "$INSTALL_DIR/$BINARY_NAME|java install-all" "$JAVA_TEST_TRACE" || fail 'incorrect backend invocation'
JAVA_TEST_EXIT=1
if step_java > "$TEST_ROOT/error" 2>&1; then fail 'provisioning failure swallowed'; fi
grep -q 'provisioning/verification failed' "$TEST_ROOT/error" || fail 'missing failure guidance'
JAVA_TEST_EXIT=0
step_update_ensure_java
grep -Fq "$BUILD_DIR/neopanel_new|java install-all" "$JAVA_TEST_TRACE" || fail 'update used obsolete installed CLI'
if step_java "$TEST_ROOT/missing" > "$TEST_ROOT/error" 2>&1; then fail 'missing binary accepted'; fi
grep -q 'Install/update NeoPanel first' "$TEST_ROOT/error" || fail 'missing actionable guidance'

for flow in run_full_installation run_node_installation; do
  body="$(declare -f "$flow")"
  [[ "$body" == *step_build_neopanel*step_java*step_config_service* ]] || fail "$flow does not verify Java between build and service start"
done
body="$(declare -f update_neopanel)"
[[ "$body" == *step_update_compile*step_update_ensure_java*step_update_swap* ]] || fail 'update restarts before Java verification'
body="$(declare -f step_java)"
[[ "$body" != *'systemctl '* && "$body" != *'curl '* && "$body" != *'apt-get '* && "$body" != *'rm -rf '* && "$body" != *'|| true'* ]] || fail 'unsafe independent shell Java installer returned'
! declare -F install_java_25 >/dev/null || fail 'legacy unverified downloader remains'
printf 'PASS: Java install/worker/update/repair delegate exact verification before existing service start/restart\n'
