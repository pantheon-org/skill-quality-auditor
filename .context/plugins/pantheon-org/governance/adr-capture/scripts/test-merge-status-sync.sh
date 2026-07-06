#!/usr/bin/env bash
# Fixture tests for merge-status-sync.sh (Phase 1 exit criteria in
# .context/plans/post-merge-status-sync-2026-07-04.md):
#   (a) PR #118 at its pre-acceptance merge commit — reproduces the real
#       ADR-032 drift this plan was written to catch.
#   (b) a file-touch-only link — must be flagged, never auto-flipped.
#   (c) a squash-merge with an empty/undersized files list — the
#       merge-commit-message fallback must still find the match.
#   (d) PR #118 today (fully synced) — must report "nothing to do".
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
PLUGIN_DIR="$ROOT/.context/plugins/pantheon-org/governance/adr-capture"
SCRIPT="$PLUGIN_DIR/scripts/merge-status-sync.sh"
PR_118_MERGE_SHA="673bef95adb3e3e9b7045d1518f170d98d4eea7d"

FAILURES=0
pass() { echo "PASS: $1"; }
fail() {
  echo "FAIL: $1"
  FAILURES=$((FAILURES + 1))
}

TMP_ROOT="$(mktemp -d)"
WORKTREES=()
cleanup() {
  for wt in "${WORKTREES[@]:-}"; do
    [ -n "$wt" ] && git -C "$ROOT" worktree remove --force "$wt" >/dev/null 2>&1 || true
  done
  rm -rf "$TMP_ROOT"
}
trap cleanup EXIT

write_stub_gh() {
  # $1 = stub bin dir. Reads STUB_GH_VIEW_JSON (file) and STUB_GH_PR_LIST
  # (raw string, default empty) from the environment at call time. Real gh
  # applies a `-q <expr>` jq filter itself before printing — this stub
  # shells out to jq (test-harness-only dependency, not part of the
  # shipped skill script) to faithfully reproduce that behavior.
  mkdir -p "$1"
  cat >"$1/gh" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail
case "$1 $2" in
  "pr view")
    shift 2
    query="."
    while [ $# -gt 0 ]; do
      case "$1" in
        -q)
          query="$2"
          shift 2
          ;;
        *) shift ;;
      esac
    done
    jq -r "$query" "$STUB_GH_VIEW_JSON"
    ;;
  "pr list")
    printf '%s' "${STUB_GH_PR_LIST:-}"
    ;;
  "pr create")
    if [ -n "${STUB_GH_MARKER:-}" ]; then
      echo "called" >>"$STUB_GH_MARKER"
    fi
    echo "https://example.invalid/pr/999"
    ;;
  *)
    echo "stub gh: unsupported invocation: $*" >&2
    exit 1
    ;;
esac
STUB
  chmod +x "$1/gh"
}

# build_pr118_view_json <out-file>
# Derives a `gh pr view --json mergedAt,files,commits,mergeCommit`-shaped
# fixture entirely from local git history (no network, no committed
# fixture file — see .agents/RULES.md "Never leak sensitive information":
# this repo's own commit metadata, e.g. author emails, must not be
# persisted into a testdata/ file). Only file paths and the already-public
# merge commit SHA are used; commit "oid" values are placeholders since
# only the count matters to the squash-merge undersized-list heuristic.
build_pr118_view_json() {
  local out="$1"
  local files_json="[" first=true f
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    if $first; then
      first=false
    else
      files_json+=","
    fi
    files_json+="{\"path\": \"$f\"}"
  done < <(git -C "$ROOT" diff-tree --no-commit-id --name-only -r "$PR_118_MERGE_SHA")
  files_json+="]"
  cat >"$out" <<EOF
{"mergedAt": "2026-07-04T06:54:15Z", "files": $files_json, "commits": [{"oid": "c1"}, {"oid": "c2"}, {"oid": "c3"}], "mergeCommit": {"oid": "$PR_118_MERGE_SHA"}}
EOF
}

