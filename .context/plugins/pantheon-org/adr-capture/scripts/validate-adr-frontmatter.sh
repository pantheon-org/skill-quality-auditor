#!/usr/bin/env bash
# Validates YAML frontmatter in docs/ADR/adr-*.md files against
# Resolved via SCRIPT_DIR relative to this skill's own assets.
# Usage: validate-adr-frontmatter.sh <file> [<file> ...]
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA="$SCRIPT_DIR/../assets/schemas/adr-frontmatter.schema.json"

python3 - "$SCHEMA" "$@" <<'PYEOF'
import sys
import json
import re
from pathlib import Path

schema_path = Path(sys.argv[1])
files = sys.argv[2:]

if not files:
    sys.exit(0)

schema = json.loads(schema_path.read_text())
required = schema.get("required", [])
props = schema.get("properties", {})

enum_fields = {k: v["enum"] for k, v in props.items() if "enum" in v}
pattern_fields = {k: re.compile(v["pattern"]) for k, v in props.items() if "pattern" in v}

errors = []

for f in files:
    p = Path(f)
    if not p.exists():
        continue
    content = p.read_text()
    if not content.startswith("---\n"):
        errors.append(f"{f}: missing frontmatter (file must begin with ---)")
        continue
    try:
        end = content.index("---\n", 4)
    except ValueError:
        errors.append(f"{f}: unclosed frontmatter (no closing ---)")
        continue
    fm_text = content[4:end]
    fm = {}
    for line in fm_text.splitlines():
        if ": " in line and not line.startswith(" "):
            k, _, v = line.partition(": ")
            fm[k.strip()] = v.strip().strip('"')

    for field in required:
        if not fm.get(field):
            errors.append(f"{f}: missing required field '{field}'")

    for field, values in enum_fields.items():
        if field in fm and fm[field] not in values:
            errors.append(f"{f}: '{field}' must be one of {values}, got '{fm[field]}'")

    for field, pattern in pattern_fields.items():
        if field in fm and not pattern.match(fm[field]):
            errors.append(f"{f}: '{field}' does not match pattern '{pattern.pattern}', got '{fm[field]}'")

if errors:
    print("ADR frontmatter validation errors:")
    for e in errors:
        print(f"  {e}")
    sys.exit(1)

print(f"OK: {len(files)} file(s) validated against schema")
PYEOF
