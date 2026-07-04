#!/usr/bin/env bash
# Posts a PR comment with the Critical gate result and a per-severity
# breakdown of every finding on this branch (code, location, source link)
# — the same detail plumber-file-issues.sh writes into the rollup issue,
# so a PR author can see exactly what to fix without opening the
# `plumber` check's log.
#
# Uses `gh pr comment --edit-last --create-if-none`, which edits the
# bot's own last comment on the PR (scoped by the GITHUB_TOKEN identity)
# instead of posting a fresh comment every run — the same "one artifact,
# regenerated in place" principle as the rollup issue (ADR-038), applied
# here without needing a custom marker search: gh already tracks "last
# comment by this identity" natively.
#
# Finding extraction mirrors plumber-file-issues.sh: Plumber has no
# unified findings[] array with a per-finding severity, so this walks
# every top-level `*Result` key's issues[] generically and joins each
# one back to its severity via plumberScore.codeLosses.
set -euo pipefail

RESULTS_PATH="${1:?usage: plumber-pr-comment.sh <results.json>}"
REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
PR_NUMBER="${PR_NUMBER:?PR_NUMBER must be set}"
ROLLUP_MARKER="<!-- plumber-compliance-backlog -->"
TMP_BODY="${RUNNER_TEMP:-/tmp}/plumber-pr-comment.md"
MAX_ROWS_PER_SECTION=25

if [ ! -f "$RESULTS_PATH" ]; then
    echo "No ${RESULTS_PATH} found; skipping PR comment."
    exit 0
fi

FINDINGS_JSON="$(jq -c '
    (.plumberScore.codeLosses // []) as $losses
    | ($losses | map({(.code): .severity}) | add // {}) as $sevmap
    | [
        to_entries[]
        | select(.key | endswith("Result"))
        | .key as $rk
        | (.value.issues // [])[]
        | {resultKey: $rk, code: (.code // "UNKNOWN"), severity: ($sevmap[.code] // "unknown"), issue: .}
    ]
' "$RESULTS_PATH")"

CRITICAL_COUNT="$(jq '.plumberScore.counts.critical // 0' "$RESULTS_PATH")"
HIGH_COUNT="$(jq '.plumberScore.counts.high // 0' "$RESULTS_PATH")"
MEDIUM_COUNT="$(jq '.plumberScore.counts.medium // 0' "$RESULTS_PATH")"
LOW_COUNT="$(jq '.plumberScore.counts.low // 0' "$RESULTS_PATH")"
TOTAL_COUNT="$((CRITICAL_COUNT + HIGH_COUNT + MEDIUM_COUNT + LOW_COUNT))"

ROLLUP_ISSUE="$(gh issue list --repo "$REPO" --search "\"${ROLLUP_MARKER}\" in:body" --state open --json number --jq '.[0].number // empty' 2>/dev/null || true)"

write_section() {
    local sev="$1" label="$2"
    local sev_findings sev_count
    sev_findings="$(jq -c --arg sev "$sev" '[.[] | select(.severity == $sev)]' <<<"$FINDINGS_JSON")"
    sev_count="$(jq 'length' <<<"$sev_findings")"
    [ "$sev_count" -eq 0 ] && return 0

    echo "### ${label} (${sev_count})"
    echo ""
    echo "| Code | Location | Source |"
    echo "| --- | --- | --- |"
    jq -r --argjson max "$MAX_ROWS_PER_SECTION" '
        .[:$max][]
        | "| \(.code) | \(.issue.jobName // .issue.job // .issue.branchName // "unknown") | \(.issue.url // "-") |"
    ' <<<"$sev_findings"
    if [ "$sev_count" -gt "$MAX_ROWS_PER_SECTION" ]; then
        echo ""
        echo "_...and $((sev_count - MAX_ROWS_PER_SECTION)) more. See \`plumber explain <code>\` or the full \`plumber-compliance\` artifact on this run for the rest._"
    fi
    echo ""
}

{
    if [ "$CRITICAL_COUNT" -gt 0 ]; then
        echo "### 🔴 ${CRITICAL_COUNT} Critical-severity finding(s) — blocking merge"
        echo ""
    else
        echo "### ✅ No Critical-severity findings"
        echo ""
    fi

    write_section critical "🔴 Critical"
    write_section high "🟠 High"
    write_section medium "🟡 Medium"
    write_section low "🔵 Low"

    if [ "$TOTAL_COUNT" -eq 0 ]; then
        echo "No findings on this branch."
        echo ""
    elif [ -n "$ROLLUP_ISSUE" ]; then
        echo "High/Medium/Low findings are tracked in #${ROLLUP_ISSUE} once this merges — that issue reflects \`main\`, not this PR's branch, and only updates on push to \`main\`, so it may not match the tables above until then."
        echo ""
    fi

    echo "<!-- plumber-pr-comment -->"
} >"$TMP_BODY"

gh pr comment "$PR_NUMBER" --repo "$REPO" --body-file "$TMP_BODY" --edit-last --create-if-none