### Fixture (a): PR #118 at its pre-acceptance commit ###
test_fixture_a() {
  local wt="$TMP_ROOT/fixture-a"
  git -C "$ROOT" worktree add --detach "$wt" 673bef9 >/dev/null 2>&1
  WORKTREES+=("$wt")

  local stub_dir="$TMP_ROOT/bin-a"
  write_stub_gh "$stub_dir"
  local view_json="$TMP_ROOT/pr-118-fixture-a.json"
  build_pr118_view_json "$view_json"

  local out
  out=$(cd "$wt" && PATH="$stub_dir:$PATH" STUB_GH_VIEW_JSON="$view_json" \
    "$SCRIPT" --dry-run 118 2>&1) || true

  if echo "$out" | grep -q "docs/ADR/adr-032-user-configurable-scoring-patterns.md \[adr\]: status proposed"; then
    pass "fixture (a): ADR-032 drift reported"
  else
    fail "fixture (a): expected ADR-032 drift, got:
$out"
  fi

  if echo "$out" | grep -q "user-configurable-scoring-patterns-2026-07-03.md \[plan\]"; then
    fail "fixture (a): plan should not be flagged — it was already status: done at this commit"
  else
    pass "fixture (a): already-done plan correctly not flagged"
  fi

  # Cross-check against check-plan-drift.sh's existing heuristic (Decision 5):
  # the plan is not `active`, so the age/related-path drift check must have
  # nothing to say about it either — the two heuristics must not disagree.
  local drift_out
  drift_out=$(cd "$wt" && "$ROOT/scripts/check-plan-drift.sh" 2>&1) || true
  if echo "$drift_out" | grep -q "user-configurable-scoring-patterns-2026-07-03.md"; then
    fail "fixture (a): check-plan-drift.sh unexpectedly flagged the same plan — heuristics disagree"
  else
    pass "fixture (a): check-plan-drift.sh agrees (no signal on the already-done plan)"
  fi
}

### Fixture (b): file-touch-only link, must be flagged not auto-flipped ###
test_fixture_b() {
  local wt="$TMP_ROOT/fixture-b"
  mkdir -p "$wt"
  git init --quiet "$wt"
  git -C "$wt" config user.email "test@example.invalid"
  git -C "$wt" config user.name "Test"
  # A user-level global gitignore may exclude .context/ — force it off for
  # this synthetic fixture repo so the plan file below actually gets tracked.
  git -C "$wt" config core.excludesfile /dev/null

  mkdir -p "$wt/.context/plans" "$wt/shared"
  cat >"$wt/shared/config.yaml" <<'EOF'
key: value
EOF
  cat >"$wt/.context/plans/unrelated-plan-2026-07-01.md" <<'EOF'
---
title: "Plan: Unrelated Plan"
type: PLAN
status: ACTIVE
date: 2026-07-01
related:
  - ../../shared/config.yaml
---

## Goal

Some unrelated plan that happens to reference a shared config file.
EOF
  git -C "$wt" add -A
  git -C "$wt" commit --quiet -m "init fixture"

  local stub_dir="$TMP_ROOT/bin-b"
  write_stub_gh "$stub_dir"
  local view_json="$TMP_ROOT/pr-b.json"
  cat >"$view_json" <<'EOF'
{"mergedAt": "2026-07-01T00:00:00Z", "files": [{"path": "shared/config.yaml"}], "commits": [{"oid": "deadbeef"}], "mergeCommit": {"oid": "deadbeef"}}
EOF

  local out
  out=$(cd "$wt" && PATH="$stub_dir:$PATH" STUB_GH_VIEW_JSON="$view_json" \
    "$SCRIPT" --dry-run 42 2>&1) || true

  if echo "$out" | grep -q "unrelated-plan-2026-07-01.md \[plan\]: status ACTIVE (signal: file-touch"; then
    pass "fixture (b): file-touch-only link flagged with the correct signal"
  else
    fail "fixture (b): expected a file-touch flag, got:
$out"
  fi

  if echo "$out" | grep -q "^Auto-flip candidates"; then
    fail "fixture (b): file-touch-only link must never be an auto-flip candidate"
  else
    pass "fixture (b): no auto-flip section printed"
  fi
}

