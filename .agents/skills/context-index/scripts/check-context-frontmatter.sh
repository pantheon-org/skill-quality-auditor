#!/usr/bin/env bash
set -euo pipefail

failed=()
for f in "$@"; do
    [[ -f "$f" ]] || continue
    first_line=$(head -1 "$f")
    if [[ "$first_line" != "---" ]]; then
        failed+=("$f")
    fi
done

if [[ ${#failed[@]} -gt 0 ]]; then
    echo "ERROR: Missing YAML frontmatter in the following .context files:"
    for f in "${failed[@]}"; do
        echo "  $f"
    done
    printf '\nAdd a frontmatter block (title, type, status, date) and re-run.\n'
    printf 'Use the context-file skill for the schema and templates.\n'
    exit 1
fi
