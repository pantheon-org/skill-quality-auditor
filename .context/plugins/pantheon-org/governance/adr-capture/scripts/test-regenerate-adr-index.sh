#!/usr/bin/env bash
# Fixture tests for regenerate-adr-index.sh --check (G2 in
# .context/plans/governance-tooling-hardening-2026-07-06.md):
#   (a) a fresh, up-to-date docs/ADR/index.yaml passes --check (exit 0).
#   (b) a stale/hand-edited docs/ADR/index.yaml fails --check (exit 1) — this is
#       the freshness gap the old `test -f` gate could never catch.
#   (c) after restoring the committed index, --check passes again.
#
# The generator resolves docs/ADR via `git rev-parse --show-toplevel`, so the
# test operates on the real index but backs it up first and restores it via a
# trap, leaving the working tree untouched even if an assertion fails.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
SCRIPT="$ROOT/.context/plugins/pantheon-org/governance/adr-capture/scripts/regenerate-adr-index.sh"
INDEX="$ROOT/docs/ADR/index.yaml"

FAILURES=0
pass() { echo "PASS: $1"; }
fail() {
  echo "FAIL: $1"
  FAILURES=$((FAILURES + 1))
}

BACKUP="$(mktemp)"
cp "$INDEX" "$BACKUP"
cleanup() {
  cp "$BACKUP" "$INDEX"
  rm -f "$BACKUP"
}
trap cleanup EXIT

# (a) fresh index passes
if "$SCRIPT" --check >/dev/null 2>&1; then
  pass "fresh ADR index passes --check"
else
  fail "fresh ADR index should pass --check"
fi

# (b) stale index fails
printf '\n# tampered by test\n' >>"$INDEX"
if "$SCRIPT" --check >/dev/null 2>&1; then
  fail "stale ADR index should fail --check"
else
  pass "stale ADR index fails --check"
fi

# (c) restored index passes again
cp "$BACKUP" "$INDEX"
if "$SCRIPT" --check >/dev/null 2>&1; then
  pass "restored ADR index passes --check"
else
  fail "restored ADR index should pass --check"
fi

echo ""
if [ "$FAILURES" -eq 0 ]; then
  echo "All fixture tests passed."
else
  echo "$FAILURES fixture test(s) failed."
  exit 1
fi
