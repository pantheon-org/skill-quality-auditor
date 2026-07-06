#!/usr/bin/env bash
# Checks plans with status:active in two ways:
# 1. Age-based: older than N days.
# 2. Related-path heuristic: related files modified since plan date.
# Informational only — exits 0.
set -euo pipefail

THRESHOLD_DAYS=${1:-60}
ROOT="$(git rev-parse --show-toplevel)"
NOW=$(date +%s)
COUNT=0

for plan in "$ROOT"/.context/plans/**/*.md "$ROOT"/.context/plans/*.md; do
    [ -f "$plan" ] || continue

    frontmatter=$(sed -n '/^---$/,/^---$/p' "$plan" | sed '1d;$d')
    [ -n "$frontmatter" ] || continue

    status=$(echo "$frontmatter" | grep -E '^status: ' | sed 's/^status: *//' | tr -d '"')
    [ "$status" = "ACTIVE" ] || continue

    plan_date=$(echo "$frontmatter" | grep -E '^date: ' | sed 's/^date: *//' | tr -d '"')
    title=$(echo "$frontmatter" | grep -E '^title: ' | sed 's/^title: *//' | tr -d '"')
    rel=${plan#"$ROOT/"}

    # ── Check 1: Age-based staleness ──
    if [ -n "$plan_date" ]; then
        plan_ts=$(date -j -f "%Y-%m-%d" "$plan_date" +%s 2>/dev/null || true)
        if [ -n "$plan_ts" ]; then
            age=$(( (NOW - plan_ts) / 86400 ))
            if [ "$age" -gt "$THRESHOLD_DAYS" ]; then
                echo "⚠  Still active after ${age}d: $rel — \"$title\""
                COUNT=$((COUNT + 1))
            fi
        fi
    fi

    # ── Check 2: Related-path modification heuristic ──
    # If related files have git activity after the plan date, the plan was likely implemented.
    PLAN_DIR=$(dirname "$plan")
    in_related=false
    while IFS= read -r line; do
        if [[ "$line" == "related:" ]]; then
            in_related=true
            continue
        fi
        if $in_related; then
            if [[ "$line" =~ ^[a-zA-Z] ]] || [[ "$line" == "---" ]]; then
                in_related=false
                break
            fi
            if [[ "$line" =~ ^[[:space:]]+-[[:space:]]+(.+) ]]; then
                raw_path="${BASH_REMATCH[1]}"
                resolved=$(realpath -m "$PLAN_DIR/$raw_path" 2>/dev/null || echo "")
                [ -n "$resolved" ] || continue
                repo_rel="${resolved#"$ROOT"/}"
                [ "$repo_rel" != "$resolved" ] || continue

                if [ -n "$plan_date" ]; then
                    git_out=$(git -C "$ROOT" log --oneline --after="$plan_date" -- "$repo_rel" 2>/dev/null || true)
                    if [ -n "$git_out" ]; then
                        commit_count=$(echo "$git_out" | wc -l | tr -d ' ')
                        echo "⚠  Active plan but related files have commits after plan date: $rel — \"$title\" ($commit_count commits touch $repo_rel)"
                        COUNT=$((COUNT + 1))
                        break
                    fi
                fi
            fi
        fi
    done < <(echo "$frontmatter")
done

if [ "$COUNT" -gt 0 ]; then
    echo "→ $COUNT stale plan(s) found. Update frontmatter status to 'DONE' if implemented."
fi
exit 0
