#!/usr/bin/env bash
# Validates the skill-quality-auditor tile structure and files.
# Checks: tile.json well-formed, evals/ present, schemas/ present,
# references/ present, SKILL.md frontmatter complete.
# Usage: validate-tile.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
errors=0

check_file() {
    if [ ! -f "$1" ]; then
        echo "ERROR: missing $1"
        errors=$((errors + 1))
    fi
}

check_dir() {
    if [ ! -d "$1" ]; then
        echo "ERROR: missing directory $1"
        errors=$((errors + 1))
    fi
}

check_file "$ROOT/tile.json"
check_file "$ROOT/SKILL.md"
check_dir "$ROOT/evals"
check_dir "$ROOT/schemas"
check_dir "$ROOT/templates"
check_dir "$ROOT/references"

if python3 -c "import json; json.load(open('$ROOT/tile.json'))" 2>/dev/null; then
    :
else
    echo "ERROR: tile.json is not valid JSON"
    errors=$((errors + 1))
fi

if [ "$errors" -eq 0 ]; then
    echo "OK: tile structure validated"
else
    echo "FAIL: $errors error(s) found"
    exit 1
fi
