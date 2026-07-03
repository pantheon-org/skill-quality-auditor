#!/usr/bin/env bash
# Checks docs/*.md and README.md against a static map of related source
# paths for drift: if related source has git commits after the doc's last
# commit, the doc has likely gone stale. Informational only — exits 0.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
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

for entry in "${MAPPINGS[@]}"; do
    doc_path="${entry%%|*}"
    globs="${entry#*|}"
    doc_file="$ROOT/$doc_path"
    [ -f "$doc_file" ] || continue

    doc_date=$(git -C "$ROOT" log -1 --format=%cI -- "$doc_path" 2>/dev/null || true)
    [ -n "$doc_date" ] || continue

    IFS=';' read -ra glob_arr <<<"$globs"
    hits=""
    for g in "${glob_arr[@]}"; do
        matches=$(git -C "$ROOT" log --oneline --after="$doc_date" -- "$g" 2>/dev/null || true)
        [ -z "$matches" ] && continue
        hits="${hits}${matches}"$'\n'
    done

    if [ -n "$hits" ]; then
        commit_count=$(echo "$hits" | grep -c . || true)
        echo "⚠  Possibly stale: $doc_path — $commit_count commit(s) touched related source since the doc was last updated ($doc_date)"
        COUNT=$((COUNT + 1))
    fi
done

if [ "$COUNT" -gt 0 ]; then
    echo "→ $COUNT doc(s) may need a review pass. Heuristic only — use judgement, not every source change needs a doc update."
fi
exit 0
