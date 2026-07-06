---
title: "Plan: Add a thematic axis to .context/ so work can be grouped, sliced, and tie-broken by area"
type: PLAN
status: DRAFT
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
a categorical fact about an entry, not a graded judgement. The one genuinely open piece
of design is ratifying the controlled vocabulary (see Open Questions), which is why this
plan ships as DRAFT pending that decision rather than going straight to ACTIVE.

**Review status:** DRAFT, not yet through `plan-review`. The vocabulary fork (Decision 1
options) should be settled by a `guided-interview` or a short `design-debate` before
promotion, the same way the `value` plan resolved its Q1-Q4 forks.

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

- A new thematic frontmatter field on the three action-candidate types (`PLAN`,
  `FINDING`, `KNOWN_ISSUE`), shape decided by Decision 1.
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
  theme as a documented tie-breaker below `value` then `effort`.
- An ADR recording the contract change and the vocabulary.

**Out of scope (deferred):**

- Backfilling `DONE`/`SUPERSEDED` entries. Unlike the `value` corpus, historical themes
  carry no learning signal worth the cost; theme is operational only.
- A generator-computed "theme with most open debt" rollup. The generator emits the field;
  any aggregation is a read-time query or a future skill concern.
- Adding the theme to `ANALYSIS`/`INSTRUCTION`/`AUDIT` types (reference material).
- Building the "what's next" skill that would consume theme as a grouping key.
- Free-form uncontrolled tags. If Decision 1 lands on tags, they are still drawn from the
  ratified vocabulary, not open text, so the axis stays queryable.

## Decisions

Decision 1 is deliberately left open for a pre-promotion interview; the rest are settled
by direct analogy to the `value` signal (ADR-049).

1. **Vocabulary shape — OPEN, resolve before promotion.** Three candidates:
   (A) single `theme` enum, one theme per entry, small ratified vocabulary;
   (B) multi-valued `themes`/`tags` list drawn from the same ratified vocabulary;
   (C) derive grouping from the `related` graph, no new field. Leaning A: it matches the
   `value`/`severity` enum pattern, keeps sort/tie-break semantics trivial, and forces a
   single primary area per entry (which is what a tie-breaker needs). B is more faithful
   to genuinely cross-cutting items but muddies the tie-break; C avoids a schema change
   but depends on link discipline the repo has already seen drift on.
2. **Seed the vocabulary from the finding's latent clusters, then ratify.** Candidate
   values: `EVAL`, `PR-TOOLING`, `DOCS`, `GOVERNANCE`, `SKILL-QUALITY`, `DISTRIBUTION`.
   The set is ratified in Phase 1 (with the interview) rather than accreted ad hoc, and
   kept deliberately small; a too-fine vocabulary is as useless as none.
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

## Phases

### Phase 1 - Vocabulary plus contract (schema and validator, field optional)

- Ratify the theme vocabulary (Decision 2) via a `guided-interview` or `design-debate`
  that also settles Decision 1's shape. Publish it as an instruction file, the artefact
  backfill grades against.
- Add the theme property to `context-frontmatter.schema.json`: the ratified UPPER_CASE
  enum, applicable to `PLAN`/`FINDING`/`KNOWN_ISSUE`, description mirroring `value`.
- Update the per-type conditional block in `validate-context-frontmatter.sh` to accept
  the field without yet requiring it.
- Add fixtures: an entry with a valid theme, one with none (both pass), and a lowercase
  value (rejected, per Decision 3).
- Exit criterion: a file with a valid theme passes, a file with none passes, a lowercase
  theme is rejected, the vocabulary instruction is committed.

### Phase 2 - Surface it, and update authoring skills early

- Update `regenerate-context-index.sh` to emit the theme on each `PLAN`/`FINDING`/
  `KNOWN_ISSUE` entry that carries it.
- Update `plan-create` and `context-file` to prompt for and emit the theme, and run
  `tessl install` in the same change so the `.tessl/` mirrors (two bundles:
  `context-mgmt` and `planning`) do not drift.
- Verify idempotency with the generator's `--check` mode.
- Exit criterion: the index shows the theme on the source finding and this plan;
  `--check` passes; both mirrors are drift-clean.

### Phase 3 - Backfill (single serialised pass)

- Enumerate the target set at execution time: every `PLAN`/`FINDING`/`KNOWN_ISSUE` with
  `status` in {`DRAFT`, `ACTIVE`}. Do not rely on baked-in counts.
- Assign a theme to each from the ratified vocabulary. Where an entry genuinely spans
  areas, apply the Decision 1 rule (single primary theme under A; the list under B).
- Regenerate the index once, in this branch.
- Exit criterion: a scripted check reports zero active/draft action-candidates missing a
  theme; the index regenerates clean.

### Phase 4 - Enforce and record

- Flip `validate-context-frontmatter.sh` to require the theme for `PLAN`/`FINDING`/
  `KNOWN_ISSUE` while `status` is `DRAFT` or `ACTIVE`. `DONE`/`SUPERSEDED` exempt.
- Document the theme field and its tie-breaker role in `ways-of-working.md`, the value
  rubric's read protocol (theme as the tie-breaker below `value` then `effort`), and the
  CLAUDE.md context-index section.
- Write an ADR (via `adr-capture`) recording the contract change and the vocabulary.
- Run the full pre-push gate including the `.tessl` mirror-drift check.
- Exit criterion: a new action-candidate without a theme fails validation with a clear
  message; the ADR is indexed; the full gate is green.

## Risks

- **Vocabulary churn.** A too-fine or contested vocabulary erodes the axis's value.
  Mitigated by ratifying a small set up front (Decision 2) and treating additions as an
  ADR amendment, not an ad-hoc edit.
- **Cross-cutting items.** Some entries genuinely belong to two themes; forcing one
  primary (Decision 1 A) loses information. Accepted as the cost of a clean tie-break;
  revisit as multi-value (B) if single-theme proves lossy in practice.
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

# Phase 2 - index surfaces the theme; idempotency via --check
bash .context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh --check
grep -A7 'context-taxonomy-gap' .context/index.yaml | grep 'theme:'

# Phase 3 - scripted completeness check (no baked-in counts)
#   expect zero active/draft plan|finding|known-issue entries missing a theme: line

# Phase 4 - enforcement bites on a new file, and the full gate passes
hk check && go test ./...
```

## Open Questions

- **Decision 1 (vocabulary shape): single enum vs multi-value tags vs derived-from-related.**
  This is the one load-bearing fork and must be resolved before promotion. Leaning single
  enum (A) for tie-break cleanliness.
- **Granularity of the vocabulary.** Six seed themes may be too coarse (governance
  currently swallows eight distinct entries) or, if split finer, too many. What is the
  right number, and what is the rule for adding one later?
- **Does theme belong below or above `effort` in the read protocol tie-break?** Leaning
  below (theme breaks ties only after `value` then `effort`), since theme expresses
  preference-of-area, not priority. Confirm in the interview.
