---
title: "Draft Plan: Periodic Cross-Reference Drift Audit"
type: plan
status: draft
date: 2026-07-04
related:
  - ../findings/cross-reference-status-drift-2026-07-04.md
  - ../findings/automate-post-merge-status-sync-2026-07-04.md
  - post-merge-status-sync-2026-07-04.md
  - ../plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh
---
# Draft plan: periodic cross-reference drift audit

Status: DRAFT for review
Date: 04-07-2026
Author: investigation by AI agent, decisions pending human owner
Reviewed: 04-07-2026 — 3-reviewer audit (Technical/Strategic/Risk lenses,
Claude Sonnet 5 for all three). Scores: Technical 6/10, Strategic 6/10,
Risk 5/10. See "Review findings" section near the end for the summary;
every critical and moderate finding has been folded directly into the
Phases, Open Questions, and a new Risks section below as concrete tasks
rather than left as freestanding notes.

> This is decision-support material. The final approach, cadence, and any
> automation scope must be decided and documented by a human maintainer
> before implementation.

## Goal

A scheduled (not per-PR-triggered) GitHub Action that runs a YAML
cross-reference audit across `.context/index.yaml` and `docs/ADR/index.yaml`,
surfacing status-drift candidates that the per-merge sync (`merge-status-sync.sh`,
proposed for automation in `automate-post-merge-status-sync-2026-07-04.md`)
structurally cannot see — because that script only resolves candidates from
the *merging PR's own* frontmatter/file-touch links, not from cross-item
references accumulated across unrelated, later PRs. The audit must only
surface candidates; it must never auto-write a status change itself.

## Scope

In scope: detecting drift where item A (plan/finding/ADR) has a stale status
because a *different*, later item B superseded it and B is already `done`/
`accepted`, but nothing ever flipped A. See
`cross-reference-status-drift-2026-07-04.md` for 8 concrete examples found by
manual audit this session.

Out of scope (see "Out of scope" section below): auto-applying any flip,
and the harder "referenced code/config was silently removed" drift class.

## Detection logic

Two checks, both pure YAML/text cross-referencing — no LLM call required for
the core pass:

1. **Check A (ADR-side):** for every ADR with `status: proposed` (or
   `status: superseded` missing a `superseded_by`), resolve its `context:`
   links against `.context/index.yaml` — if any linked plan/finding has
   `status: done`, flag the ADR as a drift candidate.
2. **Check B (plan-side):** for every plan with `status: done`, walk its
   `related:` list — any `active`/`draft` plan or finding appearing there is
   a drift candidate (signal: a done plan citing an active/draft one as a
   predecessor it fulfilled is a strong "this got superseded" signal).
3. **Known limitation, explicitly not Phase 1:** a third drift class exists
   where a finding's subject becomes moot because referenced code/config was
   removed (e.g. the `tessl-eval-criteria-schema-2026-06-30.md` case in the
   companion finding, where the CI step it describes no longer exists). That
   class requires reading file content or diffing against source, not just
   cross-referencing frontmatter. Track as a Phase 2 idea (an LLM-assisted
   content check), not a blocking requirement for Phase 1.

## Phases

### Phase 1 — Detection script + manual run

**Exit criterion (blocking, see Phase 2 gate below):** the script reproduces
— or clearly explains which detection class handles — all 8 known-drift
items from the regression fixture in Task 5.

1. **Resolve tooling first, before writing any parsing code (was previously
   sequenced as a sub-step; promoted to a blocking first task per review).**
   The manual audit this session used `python3` + `PyYAML`, which is not
   guaranteed present on the Actions runner or on every contributor's
   machine, and this repo's existing convention (documented via
   `merge-status-sync.sh`) is "no python/node in skills." Timebox a spike
   (~30 minutes) checking `yq` availability on GitHub-hosted runners;
   default to `yq` if present, only build a vendored minimal parser if it
   isn't. This must be settled before Task 2, since bash-only YAML parsing
   constrains what frontmatter structures (nested lists under `related:`,
   multi-line values) can even be supported.
2. Add a new script, e.g. `drift-audit.sh`, under
   `.context/plugins/pantheon-org/governance/adr-capture/scripts/`,
   implementing checks A and B above against `.context/index.yaml` and
   `docs/ADR/index.yaml`, using the tooling chosen in Task 1.
3. Before writing Check A's matching logic, write down the exact algorithm
   with a worked example: pick one real ADR (e.g. ADR-001) and its
   `context:` field, and show precisely how it resolves against the
   corresponding `.context/index.yaml` entry. The ADR index already stores
   `context:` paths as relative `.context/...` paths matching the index's
   own `path:` keys, so this should be a direct string match — write that
   down explicitly so a parsing or matching bug is caught in review, not
   discovered later as a silent false negative.
