---
title: "Plan: add an execution-location / coverage lens to plan-review reviewer prompts"
type: PLAN
status: DRAFT
date: 2026-07-07
effort: S
value: MEDIUM
themes:
  - SKILL-QUALITY
related:
  - ../known-issues/plan-review-execution-location-blind-spot-2026-07-06.md
  - ../findings/docs-drift-cumulative-mode-ci-gap-2026-07-06.md
  - ../plugins/pantheon-org/planning/plan-review/SKILL.md
  - ../../docs/ADR/adr-058-plan-review-execution-location-lens.md
---

# Plan: add an execution-location / coverage lens to plan-review

Status: DRAFT for review
Date: 07-07-2026
Branch: `feat/plan-review-execution-location-lens`

> Decision-support material. The lens-ownership choice in the Decisions section
> is proposed, not final; a human maintainer confirms it before implementation.

## Goal

`plan-review`'s three reviewer lenses (Technical, Strategic, Risk) are thorough
on a plan's *internal* mechanics but none is prompted to trace a stated
invocation/trigger/execution fact to its *external reach* — who or what bypasses
that path, and whether the plan's stated goal is actually met given that reach.
That gap let the docs-drift reviewed-baseline plan ship a mechanism that could
never run in CI (its own Goal section stated it was consumed only by the
`pre-push` hook, and no lens examined the implication). See
[`plan-review-execution-location-blind-spot-2026-07-06`](../known-issues/plan-review-execution-location-blind-spot-2026-07-06.md).

Success: at least one `plan-review` reviewer prompt explicitly directs the
reviewer to trace any stated execution/invocation fact to its coverage and
bypass implications, the synced `.tessl` copy matches the source, and an eval
scenario exercises the new check so the behaviour cannot silently regress.

## Scope

**In scope:**

- One additional instruction in the Risk reviewer prompt (primary owner), plus a
  lighter completeness cross-reference in the Strategic prompt, in
  `.context/plugins/pantheon-org/planning/plan-review/SKILL.md`.
- Re-syncing the generated copy under `.tessl/plugins/` via
  `tessl install file:.context/plugins/pantheon-org/planning/plan-review`.
- One new eval scenario (`scenario-04`) that presents a plan stating a narrow
  execution path and expects the reviewer to flag the coverage/bypass gap. Eval
  scenarios are discovered by **directory convention** — confirmed by inspection:
  `evals/instructions.json` is a generic procedure list, not a scenario manifest,
  so adding a `scenario-04/` directory with the sibling file set is sufficient
  (no manifest to update). Each scenario carries `task.md`, `criteria.json` (a
  `weighted_checklist`), and `capability.txt`.
- Moving the known-issue to `status: DONE` and regenerating the context index.

**Out of scope:**

- Restructuring the three-lens model or adding a fourth reviewer — this is a
  prompt refinement, not an architecture change.
- Any change to the docs-drift mechanism itself (tracked separately by the
  docs-drift known-issues and the `tessl-eval-decommission` plan).
- Re-scoring `plan-review` against the D1-D9 framework.

## Decisions

Resolved during `plan-review` (2026-07-07, 3 reviewers: Technical + Strategic on
Sonnet 5, Risk on Haiku 4.5).

