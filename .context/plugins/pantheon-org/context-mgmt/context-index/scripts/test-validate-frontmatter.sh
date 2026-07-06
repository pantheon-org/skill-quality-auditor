#!/usr/bin/env bash
# Fixture tests for validate-context-frontmatter.sh, covering the `value`
# field contract (Phase 1 exit criteria in
# .context/plans/context-prioritisation-signal-2026-07-06.md):
#   (a) value: high on a plan validates (field accepted).
#   (b) value: high on a finding validates (accepted on all action-candidate types).
#   (c) a file with no value still validates (field is optional in Phase 1).
#   (d) value: HIGH is rejected (enum is case-sensitive).
#   (e) value: urgent is rejected (not in the enum).
# The validator derives its enum checks from the schema, so these cases exercise
# the schema `value` enum without any per-type conditional in Phase 1.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VALIDATOR="$SCRIPT_DIR/validate-context-frontmatter.sh"

FAILURES=0
pass() { echo "PASS: $1"; }
fail() {
  echo "FAIL: $1"
  FAILURES=$((FAILURES + 1))
}

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

# write_fixture <name> <type> <extra-frontmatter-lines...>
write_fixture() {
  local name="$1" type="$2"
  shift 2
  local f="$TMP_DIR/$name.md"
  {
    echo "---"
    echo "title: \"Fixture: $name\""
    echo "type: $type"
    echo "status: active"
    echo "date: 2026-07-06"
    [ "$type" = "plan" ] && echo "effort: S"
    [ "$type" = "known-issue" ] && echo "severity: low"
    for line in "$@"; do echo "$line"; done
    echo "---"
    echo ""
    echo "# Fixture: $name"
  } >"$f"
  echo "$f"
}

# expect_pass <label> <file>
expect_pass() {
  if "$VALIDATOR" "$2" >/dev/null 2>&1; then pass "$1"; else fail "$1 (expected valid, got error)"; fi
}

# expect_fail <label> <file>
expect_fail() {
  if "$VALIDATOR" "$2" >/dev/null 2>&1; then fail "$1 (expected error, got valid)"; else pass "$1"; fi
}

expect_pass "plan with value: high validates"       "$(write_fixture plan-high plan 'value: high')"
expect_pass "finding with value: high validates"    "$(write_fixture finding-high finding 'value: high')"
expect_pass "plan with no value still validates"    "$(write_fixture plan-none plan)"
expect_fail "value: HIGH is rejected (case)"        "$(write_fixture plan-upper plan 'value: HIGH')"
expect_fail "value: urgent is rejected (enum)"      "$(write_fixture plan-bad plan 'value: urgent')"

echo ""
if [ "$FAILURES" -eq 0 ]; then
  echo "All fixture tests passed."
else
  echo "$FAILURES fixture test(s) failed."
  exit 1
fi
