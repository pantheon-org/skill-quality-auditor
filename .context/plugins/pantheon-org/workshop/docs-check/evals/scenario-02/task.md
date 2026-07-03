# Scenario 02: Orphan Detection

## User Prompt

"Find any orphan pages in the docs — files under docs/ that aren't linked from the index."

## Expected Behavior

1. List all `.md` files under `docs/` (excluding `docs/ADR/index.yaml`).
2. Extract all linked file paths from `docs/index.md` (relative links in parentheses).
3. Compare the two sets to find unlinked files.
4. Report any orphans found, with their paths.
5. Suggest adding a link to `docs/index.md` for each orphan.
6. If no orphans exist, report that the site is fully linked.

## Success Criteria

- All `.md` files under `docs/` are discovered.
- Links from `docs/index.md` are extracted.
- Comparison correctly identifies orphans (or confirms none exist).
- User receives a clear list of orphan paths or confirmation of completeness.
- Remediation is suggested when orphans are found.

## Failure Conditions

- Only a subset of files is checked.
- Link extraction is missing or incorrect.
- Orphans are reported as linked, or linked pages reported as orphans.
- No remediation guidance given.
- User receives raw output without interpretation.

**Context:**

- Repository root: current working directory
- `docs/index.md` is the homepage and primary navigation hub
- Subsections (architecture/, development/, reference/, ADR/) do not have their own index pages — all links are centralised in `docs/index.md`
