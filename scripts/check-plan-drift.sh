#!/usr/bin/env bash
# Checks for plans with status:active that are older than N days.
# Informational only — exits 0.
set -euo pipefail

THRESHOLD_DAYS=${1:-60}
ROOT="$(git rev-parse --show-toplevel)"
NOW=$(date +%s)
COUNT=0

for plan in "$ROOT"/.context/plans/**/*.md "$ROOT"/.context/plans/*.md; do
    [ -f "$plan" ] || continue
    status=$(grep -E '^status: ' "$plan" | sed 's/^status: *//' | tr -d '"')
    [ "$status" = "active" ] || continue

    plan_date=$(grep -E '^date: ' "$plan" | sed 's/^date: *//' | tr -d '"')
    plan_ts=$(date -j -f "%Y-%m-%d" "$plan_date" +%s 2>/dev/null || true)
    [ -n "$plan_ts" ] || continue

    age=$(( (NOW - plan_ts) / 86400 ))
    if [ "$age" -gt "$THRESHOLD_DAYS" ]; then
        title=$(grep -E '^title: ' "$plan" | sed 's/^title: *//' | tr -d '"')
        rel=${plan#"$ROOT/"}
        echo "⚠  Plan still active after ${age}d: $rel — \"$title\""
        COUNT=$((COUNT + 1))
    fi
done

if [ "$COUNT" -gt 0 ]; then
    echo "→ $COUNT stale plan(s) found. Update frontmatter status to 'done' if implemented."
fi
exit 0
