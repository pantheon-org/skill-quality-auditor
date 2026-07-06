---
title: "Known Issue: the adr-index gate only checks the file exists, so a stale ADR index passes CI"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-06
severity: MEDIUM
value: MEDIUM
themes:
  - GOVERNANCE
related:
  - ../../hk.pkl
  - ../plugins/pantheon-org/governance/adr-capture/scripts/regenerate-adr-index.sh
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh
---

The `adr-index` step in `hk.pkl` is:

    check = "test -f docs/ADR/index.yaml"

It verifies only that the index file **exists** — not that it is **fresh**. A new
ADR can be committed while `docs/ADR/index.yaml` still lacks its entry, and the gate
passes green. Compounding this, `regenerate-adr-index.sh` has no `--check` mode: it
takes `adr_dir` and `index_path` and always writes, so there is no read-only
freshness comparison to wire the gate to.

## Why it matters

This is a false-confidence gate. It looks like the ADR index is validated the way
the context index is, but it is not. The sibling `context-index` gate does it
correctly — `regenerate-context-index.sh --check` regenerates in memory and fails if
the output differs from the committed file. The ADR index has no equivalent, so
drift (a missing or reworded ADR entry) ships silently.

## Discovered

Observed directly while shipping ADR-049: `docs/ADR/index.yaml` did not yet contain
the ADR-049 entry (a manual `grep -c` returned 0), yet `hk check`'s `adr-index` step
reported green. Regenerating the index then produced a real diff, confirming the gate
had passed against a stale index.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Add a `--check` mode to `regenerate-adr-index.sh` mirroring
`regenerate-context-index.sh` (regenerate into a string, compare to the committed
file, exit non-zero with a "run to regenerate" message on mismatch), then point the
`hk.pkl` `adr-index` `check` at it instead of `test -f`. Keep `fix` pointing at the
plain regeneration invocation so `hk fix` still repairs drift.