4. Define a structured output contract for the script — e.g. one candidate
   per line, tab-separated: `check_id<TAB>item_path<TAB>current_status<TAB>evidence`
   — so Phase 2's Action has a stable, parseable interface instead of
   free-text stdout (this is a hard dependency Phase 2 needs and the
   original draft left implicit).
5. Snapshot the 8 known-drift items from `cross-reference-status-drift-2026-07-04.md`
   as a static, checked-in test fixture (small excerpt files mirroring the
   shape of `.context/index.yaml` / `docs/ADR/index.yaml`), decoupled from
   live repo state. The live items may get fixed (see the companion
   finding's own recommendation to fix them by hand) before this script
   ships, and a fixture tied to live state would silently lose its
   regression coverage the moment that happens.
6. Validate the script against both the live repo and the static fixture
   from Task 5, confirming it flags all 8 known items — or, for the one
   item outside checks A/B (a finding whose subject became moot because
   referenced CI config was removed), confirms the script correctly reports
   it as "needs manual content check" rather than silently missing it.
7. Ship as a `--dry-run`-only local CLI (printing candidates in the Task 4
   format), no GitHub Action yet — mirrors how `merge-status-sync.sh`
   shipped on-demand before its own Action-wiring follow-up.

### Phase 2 — Scheduled Action + tracking issue

**Gate (new, per review): do not begin Phase 2 until Phase 1 Task 6 passes.**
Nothing previously stated this explicitly, which left room for Action work
to start in parallel with unproven detection logic.

1. Add a GitHub Action triggered on `schedule: cron` (cadence: Open Question
   1) plus `workflow_dispatch` for manual runs, with a `concurrency:` group
   (e.g. `drift-audit`) so an overlapping scheduled run and manual trigger
   cannot create duplicate issues or race on the same edit.
2. Declare `permissions: { issues: write }` only, and SHA-pin any
   third-party actions the workflow uses — consistent with this repo's
   existing convention (PR #171 pinned third-party actions by commit SHA
   and declared workflow permissions after the same class of gap was found
   in `ci.yml`/`skill-quality.yml`).
3. On each run, invoke the Phase 1 script and open or update a single
   tracking issue (e.g. titled `Context drift audit — <date>`). Reuse —
   don't reimplement or refactor — the "find existing open issue, edit in
   place" pattern from this repo's Plumber rollup-issue work
   (`feat(ci): single rollup issue for Plumber findings`, PR #154; see also
   ADR-039) as the reference implementation for this step.
4. Design the issue body to include, per candidate: file path, current
   status, the specific evidence signal (which check flagged it and what it
   points to), and a suggested next action (e.g. "run
   `merge-status-sync.sh <pr-number>`" or "confirm and hand-edit + PR"). A
   flagged candidate with no suggested action is a list nobody acts on.
5. Add a suppression mechanism: a per-candidate way for a human to mark
   "confirmed not drift" (e.g. a checkbox in the issue body, or an
   ignore-list file the script reads before flagging), so a dismissed item
   is not re-flagged on every subsequent run.
6. Idempotency: a re-run with no new, non-suppressed drift should not touch
   the issue, or should note "no new drift since `<date>`" rather than
   resetting it.
7. Add a lightweight health signal for the audit itself — e.g. a "last
   successful run" line in the issue, or a workflow-failure notification —
   so a broken cron expression or workflow syntax error doesn't go
   unnoticed for months. A watchdog needs a watchdog of its own.
8. The Action must be read-only against `.context/` and `docs/ADR/` — its
   only write surface is the GitHub issue it manages via `gh`.

## Open Questions

1. Cron cadence — weekly (e.g. Monday mornings), monthly, or triggered only
   on pushes that touch `.context/index.yaml` or `docs/ADR/index.yaml`?
   Still open; must be resolved before Phase 2 Task 1.
2. Script tooling — **addressed above as Phase 1 Task 1**, promoted to a
   blocking spike before any parsing code is written (was previously an
   open question resolved too late in the sequencing).
3. Does Phase 1 ship checks A+B only, or also attempt the harder
   content-check class from the companion finding? Recommendation
   unchanged: A+B only for Phase 1; content-check is an explicit,
   separately-scoped Phase 3.
4. Issue-tracking conventions — a dedicated label (e.g. `context-drift`)?
   Assignee or team routing? Still open, and now explicitly blocking:
   Phase 2 must not ship without an answer — a tracking issue with no owner
   risks becoming exactly the kind of ignored, stale signal this plan
   exists to eliminate.
5. **Resolved:** age-based staleness (findings/plans that have sat
   `active`/`draft` past some threshold with no superseding item at all) is
   **out of scope** for this plan. It duplicates the existing pre-push
   "plans marked active for more than 60 days" check in `ways-of-working.md`.
   This plan only detects supersession-based drift (item A superseded by a
   different, later item B), not age-based staleness with no successor.

## Verification (once implemented)

- Manually re-run the script against both the live repo and the static
  fixture (Phase 1 Task 5) — it should flag all 8 known-drift items from
  `cross-reference-status-drift-2026-07-04.md`, or clearly explain which
  class each falls into (e.g. the `tessl-eval-criteria-schema` item flagged
  as "needs manual content check" rather than silently missed).
- Confirm a clean run (no drift) produces no spurious issue/comment noise,
  **and** confirm a run with only previously-suppressed candidates also
  produces no noise — this exercises the suppression mechanism (Phase 2
  Task 5), not just the trivial zero-drift case.
- Confirm the script never mutates `.context/` or `docs/ADR/` files — the
  only write path is the GitHub issue body via `gh`.
- Confirm the Phase 2 workflow's `concurrency:` group actually prevents a
  duplicate issue when a scheduled and manual run overlap (test via two
  rapid manual `workflow_dispatch` triggers, or a manual trigger during a
  scheduled run).
- All 5 Open Questions above are resolved or explicitly deferred before
  Phase 1 implementation begins.

## Risks

1. **Silent false negatives from bash-only YAML parsing.** A hand-rolled or
   misconfigured parser can mishandle valid-but-unusual YAML (flow-style vs
   block-style lists, multi-line strings) and produce "no drift found"
   identically to a genuinely clean run — the worst failure mode for a
   detection tool, since it erodes trust silently rather than loudly.
   Mitigated by the worked-example matching algorithm (Phase 1 Task 3) and
   the static fixture (Phase 1 Task 5), but not eliminated.
2. **Coupling to `merge-status-sync.sh`'s data model.** Both scripts
   independently parse the same index files and frontmatter conventions
   with no shared schema or contract. A future change to one (e.g. adding a
   `supersedes:` field to make Check B more precise) can silently desync
   the other. Not addressed by this plan; flag as a follow-up if it becomes
   a real maintenance burden.
3. **Alert fatigue.** If checks A/B produce even a modest false-positive
   rate on a growing `.context/` corpus, the tracking issue becomes noisy
   and gets ignored — a common failure pattern for automated-audit tooling.
   Mitigated by the suppression mechanism (Phase 2 Task 5), but that
   mechanism's effectiveness is unproven until used in practice.
4. **Ownership.** Without an assigned owner or routing (Open Question 4),
   the audit risks becoming the same kind of ignored, stale signal this
   plan exists to fix, one layer up — detecting drift forever without
   anyone acting on it.
5. **Concurrent runs.** A scheduled run overlapping a manual
   `workflow_dispatch` run could create duplicate issues or race on the
   same edit if the `concurrency:` group (Phase 2 Task 1) is misconfigured
   or omitted.

## Review findings (04-07-2026)

This plan was reviewed by 3 independent Claude Sonnet 5 subagents (Technical,
Strategic, Risk lenses) before implementation began. Scores: Technical 6/10,
Strategic 6/10, Risk 5/10 — sound direction and a correctly scoped safety
boundary (no auto-apply, defers all status flips to `merge-status-sync.sh`'s
existing human-in-the-loop, branch+PR-gated model), but the initial draft had
unresolved sequencing (tooling decided too late), an undefined cross-phase
output contract, no remediation-to-issue linkage, no false-positive
suppression mechanism, and no named issue owner. All critical and moderate
findings have been folded directly into the Phases, Open Questions, and
Risks sections above as concrete tasks and gates, rather than left as
freestanding review notes.

## Out of scope

- Auto-applying any status flip discovered by this audit. That decision
  stays with `merge-status-sync.sh`'s existing human-in-the-loop pattern:
  once a human confirms a flagged item should flip, run
  `merge-status-sync.sh` (or a manual edit + PR) against it directly.
- The content-check class of drift (Phase 2 idea in the companion finding,
  renumbered Phase 3 here to avoid clashing with this plan's own Phase 2) —
  flag as future work only, not committed scope.
- Age-based staleness detection with no superseding item (see Open Question
  5, resolved above) — duplicates the existing 60-day pre-push staleness
  check and is not this plan's concern.
