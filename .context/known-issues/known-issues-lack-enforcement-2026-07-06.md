---
title: "Known Issue: known-issues has no enforcement, forcing function, or expiry"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-06
value: MEDIUM
severity: HIGH
themes:
  - GOVERNANCE
related:
  - ../../docs/ADR/adr-046-known-issues-context-type.md
  - ../../.context/plugins/pantheon-org/planning/design-debate/SKILL.md
---
# Known Issue: known-issues has no enforcement, forcing function, or expiry

> Originally surfaced as the sharpest finding of a session-reflection sub-agent: the `known-issues` mechanism (ADR-046) has `severity`/`status` fields and sorts to the top of `.context/index.yaml` — that's it. No CI gate fails on a `severity: critical` entry sitting `active`. Run through `design-debate` (Advocate / Skeptic / Migration-Risk) to decide whether to add enforcement now. **Verdict: `do_not_proceed_for_now`.**

## Why this exists

The Advocate proposed reusing `scripts/check-plan-drift.sh`'s proven 60-day age-threshold pattern, wired into `ci.yml` (unlike `check-plan-drift.sh` itself, which has the identical "advisory, pre-push-only, never CI" gap). Three findings against proceeding now, none rebutted:

1. **Severity is self-assigned with no calibration track record** (Skeptic) — the only two real examples so far (the jq bug rated `critical`, the CI-visibility gap rated `high`) are arguably backwards in actual impact, and a gate tied to `severity: critical` creates a rating-gaming incentive that doesn't exist while the field stays advisory.
2. **Age-since-creation can't distinguish "actively being fixed" from "neglected"** (Migration-Risk) — `context-file`'s own convention is that `date` is creation date, never updated on edits, so a critical issue worked on for 10 days looks identical to one abandoned for 10 days. This is a design flaw in the proposed mechanism, not an implementation detail.
3. **A date-threshold CI gate is non-deterministic** (Migration-Risk) — a PR that passed yesterday could fail today with zero new commits, purely from wall-clock aging. Every other job in `ci.yml` is diff-triggered; this would be a real departure from that invariant, and the reused `check-plan-drift.sh` pattern doesn't even run cross-platform as-is (`date -j` is macOS-only; GitHub's `ubuntu-latest` runners need a GNU `date -d` branch that doesn't exist yet).
4. Also: the gate would be **trivially, undetectably silenceable** by bumping the `date` field forward — nothing validates date-immutability beyond the documented convention.
5. **No usage evidence yet** — the mechanism is hours old, three entries total, zero currently at any proposed trigger condition. The one precedent for "advisory failed" (`check-docs-drift.sh`'s cumulative mode) was a different mechanism, fixed the same day it was found — not evidence of multi-week neglect.

## Impact if unfixed

The mechanism built specifically to avoid "a list nobody reads" risks becoming exactly that over time: visible, not enforced, the same pattern already shown insufficient earlier in the same session for `check-docs-drift.sh` before that got a CI-visibility fix (which itself still isn't a hard gate — see `.context/known-issues/docs-drift-cumulative-not-enforced-2026-07-06.md`, the sibling gap this one echoes for a different mechanism).

## Second design-debate (2026-07-06) — verdict reaffirmed

Re-opened on the premise that ADR-045's reviewed-baseline pattern now satisfied
revisit trigger (b). A second 3-role `design-debate` (Advocate / Skeptic /
Migration-Risk) **reaffirmed `do_not_proceed_for_now`**. Why the reopening did not hold:

- **The premise was not new information.** ADR-045 was accepted at 08:41; this
  known-issue and its original verdict were written at 10:45 the same day — the
  first debate already knew the pattern existed (its trigger text says "already
  built this same session"). ADR-052 then *reaffirmed* the rejection at 18:50 with
  ADR-045 fully in hand. Trigger (b) requires the **known-issues design to be
  reworked** around a reviewed timestamp (the proceed-action itself), not merely
  the pattern existing for docs.
- **Reviewed-baseline moves only objection #2.** #1 (self-assigned severity is
  gameable once it gates), #3 (a wall-clock gate is non-deterministic — every other
  `ci.yml` job is diff-triggered), #4 (a reviewed-epoch bump is as silenceable as a
  date bump, and the sidecar's own designers left marking un-authenticated *because*
  it was advisory), and #5 (a day on, zero neglect observed; the only `CRITICAL` is
  already `DONE`) all still stand. Migration-Risk confirmed the pattern mitigates
  only #2 and adds a new risk (orphaned sidecar entries as issues churn).
- **Even an advisory-only step is not worth it now.** Known-issues already sort to
  the top of `.context/index.yaml`, so they are already visible; a second sidecar +
  `mark-reviewed` + `check` scripts mostly duplicate the sibling
  `docs-drift-cumulative-not-enforced` pattern — itself an unfixed HIGH, i.e. live
  proof that a CI-visible advisory step has not delivered enforcement even where
  already tried. Adding a third instance before resolving the second is backwards.

**Carry-forward for any future attempt:** derive `created_epoch` from
`git log --diff-filter=A -- <file>` rather than parsing the frontmatter `date:` with
`date -j` — this sidesteps the macOS-only portability bug that stops
`check-plan-drift.sh` running on `ubuntu-latest`.

## Revisit trigger

Revisit when **either** holds:

- **(a) Evidence of real neglect** — a `HIGH` or `CRITICAL` known-issue is observed
  sitting `status: ACTIVE` for **30+ days** of real calendar time (now measurable as
  entries accrue), i.e. the problem this mechanism guards against actually manifests
  once.
- **(b) The sibling gap is resolved** — `docs-drift-cumulative-not-enforced` lands a
  chosen advisory-vs-hard-gate stance that a known-issues check could then mirror,
  so this is not the first place a new enforcement pattern is proven.
