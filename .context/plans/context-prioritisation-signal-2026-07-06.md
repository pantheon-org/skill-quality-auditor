---
title: "Plan: Add a prioritisation signal to .context/ so \"highest value\" is queryable"
type: plan
status: draft
date: 2026-07-06
effort: L
related:
  - ../findings/prioritisation-signal-gap-2026-07-06.md
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh
  - ../instructions/ways-of-working.md
---

**Effort:** L. No deep engineering, but it touches the frontmatter contract (schema plus validator), the index generator, a live backfill across active/draft files with a calibration pass, a full historical backfill of done/superseded entries as a learning corpus, two skills in two separate plugin bundles, an ADR, and a saved learnings finding, with a migration sequence that must be ordered correctly to avoid breaking the 100+ existing `.context/` files.

**Review status:** amended after a 3-reviewer `plan-review` (Sonnet Technical + Strategic, Haiku Risk) on 2026-07-06, then resolved through a `guided-interview` on the four contentious forks. All design forks are now settled (see Decisions 10-13). The plan is ready for promotion to `active` pending a final read.

**Interview outcomes (2026-07-06):** (Q2) `value` is an authoritative sort key, not an advisory label; (Q3) enum `high`/`medium`/`low` now, numeric only if ties prove blocking; (Q1) the generator exposes fields only and the consumer sorts, no derived hint; (Q4) `known-issue` carries `value` too, so all three action types sort on one axis. A post-interview addendum added a full historical backfill of `done`/`superseded` entries, graded by a lesser model as a learning corpus whose findings are saved to inform a future "what's next" skill (Decision 14, Phase 5).

## Goal

Make "which item is the highest value to do next?" answerable by reading `.context/index.yaml`, instead of re-derived from out-of-band agent judgement every time. The end state: plans and findings carry a stored `value` signal in frontmatter, graded against a published rubric; the index generator surfaces it; existing active and draft entries are backfilled consistently; and the frontmatter contract change is recorded in an ADR. See `.context/findings/prioritisation-signal-gap-2026-07-06.md` for the gap this plan closes.

## Scope

**In scope:**

- A new top-level `value` frontmatter field, an enum `high` / `medium` / `low` (Decision 12), on `type: plan`, `type: finding`, and `type: known-issue` (Decision 13).
- A published **value rubric** (what makes something `high` vs `medium` vs `low`), authored in Phase 1 so backfill grades against it, not retrofitted later.
- A documented **read protocol** (Decision 10): consumers sort by `value` descending, then `effort` ascending where present, and act on the top item without re-judging.
- Schema update in `context-frontmatter.schema.json` (currently `additionalProperties: false`, so the field must be declared explicitly or every file fails validation).
- `validate-context-frontmatter.sh` update, including its hardcoded per-type conditional block, sequenced so the field is accepted first and required only after backfill.
- `regenerate-context-index.sh` update to emit `value` into each plan, finding, and known-issue entry (fields only, no derived hint - Decision 11).
- A live backfill of `value` on all active and draft plans, findings, and known-issues, by a stated method with a calibration pass.
- A full historical backfill of `value` on all `done` and `superseded` entries, graded by a lesser model as a learning corpus (Decision 14, Phase 5).
- A saved learnings finding capturing what the historical corpus reveals about the rubric and about prioritisation patterns, explicitly intended to inform a future "what's next" skill.
- Updates to `plan-create` and `context-file` skills so new files prompt for `value`, applied early (Phase 2) so content authored mid-migration is not missed.
- A re-grading trigger tied to status transitions, documented in `ways-of-working.md`.
- An ADR recording the contract change, the rubric, and the read protocol.

**Out of scope (deferred):**

- A persisted ROI field (`value` divided by `effort`) or any generator-computed priority hint. The generator exposes fields only; the consumer sorts (Decision 11).
- A numeric value scale. Enum ships first; numeric is revisited only if bucket-ties prove to make the sort useless (Decision 12).
- Collapsing `effort`, `severity`, and `value` into one unified cross-type axis (rejected, Decision 2). Note `value` itself does become a shared axis across the three action types; the point rejected here is merging the three *different* axes into one number.
- Making `value` a *required* field on `done` or `superseded` entries. They are backfilled once as a learning corpus (Decision 14) but are exempt from the Phase 4 required-check; the operational sort only ever reads active/draft (Decision 7).
- Adding `value` to `analysis`, `instruction`, or `audit` types. These are reference material, not action candidates.
- Building the "what's next" skill itself. It is a separate follow-on; this plan only produces the signal it will consume and the learnings finding (Phase 5) that will inform its design.

## Decisions

All decisions are settled. Decisions 10-13 record the outcomes of the guided-interview on the forks the plan-review surfaced.

