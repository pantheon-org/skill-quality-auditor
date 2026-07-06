---
title: "Known Issue: cumulative docs-drift is now visible in CI but still never enforced"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-06
value: MEDIUM
severity: HIGH
related:
  - ../findings/docs-drift-cumulative-mode-ci-gap-2026-07-06.md
  - ../../docs/ADR/adr-044-docs-drift-pr-gate.md
  - ../../.github/workflows/ci.yml
---
# Known Issue: cumulative docs-drift is now visible in CI but still never enforced

> `.context/findings/docs-drift-cumulative-mode-ci-gap-2026-07-06.md` fixed cumulative mode's total invisibility to CI by adding it as a second step in `ci.yml`'s `docs-drift` job. That step always `exit 0`s by design (matches the script's existing informational character), so a PR can still merge with real, flagged doc drift — the warning is now visible in an Actions log nobody is required to read, but nothing blocks the merge. The bypass vectors that motivated the original investigation (`--no-verify`, GitHub web-UI merges, API commits, fork PRs where local hooks were never installed) remain fully unenforced for cumulative-mode drift.

## Why this exists

Only *gate mode* (ADR-044, the `pull_request`-scoped diff-against-base check) blocks anything, and it's deliberately scoped to the current PR's own diff — it was never meant to, and still doesn't, catch accumulated historical drift. Making cumulative mode itself a blocking gate was never proposed, reviewed, or decided; it would immediately raise the same "how do we not fail every PR on pre-existing debt" problem gate mode was designed to sidestep, this time for a check that's supposed to also catch old debt, not just new drift.

## Impact if unfixed

The stated goal implied by having a docs-drift check at all ("stale docs shouldn't silently ship") isn't actually met — it's now *observable* in CI, but observability isn't enforcement. A team that doesn't habitually read Actions logs gets no practical benefit over the pre-existing local-pre-push-only state, beyond a paper trail after the fact.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Not a mechanical fix — a policy decision. Options, none decided: (a) leave as informational-only permanently, accepting the trade-off explicitly rather than by default; (b) make cumulative mode blocking, but only for docs whose sidecar-eligible reviewed-baseline is stale beyond some grace period (avoids day-one mass failure); (c) post cumulative-mode output as a PR comment (this repo already has a marker-based upsert pattern from ADR-043) so it's visible without needing to open Actions logs, without making it blocking. Needs a deliberate decision, not a default.
