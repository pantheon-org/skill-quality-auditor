---
title: "Plan: Add a thematic axis to .context/ so work can be grouped, sliced, and tie-broken by area"
type: PLAN
status: ACTIVE
date: 2026-07-06
effort: M
value: MEDIUM
related:
  - ../findings/context-taxonomy-gap-2026-07-06.md
  - ../instructions/value-rubric.md
  - ../plans/context-prioritisation-signal-2026-07-06.md
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh
  - ../instructions/ways-of-working.md
---

**Effort:** M. No new engineering shape — this reuses the exact migration groove the
`value` signal cut (schema, validator, index generator, two authoring skills in two
plugin bundles, a single serialised backfill, an ADR). It is smaller than `value` (L)
because there is no historical-corpus phase and no calibration-pass subtlety: a theme is
a categorical fact about an entry, not a graded judgement. The multi-valued shape adds a
small authoring discipline (order tags by primary) but no engineering complexity.

**Review status:** ACTIVE. The three design forks were resolved by a `guided-interview`
on 2026-07-06 (see Decisions 1, 8, 9). Not yet through a full `plan-review`; the forks
that would have blocked promotion are closed, so it is ready for execution.

**Interview outcomes (2026-07-06):** (Q1) the axis is a **multi-valued `themes` list**,
not a single enum — an entry can belong to several areas (Decision 1). (Q2) the list is
**ordered and `themes[0]` is the primary theme**, which breaks read-protocol ties below
`value` then `effort` (Decision 8, resolving the old below-or-above-`effort` fork). (Q3)
**ship the coarse 6 seed themes and split on evidence** — split a theme only once backfill
shows it dominating (rough threshold ~30% of entries), recorded in the ADR, mirroring how
the `value` plan deferred a numeric scale (Decision 9).

## Goal

Make "which area is this about?" a stored, queryable fact so the backlog can be sliced
by theme, interrelated clusters can be closed as batches, and same-`value`/`effort` ties
can be broken by area instead of arbitrarily. The end state: action-candidate entries
carry a thematic signal in frontmatter drawn from a ratified vocabulary; the index
surfaces it; active/draft entries are backfilled; and the contract change is recorded in
an ADR. See `.context/findings/context-taxonomy-gap-2026-07-06.md` for the gap this
closes.

## Scope

**In scope:**

- A new multi-valued `themes` frontmatter field (an ordered list, Decision 1) on the
  three action-candidate types (`PLAN`, `FINDING`, `KNOWN_ISSUE`).
- A **ratified controlled vocabulary** of themes, seeded from the latent clusters the
  finding identified, published as an instruction file so backfill grades against it.
- Schema update in `context-frontmatter.schema.json` (`additionalProperties: false`, so
  the field must be declared or every file fails validation).
- `validate-context-frontmatter.sh` update, optional-first then required, mirroring the
  `value` sequencing.
- `regenerate-context-index.sh` update to emit the theme field per entry.
- A single serialised backfill of active/draft plans, findings, and known-issues.
- `plan-create` and `context-file` skill updates to prompt for the theme, plus
  `tessl install` to keep both plugin-bundle mirrors in sync.
- Read-protocol update in `ways-of-working.md` and the value rubric / read protocol:
  `themes[0]` (the primary theme) as a documented tie-breaker below `value` then
  `effort` (Decision 8).
- An ADR recording the contract change and the vocabulary.

**Out of scope (deferred):**

- Backfilling `DONE`/`SUPERSEDED` entries. Unlike the `value` corpus, historical themes
  carry no learning signal worth the cost; theme is operational only.
- A generator-computed "theme with most open debt" rollup. The generator emits the field;
  any aggregation is a read-time query or a future skill concern.
- Adding the theme to `ANALYSIS`/`INSTRUCTION`/`AUDIT` types (reference material).
- Building the "what's next" skill that would consume theme as a grouping key.
- Free-form uncontrolled tags. The `themes` list is multi-valued (Decision 1) but every
  member is drawn from the ratified vocabulary, not open text, so the axis stays queryable.

## Decisions

All decisions are settled. Decisions 1, 8, and 9 record the outcomes of the
2026-07-06 guided-interview on the forks this plan was left DRAFT to resolve.