### Fixture (c): squash-merge with empty files list, message-body fallback ###
test_fixture_c() {
  local wt="$TMP_ROOT/fixture-c"
  mkdir -p "$wt"
  git init --quiet "$wt"
  git -C "$wt" config user.email "test@example.invalid"
  git -C "$wt" config user.name "Test"
  git -C "$wt" config core.excludesfile /dev/null

  mkdir -p "$wt/.context/plans"
  cat >"$wt/.context/plans/squash-target-2026-07-01.md" <<'EOF'
---
title: "Plan: Squash Target"
type: PLAN
status: ACTIVE
date: 2026-07-01
related:
  - ../../src/widget.go
---

## Goal

A plan whose implementation PR was squash-merged with an undersized files list.
EOF
  git -C "$wt" add -A
  git -C "$wt" commit --quiet -m "init fixture"
  git -C "$wt" commit --quiet --allow-empty -m "feat: implement widget (#42)

* abc1234 add src/widget.go
* def5678 wire up widget in main"
  local merge_sha
  merge_sha=$(git -C "$wt" rev-parse HEAD)

  local stub_dir="$TMP_ROOT/bin-c"
  write_stub_gh "$stub_dir"
  local view_json="$TMP_ROOT/pr-c.json"
  cat >"$view_json" <<EOF
{"mergedAt": "2026-07-01T00:00:00Z", "files": [], "commits": [{"oid": "abc1234"}, {"oid": "def5678"}], "mergeCommit": {"oid": "$merge_sha"}}
EOF

  local out
  out=$(cd "$wt" && PATH="$stub_dir:$PATH" STUB_GH_VIEW_JSON="$view_json" \
    "$SCRIPT" --dry-run 42 2>&1) || true

  if ! echo "$out" | grep -q "falling back to merge-commit message parsing"; then
    fail "fixture (c): expected the fallback note, got:
$out"
  else
    pass "fixture (c): squash-merge fallback triggered"
  fi

  if echo "$out" | grep -q "squash-target-2026-07-01.md \[plan\]: status ACTIVE (signal: file-touch"; then
    pass "fixture (c): fallback found the match via the merge-commit message body"
  else
    fail "fixture (c): expected the fallback to surface the plan, got:
$out"
  fi
}

