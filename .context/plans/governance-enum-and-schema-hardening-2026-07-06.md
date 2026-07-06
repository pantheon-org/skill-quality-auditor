---
title: "Plan: Deferred — effort full-word enum migration + real JSON-schema validator (G6, G5)"
type: PLAN
status: DRAFT
date: 2026-07-06
effort: L
value: LOW
themes:
  - GOVERNANCE
  - SKILL-QUALITY
related:
  - ../known-issues/effort-single-letter-values-2026-07-06.md
  - ../known-issues/schema-not-enforced-by-real-validator-2026-07-06.md
  - governance-tooling-hardening-2026-07-06.md
  - ../../docs/ADR/adr-052-governance-known-issue-triage.md
  - ../../docs/ADR/adr-050-uppercase-frontmatter-enums.md
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
---

**Effort:** L. Two `value: LOW` gaps with real blast radius: an enum migration touching schemas, validator, Go code, ~15 live values, docs and eval fixtures (needs an ADR), followed by a first real JSON-schema validation step that must not break generated remediation plans. Deliberately deferred, not urgent.

**Review status:** DRAFT, split out of the governance-hardening review (2026-07-06) because both gaps are LOW-value and were inflating an otherwise-MEDIUM plan and gating a valuable fix. Not yet reviewed on its own. Should get its own `plan-review` before promotion — several load-bearing decisions are still open (effort vocabulary, ADR form, schema-validator runner, remediation-plan schema scope).

## Goal

Bring the two remaining LOW-value GOVERNANCE consistency/enforcement gaps to closure when they rise to the top of the queue: make `effort` a full-word UPPER_CASE enum consistent with ADR-050, and add a real JSON-schema validation step so `additionalProperties: false` and typo'd frontmatter keys are actually caught rather than silently passing the partial shell validator. See the two `known-issues/` entries in `related`.

## Scope

**In scope:**

- **G6 — effort full-word enum.** Migrate `effort` from `S/M/L/TBD` to full UPPER_CASE words across: `context-frontmatter.schema.json` and `plan-scaffold.schema.json` (both `additionalProperties: false`), `validate-context-frontmatter.sh`, `gapEffort()` in `reporter/remediation_plan_generate.go`, the plan-scaffold template and `plan-create`/`context-file` skill docs, every existing `effort:` value across `.context/`, docs (`remediation-flow.md`, `aggregation-flow.md`), and the eval fixtures under `cmd/assets/evals/**` — including `cmd/assets/evals/instructions.json`, which instructs agents to produce "S/M/L" sizing in prose (flagged by the Technical reviewer as a false-green blind spot). Recorded as an ADR or ADR-050 amendment.
- **G5 — real schema enforcement.** Add a genuine JSON-schema validation step (runner per Open Question) to the gate so `context-frontmatter.schema.json`'s `additionalProperties: false` and typo'd keys are enforced against `.context/**/*.md`. Gated on G6 (enum must be final) and on a remediation-plan schema-scope decision (see Decisions).

**Out of scope:**

- The three quick gaps (adr-index freshness, undocumented-decisions false-negative, remediation-plan `value`) — done in `governance-tooling-hardening-2026-07-06.md`.
- `plan-scaffold.schema.json`'s existing `additionalProperties: false` is unchanged; it governs only the plan-create authoring scaffold artefact, not `.context/*.md` files or generated remediation plans, so it is not the surface G5 changes.

## Decisions

1. **G6 lands before G5.** The `effort` enum must be finalised before a real schema validator locks it in; both schemas carry `additionalProperties: false`, and G6 already forces the schema edits G5 would then police.
2. **G5 must not enable `additionalProperties: false` enforcement against generated remediation plans until they have a home.** A strict validator over `context-frontmatter.schema.json` would reject generated remediation plans, which emit extra keys (`skill_name`, `source_audit`, `executive_summary`, …) not in that schema. Resolve the schema-scope Open Question (dedicated `remediation-plan.schema.json` vs a path-glob exemption) before enabling. Note: the sibling plan's G4 adds the required `value` field, so by the time this plan runs, the *missing-field* half is already handled; only the *extra-keys* half remains.
3. **G6 migration is one serialised pass with a content-grep completeness check**, not a frontmatter-line grep. Verification greps for the string `S/M/L` and bare `effort:` values across `.context`, `docs`, and `cmd/assets/evals` — the line-anchored `effort: S` pattern misses prose occurrences (Technical reviewer finding). Land as a single PR so a failed completeness check blocks merge rather than leaving a half-migrated enum on `main`.

