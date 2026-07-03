#!/usr/bin/env bash
# Scans .context/**/*.md for decision indicators and cross-references
# against ADR context: links. Reports any context files that appear
# to contain decisions but are not documented as an ADR.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
CONTEXT_DIR="$ROOT/.context"
ADR_DIR="$ROOT/docs/ADR"

python3 - "$CONTEXT_DIR" "$ADR_DIR" <<'PYEOF'
import sys
import re
from pathlib import Path

context_dir = Path(sys.argv[1])
adr_dir = Path(sys.argv[2])

# --- Collect all context file paths referenced by ADRs ---
referenced = set()

root = context_dir.parent

for adr_file in adr_dir.glob("adr-*.md"):
    content = adr_file.read_text()
    # find context: list blocks
    m = re.search(r"context:\n((?:  - .+\n?)+)", content)
    if m:
        for line in m.group(1).splitlines():
            line = line.strip()
            if line.startswith("- path:"):
                path_val = line.split(":", 1)[1].strip().strip('"').strip("'")
                resolved = (root / path_val).resolve()
                if resolved.exists():
                    referenced.add(str(resolved))

# --- Decision-indicating keywords in context files ---
DECISION_KEYWORDS = [
    r"## Recommended Action",
    r"## Decision",
    r"## Proposed Approach",
    r"## Recommendation",
    r"### Decision:",
    r"### Recommendation",
    r"\*\*Decision:\*\*",
    r"Adopt Option",
    r"Option A.*recommended",
    r"recommended approach",
    r"recommended path",
]

# --- Scan context files ---
undocumented = []

for md_file in sorted(context_dir.rglob("*.md")):
    resolved = str(md_file.resolve())
    if resolved in referenced:
        continue  # already tracked by an ADR

    # skip plugin skill files — they contain decision keywords in evals/docs
    if "/plugins/" in str(md_file):
        continue

    content = md_file.read_text()
    # skip index.yaml reference files
    if "index.yaml" in content:
        continue

    # look for decision indicators
    found_keywords = []
    for kw in DECISION_KEYWORDS:
        if re.search(kw, content, re.MULTILINE):
            found_keywords.append(kw)
            break  # one match per file is enough

    if found_keywords:
        rel = str(md_file.relative_to(context_dir.parent))
        undocumented.append((rel, found_keywords[0]))

# --- Report ---
if not undocumented:
    print("All context files with decisions are documented by ADRs.")
    sys.exit(0)

print("WARNING: The following context files contain decision indicators but")
print("are NOT referenced by any ADR. Consider creating an ADR for each:")
print()
for path, keyword in undocumented:
    print(f"  {path}")
    print(f"    Indicator: {keyword}")
    print()

print(f"Total: {len(undocumented)} undocumented decision(s)")
print("Run .context/plugins/pantheon-org/adr-capture/scripts/regenerate-adr-index.sh after creating ADRs.")
sys.exit(2)
PYEOF
