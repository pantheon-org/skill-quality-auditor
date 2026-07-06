#!/usr/bin/env bash
# Fixture tests for validate-context-frontmatter.sh, covering the `value`
# field contract from .context/plans/context-prioritisation-signal-2026-07-06.md.
#
# Enum values are UPPER_CASE (ADR-050). value is required for draft/active
# action-candidate types:
#   (a) value: HIGH on a draft/active plan validates.
#   (b) value: HIGH on a finding validates (all action-candidate types).
#   (c) a draft/active plan with NO value is REJECTED (value now required).
#   (d) a DONE plan with no value validates (DONE/SUPERSEDED are exempt).
#   (e) an INSTRUCTION with no value validates (value not applicable).
#   (f) value: high is rejected (enum is case-sensitive; UPPER_CASE only).
#   (g) value: urgent is rejected (not in the enum).
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

# write_fixture <name> <type> <status> <extra-frontmatter-lines...>
write_fixture() {
  local name="$1" type="$2" status="$3"
  shift 3
  local f="$TMP_DIR/$name.md"
  {
    echo "---"
    echo "title: \"Fixture: $name\""
    echo "type: $type"
    echo "status: $status"
    echo "date: 2026-07-06"
    [ "$type" = "PLAN" ] && echo "effort: S"
    [ "$type" = "KNOWN_ISSUE" ] && echo "severity: LOW"
    for line in "$@"; do echo "$line"; done
    echo "---"
    echo ""
    echo "# Fixture: $name"
  } >"$f"
  echo "$f"
}

# expect_pass <label> <file> ; expect_fail <label> <file>
expect_pass() {
  if "$VALIDATOR" "$2" >/dev/null 2>&1; then pass "$1"; else fail "$1 (expected valid, got error)"; fi
}
expect_fail() {
  if "$VALIDATOR" "$2" >/dev/null 2>&1; then fail "$1 (expected error, got valid)"; else pass "$1"; fi
}

expect_pass "draft plan with value: HIGH validates"       "$(write_fixture plan-high PLAN ACTIVE 'value: HIGH')"
expect_pass "finding with value: HIGH validates"          "$(write_fixture finding-high FINDING ACTIVE 'value: HIGH')"
expect_fail "draft/active plan with no value is rejected"  "$(write_fixture plan-none PLAN ACTIVE)"
expect_pass "DONE plan with no value is exempt"           "$(write_fixture plan-done PLAN DONE)"
expect_pass "INSTRUCTION with no value validates"         "$(write_fixture instr INSTRUCTION ACTIVE)"
expect_fail "value: high is rejected (case)"              "$(write_fixture plan-lower PLAN ACTIVE 'value: high')"
expect_fail "value: urgent is rejected (enum)"            "$(write_fixture plan-bad PLAN ACTIVE 'value: urgent')"

echo ""
if [ "$FAILURES" -eq 0 ]; then
  echo "All fixture tests passed."
else
  echo "$FAILURES fixture test(s) failed."
  exit 1
fi
