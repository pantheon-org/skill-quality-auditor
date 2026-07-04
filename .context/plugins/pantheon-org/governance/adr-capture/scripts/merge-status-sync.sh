#!/usr/bin/env bash
# Detects .context/plans/*.md and docs/ADR/*.md whose status is still
# active/draft/proposed after the PR that implemented them has merged, and
# (unless --dry-run) auto-flips the safe cases via a branch + PR.
#
# Pure bash + gh's built-in `-q` JSON query — no python/node (see
# .agents/RULES.md "Avoid Python/Node.js scripts in skills").
#
# See .context/plans/post-merge-status-sync-2026-07-04.md for the design
# decisions this script implements (Decisions 1-10).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/merge-status-sync-lib.sh
source "$SCRIPT_DIR/lib/merge-status-sync-lib.sh"

usage() {
  cat <<'EOF'
Usage: merge-status-sync.sh [--dry-run|-n] <pr-number>

  --dry-run, -n   Report drift and proposed actions; write nothing.
  (no flag)       Auto-flip single-phase plans directly/frontmatter-linked
                  to the PR via a branch + PR (never pushes to main); always
                  prints the flagged summary for multi-phase plans, ADRs,
                  and file-touch-only links, which require confirmation.

Examples:
  merge-status-sync.sh --dry-run 118
  merge-status-sync.sh 118
EOF
}

DRY_RUN=false
PR_NUMBER=""
for arg in "$@"; do
  case "$arg" in
    --dry-run | -n) DRY_RUN=true ;;
    -h | --help)
      usage
      exit 0
      ;;
    *[!0-9]* | "")
      echo "error: expected a numeric PR number, got '$arg'" >&2
      usage
      exit 1
      ;;
    *) PR_NUMBER="$arg" ;;
  esac
done

if [ -z "$PR_NUMBER" ]; then
  echo "error: PR number required" >&2
  usage
  exit 1
fi

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh CLI not found — required to inspect PR state" >&2
  exit 1
fi

ERR_FILE="$(mktemp)"
TOUCHED_FILE="$(mktemp)"
CANDIDATES_FILE="$(mktemp)"
AUTO_PLANS_FILE="$(mktemp)"
NEED_RESTORE_BRANCH=""
ORIGINAL_BRANCH=""

cleanup() {
  rm -f "$ERR_FILE" "$TOUCHED_FILE" "$CANDIDATES_FILE" "$AUTO_PLANS_FILE"
  if [ -n "$NEED_RESTORE_BRANCH" ]; then
    git checkout --quiet "$ORIGINAL_BRANCH" 2>/dev/null || true
  fi
}
trap cleanup EXIT

if ! MERGED_AT=$(gh pr view "$PR_NUMBER" --json mergedAt -q '.mergedAt // empty' 2>"$ERR_FILE"); then
  echo "error: could not fetch PR #$PR_NUMBER — $(cat "$ERR_FILE")" >&2
  exit 1
fi
if [ -z "$MERGED_AT" ]; then
  echo "error: PR #$PR_NUMBER is not merged yet" >&2
  exit 1
fi

FILE_COUNT=$(gh pr view "$PR_NUMBER" --json files -q '.files | length')
COMMIT_COUNT=$(gh pr view "$PR_NUMBER" --json commits -q '.commits | length')

