---
title: "Plan: reconcile ADR immutability wording to 'immutable from acceptance'"
type: PLAN
status: DRAFT
date: 2026-07-07
effort: S
value: MEDIUM
themes:
  - GOVERNANCE
related:
  - ../findings/adr-immutability-wording-discrepancy-2026-07-07.md
  - ../../docs/ADR/adr-060-deferred-lifecycle-status.md
  - ../plugins/pantheon-org/governance/adr-capture/SKILL.md
---

# Plan: reconcile ADR immutability wording to "immutable from acceptance"

## Goal

The `adr-capture` skill tells agents an ADR is immutable **once created**, but the
maintainer's actual rule is that immutability begins **at acceptance** — a
`proposed`, unmerged ADR is still a draft and may be edited in place. Bring the
skill's stated rule, its supporting docs, and the distributed mirror into line with
the real rule, and record the loosened rule as an ADR so the governance change is
itself documented. Success: every place that states the immutability rule says it
applies from `accepted` onward (edit only `status`/`superseded_by`, otherwise
supersede), `proposed`/unmerged ADRs are explicitly editable, and the change is
captured by a new ADR; `hk check` stays green.

## Scope

**In scope:**

- The three statements of the rule in
  `.context/plugins/pantheon-org/governance/adr-capture/SKILL.md`: the immutability
  banner (`SKILL.md:10`), the mindset bullet (`SKILL.md:113`), and the
  "NEVER edit or delete an ADR after creation" anti-pattern (`SKILL.md:133-134`).
- Verification (and alignment only where actually overreaching) of
  `references/adr-frontmatter-schema.md`, `references/adr-supersession.md`, and
  `assets/templates/adr-template.yaml`.
- A new ADR recording that immutability applies from `accepted`.
- Re-syncing the `.tessl` mirror of the skill from source.

**Out of scope:**

- The immutability of `accepted`/`superseded` ADRs — that rule is unchanged; this
  plan only clarifies that it does not apply to `proposed`/unmerged drafts.
- `references/adr-supersession.md:60` ("Never edit body of superseded ADR") — this is
  about superseded ADRs, which remain immutable; leave unless it reads as applying
  pre-acceptance.
- Any retroactive change to existing ADRs.

## Phases

### Phase 1 — Capture the reconciled rule as an ADR

Exit criterion: a new `proposed` ADR states that ADR immutability begins at
acceptance and that `proposed`/unmerged ADRs may be amended in place; it links this
plan and the source finding; the ADR index is fresh and frontmatter validates.

- Task 1.1: Create the ADR via the `adr-capture` skill (next free number), with
  `context:` linking this plan and
  `../findings/adr-immutability-wording-discrepancy-2026-07-07.md`. State the rule
  (immutable from `accepted`; before that, editable), the rationale (avoid needless
  supersession churn on unmerged drafts — as nearly happened with ADR-060), and the
  consequence (agents may refine a `proposed` ADR in place).
- Task 1.2: Run the ADR index regeneration and frontmatter validation scripts;
  confirm the new ADR appears and validates.

### Phase 2 — Reconcile the skill wording to match

Exit criterion: all three rule statements in `adr-capture/SKILL.md` say immutability
applies from `accepted`; supporting docs verified and aligned only where they
overreach; wording references the new ADR where useful.

- Task 2.1: Edit `SKILL.md:10` (banner), `SKILL.md:113` (mindset bullet), and
  `SKILL.md:133-134` (anti-pattern) so each scopes immutability to `accepted`
  onward and explicitly permits editing a `proposed`/unmerged ADR in place. Keep the
  supersession rule intact for `accepted` ADRs.
- Task 2.2: Inspect `references/adr-frontmatter-schema.md`,
  `references/adr-supersession.md`, and `assets/templates/adr-template.yaml`; align
  only phrasing that wrongly implies immutability from creation. Leave
  superseded-ADR immutability wording as is.
- Task 2.3: Grep the repo-level guidance (`AGENTS.md`, its `CLAUDE.md` symlink,
  `.context/instructions/ways-of-working.md`) for any "ADR immutable once created"
  phrasing and align it to the ADR from Phase 1.

### Phase 3 — Sync mirror and verify

Exit criterion: the `.tessl` mirror matches source and all checks are green.

- Task 3.1: Re-sync the mirror with
  `tessl install file:.context/plugins/pantheon-org/governance/adr-capture` and
  confirm only the intended lines changed (the mirror is gitignored and CI
  regenerates it, so this is a local consistency check, not a committed artifact).
- Task 3.2: Run `hk check` (adr-frontmatter, adr-index, adr-undocumented,
  markdownlint) and `go test ./...`; confirm green. The Go CLI is unaffected, so
  `go test` is a regression guard only.

## Risks

- **Over-broadening the loosening.** Wording that permits editing `proposed` ADRs
  must not read as permitting edits to `accepted`/`superseded` ones. Mitigation:
  Task 2.1 keeps the supersession rule explicit for `accepted` ADRs; Phase 1 ADR
  states the boundary precisely.
- **Mirror drift.** Editing source without re-syncing `.tessl` leaves the CI
  mirror-drift job to fail. Mitigation: Task 3.1 re-syncs and diffs.
- **Undocumented-decision gate.** The rule change is a governance decision; shipping
  the wording edit without the ADR would fail `check-undocumented-decisions.sh`.
  Mitigation: Phase 1 (ADR) precedes Phase 2 (wording), and the ADR links this plan.

## Verification

- `bash .context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/adr-immutability-wording-2026-07-07.md`
  passes.
- The new ADR validates and appears in `docs/ADR/index.yaml`.
- `grep -rniE "immutab|once created|after creation" .context/plugins/pantheon-org/governance/adr-capture/`
  shows only wording consistent with "from acceptance" (plus the untouched
  superseded-ADR rule).
- `hk check` and `go test ./...` are green.

## Open Questions

- Should the loosened rule be encoded in the `adr-frontmatter.schema.json` status
  description (which currently only defines the status lifecycle), or is the SKILL.md
  + ADR statement sufficient? Leaning sufficient; the schema is not where behavioural
  rules live.
- Does any other skill or doc outside `adr-capture` restate the immutability rule and
  need the same alignment? Phase 2 Task 2.3 scopes the known locations; a repo-wide
  grep during execution confirms there are no others.