## Phases

### Wave A — effort full-word enum (G6) — needs an ADR first

- Resolve the vocabulary Open Question, then migrate in one serialised pass: both schemas, `validate-context-frontmatter.sh`, `gapEffort()`, plan-scaffold template, `plan-create`/`context-file` docs, all existing `effort:` values, docs, and `cmd/assets/evals/**` (incl. `instructions.json`). Regenerate the index once. Run `tessl install` for the two mirrored schema/script bundles.
- Write the ADR (or ADR-050 amendment) recording the vocabulary and the migration.
- Exit criterion: `grep -rn 'S/M/L' .context docs cmd/assets/evals` and a bare-`effort:`-value scan both return zero legacy values; both schemas validate; `go test ./...` green; index fresh; ADR indexed; mirror-drift clean.

### Wave B — real JSON-schema validator (G5) — last, gated by A + scope decision

- Resolve the remediation-plan schema-scope Open Question (dedicated schema vs path-glob exemption). If exemption, specify the detection rule (e.g. path glob `.context/plans/*-remediation-plan-*.md`, or a marker field) so the validator can identify a generated plan.
- Add the JSON-schema validation step (runner per Open Question) wired into `hk.pkl`/CI *alongside* — not replacing — the existing shell checks during a proving period.
- Exit criterion: an intentionally malformed fixture (typo'd key `valeu: HIGH`, unknown property) fails the schema step; every real `.context/` file — including generated remediation plans (post-G4) — passes; full gate green.

## Risks

- **G6 blast radius / missed site.** Many surfaces; a missed one fails validation. Mitigated by Decision 3's content-grep completeness check and single-PR landing.
- **Enum collision (readability, not correctness).** Full-word `effort` makes `MEDIUM` a legal value for `effort`, `value`, and `severity`. The validator's per-field enum checks keep these correct, but `effort: MEDIUM` next to `value: MEDIUM` is more ambiguous to a human reader. Weigh in the vocabulary Open Question (e.g. `SHORT/MEDIUM/LONG` or keeping distinct words) — this is a real reason the migration may not be worth it.
- **G5 rejects remediation plans (extra-keys half).** The central hazard; mitigated by Decision 2 and by ordering after the sibling plan's G4.
- **Docs-drift gate.** Schema/Go/doc changes will trip the docs-drift PR gate (observed on #208); update mapped docs in the same PR.
- **Mirror drift** on edited plugin schemas/scripts — run `tessl install` in the same change.

## Verification

```bash
# Wave A — no legacy effort abbreviations remain anywhere (content grep, not line-anchored)
! grep -rn 'S/M/L' .context docs cmd/assets/evals 2>/dev/null
go test ./...

# Wave B — schema step catches a bad key; all real files (incl. generated plans) pass
hk check && go test ./...
```

## Open Questions

- **G6 worth it at all?** This is `value: LOW` with real blast radius and a human-readability collision (`effort: MEDIUM` vs `value: MEDIUM`). Confirm it is worth doing versus (a) dropping it, or (b) doing only the schema/validator half and mapping `gapEffort()`'s output. (Leaning: revisit when this plan reaches the top of the queue; do not do it just for tidiness.)
- **G6 vocabulary + ADR form:** `SMALL/MEDIUM/LARGE` (collides with `value`/`severity` MEDIUM) vs `SHORT/MEDIUM/LONG` vs another set; keep `TBD` or rename to `UNKNOWN`? New ADR vs ADR-050 amendment? (Leaning: amend ADR-050, since `effort` is the enum it should have covered; pick a vocabulary that minimises the MEDIUM collision.)
- **G5 remediation-plan schema scope:** dedicated `remediation-plan.schema.json` vs an exemption list keyed on a path glob? (Leaning: dedicated schema — remediation plans have a genuinely different shape.)
- **G5 runner:** Python `jsonschema` (python3 already in the gate) vs Go `santhosh-tekuri/jsonschema` (no new CI runtime dep, testable in `go test`)? (Leaning: Go, to keep enforcement in the same test suite as the code.)