# Decision 10: gh's files list is the primary source. Only fall back to
# parsing the merge commit message when it looks empty or undersized
# relative to the PR's own commit count (e.g. a squash merge elsewhere
# that summarised commits into the message body instead of the API
# reflecting them individually).
if [ "$FILE_COUNT" -eq 0 ] || [ "$FILE_COUNT" -lt "$COMMIT_COUNT" ]; then
  MERGE_SHA=$(gh pr view "$PR_NUMBER" --json mergeCommit -q '.mergeCommit.oid // empty')
  if [ -n "$MERGE_SHA" ] && git cat-file -e "${MERGE_SHA}^{commit}" 2>/dev/null; then
    echo "note: PR #$PR_NUMBER files list looks empty/undersized ($FILE_COUNT files vs $COMMIT_COUNT commits) — falling back to merge-commit message parsing" >&2
    git log --format=%B -n 1 "$MERGE_SHA" | grep -oE '[A-Za-z0-9_./-]+\.[A-Za-z0-9]+' | sort -u >"$TOUCHED_FILE"
  else
    echo "note: PR #$PR_NUMBER files list looks empty/undersized and no merge commit is reachable locally — proceeding with an empty file-touch set" >&2
    : >"$TOUCHED_FILE"
  fi
else
  gh pr view "$PR_NUMBER" --json files -q '.files[].path' >"$TOUCHED_FILE"
fi

detect_candidates "$ROOT" "$TOUCHED_FILE" >"$CANDIDATES_FILE"

print_report() {
  local file="$1"
  if [ ! -s "$file" ]; then
    echo "No linked plan/ADR found for this PR."
    return
  fi
  local printed=false type path status signal phases auto_flip target
  while IFS=$'\t' read -r type path status signal phases auto_flip target; do
    [ "$auto_flip" = "1" ] || continue
    if ! $printed; then
      echo "Auto-flip candidates (single-phase plan, direct/frontmatter link):"
      printed=true
    fi
    echo "  $path: status $status -> $target (signal: $signal)"
  done <"$file"
  printed=false
  while IFS=$'\t' read -r type path status signal phases auto_flip target; do
    [ "$auto_flip" = "1" ] && continue
    if ! $printed; then
      echo "Flagged for confirmation (requires a human decision):"
      printed=true
    fi
    local extra=""
    [ "$type" = "plan" ] && extra=", ${phases} phases"
    echo "  $path [$type]: status $status (signal: $signal${extra})"
  done <"$file"
}

print_report "$CANDIDATES_FILE"

if [ ! -s "$CANDIDATES_FILE" ]; then
  exit 0
fi

if [ "$DRY_RUN" = true ]; then
  exit 0
fi

awk -F'\t' '$6 == 1 {print $2}' "$CANDIDATES_FILE" >"$AUTO_PLANS_FILE"

if [ ! -s "$AUTO_PLANS_FILE" ]; then
  echo "Nothing eligible for auto-flip; see the flagged items above."
  exit 0
fi

BRANCH="chore/status-sync-pr-${PR_NUMBER}"

EXISTING_PR=$(gh pr list --head "$BRANCH" --state open --json url -q '.[0].url // empty' 2>/dev/null || true)
if [ -n "$EXISTING_PR" ]; then
  echo "sync PR already open for PR #$PR_NUMBER: $EXISTING_PR — nothing to do"
  exit 0
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "error: working tree has uncommitted changes — commit or stash before applying" >&2
  exit 1
fi

ORIGINAL_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
NEED_RESTORE_BRANCH=1

git fetch origin main
git checkout -b "$BRANCH" origin/main

CHANGED_PLAN_FILES=()
while IFS= read -r plan; do
  [ -n "$plan" ] || continue
  sed -i.bak -E 's/^status: (active|draft)$/status: done/' "$plan"
  rm -f "${plan}.bak"
  CHANGED_PLAN_FILES+=("$plan")
done <"$AUTO_PLANS_FILE"

"$ROOT/.context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh"

git add "${CHANGED_PLAN_FILES[@]}" .context/index.yaml
git commit --quiet -m "chore(status-sync): flip plan status to done for PR #${PR_NUMBER}"
git push --quiet -u origin "$BRANCH"

PR_BODY="Auto-generated by merge-status-sync.sh for PR #${PR_NUMBER}.

$(print_report "$CANDIDATES_FILE")"

gh pr create --base main --head "$BRANCH" \
  --title "chore: sync plan status for PR #${PR_NUMBER}" \
  --body "$PR_BODY"