### Fixture (e): apply mode auto-flips via branch + PR, idempotent on rerun ###
test_fixture_e() {
  local origin="$TMP_ROOT/origin-e.git"
  local wt="$TMP_ROOT/fixture-e"
  git init --quiet --bare "$origin"
  mkdir -p "$wt"
  git init --quiet --initial-branch=main "$wt"
  git -C "$wt" config user.email "test@example.invalid"
  git -C "$wt" config user.name "Test"
  git -C "$wt" config core.excludesfile /dev/null
  git -C "$wt" remote add origin "$origin"

  mkdir -p "$wt/.context/plans"
  cat >"$wt/.context/plans/single-phase-2026-07-01.md" <<'EOF'
---
title: "Plan: Single Phase"
type: PLAN
status: ACTIVE
date: 2026-07-01
---

## Goal

A single-phase plan whose implementation PR touches the plan file directly.
EOF
  # merge-status-sync.sh shells out to context-index's regenerate script at
  # a path relative to the repo it's running in — stub it here since this
  # fixture repo doesn't carry the full plugin tree; the index-regeneration
  # behavior itself is context-index's concern, not this test's.
  mkdir -p "$wt/.context/plugins/pantheon-org/context-mgmt/context-index/scripts"
  cat >"$wt/.context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh" <<'EOF'
#!/usr/bin/env bash
echo "stub index" >.context/index.yaml
EOF
  chmod +x "$wt/.context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh"

  git -C "$wt" add -A
  git -C "$wt" commit --quiet -m "init fixture"
  git -C "$wt" push --quiet -u origin main

  local stub_dir="$TMP_ROOT/bin-e"
  write_stub_gh "$stub_dir"
  local view_json="$TMP_ROOT/pr-e.json"
  cat >"$view_json" <<'EOF'
{"mergedAt": "2026-07-01T00:00:00Z", "files": [{"path": ".context/plans/single-phase-2026-07-01.md"}], "commits": [{"oid": "c1"}], "mergeCommit": {"oid": "deadbeef"}}
EOF
  local marker="$TMP_ROOT/pr-create-marker-e.txt"

  local out rc
  set +e
  out=$(cd "$wt" && PATH="$stub_dir:$PATH" STUB_GH_VIEW_JSON="$view_json" STUB_GH_MARKER="$marker" \
    "$SCRIPT" 77 2>&1)
  rc=$?
  set -e

  if [ "$rc" -eq 0 ]; then
    pass "fixture (e): apply mode exits 0"
  else
    fail "fixture (e): apply mode exited $rc, output:
$out"
  fi

  if [ -f "$marker" ] && [ "$(wc -l <"$marker" | tr -d ' ')" = "1" ]; then
    pass "fixture (e): gh pr create invoked exactly once"
  else
    fail "fixture (e): expected exactly one pr create call, marker: $(cat "$marker" 2>/dev/null || echo '<missing>')"
  fi

  local current_branch
  current_branch=$(git -C "$wt" rev-parse --abbrev-ref HEAD)
  if [ "$current_branch" = "main" ]; then
    pass "fixture (e): local checkout restored to the original branch"
  else
    fail "fixture (e): expected to be back on main, got $current_branch"
  fi

  if grep -q '^status: ACTIVE$' "$wt/.context/plans/single-phase-2026-07-01.md"; then
    pass "fixture (e): local main working copy left untouched"
  else
    fail "fixture (e): local main working copy was mutated in place"
  fi

  git -C "$wt" fetch --quiet origin "chore/status-sync-pr-77" || true
  if git -C "$wt" show "origin/chore/status-sync-pr-77:.context/plans/single-phase-2026-07-01.md" 2>/dev/null | grep -q '^status: DONE$'; then
    pass "fixture (e): pushed sync branch has the flipped status"
  else
    fail "fixture (e): pushed sync branch does not have the flipped status"
  fi

  # Idempotency: rerun with gh pr list now reporting the PR as already open.
  local out2 rc2
  set +e
  out2=$(cd "$wt" && PATH="$stub_dir:$PATH" STUB_GH_VIEW_JSON="$view_json" STUB_GH_MARKER="$marker" \
    STUB_GH_PR_LIST="https://example.invalid/pr/999" "$SCRIPT" 77 2>&1)
  rc2=$?
  set -e

  if [ "$rc2" -eq 0 ] && echo "$out2" | grep -q "sync PR already open"; then
    pass "fixture (e): rerun detects the open sync PR and is a no-op"
  else
    fail "fixture (e): expected an idempotent no-op on rerun, got (rc=$rc2):
$out2"
  fi

  if [ "$(wc -l <"$marker" | tr -d ' ')" = "1" ]; then
    pass "fixture (e): rerun did not open a second PR"
  else
    fail "fixture (e): rerun opened a duplicate PR"
  fi
}

### Fixture (d): PR #118 today — fully synced, must be a no-op ###
test_fixture_d() {
  local stub_dir="$TMP_ROOT/bin-d"
  write_stub_gh "$stub_dir"
  local view_json="$TMP_ROOT/pr-118-fixture-d.json"
  build_pr118_view_json "$view_json"

  local out
  out=$(cd "$ROOT" && PATH="$stub_dir:$PATH" STUB_GH_VIEW_JSON="$view_json" \
    "$SCRIPT" --dry-run 118 2>&1) || true

  if echo "$out" | grep -q "adr-032-user-configurable-scoring-patterns.md \[adr\]"; then
    fail "fixture (d): PR #118 is fully synced today — ADR-032 should not be flagged, got:
$out"
  else
    pass "fixture (d): today's fully-synced state reports no ADR-032 drift"
  fi
}

test_fixture_a
test_fixture_b
test_fixture_c
test_fixture_e
test_fixture_d

echo
if [ "$FAILURES" -gt 0 ]; then
  echo "$FAILURES test(s) failed."
  exit 1
fi
echo "All merge-status-sync.sh fixture tests passed."