1. **Add a top-level `value` field on action-candidate types (Option A), not a derived ROI field (Option B) and not a merged single-number axis (Option C).** ROI presupposes a value input anyway, so B depends on A. A is the minimal change that makes "highest value" queryable and reuses the existing enum pattern used by `effort` and `severity`. Scope of "action-candidate types" is set by Decision 13.

2. **Keep `effort`, `severity`, and `value` as three distinct axes; do not unify (reject Option C).** They measure different things: `severity` is risk-of-inaction, `effort` is cost-of-action, `value` is benefit-of-action. Cross-type comparability is met by making all three visible in the index, not by collapsing to one number.

3. **The value rubric is authored in Phase 1, before any backfill.** The review found the original sequencing inverted (rubric in the Phase 4 ADR, used by the Phase 3 backfill). Grading 30-plus files against an undefined standard guarantees inconsistency and a second pass. The rubric leads.

4. **Migrate field-optional-first: schema accepts `value` and the validator does not require it (Phase 1), backfill (Phase 3), then flip the validator to required (Phase 4).** Requiring `value` before backfill would fail-validate every existing file at once. Optional-first keeps the tree green throughout.

5. **`value` scoped to the action-candidate types.** Per Decision 13 these are `plan`, `finding`, and `known-issue`. `analysis` / `instruction` / `audit` are reference material, not action items, and are excluded.

6. **Backfill reads existing signals but frontmatter is the single source of truth.** Where an active or draft remediation plan carries a body-level `priority: Critical/High/Medium/Low`, it *informs* the backfilled top-level `value` but does not override a deliberate grade. Mapping rule: body `Critical` and `High` to `value: high`, `Medium` to `medium`, `Low` to `low`; the remediation-plan bodies themselves are left untouched. If a body priority and a hand-set frontmatter `value` disagree, frontmatter wins.

7. **The operational sort only ever reads active/draft; `done` and `superseded` stay exempt from the Phase 4 required-check.** New work carries `value` from creation (Phase 4 makes it required for draft/active), so items acquire `value` before they reach `done`. Historical done/superseded entries that predate the field are backfilled once as a learning corpus (Decision 14), not because the live sort needs them.

