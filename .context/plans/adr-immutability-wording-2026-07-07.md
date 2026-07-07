---
title: "Plan: reconcile ADR immutability wording to 'immutable from acceptance'"
type: PLAN
status: DONE
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

> **Implemented (2026-07-07), branch `chore/adr-immutability-wording`.** Phase 1:
> ADR-061 created (`proposed`), linking this plan + the finding; ADR index regenerated.
> Phase 2: the four sites reconciled — `SKILL.md` banner/mindset/anti-pattern and
> `evals/instructions.json:11`; references/template confirmed no-op; the repo-wide grep
> surfaced only one historical hit (`plumber-advisory-workflow-2026-07-04.md:133`,
> status DONE) left as a record. Phase 3: coherence gate clean, `.tessl` mirror
> re-synced (gitignored), skill re-scored 114/140 B (no regression vs the 85/F
> pre-remediation baseline), `hk check` + `go test` green.

## Goal

The `adr-capture` skill tells agents an ADR is immutable **once created**, but the
maintainer's actual rule is that immutability begins **at acceptance** — a
`proposed`, unmerged ADR is still a draft and may be edited in place. Bring the
skill's stated rule, its supporting docs, its eval assets, and the distributed mirror
into line with the real rule, and record the loosened rule as an ADR so the
governance change is itself documented. Success: the enumerated site list —
`SKILL.md` (banner, mindset bullet, anti-pattern), `evals/instructions.json`, plus
any of the references/template that turn out to overreach — all say immutability
applies from `accepted` onward (edit only `status`/`superseded_by`, otherwise
supersede), `proposed`/unmerged ADRs are explicitly editable, the change is captured
by a new ADR, and both `hk check` and `go test ./...` are green.

All three phases run in a **single feature branch** (not separate PRs), so the
Phase 1 ADR is indexed before the Phase 2 wording edit in the same working tree.

## Scope

**In scope:**

- The three statements of the rule in
  `.context/plugins/pantheon-org/governance/adr-capture/SKILL.md`: the immutability
  banner (`SKILL.md:10`), the mindset bullet (`SKILL.md:113`), and the
  "NEVER edit or delete an ADR after creation" anti-pattern (`SKILL.md:133-134`).
- `evals/instructions.json:11`, which restates the same "NEVER edit or delete an ADR
  after creation" overreach verbatim and is injected into agents as "new knowledge"
  during eval runs. This is a required edit, not just an inspection: left stale, an
  eval-graded agent that correctly amends a `proposed` ADR would be judged against a
  contradictory instruction from the same skill. Spot-check the other `evals/*` task
  and criteria files for the same phrasing while here (scenario-03 concerns an
  `accepted` ADR being superseded — correctly out of scope).
- Verification (and alignment only where actually overreaching) of
  `references/adr-frontmatter-schema.md`, `references/adr-supersession.md`, and
  `assets/templates/adr-template.yaml`. Note: a scan during planning found **zero**
  immutability-wording matches in the frontmatter-schema and template, so Task 2.2 is
  expected to be a no-op for those two — the real second site is the eval file above.
- A new ADR recording that immutability applies from `accepted`.
- Re-syncing the `.tessl` mirror of the skill from source.

**Out of scope:**

- The immutability of `accepted`/`superseded` ADRs — that rule is unchanged; this
  plan only clarifies that it does not apply to `proposed`/unmerged drafts.
- `references/adr-supersession.md:60` ("Never edit body of superseded ADR") — this is
  about superseded ADRs, which remain immutable; leave unless it reads as applying
  pre-acceptance.
- Encoding the rule in `adr-frontmatter.schema.json`'s status description
  (resolves Open Question 1). Decided against: behavioural rules live in the skill
  docs + ADR, and the schema stays a pure structural validator. The residual risk —
  future tooling that reads only the schema re-implementing the old rule — is
  accepted; if such tooling is ever built, encode the rule then.
- Any retroactive change to existing ADRs.

## Phases

### Phase 1 — Capture the reconciled rule as an ADR

Exit criterion: a new `proposed` ADR states that ADR immutability begins at
acceptance and that `proposed`/unmerged ADRs may be amended in place; it links this
plan and the source finding; the ADR index is fresh and frontmatter validates.

- Task 1.1: Create the ADR via the `adr-capture` skill, with `context:` linking this
  plan and `../findings/adr-immutability-wording-discrepancy-2026-07-07.md`. State the
  rule (immutable from `accepted`; before that, editable), the rationale (avoid
  needless supersession churn on unmerged drafts — as nearly happened with ADR-060),
  and the consequence (agents may refine a `proposed` ADR in place). **Re-check
  `docs/ADR/` for the next free number immediately before creating** — 058, 059, and
  060 were all minted on 2026-07-07, so a stale "next is 061" assumption is race-prone
  if another session claims it first.
- Task 1.2: Run the ADR index regeneration (`regenerate-adr-index.sh`) and the ADR
  frontmatter validator (`validate-adr-frontmatter.sh`); confirm the new ADR appears
  and validates. Note the binding repo gate is `hk`'s `adr-frontmatter` step
  (Task 3.2) — 1.2's scripts are the skill-local check, not a substitute for it.

### Phase 2 — Reconcile the skill wording to match

Exit criterion: all three rule statements in `adr-capture/SKILL.md` say immutability
applies from `accepted`; supporting docs verified and aligned only where they
overreach; wording references the new ADR where useful.