1. **Vocabulary shape — multi-valued `themes` list (Option B), not a single enum (A) or
   derive-from-`related` (C) (resolves the interview's Q1).** An entry can genuinely
   belong to several areas (e.g. `cross-reference-drift-audit` is both `GOVERNANCE` and
   tooling), and a single enum would force a lossy choice. The list draws every member
   from the ratified vocabulary, so it stays queryable; the tie-break complication that
   made B look costly is resolved by Decision 8 (ordered list, primary is `themes[0]`).
   C is rejected: it depends on `related`-link discipline the repo has already seen drift
   on, and it cannot express "this is about X" independently of "this links to Y".
2. **Ship the six seed themes drawn from the finding's latent clusters:** `EVAL`,
   `PR-TOOLING`, `DOCS`, `GOVERNANCE`, `SKILL-QUALITY`, `DISTRIBUTION`. The set is kept
   deliberately coarse; a too-fine vocabulary is as useless as none. Splitting a theme
   later is evidence-driven, not speculative (Decision 9).
3. **UPPER_CASE enum values (ADR-050).** Whatever the vocabulary, values follow the
   established frontmatter casing convention.
4. **Migrate field-optional-first.** Schema accepts the field and the validator does not
   require it (Phase 1), backfill (Phase 3), then flip to required for draft/active
   (Phase 4). Requiring before backfill would fail-validate every existing file.
5. **Scoped to action-candidate types only** (`PLAN`, `FINDING`, `KNOWN_ISSUE`), matching
   `value`. Reference types are excluded.
6. **Theme is a fact, not a graded judgement, so no calibration pass.** Unlike `value`,
   two authors should assign the same theme to the same entry; a light single pass with
   the ratified vocabulary in hand is sufficient.
7. **Single serialised backfill, one PR, one index regeneration**, to avoid the
   `index.yaml` merge conflicts already seen on PRs #200/#201.
8. **The `themes` list is ordered; `themes[0]` is the primary theme and is the sole
   tie-breaker, sitting below `value` then `effort` (resolves the interview's Q2 and the
   old below-or-above-`effort` fork).** The full sort is `value` descending, then `effort`
   ascending, then `themes[0]` as a final tie-break. Only the primary participates in the
   sort; the remaining tags are for filtering and cluster views, never for ordering.
   Theme sits *below* `effort` because it expresses preference-of-area, not priority.
   Authors order the list primary-first; this is the one authoring discipline the
   multi-valued shape adds, documented in the authoring skills and `ways-of-working.md`.
9. **Ship the coarse six and split a theme only on evidence (resolves the interview's
   Q3).** No theme is subdivided pre-emptively. Once the Phase 3 backfill is in, if a
   single theme carries a disproportionate share of entries (rough guide ~30%, with
   `GOVERNANCE` the likely first candidate) it is split, and the split is recorded as an
   ADR amendment to the vocabulary. This mirrors the `value` plan's deferral of a numeric
   scale until ties proved blocking: ship the simple thing, refine on observed need.

## Phases

### Phase 1 - Vocabulary plus contract (schema and validator, field optional)

- Publish the ratified six-theme vocabulary (Decision 2) as an instruction file, the
  artefact backfill grades against. The vocabulary and shape are already settled by the
  interview, so no further ratification step is needed.
- Add the `themes` property to `context-frontmatter.schema.json`: an array whose items
  are the ratified UPPER_CASE enum, applicable to `PLAN`/`FINDING`/`KNOWN_ISSUE`,
  description mirroring `value`. Constrain to a non-empty array of unique enum members.
- Update the per-type conditional block in `validate-context-frontmatter.sh` to accept
  the field without yet requiring it.
- Add fixtures: an entry with a valid `themes` list, one with none (both pass), a
  lowercase member (rejected, per Decision 3), and an unknown-theme member (rejected).
- Exit criterion: a file with a valid `themes` list passes, a file with none passes, a
  lowercase or unknown member is rejected, the vocabulary instruction is committed.

### Phase 2 - Surface it, and update authoring skills early

- Update `regenerate-context-index.sh` to emit the `themes` list (order preserved, so
  `themes[0]` stays the primary) on each `PLAN`/`FINDING`/`KNOWN_ISSUE` entry that
  carries it.
- Update `plan-create` and `context-file` to prompt for and emit `themes` primary-first
  (Decision 8), and run
  `tessl install` in the same change so the `.tessl/` mirrors (two bundles:
  `context-mgmt` and `planning`) do not drift.
- Verify idempotency with the generator's `--check` mode.
- Exit criterion: the index shows the `themes` list on the source finding and this plan;
  `--check` passes; both mirrors are drift-clean.

### Phase 3 - Backfill (single serialised pass)

- Enumerate the target set at execution time: every `PLAN`/`FINDING`/`KNOWN_ISSUE` with
  `status` in {`DRAFT`, `ACTIVE`}. Do not rely on baked-in counts.
- Assign a `themes` list to each from the ratified vocabulary, ordered primary-first
  (Decision 8). An entry that genuinely spans areas carries all its areas; the primary is
  the one that best answers "what is this mainly about?".
- Regenerate the index once, in this branch.
- Exit criterion: a scripted check reports zero active/draft action-candidates with an
  empty or missing `themes` list; the index regenerates clean.

### Phase 4 - Enforce and record

- Flip `validate-context-frontmatter.sh` to require a non-empty `themes` list for
  `PLAN`/`FINDING`/`KNOWN_ISSUE` while `status` is `DRAFT` or `ACTIVE`.
  `DONE`/`SUPERSEDED` exempt.
- Document the `themes` field and the `themes[0]` tie-breaker role in `ways-of-working.md`,
  the value rubric's read protocol (`themes[0]` as the tie-breaker below `value` then
  `effort`, per Decision 8), and the CLAUDE.md context-index section.
