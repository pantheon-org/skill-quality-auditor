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

1. Add a new script, e.g. `drift-audit.sh`, under
   `.context/plugins/pantheon-org/governance/adr-capture/scripts/`,
   implementing checks A and B above against `.context/index.yaml` and
   `docs/ADR/index.yaml`.
2. Resolve the tooling question (see Open Decisions #2): the manual audit
   this session used `python3` + `PyYAML`, which is not guaranteed present
   on the Actions runner or on every contributor's machine. Confirm an
   available equivalent (`yq`, or a vendored minimal parser) before
   committing to an implementation language, consistent with
   `merge-status-sync.sh`'s existing "no python/node in skills" convention.
3. Validate the script by running it against the current repo state and
   confirming it reproduces (or clearly explains why it doesn't reach) the
   8 known-drift items from `cross-reference-status-drift-2026-07-04.md`.
4. Ship as a `--dry-run`-only local CLI first (prints candidates to stdout),
   no GitHub Action yet — mirrors how `merge-status-sync.sh` shipped
   on-demand before the Action-wiring follow-up.

### Phase 2 — Scheduled Action + tracking issue

1. Add a GitHub Action triggered on `schedule: cron` (cadence: Open Decision
   #1) plus `workflow_dispatch` for manual runs.
2. On each run, invoke the Phase 1 script and open or update a single
   tracking issue (e.g. titled `Context drift audit — <date>`) listing each
   candidate with its file path, current status, and the specific evidence
   signal (which check flagged it, and what it points to).
3. If a tracking issue from a previous run is already open, edit it in place
   rather than opening a duplicate — mirrors the ADR-039 "single comment/issue,
   edited in place" pattern already established for the Plumber rollup issue.
4. Idempotency: a re-run with no new drift should not touch the issue, or
   should note "no new drift since `<date>`" rather than resetting it.
5. The Action must be read-only against `.context/` and `docs/ADR/` — its
   only write surface is the GitHub issue it manages via `gh`.

## Open Questions

1. Cron cadence — weekly (e.g. Monday mornings), monthly, or triggered only
   on pushes that touch `.context/index.yaml` or `docs/ADR/index.yaml`?
2. Script tooling — `bash` + `yq`, or a small vendored parser, given the
   Actions runner's default toolchain doesn't guarantee `python3`+`PyYAML`?
3. Does Phase 1 ship checks A+B only, or also attempt the harder
   content-check class from the companion finding? Recommendation: A+B only
   for Phase 1; content-check is an explicit, separately-scoped Phase 3.
4. Issue-tracking conventions — a dedicated label (e.g. `context-drift`)?
   Assignee or team routing?
5. Should the audit also flag findings/plans that have sat `active`/`draft`
   past some age threshold with *no* superseding item at all (a different,
   simpler staleness signal already partially covered by the pre-push
   "plans marked active for more than 60 days" check in `ways-of-working.md`)
   — or is that a separate, already-solved concern this plan should not
   duplicate?

## Verification (once implemented)

- Manually re-run the script against the 8 known-drift items from
  `cross-reference-status-drift-2026-07-04.md` as a regression fixture — it
  should flag all 8, or clearly explain which class each falls into (e.g.
  item 3, `tessl-eval-criteria-schema`, flagged as "needs manual content
  check" rather than silently missed).
- Confirm a clean run (no drift) produces no spurious issue/comment noise.
- Confirm the script never mutates `.context/` or `docs/ADR/` files — the
  only write path is the GitHub issue body via `gh`.

## Out of scope

- Auto-applying any status flip discovered by this audit. That decision
  stays with `merge-status-sync.sh`'s existing human-in-the-loop pattern:
  once a human confirms a flagged item should flip, run
  `merge-status-sync.sh` (or a manual edit + PR) against it directly.
- The content-check class of drift (Phase 2 idea in the companion finding,
  renumbered Phase 3 here to avoid clashing with this plan's own Phase 2) —
  flag as future work only, not committed scope.
