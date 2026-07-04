#!/usr/bin/env bash
# Posts a PR comment summarizing the Critical gate result and, if any,
# the High/Medium/Low backlog counts with a link to the persistent
# rollup issue (see plumber-file-issues.sh / ADR-038).
#
# Uses `gh pr comment --edit-last --create-if-none`, which edits the
# bot's own last comment on the PR (scoped by the GITHUB_TOKEN identity)
# instead of posting a fresh comment every run — the same "one artifact,
# regenerated in place" principle as the rollup issue, applied here
# without needing a custom marker search: gh already tracks "last
# comment by this identity" natively.
set -euo pipefail

RESULTS_PATH="${1:?usage: plumber-pr-comment.sh <results.json>}"
REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
PR_NUMBER="${PR_NUMBER:?PR_NUMBER must be set}"
ROLLUP_MARKER="<!-- plumber-compliance-backlog -->"
TMP_BODY="${RUNNER_TEMP:-/tmp}/plumber-pr-comment.md"

if [ ! -f "$RESULTS_PATH" ]; then
    echo "No ${RESULTS_PATH} found; skipping PR comment."
    exit 0
fi

CRITICAL_COUNT="$(jq '.plumberScore.counts.critical // 0' "$RESULTS_PATH")"
HIGH_COUNT="$(jq '.plumberScore.counts.high // 0' "$RESULTS_PATH")"
MEDIUM_COUNT="$(jq '.plumberScore.counts.medium // 0' "$RESULTS_PATH")"
LOW_COUNT="$(jq '.plumberScore.counts.low // 0' "$RESULTS_PATH")"
BACKLOG_TOTAL="$((HIGH_COUNT + MEDIUM_COUNT + LOW_COUNT))"

ROLLUP_ISSUE="$(gh issue list --repo "$REPO" --search "\"${ROLLUP_MARKER}\" in:body" --state open --json number --jq '.[0].number // empty' 2>/dev/null || true)"

{
    if [ "$CRITICAL_COUNT" -gt 0 ]; then
        echo "### 🔴 Plumber: ${CRITICAL_COUNT} Critical-severity finding(s) — blocking merge"
        echo ""
        echo "Fix these before merging; see the \`plumber\` check's log for detail."
    else
        echo "### ✅ Plumber: no Critical-severity findings"
    fi
    echo ""
    if [ "$BACKLOG_TOTAL" -gt 0 ]; then
        echo "Non-blocking backlog on this branch: ${HIGH_COUNT} High, ${MEDIUM_COUNT} Medium, ${LOW_COUNT} Low."
        if [ -n "$ROLLUP_ISSUE" ]; then
            echo "Tracked in #${ROLLUP_ISSUE} — that issue reflects \`main\`, not this PR's branch, and only updates on push to \`main\`, so counts here may differ until this merges."
        fi
    else
        echo "No High/Medium/Low backlog findings on this branch."
    fi
    echo ""
    echo "<!-- plumber-pr-comment -->"
} >"$TMP_BODY"

gh pr comment "$PR_NUMBER" --repo "$REPO" --body-file "$TMP_BODY" --edit-last --create-if-none