- Write an ADR (via `adr-capture`) recording the contract change, the multi-valued
  ordered-list shape, the six-theme vocabulary, and the evidence-driven split rule
  (Decision 9).
- Run the full pre-push gate including the `.tessl` mirror-drift check.
- Exit criterion: a new action-candidate with no `themes` fails validation with a clear
  message; the ADR is indexed; the full gate is green.

## Risks

- **Vocabulary churn.** A too-fine or contested vocabulary erodes the axis's value.
  Mitigated by ratifying a small set up front (Decision 2) and treating additions as an
  ADR amendment, not an ad-hoc edit.
- **Primary-tag drift.** Because `themes[0]` is load-bearing for the tie-break, an entry
  whose primary is set carelessly sorts wrongly within its bucket. Lower stakes than a
  `value` misgrade (it only affects same-`value`/`effort` ties), but mitigated by the
  authoring-skill prompt asking for the primary explicitly and by `ways-of-working.md`
  stating the primary-first convention.
- **Over-tagging.** A multi-valued list invites attaching every plausible theme, diluting
  the filter. Mitigated by keeping the vocabulary coarse (Decision 2) and by guidance to
  tag only genuine areas, not tangential links.
- **Big-bang validation breakage.** Mitigated by optional-first sequencing (Decision 4).
- **Validator field-specific logic.** Schema alone will not make the validator accept or
  require the field. Mitigated by Phase 1 editing the conditional block explicitly.
- **Concurrent index.yaml regeneration conflicts.** Mitigated by the single-serialised
  backfill (Decision 7).
- **Mirror drift.** Editing source skills diverges from `.tessl/` until `tessl install`.
  Mitigated by running it in the same Phase 2 change.

## Verification

```bash
# Phase 1 - field accepted (not required), vocabulary committed, casing enforced
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh \
  .context/findings/context-taxonomy-gap-2026-07-06.md

# Phase 2 - index surfaces the themes list; idempotency via --check
bash .context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh --check
grep -A7 'context-taxonomy-gap' .context/index.yaml | grep 'themes:'

# Phase 3 - scripted completeness check (no baked-in counts)
#   expect zero active/draft plan|finding|known-issue entries with an empty/missing themes list

# Phase 4 - enforcement bites on a new file, and the full gate passes
hk check && go test ./...
```

## Open Questions

The three forks this plan was left DRAFT to resolve (vocabulary shape, granularity/growth
rule, tie-break position) are settled in Decisions 1, 8, and 9. Remaining, non-blocking:

- **Split threshold precision.** Decision 9 uses a rough ~30% guide for when a theme is
  dominant enough to split. What counts precisely is left to the evidence at split time,
  but a firmer threshold would stop it being litigated ad hoc.
- **Primary-theme guidance.** Decision 8 makes `themes[0]` load-bearing but "what is this
  mainly about?" is still a judgement. Does the vocabulary instruction need worked
  examples of primary selection, or is the one-line rule enough? (Leaning: add two or
  three examples, as the value rubric did.)
- **Should the index emit a derived primary-theme convenience field** (e.g. `theme:` =
  `themes[0]`) for cheaper grepping, or is reading `themes[0]` at query time sufficient?
  (Leaning: no derived field, matching the `value` plan's Decision 11 that the generator
  emits fields only.)
