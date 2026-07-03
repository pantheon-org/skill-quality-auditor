# Scenario 03: ADR Index Freshness

## User Prompt

"Check the ADR index — make sure every ADR file has a corresponding entry in index.yaml and there are no stale entries."

## Expected Behavior

1. List all `adr-*.md` files under `docs/ADR/`.
2. Read `docs/ADR/index.yaml` and extract the registered ADR filenames from the `adr:` field.
3. Compare the two sets:
   - ADR files on disk but NOT in index.yaml are unregistered.
   - Entries in index.yaml referencing files NOT on disk are stale.
4. Report any discrepancies found.
5. For unregistered ADRs, suggest running the `adr-capture` skill to generate the index entry.
6. For stale entries, suggest removing the orphaned entry from index.yaml.
7. If everything is in sync, confirm freshness.

## Success Criteria

- All ADR files on disk are enumerated.
- All index.yaml entries are parsed.
- Both unregistered ADRs and stale index entries are correctly detected.
- User receives a clear report with paths.
- Remediation is suggested for each discrepancy.

## Failure Conditions

- Only one direction is checked (e.g., only files on disk, not stale index entries).
- Comparison is incorrect — discrepancies are missed or misreported.
- No remediation guidance given.
- User receives raw output without interpretation.
- The tool output for reading index.yaml is not parsed correctly.

**Context:**

- Repository root: current working directory
- ADR files follow the naming convention `adr-NNN-title.md`
- Index file is at `docs/ADR/index.yaml` with YAML entries containing `adr:` fields referencing filenames
