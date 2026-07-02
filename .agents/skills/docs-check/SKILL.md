---
name: docs-check
description: "Validate the GitHub Pages documentation site built by @docmd/core. Covers build verification, orphan detection, ADR index freshness, cross-reference integrity, and LLM output audit. Use when editing docs, after merging doc changes, before deploying, or when troubleshooting a broken site build. Triggers: 'check docs', 'build docs', 'verify site', 'docs audit', 'orphan detection', 'docmd check', 'preview docs', 'LLM output check', 'ADR index check', 'documentation validation'."
---

# Docs Check

Validate the GitHub Pages documentation site produced by `@docmd/core`.

## Prerequisites

- Node.js 18+ (for `npx @docmd/core`)
- The repo root as working directory
- Optional: `npx serve` or similar for local preview

## When to Use

- After editing or adding `docs/` content
- Before pushing doc changes to `main`
- When the docs CI workflow fails
- When adding or renaming pages
- When ADRs are added without updating `docs/ADR/index.yaml`
- Before a release to verify the published site is complete

## When Not to Use

- Do not use for checking the content quality of individual pages — use the `markdown-authoring` skill instead
- Do not use for checking agent skill documentation — use the `skill-quality-auditor` skill instead

## Checks

### 1. Build check

```bash
npx @docmd/core build
```

Expected result: exit code 0, output written to `./site/`.

If the build fails, inspect the error output — common causes include broken links, invalid YAML frontmatter, or missing referenced files.

### 2. Preview

```bash
npx @docmd/core dev
```

Opens a dev server at `http://localhost:3000` with hot reload. Browse the site and verify:
- All navigation links work
- New pages appear in the expected section
- No broken internal links
- Content renders correctly

Stop the dev server with `Ctrl+C`.

### 3. Orphan detection

Pages in `docs/` that are not linked from any index are invisible to site visitors. Check for orphans:

```bash
# List all markdown files under docs/ (excluding ADR index.yaml)
find docs/ -name '*.md' ! -name 'index.md' | sort > /tmp/all-pages

# Extract linked references from the index page
rg -o '\([^)]+\)' docs/index.md | rg -o '[^()]+' > /tmp/linked-pages

# Find orphans
comm -23 /tmp/all-pages /tmp/linked-pages
```

Any file in the output is an orphan — add a link from `docs/index.md` or a subsection index.

### 4. ADR index freshness

Every ADR file in `docs/ADR/` should be registered in `docs/ADR/index.yaml`:

```bash
# List ADR files (expect adr-NNN-*.md)
ls docs/ADR/adr-*.md 2>/dev/null | sed 's|docs/ADR/||' > /tmp/adr-files

# Extract registered ADRs from index.yaml
grep 'adr:' docs/ADR/index.yaml | sed 's/.*adr: //' > /tmp/adr-indexed

# Find unregistered ADRs
comm -23 <(sort /tmp/adr-files) <(sort /tmp/adr-indexed)
```

Unregistered ADRs should be added to `docs/ADR/index.yaml`. Use the `adr-capture` skill to generate the entry.

Also check for stale index entries (ADRs listed in `index.yaml` but missing from disk):

```bash
comm -13 <(sort /tmp/adr-files) <(sort /tmp/adr-indexed)
```

Entries in this output reference files that no longer exist — remove them from `index.yaml` or restore the files.

### 5. LLM output audit

When `llms.enabled` is true in `docmd.config.json`, the build generates `llms.txt`, `llms-full.txt`, and `llms.json` in the output directory. Verify they exist and contain expected entries:

```bash
# Check files exist
test -f site/llms.txt && echo "llms.txt: OK" || echo "llms.txt: MISSING"
test -f site/llms-full.txt && echo "llms-full.txt: OK" || echo "llms-full.txt: MISSING"
test -f site/llms.json && echo "llms.json: OK" || echo "llms.json: MISSING"

# Count entries
grep -c '^- \[' site/llms.txt 2>/dev/null || echo "llms.txt has no entries or is missing"
```

### 6. Cross-reference integrity

Reference files in `cmd/assets/references/` are embedded assets used by the CLI at runtime. If any should also appear on the doc site, verify they have corresponding pages under `docs/`:

```bash
ls cmd/assets/references/*.md | sed 's|cmd/assets/references/||' > /tmp/asset-refs
find docs/ -name '*.md' -exec basename {} \; | sort > /tmp/docs-pages
comm -12 /tmp/asset-refs /tmp/docs-pages
```

The output shows files present in both locations — these should be kept in sync.

## Workflow

1. Run **build check** to confirm the site compiles
2. Run **orphan detection** and link any found orphans
3. Run **ADR index freshness** and fix discrepancies
4. Run **LLM output audit** to verify machine-readable outputs
5. Run **preview** for a final manual review
6. Commit and push — the CI workflow deploys automatically

## Anti-Patterns

### NEVER skip the build check before pushing

**WHY:** A broken build blocks the deployment CI job and requires a separate fix commit.

**GOOD:** Run `npx @docmd/core build` locally before pushing doc changes.

### NEVER leave orphan pages unlinked

**WHY:** Orphan pages are invisible to navigation and search, wasting the effort spent writing them.

**GOOD:** Add a link from `docs/index.md` or the relevant subsection index.

### NEVER add an ADR without updating the index

**WHY:** The ADR index is the canonical discovery mechanism; a missing entry means the ADR may as well not exist.

**GOOD:** Add the ADR file and register it in `docs/ADR/index.yaml` in the same commit.

### NEVER reference a page that does not exist

**WHY:** Broken links degrade trust in the documentation and produce 404s for visitors.

**GOOD:** Use relative paths and verify the target file exists before committing.

## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| Available docmd configuration options | [docmd Configuration](references/docmd-config.md) | Modifying or troubleshooting the build config |
| Site structure and conventions | [Site Layout](references/site-layout.md) | Deciding where to place a new doc page |

## Troubleshooting

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| Build fails with "ENOENT" | Missing referenced file or broken symlink | Check the error path and restore or fix the reference |
| Dev server does not start | Port 3000 already in use | Kill the existing process or use `npx @docmd/core dev --port 3001` |
| Orphan detection shows false positives | A page linked from a sub-index or sidebar | Add the link target to the `rg` pattern or manually verify |
| `site/` is missing after build | `.gitignore` excludes `site/` (intentional) | The directory only exists after a local build; it is not tracked in git |
