#!/usr/bin/env bash
# Scans .context/**/*.md for decision indicators and cross-references
# against ADR context: links. Reports any context files that appear
# to contain decisions but are not documented as an ADR.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
# CONTEXT_DIR/ADR_DIR default to the repo's real trees; they can be overridden
# via the environment so the fixture test can point at a temporary tree.
CONTEXT_DIR="${CONTEXT_DIR:-$ROOT/.context}"
ADR_DIR="${ADR_DIR:-$ROOT/docs/ADR}"

python3 - "$CONTEXT_DIR" "$ADR_DIR" <<'PYEOF'
import sys
import re
from pathlib import Path

context_dir = Path(sys.argv[1])
adr_dir = Path(sys.argv[2])


def strip_code(text):
    """Remove fenced code blocks and inline code spans so a decision marker
    merely *quoted* in prose (e.g. a doc explaining this gate) does not count as
    a real decision heading. See
    .context/known-issues/undocumented-decision-detector-false-positive-2026-07-07.md
    """
    text = re.sub(r"```.*?```", "", text, flags=re.DOTALL)
    text = re.sub(r"~~~.*?~~~", "", text, flags=re.DOTALL)
    text = re.sub(r"`[^`]*`", "", text)
    return text

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
# Binding-decision indicators only. In this repo, plans DECIDE (a "## Decisions"
# section, captured as an ADR) while findings/known-issues RECOMMEND (a
# "## Recommended Action" / "## Recommendation" section, which is not itself
# ADR-triggering — see adr-capture: "DO NOT use for observational findings
# without decisions"). Matching the soft recommendation headings produced false
# positives on findings, which the old blanket `index.yaml` skip masked; both are
# fixed here (G3, .context/plans/governance-tooling-hardening-2026-07-06.md).
# Heading/bold markers are anchored to line start (with re.MULTILINE) so a marker
# embedded mid-prose cannot match; "Adopt Option" is decision *phrasing* rather than
# a heading, so it stays unanchored. Both are additionally protected by strip_code(),
# which removes code spans/fences before matching.
DECISION_KEYWORDS = [
    r"^## Decision",  # also matches "## Decisions"
    r"^### Decision:",
    r"^\*\*Decision:\*\*",
    r"^## Proposed Approach",
    r"Adopt Option",
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

    content = strip_code(md_file.read_text())

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
print("Run .context/plugins/pantheon-org/governance/adr-capture/scripts/regenerate-adr-index.sh after creating ADRs.")
sys.exit(2)
PYEOF
