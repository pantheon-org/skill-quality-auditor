#!/usr/bin/env bash
# Checks docs/*.md and README.md against a static map of related source
# paths for drift.
#
# Default mode (no args): cumulative check used by the pre-push hook —
# flags a doc as possibly stale if related source has git commits after
# the doc's last commit, no matter how long ago. Informational only —
# always exits 0.
#
# A doc reviewed and confirmed current via mark-docs-reviewed.sh (see
# ADR-045) has its baseline raised to max(doc_last_touch_epoch,
# reviewed_epoch) instead of just doc_last_touch_epoch, read from the
# scripts/docs-drift-reviewed.jsonl sidecar. Comparison is done on epoch
# integers, never ISO8601 strings — string comparison of differently
# timezone-offset dates sorts wrong; epoch comparison never does.
#
# Gate mode (pass a base ref, e.g. `check-docs-drift.sh origin/main`):
# used by CI on pull_request — flags only source changes made between
# the base ref and HEAD that touch a mapped source glob without this
# same range also touching the corresponding doc. Exits 1 on any gap,
# so it can block a PR that introduces new drift without also failing
# on whatever pre-existing drift the cumulative mode already knows about.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
BASE_REF="${1:-}"
COUNT=0

# doc_path|source_glob1;source_glob2;...
MAPPINGS=(
    "README.md|cmd/*.go;agents/registry.go"
    "docs/index.md|cmd/*.go;scorer/*.go;internal/*/*.go"
    "docs/architecture/overview.md|cmd/*.go;scorer/*.go;internal/*/*.go;reporter/*.go;duplication/*.go;analysis/*.go;agents/registry.go"
    "docs/architecture/batch-flow.md|cmd/batch.go"
    "docs/architecture/duplication-flow.md|cmd/duplication.go;duplication/*.go"
    "docs/architecture/aggregation-flow.md|cmd/aggregate.go"
    "docs/architecture/remediation-flow.md|cmd/remediate.go;reporter/remediation*.go"
    "docs/architecture/trend-flow.md|cmd/trend.go"
    "docs/architecture/eval-runner.md|cmd/eval.go"
    "docs/architecture/init-update-prune.md|cmd/init.go;cmd/update.go;cmd/prune.go;agents/registry.go"
    "docs/architecture/validate-analyze.md|cmd/validate.go;cmd/analyze.go;analysis/*.go"
    "docs/reference/d1-knowledge-delta.md|scorer/d1_*.go;cmd/assets/assets/config/scoring-patterns.yaml"
    "docs/reference/d2-mindset-procedures.md|scorer/d2_*.go"
    "docs/reference/d3-anti-pattern-coverage.md|scorer/d3_*.go"
    "docs/reference/d4-specification-compliance.md|scorer/d4_*.go"
    "docs/reference/d5-progressive-disclosure.md|scorer/d5_*.go"
    "docs/reference/d6-freedom-calibration.md|scorer/d6_*.go;cmd/assets/assets/config/scoring-patterns.yaml"
    "docs/reference/d7-pattern-recognition.md|scorer/d7_*.go"
    "docs/reference/d8-practical-usability.md|scorer/d8_*.go"
    "docs/reference/d9-eval-validation.md|scorer/d9_*.go"
    "docs/reference/scoring-dimensions.md|scorer/dimensions.go;scorer/thresholds.go"
    "docs/development/adding-a-scorer.md|scorer/dimensions.go;scorer/scorer.go"
    "docs/development/skills-and-rules.md|.context/plugins/**;.agents/RULES.md;tessl.json"
    "docs/development/setup.md|hk.pkl;mise.toml"
)

