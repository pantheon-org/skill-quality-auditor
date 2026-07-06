#!/usr/bin/env bash
# Validates YAML frontmatter in .context/*.md files against
# Resolved via SCRIPT_DIR relative to sibling skill's assets.
# Usage: validate-context-frontmatter.sh <file> [<file> ...]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA="$SCRIPT_DIR/../../context-file/assets/schemas/context-frontmatter.schema.json"

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

    if fm.get("type") == "PLAN" and fm.get("status") in ("DRAFT", "ACTIVE") and not fm.get("effort"):
        errors.append(
            f"{f}: type: PLAN with status: {fm.get('status')} must set 'effort' "
            f"(S/M/L, or TBD if genuinely blocked on an Open Question)"
        )

    if fm.get("type") == "KNOWN_ISSUE" and not fm.get("severity"):
        errors.append(
            f"{f}: type: KNOWN_ISSUE must set 'severity' (CRITICAL/HIGH/MEDIUM/LOW)"
        )

    if (
        fm.get("type") in ("PLAN", "FINDING", "KNOWN_ISSUE")
        and fm.get("status") in ("DRAFT", "ACTIVE")
        and not fm.get("value")
    ):
        errors.append(
            f"{f}: type: {fm.get('type')} with status: {fm.get('status')} must set 'value' "
            f"(HIGH/MEDIUM/LOW, graded against .context/instructions/value-rubric.md). "
            f"DONE/SUPERSEDED entries are exempt."
        )

    for field, values in enum_fields.items():
        if field in fm and fm[field] not in values:
            errors.append(f"{f}: '{field}' must be one of {values}, got '{fm[field]}'")

    for field, pattern in pattern_fields.items():
        if field in fm and not pattern.match(fm[field]):
            errors.append(f"{f}: '{field}' does not match pattern '{pattern.pattern}', got '{fm[field]}'")

    # themes is an array with a nested item enum; the enum_fields check above only
    # covers top-level scalar enums, so validate its members explicitly. Block YAML
    # style (like related) is the documented standard; inline [A, B] is tolerated.
    themes_enum = props.get("themes", {}).get("items", {}).get("enum")
    if themes_enum is not None:
        inline = fm.get("themes", "")
        if inline.startswith("["):
            members = [x.strip().strip('"').strip("'") for x in inline.strip("[]").split(",") if x.strip()]
        else:
            m = re.search(r"^themes:\n((?:  - .+\n?)+)", fm_text, re.MULTILINE)
            members = (
                [ln.strip()[2:].strip().strip('"') for ln in m.group(1).splitlines() if ln.strip().startswith("- ")]
                if m else []
            )
        if (
            fm.get("type") in ("PLAN", "FINDING", "KNOWN_ISSUE")
            and fm.get("status") in ("DRAFT", "ACTIVE")
            and not members
        ):
            errors.append(
                f"{f}: type: {fm.get('type')} with status: {fm.get('status')} must set a non-empty "
                f"'themes' list (ordered, primary-first, drawn from "
                f".context/instructions/theme-vocabulary.md). DONE/SUPERSEDED entries are exempt."
            )
        for tm in members:
            if tm not in themes_enum:
                errors.append(f"{f}: 'themes' member must be one of {themes_enum}, got '{tm}'")
        if len(members) != len(set(members)):
            errors.append(f"{f}: 'themes' members must be unique")

if errors:
    print("Frontmatter validation errors:")
    for e in errors:
        print(f"  {e}")
    sys.exit(1)

print(f"OK: {len(files)} file(s) validated against schema")
PYEOF
