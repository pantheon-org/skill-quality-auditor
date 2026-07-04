#!/usr/bin/env bash
# Fails the job if the Plumber JSON report (results.json from
# `getplumber/plumber`) contains any Critical-severity finding. Plumber's own
# exit code reflects overall compliance against --threshold, not per-finding
# severity, so gating reads plumberScore.counts.critical directly instead.
set -euo pipefail

RESULTS_PATH="${1:?usage: plumber-gate.sh <results.json>}"

if [ ! -f "$RESULTS_PATH" ]; then
    echo "::error::${RESULTS_PATH} not found; Plumber did not produce a report."
    exit 1
fi

CRITICAL_COUNT="$(jq '.plumberScore.counts.critical // 0' "$RESULTS_PATH")"

if [ "$CRITICAL_COUNT" -gt 0 ]; then
    echo "::error::${CRITICAL_COUNT} Critical-severity Plumber finding(s) present — see the plumber-compliance artifact or job summary for detail."
    jq -r '
        (.plumberScore.codeLosses // [])[]
        | select(.severity == "critical")
        | "::error::[\(.code)] \(.count) Critical-severity finding(s)"
    ' "$RESULTS_PATH"
    exit 1
fi

echo "No Critical-severity Plumber findings. Gate passed."