- Task 2.1: Edit `SKILL.md:10` (banner), `SKILL.md:113` (mindset bullet), and
  `SKILL.md:133-134` (anti-pattern), plus `evals/instructions.json:11`, so each scopes
  immutability to `accepted` onward and explicitly permits editing a `proposed`/
  unmerged ADR in place. Use precise wording to avoid the "from acceptance" ambiguity:
  a `proposed`/unmerged ADR is fully editable; once **accepted**, its title, body, and
  context are frozen and only `status`/`superseded_by` change (otherwise supersede).
- Task 2.2: Inspect `references/adr-frontmatter-schema.md`,
  `references/adr-supersession.md`, and `assets/templates/adr-template.yaml`; align
  only phrasing that wrongly implies immutability from creation (expected no-op for the
  first and third — see Scope). Leave superseded-ADR immutability wording as is.
- Task 2.3: Run an exhaustive repo-wide grep, not a fixed file list —
  `grep -rniE "immutab|once created|after creation" . --include='*.md' --include='*.json'`
  excluding `.git`, `.tessl`, `node_modules`, `dist`, `site` — and for every hit that
  states the rule, either align it to the Phase 1 ADR or confirm it is a
  correctly-out-of-scope superseded-ADR reference. The known repo-level candidates are
  `AGENTS.md`, its `CLAUDE.md` symlink, and `.context/instructions/ways-of-working.md`;
  the grep is the authority, the list is a hint. Exit only when the grep shows no
  unaligned "from creation" phrasing remains.

### Phase 3 — Sync mirror and verify

Exit criterion: the coherence grep is clean, the `.tessl` mirror matches source, and
all checks are green.

- Task 3.1: **Coherence gate (golden-rule check).** Run
  `grep -rniE "immutab|once created|after creation" .context/plugins/pantheon-org/governance/adr-capture/ | grep -viE "from acceptance|accepted|superseded"`
  and confirm it returns nothing. This catches a partial edit (e.g. 3 of 4 sites
  reconciled) before merge — the plan's own success check, promoted to a hard exit
  gate rather than a passive Verification line.
- Task 3.2: Re-sync the mirror with **bare `tessl install`** (which installs all
  plugins registered in `tessl.json`; the skill is registered at
  `pantheon-org/governance`, not as an `adr-capture` subdirectory, so the earlier
  `file:...adr-capture` target was invalid), then diff
  `.tessl/plugins/pantheon-org/governance/adr-capture/` for only the intended lines.
  The mirror is gitignored and CI regenerates it, so this is a local consistency
  check, not a committed artifact.
- Task 3.3: Re-score the skill — `./dist/skill-auditor evaluate <adr-capture path> --store` —
  and confirm no grade regression against the prior audit in
  `.context/audits/adr-capture/`. Repo discipline re-scores a skill on any SKILL.md
  content change; nearly free at this size.
- Task 3.4: Run `hk check` (adr-frontmatter, adr-index, adr-undocumented,
  markdownlint) and `go test ./...`; confirm green. The Go CLI is unaffected, so
  `go test` is a regression guard only.

## Risks

- **Over-broadening the loosening.** Wording that permits editing `proposed` ADRs
  must not read as permitting edits to `accepted`/`superseded` ones. Mitigation:
  Task 2.1 keeps the supersession rule explicit for `accepted` ADRs; Phase 1 ADR
  states the boundary precisely.
- **Mirror drift.** Editing source without re-syncing `.tessl` leaves the CI
  mirror-drift job to fail. Mitigation: Task 3.1 re-syncs and diffs.
- **Undocumented-decision gate.** `check-undocumented-decisions.sh` fires only on
  explicit decision-heading markers (a decision-prefixed H2/H3, a proposed-approach
  heading, or adopt-the-option phrasing), which neither this plan nor its finding
  uses — so it will not actually block the current files. (This plan deliberately
  avoids reproducing those exact marker strings in prose, since the detector matches
  them literally regardless of surrounding backticks or context.) Phase 1 (ADR) still
  precedes Phase 2 (wording) on governance grounds (a binding rule change warrants its
  ADR) and to keep the changeset self-consistent, but do not rely on the automated
  gate as the enforcement mechanism; it is a soft policy reason here, not a hard CI
  failure.
- **Partial reconciliation shipping silently.** If only some of the sites are edited,
  `hk check` still passes (it validates frontmatter/index, not rule coherence).
  Mitigation: Task 3.1's coherence grep is a hard exit gate.
- **Wording ambiguity.** "From acceptance" can be misread as "at the instant of
  acceptance". Mitigation: Task 2.1 pins the precise phrasing (proposed = fully
  editable; accepted = frozen except status/superseded_by).

## Verification

- `bash .context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/adr-immutability-wording-2026-07-07.md`
  passes.
- The new ADR validates and appears in `docs/ADR/index.yaml`.
- All enumerated sites are reconciled: `SKILL.md` ×3 (banner/mindset/anti-pattern) and
  `evals/instructions.json`, plus any references/template hit found by Task 2.3.
- The Task 3.1 coherence grep over the adr-capture directory returns nothing (only the
  untouched superseded-ADR rule remains).
- The Task 2.3 repo-wide grep shows no unaligned "from creation" phrasing anywhere.
- `skill-auditor evaluate` on adr-capture shows no grade regression.
- `hk check` and `go test ./...` are green.

## Open Questions

- ~~Should the loosened rule be encoded in `adr-frontmatter.schema.json`?~~ **Resolved
  (2026-07-07):** no — behavioural rules live in the skill docs + ADR; the schema stays
  a pure structural validator. Recorded in Scope (out of scope) with the accepted
  residual risk.
- ~~Does any doc outside `adr-capture` restate the rule?~~ **Folded into Task 2.3**,
  which is now an exhaustive repo-wide grep rather than a fixed file list, so coverage
  is determined at execution rather than guessed here.

None outstanding.