1. **Risk lens owns the new instruction; Strategic gets a light cross-reference;
   Technical stays unchanged.** Blind spots, unstated assumptions, and "system
   boundaries not considered" are already the Risk lens's remit (question 1,
   BLIND SPOTS), so the execution-reach question extends an existing focus rather
   than diluting a new one. Strategic gets only a light completeness cross-reference
   (its COMPLETENESS question already asks "what would a reasonable person expect
   that isn't here"). *Alternative considered and rejected:* the Strategic reviewer
   argued the Technical lens ("feasibility — does this work given where it's
   invoked") is an equally natural home and that adding it there would give
   two-angle coverage of the docs-drift-shaped miss. Rejected to keep the change
   minimal and avoid editing a third prompt; the Risk+Strategic pairing already
   covers the "who bypasses this / is the goal met" question, and a later plan can
   extend Technical if evals show the pairing misses feasibility-framed cases.
2. **Phrase as an extension of Risk question 1, not a sixth question.** Adding a
   whole new numbered question risks reviewers treating it as optional boilerplate;
   folding it into the existing BLIND SPOTS item keeps the prompt tight. Wording:
   *"If the plan states where or how a mechanism executes (a hook, a CI job, a
   trigger, a caller), explicitly assess who or what bypasses that path and whether
   the stated goal is actually met given that reach."*
3. **Add one eval scenario rather than amending an existing one**, so the
   regression signal is isolated and the existing scenarios keep their current
   expected findings.
4. **The Strategic cross-reference is a one-clause aside, not a full sub-bullet**
   (resolves prior Open Question 2). A full sixth-style sub-bullet on Strategic
   would reintroduce the dilution risk Decision 1 avoids on the Risk side. Wording:
   append to Strategic's COMPLETENESS question — *"...including whether stated
   execution paths actually reach far enough to meet the goal."*

## Phases

### Phase 0 — Baseline (pre-flight)

Exit criterion: a clean `hk check` baseline is captured so any Phase 3 failure can
be attributed to this change rather than pre-existing debt.

- Task 0.1: Run `hk check` on the untouched branch and record the result. If it is
  already red for unrelated reasons, note which checks so Phase 3 compares like for
  like.

### Phase 1 — Amend the reviewer prompts (source of truth)

Exit criterion: the source `SKILL.md` carries the execution-reach instruction on
the Risk lens and the one-clause aside on Strategic; wording matches the Decisions
section; both prompt JSON blocks still parse.

- Task 1.1: In a **single atomic edit**, extend the Risk reviewer prompt
  (`SKILL.md:454-465`, BLIND SPOTS item) with the Decision 2 clause **and** append
  the Decision 4 aside to the Strategic reviewer prompt (`SKILL.md:444`,
  COMPLETENESS question). Editing both prompts together avoids version drift
  between them. The prompts are single `\n`-delimited JSON string values, so every
  added `"` and newline must be escaped as `\"` and `\n`.
- Task 1.2: Validate both edited prompt strings parse as JSON — extract each
  `prompt` value and run it through a parser, e.g.
  `python3 -c "import json,sys; json.loads(sys.stdin.read())"` (or `jq .` on the
  enclosing JSON block). This is a scripted gate, not a visual re-read.

### Phase 2 — Sync the generated copy and add an eval scenario

Exit criterion: `.tessl` copy matches source; a new eval scenario asserts the
coverage/bypass finding **and passes when run**. Waves A and B are independent and
can run concurrently.

- Task 2.1 (Wave A): `git stash` any unrelated work, then run
  `tessl install file:.context/plugins/pantheon-org/planning/plan-review` so
  `.tessl/plugins/pantheon-org/planning/plan-review/SKILL.md` matches the edited
  source. Confirm with `diff` (or `git diff --stat`) that only the intended prompt
  lines changed in the `.tessl` copy. If `tessl` reorders or reformats the file
  beyond the intended lines, treat the sync as failed and reconcile before
  proceeding — do not accept a divergent copy.
- Task 2.2 (Wave B): Create `evals/scenario-04/` with the sibling file set
  (`task.md`, `criteria.json` as a `weighted_checklist`, `capability.txt`). The
  `task.md` fixture MUST state a narrow execution path in terms the reviewer has to
  reason about (e.g. a plan whose mechanism "is consumed only by the pre-push
  hook") **without** using the instruction's own vocabulary (bypass / reach /
  coverage), so a pass reflects reasoning, not keyword echo. The `criteria.json`
  checklist MUST score whether the Risk (and/or Strategic) reviewer actually flags
  that the mechanism never runs in CI and the stated goal is therefore unmet.
- Task 2.3 (Wave B): Run the scenario through the eval runner —
  `tessl eval list` to confirm `scenario-04` is discovered, then `tessl eval run`
  (scoped to this plugin) to confirm it executes and the expected finding scores
  above threshold. Capture stderr; a discovered-but-skipped scenario is a failure.
- Task 2.4 (Wave B): **Regression check** — re-run scenarios 01-03 and confirm
  their scores are unchanged, proving the Risk-prompt edit did not degrade the
  existing BLIND SPOTS behaviour (the "prompt dilution" risk).

### Phase 3 — Close out the known-issue

Exit criterion: all checks green **first**, then the known-issue is `DONE` and the
index is regenerated cleanly. Status is flipped only after verification passes, so
a red check never leaves a prematurely-resolved issue.

- Task 3.1: Run `hk check` (or the repo pre-push equivalent) and confirm green
  relative to the Phase 0 baseline; confirm Phase 2's eval tasks passed. Do not
  proceed to 3.2 until both hold.
- Task 3.2: Verify
  `.context/known-issues/plan-review-execution-location-blind-spot-2026-07-06.md`
  exists and is `status: ACTIVE`; then flip it to `status: DONE` with a one-line
  resolution note linking this plan. If it is not `ACTIVE`, stop and reconcile.
- Task 3.3: Regenerate the context index via
  `.context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh`
  and confirm no stderr warnings and that the known-issue now shows `DONE` in
  `.context/index.yaml`.

## Risks

- **Prompt dilution** — piling reach-analysis onto the Risk lens could crowd out
  its existing five questions. Mitigation: fold into the existing BLIND SPOTS
  item (Decision 2) rather than adding a sixth question.
- **Eval fixture is a self-fulfilling prophecy** — a scenario written to match
  the exact wording of the new instruction proves nothing about real behaviour.
  Mitigation: phrase the fixture in terms the reviewer must *reason about* (a
  plausible narrow-execution plan), not in the instruction's own vocabulary.
- **Sync drift** — editing the source but forgetting `tessl install` leaves the
  `.tessl` copy stale, and vice versa. Mitigation: Task 2.1 diffs the two copies
  as its exit check.

## Verification

- `bash .context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/plan-review-execution-location-lens-2026-07-07.md`
  passes.
- Both edited `prompt` strings parse as JSON (Task 1.2 gate).
- `git diff` on both `SKILL.md` copies shows only the intended prompt lines
  changed and the two copies are byte-identical on those lines.
- `tessl eval list` shows `scenario-04`, and `tessl eval run` scores its expected
  coverage/bypass finding above threshold (Task 2.3).
- Scenarios 01-03 re-run with unchanged scores (Task 2.4 regression check).
- The known-issue appears as `DONE` in a regenerated `.context/index.yaml` with
  no stderr warnings.
- `hk check` passes, green relative to the Phase 0 baseline.

## Open Questions

None outstanding. The two prior open questions were resolved during `plan-review`:
eval-scenario discovery is by directory convention (confirmed by inspecting
`evals/instructions.json` — see Scope), and the Strategic cross-reference is a
one-clause aside (Decision 4).