if [ -n "$BASE_REF" ]; then
    for entry in "${MAPPINGS[@]}"; do
        doc_path="${entry%%|*}"
        globs="${entry#*|}"
        doc_file="$ROOT/$doc_path"
        [ -f "$doc_file" ] || continue

        doc_touched=$(git -C "$ROOT" log --oneline "$BASE_REF"...HEAD -- "$doc_path" 2>/dev/null || true)
        [ -n "$doc_touched" ] && continue

        IFS=';' read -ra glob_arr <<<"$globs"
        hits=""
        for g in "${glob_arr[@]}"; do
            matches=$(git -C "$ROOT" log --oneline "$BASE_REF"...HEAD -- "$g" 2>/dev/null || true)
            [ -z "$matches" ] && continue
            hits="${hits}${matches}"$'\n'
        done

        if [ -n "$hits" ]; then
            hits="${hits%$'\n'}"
            echo "✗ $doc_path is likely out of date — this PR changes related source without updating it:"
            echo "    ${hits//$'\n'/$'\n    '}"
            COUNT=$((COUNT + 1))
        fi
    done

    if [ "$COUNT" -gt 0 ]; then
        echo
        echo "→ $COUNT doc(s) likely need an update in this PR. Update the doc(s) above, or if this change genuinely doesn't warrant one, explain why in the PR description."
        exit 1
    fi
    echo "No new doc drift introduced by this PR."
    exit 0
fi

REVIEWED_SIDECAR="$ROOT/scripts/docs-drift-reviewed.jsonl"

# Looks up doc_path's reviewed entry in the sidecar, if any. JSONL (one JSON
# object per line), parsed via jq — already an assumed-available dependency
# elsewhere in this repo's scripts/ (plumber-*.sh). Deliberately not a bash
# associative array (declare -A): bash 3.2 (macOS's default /bin/bash, still
# what `env bash` resolves to on a stock Mac) predates bash 4 and has none.
lookup_reviewed() {
    local doc="$1"
    [ -f "$REVIEWED_SIDECAR" ] || return 0
    jq -c --arg d "$doc" 'select(.doc == $d)' "$REVIEWED_SIDECAR" 2>/dev/null
}

for entry in "${MAPPINGS[@]}"; do
    doc_path="${entry%%|*}"
    globs="${entry#*|}"
    doc_file="$ROOT/$doc_path"
    [ -f "$doc_file" ] || continue

    doc_meta=$(git -C "$ROOT" log -1 --format='%cI|%ct' -- "$doc_path" 2>/dev/null || true)
    [ -n "$doc_meta" ] || continue
    doc_date="${doc_meta%%|*}"
    doc_epoch="${doc_meta#*|}"

    effective_date="$doc_date"
    effective_epoch="$doc_epoch"
    reviewed_line=$(lookup_reviewed "$doc_path")
    if [ -n "$reviewed_line" ]; then
        reviewed_iso=$(printf '%s' "$reviewed_line" | jq -r '.reviewed_iso // empty')
        reviewed_epoch=$(printf '%s' "$reviewed_line" | jq -r '.reviewed_epoch // empty')
        case "$reviewed_epoch" in
            '' | *[!0-9]*)
                echo "⚠  Skipping malformed docs-drift-reviewed.jsonl entry for '$doc_path' (non-numeric epoch)" >&2
                ;;
            *)
                if [ "$reviewed_epoch" -gt "$doc_epoch" ]; then
                    effective_epoch="$reviewed_epoch"
                    effective_date="$reviewed_iso (reviewed)"
                fi
                ;;
        esac
    fi

    IFS=';' read -ra glob_arr <<<"$globs"
    hits=""
    for g in "${glob_arr[@]}"; do
        matches=$(git -C "$ROOT" log --oneline --after="@$effective_epoch" -- "$g" 2>/dev/null || true)
        [ -z "$matches" ] && continue
        hits="${hits}${matches}"$'\n'
    done

    if [ -n "$hits" ]; then
        commit_count=$(echo "$hits" | grep -c . || true)
        echo "⚠  Possibly stale: $doc_path — $commit_count commit(s) touched related source since the doc was last updated ($effective_date)"
        COUNT=$((COUNT + 1))
    fi
done

if [ "$COUNT" -gt 0 ]; then
    echo "→ $COUNT doc(s) may need a review pass. Heuristic only — use judgement, not every source change needs a doc update."
fi
exit 0
