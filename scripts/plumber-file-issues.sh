#!/usr/bin/env bash
# Files a deduplicated GitHub issue for every High/Medium/Low-severity finding
# in a Plumber JSON report (results.json from `getplumber/plumber`).
#
# Plumber has no native "findings" array with a per-finding severity field —
# each control writes its own `<control>Result.issues[]` with a different
# shape (jobName vs job vs branchName, etc). Severity per issue code is only
# available in `plumberScore.codeLosses[]`, so this walks every top-level
# `*Result` key generically, joins each issue back to its severity by code,
# and skips Critical findings (those block the pipeline via plumber-gate.sh
# instead of being filed here).
#
# Dedup key: sha256(resultKey|code|canonicalized issue JSON, with the
# commit-ref segment of any blob URL normalized out). Plumber's `url` field
# is a live GitHub blob link (`.../blob/<commit-sha>/path#Lline`) that
# embeds the *current* commit — hashing it verbatim means every single run
# produces a fresh fingerprint for the same underlying finding, since the
# SHA changes on every commit. Stripping that segment keeps the key stable
# across reruns while the display body still links to the real commit.
set -euo pipefail

RESULTS_PATH="${1:?usage: plumber-file-issues.sh <results.json>}"
REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
TMP_BODY="${RUNNER_TEMP:-/tmp}/plumber-issue-body.md"

if [ ! -f "$RESULTS_PATH" ]; then
    echo "No ${RESULTS_PATH} found; skipping issue filing."
    exit 0
fi

jq -c '
    (.plumberScore.codeLosses // []) as $losses
    | ($losses | map({(.code): .severity}) | add // {}) as $sevmap
    | to_entries[]
    | select(.key | endswith("Result"))
    | .key as $rk
    | (.value.issues // [])[]
    | . as $iss
    | {
        resultKey: $rk,
        code: ($iss.code // "UNKNOWN"),
        severity: ($sevmap[$iss.code] // "unknown"),
        issue: $iss,
        dedupIssue: (if ($iss | has("url")) then ($iss | .url |= sub("/blob/[^/]+/"; "/blob/_/")) else $iss end)
      }
    | select(.severity == "high" or .severity == "medium" or .severity == "low")
' "$RESULTS_PATH" |
    while IFS= read -r entry; do
        result_key="$(jq -r '.resultKey' <<<"$entry")"
        code="$(jq -r '.code' <<<"$entry")"
        severity="$(jq -r '.severity' <<<"$entry")"
        issue_json="$(jq -cS '.issue' <<<"$entry")"
        dedup_json="$(jq -cS '.dedupIssue' <<<"$entry")"
        location="$(jq -r '.issue.jobName // .issue.job // .issue.branchName // "unknown"' <<<"$entry")"
        doc_url="$(jq -r '.issue.docUrl // ""' <<<"$entry")"
        source_url="$(jq -r '.issue.url // ""' <<<"$entry")"

        fingerprint="$(printf '%s|%s|%s' "$result_key" "$code" "$dedup_json" | shasum -a 256 | cut -c1-16)"
        marker="<!-- plumber-dedup:${fingerprint} -->"

        existing="$(gh issue list --repo "$REPO" --search "\"${marker}\" in:body" --state all --json number --jq '.[0].number' 2>/dev/null || true)"
        if [ -n "$existing" ]; then
            echo "Already tracked (#${existing}): ${code} @ ${location}"
            continue
        fi

        title="[Plumber] ${code} (${severity}): ${location}"
        {
            echo "Plumber flagged a ${severity}-severity finding."
            echo ""
            echo "- **Code:** ${code}"
            echo "- **Location:** ${location}"
            [ -n "$source_url" ] && echo "- **Source:** ${source_url}"
            [ -n "$doc_url" ] && echo "- **Docs:** ${doc_url}"
            echo ""
            echo '```json'
            echo "$issue_json"
            echo '```'
            echo ""
            echo "$marker"
        } >"$TMP_BODY"

        gh issue create --repo "$REPO" --title "$title" --body-file "$TMP_BODY"
        echo "Filed issue: ${code} @ ${location}"
    done
