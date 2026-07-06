#!/usr/bin/env bash
# Fixture tests for check-undocumented-decisions.sh (G3 in
# .context/plans/governance-tooling-hardening-2026-07-06.md):
#   (a) a file with a real "## Decision" section that also mentions index.yaml
#       IS flagged — proving the old blanket `index.yaml` substring skip (which
#       silently exempted any file mentioning index.yaml) is gone.
#   (b) a finding with a "## Recommended Action" section that mentions index.yaml
#       is NOT flagged — recommendation language is not a binding decision, so it
#       must not be a false positive (this is what the old skip was masking).
#
# The script reads CONTEXT_DIR/ADR_DIR from the environment, so the test points
# it at a throwaway tree and never touches the real .context/.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
SCRIPT="$ROOT/.context/plugins/pantheon-org/governance/adr-capture/scripts/check-undocumented-decisions.sh"

FAILURES=0
pass() { echo "PASS: $1"; }
fail() {
  echo "FAIL: $1"
  FAILURES=$((FAILURES + 1))
}

TMP="$(mktemp -d)"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

mkdir -p "$TMP/.context/plans" "$TMP/.context/findings" "$TMP/docs/ADR"

cat >"$TMP/.context/plans/fixture-decision.md" <<'EOF'
---
title: "Plan: fixture with a real decision"
type: PLAN
status: ACTIVE
date: 2026-07-06
---
## Decision
We will split the work. See `.context/index.yaml` for tracking.
EOF

cat >"$TMP/.context/findings/fixture-recommendation.md" <<'EOF'
---
title: "Finding: fixture with only a recommendation"
type: FINDING
status: ACTIVE
date: 2026-07-06
---
## Recommended Action
Consider automating this. Related: `.context/index.yaml`.
EOF

OUT="$(CONTEXT_DIR="$TMP/.context" ADR_DIR="$TMP/docs/ADR" bash "$SCRIPT" 2>&1 || true)"

# (a) the decision file mentioning index.yaml is now scanned and flagged
if grep -q "fixture-decision.md" <<<"$OUT"; then
  pass "decision file mentioning index.yaml is flagged (skip removed)"
else
  fail "decision file mentioning index.yaml should be flagged; output was: $OUT"
fi

# (b) the recommendation-only finding is NOT flagged (no false positive)
if grep -q "fixture-recommendation.md" <<<"$OUT"; then
  fail "recommendation-only finding should NOT be flagged; output was: $OUT"
else
  pass "recommendation-only finding is not a false positive"
fi

echo ""
if [ "$FAILURES" -eq 0 ]; then
  echo "All fixture tests passed."
else
  echo "$FAILURES fixture test(s) failed."
  exit 1
fi
