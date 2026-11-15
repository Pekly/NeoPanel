#!/usr/bin/env bash
# Compiler-selection regressions; no downloads, package managers or system changes.
set -eu
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source <(sed '/^# Entrypoint$/,$d; /^: > "\$LOG_FILE"$/d' "$ROOT/install.sh")
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }

(
  original_path="$PATH"
  go_compiler_supported() { return 0; }
  detect_os_pkg_manager() { fail 'unexpected package install'; }
  step_go_compiler
  [ "$PATH" = "$original_path" ] || fail 'supported compiler PATH changed'
)
echo 'PASS: supported compiler is preserved'

(
  PATH=/mock/old-go
  go_compiler_supported() { [[ "$PATH" == /usr/local/go/bin:* ]]; }
  detect_os_pkg_manager() { fail 'fallback was not reused'; }
  (step_go_compiler)
  [ "$PATH" = /mock/old-go ] || fail 'subshell unexpectedly changed parent'
  step_go_compiler
  [ "$PATH" = /usr/local/go/bin:/mock/old-go ] || fail 'fallback PATH not selected for build'
)
echo 'PASS: fallback compiler is reselected after a background step'

(
  step_go_compiler() { return 1; }
  mkdir() { fail 'build proceeded with unsupported compiler'; }
  if step_build_neopanel; then fail 'compiler failure was ignored'; fi
)
echo 'PASS: fresh build stops when compiler selection fails'

for name in step_build_neopanel step_update_compile; do
  body="$(declare -f "$name")"
  [[ "$body" == *step_go_compiler* ]] || fail "$name does not select compiler"
  [[ "$body" == *GOWORK=off* && "$body" == *-mod=readonly* ]] || fail "$name does not isolate the root module"
done
echo 'PASS: install and update builds select compiler and exclude workspace modules'