8. **Backfill method: one serialised pass, one PR, one index regeneration.** `index.yaml` is git-tracked and this repo has already hit merge conflicts from parallel branches regenerating it (PRs #200, #201). The backfill is a single branch that touches the files and regenerates the index once. `value` is assigned by an author pass grounded in the Phase 1 rubric, followed by a calibration second-look (see Phase 3) rather than a single unchecked self-grade.

9. **Re-grade `value` on status transitions.** Unlike `date`, `value` can go stale as context changes. The re-grading trigger (revisit `value` when a plan moves draft to active, or when scope materially changes) is documented alongside the existing active-to-done sync rule in `ways-of-working.md`.

10. **`value` is an authoritative sort key, not an advisory label (resolves Q2).** The read protocol: rank candidates by `value` descending, then `effort` ascending where present, and act on the top item without re-forming an independent judgement. `effort` is plan-only, so findings and known-issues (which have no `effort`) sort by `value` alone within a bucket. This is what makes the field close the finding rather than relocate the judgement from query-time to authoring-time - but it only holds if the grades are trustworthy, which is why the rubric-first sequencing (Decision 3), the calibration pass (Decision 8), and the re-grade trigger (Decision 9) are load-bearing, not optional. The read protocol is documented in `ways-of-working.md` and the ADR.

11. **The index generator exposes fields only; it computes no derived hint (resolves Q1).** `regenerate-context-index.sh` emits `value` and `effort` as separate fields per entry; the consumer performs the value-then-effort sort at read time. No materialised hint, no composite rank key, no pre-sorted list. This keeps the index purely declarative, keeps Decision 2 clean (the generator never combines the axes), and sidesteps the review's objections that a hint had no specified formula and could only ever fire for plans.

12. **`value` is an enum `high` / `medium` / `low`, shipping now; numeric is deferred (resolves Q3).** The enum matches the `severity` vocabulary and keeps grading judgemental rather than falsely precise. Bucket-ties are broken by `effort` per Decision 10. Revisit a numeric scale only if, in practice, within-bucket ties are frequent enough that the sort stops being useful; that is a future decision with its own evidence, not a Phase 1 concern.

13. **`known-issue` carries `value` too, alongside `severity` (resolves Q4).** All three action-candidate types (plan, finding, known-issue) share the `value` axis, so "what's next" ranks a critical bug against a plan on one scale rather than comparing `severity` against `value` across a type silo, which was the finding's original complaint. A known-issue therefore carries both `severity` (its required risk-of-inaction axis) and `value` (its benefit-of-action axis); they are distinct per Decision 2 and the redundancy risk is accepted as the cost of a genuinely unified sort.

14. **Backfill the full history as a learning corpus, graded by a lesser model, and save the learnings (post-interview addendum).** Every `done` and `superseded` entry is also graded, using a cheaper model since these grades never feed the live sort and precision matters less than coverage. The exercise has two payoffs: it validates and refines the rubric against known outcomes, and it produces a labelled corpus of "what did we consider valuable, and what did we actually complete." The learnings are written up as a finding (Phase 5) explicitly intended to inform the design of a future "what's next" skill, the natural consumer of this signal. Hindsight bias (grading a shipped item `high` because it shipped) is expected and is itself a thing to observe and note, not a reason to skip the exercise. These historical grades are calibration and training signal only; they never enter the authoritative live sort (Decision 10), which reads active/draft.

## Phases

### Phase 1 - Rubric plus contract (schema and validator, field optional)

- Draft the **value rubric**: concrete criteria distinguishing `high` / `medium` / `low` (for example leverage across future work, number of consumers unblocked, reversibility), with two or three worked examples from existing `.context/` items. This is the artefact backfill grades against. Include the read protocol (Decision 10) so the rubric and the way the grade is consumed are defined together.
- Add a `value` property to `context-frontmatter.schema.json`: enum `[high, medium, low]` (Decision 12), applicable to `plan`, `finding`, and `known-issue` (Decision 13), with a description mirroring the `effort` / `severity` entries.
- Update the hardcoded per-type conditional block in `validate-context-frontmatter.sh` to *accept* `value` without yet *requiring* it (the review flagged that simply adding to the schema is insufficient because the validator has field-specific logic).
- Add fixtures carrying `value` and carrying none, and confirm both validate; include a case-sensitivity fixture (`value: HIGH` must be rejected).
- Exit criterion: a file with `value: high` passes validation, a file with no `value` still passes, an uppercase value is rejected, and the rubric is committed.

### Phase 2 - Surface it, and update authoring skills early

- Update `regenerate-context-index.sh` to emit `value` on each `plan`, `finding`, and `known-issue` entry that carries it (absent, not empty, on those that do not). Fields only, no derived hint or composite key (Decision 11).
- Update the `plan-create` and `context-file` skills to prompt for and emit `value`, and run `tessl install` in the same change so the `.tessl/` mirrors do not drift. These two skills live in two separate plugin bundles (`pantheon-org/context-mgmt` and `pantheon-org/planning`); both need syncing.
- Verify idempotency using the generator's existing `--check` mode (not a manual double-run diff), against files that carry only `value`, only `effort`, and both.
- Exit criterion: `.context/index.yaml` shows `value` on the source finding and this plan; `regenerate-context-index.sh --check` passes; both mirrors are drift-clean.

### Phase 3 - Backfill (single serialised pass)

- Enumerate the target set at execution time with a query (do not rely on baked-in counts; the population churns daily). Target = every `plan`, `finding`, and `known-issue` with `status` in {draft, active}.
- Assign `value` to each target file per the Phase 1 rubric, folding in any remediation-plan body priority per Decision 6. Known-issues get `value` in addition to their existing `severity` (Decision 13); grading the two axes independently is expected, not redundant.
- Run a calibration second-look across the assigned grades in one place (all grades visible together) to catch drift and self-grading inflation before committing, rather than trusting each file's grade in isolation. Because `value` is an authoritative sort key (Decision 10), this pass is load-bearing, not cosmetic.
- Verify `effort` coverage at execution time and resolve any genuine gap; note that a prior same-day rollout may already have completed this, so treat it as verify-not-assume rather than assumed work.
- Regenerate the index once, in this same branch.
- Exit criterion: a scripted check reports zero active or draft `plan`/`finding`/`known-issue` files missing `value`; `effort` gaps on active/draft plans are closed or explicitly recorded; the index regenerates clean.

### Phase 4 - Enforce and record

- Flip `validate-context-frontmatter.sh` to require `value` for `type: plan`, `type: finding`, and `type: known-issue` while `status` is `draft` or `active`, mirroring how `effort` is required for plans. `done`/`superseded` remain exempt (Decision 7).
- Add the re-grading trigger (Decision 9) and the read protocol (Decision 10) to `ways-of-working.md`, and document the new field in the CLAUDE.md context-index section so authors (not just ADR readers) see it.
- Write an ADR (via `adr-capture`) recording the contract change, the three-axes decision, the rubric, and the read protocol. Note in the ADR why `value` is ADR'd when the earlier `effort` field was not (we are correcting practice going forward).
- Run the full pre-push gate including the `.tessl` mirror-drift check.
- Exit criterion: a newly created plan without `value` fails validation with a clear message; the ADR is indexed in `docs/ADR/index.yaml`; the full gate is green.

### Phase 5 - Historical calibration corpus (learning exercise)

- Enumerate all `done` and `superseded` entries at execution time (all types that will carry `value`: plan, finding, known-issue).
- Grade `value` on each using a lesser model (a cheaper tier is deliberate; these grades never feed the live sort, so coverage matters more than precision). Batch the grading so the model sees the rubric and one entry at a time.
- Regenerate the index once so the historical grades are visible, in a change kept separate from the operational Phase 3 backfill so the two are auditable independently.
- Write a **learnings finding** capturing: how well the rubric held against known outcomes, where hindsight bias showed up, any mismatch between what was graded high and what was actually completed, and any prioritisation patterns worth encoding. State explicitly that the finding is an input to the design of the future "what's next" skill.
- Exit criterion: every done/superseded entry (of the three graded types) carries a `value`; the learnings finding is written, validated, and indexed; the index regenerates clean.

## Risks

- **Big-bang validation breakage.** Requiring `value` before backfill fails every existing file. Mitigated by the optional-first sequence (Decision 4) and by enforcement (Phase 4) landing only after coverage (Phase 3).
- **Validator has field-specific logic, not just schema-driven checks.** Adding `value` to the schema alone would not make the validator accept or later require it. Mitigated by Phase 1 explicitly editing the conditional block and testing both accept-optional and (Phase 4) require paths.
- **Self-graded, possibly stale value erodes trust.** Two authors grade differently; an author inflates their own plan; a grade goes stale. Mitigated by the rubric-first sequencing (Decision 3), the calibration pass (Decision 8), and the re-grade trigger (Decision 9). The residual question of how much to trust the field is Q2.
- **Derived hint scope creep or plan-only blind spot.** If the hint survives Q1, it fires only for plans and risks being read as authoritative. Mitigated by resolving Q1 explicitly and, if kept, documenting it as a plan-only convenience with a specified formula.
- **Concurrent index.yaml regeneration conflicts.** Already observed on PRs #200, #201. Mitigated by the single-serialised-pass backfill (Decision 8).
- **Mirror drift, including transient.** Editing source skills under `.context/plugins/` diverges from `.tessl/` until `tessl install` runs. The drift check will fail on a source-only commit, not just at the end. Mitigated by running `tessl install` in the same Phase 2 change as the source edits, and by the Phase 4 gate covering both plugin bundles.
- **Index generator regression.** Mitigated by the Phase 2 `--check` idempotency test and the generator's existing eval scenarios.
- **Historical grades leak into the live sort or are mistaken for ground truth.** Hindsight bias and the lesser model make the Phase 5 grades noisier. Mitigated by Decision 14 and Decision 7 keeping done/superseded out of the operational sort, by shipping Phase 5 as a change separate from the Phase 3 operational backfill, and by the learnings finding naming the bias rather than hiding it.
- **Lesser-model grading quality.** A cheap model may grade inconsistently. Accepted, because Phase 5 output is calibration and training signal, not a gate; the finding treats inconsistency as a datum about the rubric's clarity, not a failure.

## Verification

```bash
# Phase 1 - field accepted (not required), rubric committed, case-sensitivity enforced
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh \
  .context/findings/prioritisation-signal-gap-2026-07-06.md

# Phase 2 - index surfaces value; idempotency via the generator's own --check mode
bash .context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh --check
grep -A6 'prioritisation-signal-gap' .context/index.yaml | grep 'value:'

# Phase 3 - scripted completeness check (no baked-in counts)
#   expect zero active/draft plan|finding entries missing a value: line

# Phase 4 - enforcement bites on a new file, and the full gate passes
hk check && go test ./...
```

## Open Questions

The four interview forks (Q1 derived hint, Q2 operational trust, Q3 enum vs numeric, Q4 known-issue scope) are resolved in Decisions 10-13. Remaining questions:

- Backfill grading is one author plus a calibration pass. Now that `value` is an authoritative sort key (Decision 10) rather than advisory, is that enough governance, or does the initial backfill warrant a second human reviewer? (Leaning: the calibration pass plus the authoritative-sort stakes argue for at least a second look; revisit if the calibration pass proves insufficient.)
- Should the CLAUDE.md context-index documentation and the ADR rubric be the same text (single source) or is the ADR the canonical rubric with CLAUDE.md pointing at it? (Leaning: ADR canonical, CLAUDE.md links.)
- Numeric-scale revisit trigger (Decision 12): what concretely counts as "ties frequent enough to block the sort"? Left to the future decision, but worth a rough threshold so it is not litigated ad hoc.
