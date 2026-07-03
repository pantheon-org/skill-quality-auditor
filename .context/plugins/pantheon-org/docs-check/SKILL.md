---
name: docs-check
description: "Validate the GitHub Pages documentation site built by @docmd/core. Covers build verification, orphan detection, ADR index freshness, cross-reference integrity, and LLM output audit. Use when editing docs, after merging doc changes, before deploying, or when troubleshooting a broken site build. Do NOT use for content quality review (use markdown-authoring) or agent skill documentation review (use skill-quality-auditor). Triggers: 'check docs', 'build docs', 'verify site', 'docs audit', 'orphan detection', 'docmd check', 'preview docs', 'LLM output check', 'ADR index check', 'documentation validation'."
---

# Docs Check

Validate the GitHub Pages documentation site produced by `@docmd/core`.

## Quick Reference

| Step | Check | Command |
| --- | --- | --- |
| 1 | Build | `npx @docmd/core build` |
| 2 | Orphans | `comm -23 <(find docs/ -name '*.md' \! -name 'index.md' \| sort) <(rg -o '\([^)]+\)' docs/index.md \| rg -o '[^()]+' \| sort)` |
| 3 | ADR index | Cross-ref `docs/ADR/adr-*.md` vs `docs/ADR/index.yaml` |
| 4 | LLM audit | Check `site/llms.txt`, `site/llms-full.txt`, `site/llms.json` |
| 5 | Preview | `npx @docmd/core dev` |

## Prerequisites

- Node.js 18+ (for `npx @docmd/core`)
- The repo root as working directory
- Optional: `npx serve` or similar for local preview

## Mindset

Documentation is only as valuable as it is discoverable. A broken build,
orphan pages, missing index entries, and stale references each silently
erode trust in the site. Treat every check as a gate: if it fails, the
user loses confidence.

- **Build first, ask questions later.** Always start with the build check — it is the fastest path to failure and catches mechanical issues early.
- **Automate what you check twice.** If you catch the same issue twice, TYPICALLY you should add a script or CI step so you NEVER check it manually again.
- **Fix the source, not the output.** When a check reveals a problem, fix the markdown source, not the built `site/` — the build is ephemeral.
- **Escalate to the right skill.** When content quality issues surface, redirect to `markdown-authoring`; when skill documentation is the concern, redirect to `skill-quality-auditor`. For quick formatting-only issues, PREFER using `markdownlint` directly.
- **Adapt to the situation.** UNLESS the CI deployment is already broken, you may skip the build check for trivial doc tweaks (single-word fixes, typo corrections).

## When to Use

- After editing or adding `docs/` content
- Before pushing doc changes to `main`
- When the docs CI workflow fails
- When adding or renaming pages
- When ADRs are added without updating `docs/ADR/index.yaml`
- Before a release to verify the published site is complete

## When Not to Use

| Situation | Use Instead |
| --- | --- |
| Checking content quality, style, or lint of individual pages | `markdown-authoring` skill |
| Reviewing agent skill documentation (SKILL.md) | `skill-quality-auditor` skill |
| Writing new documentation from scratch | `markdown-authoring` skill |
| Capturing an ADR decision | `adr-capture` skill |
| Reviewing formatting or lint issues only | Skip if prose quality is fine — use `markdownlint` directly |

## Checks

### 1. Build check

```bash
npx @docmd/core build
```

Expected result: exit code 0, output written to `./site/`.

If the build fails, inspect the error output — common causes include broken links, invalid YAML frontmatter, or missing referenced files.

**After a successful build**, confirm the page count is reasonable (typically 45-55 pages for this project). A sudden drop signals missing content.

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

1. **Build check** — confirm the site compiles. If it fails, triage the error and fix before proceeding.
2. **Orphan detection** — find and link any unlinked pages. If orphans exist, add links to `docs/index.md`.
3. **ADR index freshness** — reconcile on-disk ADRs with the index. Use `adr-capture` to register missing entries.
4. **LLM output audit** — verify machine-readable outputs exist with non-zero entries.
5. **Preview** — final manual review in browser before pushing.
6. **Commit and push** — the CI workflow deploys automatically when `docs/**` or `docmd.config.json` changes on `main`.

## Anti-Patterns

**NEVER** skip the build check before pushing

**WHY:** A broken build blocks the deployment CI job and requires a separate fix commit, wasting time and polluting git history.

**SYMPTOM:** Doc changes are pushed without a local build, and the CI deployment step fails with a cryptic error.

**CONSEQUENCE:** The deployment is blocked until a fix commit is pushed, adding latency and noise to git history.

**BAD:** Editing `docs/index.md`, committing, and pushing without running the build first.

**GOOD:** Run `npx @docmd/core build` locally before pushing doc changes.

---

**NEVER** leave orphan pages unlinked

**WHY:** Orphan pages are invisible to navigation and search, wasting the effort spent writing them.

**SYMPTOM:** A documentation page you wrote never appears in search results or navigation menus.

**CONSEQUENCE:** Site visitors cannot discover the content, so the writing effort is wasted.

**BAD:** A new `docs/architecture/deployment-flow.md` exists but is not linked from `docs/index.md` — it compiles but no visitor can reach it.

**GOOD:** Add a link from `docs/index.md` or the relevant subsection index. Run the orphan detection command to confirm.

---

**NEVER** add an ADR without updating the index

**WHY:** The ADR index is the canonical discovery mechanism; a missing entry means the ADR may as well not exist.

**SYMPTOM:** A reviewer asks "is there an ADR about this?" and the index query returns empty, even though the file exists on disk.

**CONSEQUENCE:** The ADR is invisible to automated discovery and humans navigating the index, defeating the purpose of documenting the decision.

**BAD:** Creating `docs/ADR/adr-026-new-format.md` without adding an `adr` entry to `docs/ADR/index.yaml`.

**GOOD:** Add the ADR file and register it in `docs/ADR/index.yaml` in the same commit. Use the `adr-capture` skill to generate the correct YAML entry.

---

**NEVER** reference a page that does not exist

**WHY:** Broken links degrade trust in the documentation and produce 404s for visitors, eroding confidence in the entire project.

**SYMPTOM:** Clicking a link on the site leads to a 404 page instead of the expected content.

**CONSEQUENCE:** Visitors lose trust in the documentation and may stop relying on it as an authoritative source.

**BAD:** Linking to `architecture/new-flow.md` in `docs/index.md` when the file was renamed to `architecture/updated-flow.md`.

**GOOD:** Use relative paths and verify the target file exists before committing. After renaming files, update all inbound links.

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
| LLM output files missing from site/ | `llms.enabled` is off in config | Enable `"llms": {"enabled": true}` in `docmd.config.json` and rebuild |
