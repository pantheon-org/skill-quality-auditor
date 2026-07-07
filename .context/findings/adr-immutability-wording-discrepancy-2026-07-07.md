---
title: "Finding: adr-capture skill says ADRs are immutable 'once created', but the rule is 'from acceptance'"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: MEDIUM
themes:
  - GOVERNANCE
related:
  - ../plugins/pantheon-org/governance/adr-capture/SKILL.md
  - ../../docs/ADR/adr-060-deferred-lifecycle-status.md
  - ./deferred-status-critical-review-2026-07-07.md
---

# Finding: ADR immutability wording is stricter than the actual rule

## The discrepancy

The `adr-capture` skill states, in both its immutability banner and its
anti-patterns, that an ADR is immutable **once created** ("Once created, an ADR is
never edited or deleted"; "NEVER edit or delete an ADR after creation"). The
maintainer's actual rule is that ADRs are immutable **from acceptance** — a
`proposed`, unmerged ADR is still a draft and may be amended in place; immutability
(edit only `status`/`superseded_by`, otherwise supersede) begins when the ADR is
`accepted`.

## Why it matters

Taken literally, the skill wording forces an agent to **supersede** an ADR it wrote
minutes ago in the same unmerged branch rather than simply refining it — spawning a
needless ADR-NNN+1 that supersedes a never-accepted ADR-NNN. That is churn the
immutability rule was never meant to create. This actually happened during ADR-060:
the ADR was amended in place across three iterations while `proposed`, which is
correct under the "from acceptance" rule but contradicts the skill's stated wording.

## Recommended follow-up

- Reconcile the wording in `adr-capture/SKILL.md` (immutability banner + the two
  relevant anti-patterns) to say immutability applies from `accepted` onward, and
  that `proposed`/unmerged ADRs may be edited in place. Mirror the change into the
  `.tessl` source if the skill is distributed from there.
- Because this loosens a stated governance rule, capture the reconciled rule as an
  ADR (the change to how ADRs themselves are governed is exactly the kind of
  process decision the ADR system exists to record).
- Check `references/adr-frontmatter-schema.md` and the ADR body template for the same
  "once created" phrasing and align them.

Left as follow-up, not actioned here, to keep the DEFERRED-status change focused.
