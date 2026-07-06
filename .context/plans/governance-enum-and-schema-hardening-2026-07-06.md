---
title: "Plan: Add a real JSON-schema validator for .context frontmatter (G5)"
type: PLAN
status: DRAFT
date: 2026-07-06
effort: M
value: MEDIUM
themes:
  - GOVERNANCE
  - SKILL-QUALITY
related:
  - ../known-issues/schema-not-enforced-by-real-validator-2026-07-06.md
  - ../known-issues/effort-single-letter-values-2026-07-06.md
  - ../../docs/ADR/adr-050-uppercase-frontmatter-enums.md
  - ../../docs/ADR/adr-052-governance-known-issue-triage.md
  - ../../.context/plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
  - ../../.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh
  - ../../cmd/validate.go
  - ../../reporter/remediation_plan_generate.go
---

> **History.** This plan originally bundled two gaps — an `effort` single-letter → full-word enum migration (G6) and a real JSON-schema validator (G5). A 3-reviewer `plan-review` on 2026-07-06 unanimously recommended **dropping G6**: it is `value: LOW`, contradicts ADR-050 (which recorded `effort` as deliberately unchanged), has a ~5× larger blast radius than first estimated (~73 sites across 28 files, including a silent split-brain in `reporter/aggregation.go`), and is net-negative — it would make `MEDIUM` a legal value for `effort`, `value`, and `severity` simultaneously, a collision that cannot occur while `effort` stays `S/M/L`. G6 is closed as won't-fix; the rationale is recorded as an amendment to ADR-050, and the `effort-single-letter-values` known-issue is marked `SUPERSEDED`. This plan now covers **only G5**.

**Effort:** M. One dedicated schema, a Go validator wired into the existing `validate` command and the hk/CI gate, a small fixture set, and a decision (already made) on how generated remediation plans are scoped. No enum migration, no repo-wide mechanical edit.

**Review status:** Reviewed (3-reviewer `plan-review`, 2026-07-06). All design forks resolved (see Decisions). Execution-ready; deferred only by queue position (`value: MEDIUM`, below the S-effort MEDIUM plans and the HIGH `migrate-off-tessl-eval`).

## Goal

Catch the class of frontmatter error the current shell validator cannot: a typo'd or unknown top-level key (`valeu: HIGH`, `stauts: DRAFT`) passes `validate-context-frontmatter.sh` silently today, because that script re-implements only a subset of the schema (required / enum / pattern / per-type rules) and never enforces the schema's `additionalProperties: false`. Add a real JSON-schema validation step so unknown keys are rejected, running **alongside** — not replacing — the existing shell validator during a proving period. See `.context/known-issues/schema-not-enforced-by-real-validator-2026-07-06.md`.

## Scope

**In scope:**

- A Go JSON-schema validator using `santhosh-tekuri/jsonschema` (Decision 3), wired into the existing `cmd/validate.go` cobra command as a new sub-check and into `hk.pkl` (built-binary gate) alongside the shell validator.
- Enforcement of `context-frontmatter.schema.json` (including `additionalProperties: false`) against hand-authored `.context/**/*.md`.
- A dedicated `remediation-plan.schema.json` for generated remediation plans (Decision 2), selected by the validator via a path glob so generated plans — which legitimately carry extra keys (`skill_name`, `source_audit`, `executive_summary`, …) — validate against their own contract.
- Fixtures proving: an unknown/typo'd key is rejected; a valid hand-authored file passes; a generated remediation plan passes against its dedicated schema.

**Out of scope:**

- **G6 (effort enum migration)** — dropped, see History and ADR-050 amendment.
- Replacing the shell validator. The shell validator keeps the per-type requiredness/applicability logic (a single shared schema with `additionalProperties: false` catches unknown keys but cannot model "FINDING must not carry `effort`" or "KNOWN_ISSUE must carry `severity`" — Decision 4). The two run together.
- Modelling per-type required fields in JSON Schema (would need six conditional sub-schemas; the shell validator already does this and stays authoritative for it).

## Decisions

1. **G5 stands alone; it never needed G6.** `additionalProperties: false` enforcement is identical whether `effort` is `L` or `LARGE`. The original G6-before-G5 gating reproduced the anti-pattern ADR-052 was written to stop, so it is removed.
2. **Generated remediation plans get a dedicated `remediation-plan.schema.json`**, selected by a path glob (`*-remediation-plan-*.md`), rather than a carve-out inside `context-frontmatter.schema.json`. Remediation plans have a genuinely different shape (a dozen extra structured keys); a separate contract is clearer than an exemption and lets that shape be validated properly rather than merely skipped. The generator already emits `value` + `themes` (G4), so the new schema requires them too.
3. **The runner is Go (`santhosh-tekuri/jsonschema`), not Python `jsonschema`.** The repo has no Python dependency-provisioning path in CI (the shell validator uses only stdlib), whereas the Go binary is already built in the pre-push/CI flow. A Go validator lives in the same `go test` suite as the code and adds no new CI runtime. It is wired as a `validate` sub-check that depends on the existing `go-build` step.
4. **Run alongside the shell validator, not as a replacement (proving period).** The JSON-schema step catches unknown keys; the shell validator keeps per-type requiredness. Both run in `hk`/CI. Revisit consolidation only after the JSON-schema step has proven stable — do not delete the shell validator in this plan.

## Phases

### Phase 1 — dedicated remediation-plan schema

- Author `remediation-plan.schema.json` (co-located with `context-frontmatter.schema.json`) capturing the generated-plan shape: the base frontmatter fields plus `plan_date`, `skill_name`, `source_audit`, `executive_summary`, `critical_issues`, `remediation_phases`, `verification_commands`, `success_criteria`, `effort_estimates`, `dependencies`, `rollback_plan`, `notes`. Require `value` and `themes` (G4 emits them).
- Validate the schema itself is well-formed and matches a freshly generated plan (round-trip a `RemediationPlan(...)` output against it in a Go test).
- Exit criterion: a generated remediation plan validates against `remediation-plan.schema.json`; a hand-authored plan does not (wrong shape).

### Phase 2 — Go validator + wiring

- Add a JSON-schema validation routine (Go, `santhosh-tekuri/jsonschema`) that, per `.context/**/*.md` file, picks `remediation-plan.schema.json` for `*-remediation-plan-*.md` paths and `context-frontmatter.schema.json` otherwise, and enforces it (including `additionalProperties: false`).
- Surface it as a `validate` sub-check in `cmd/validate.go`; add a `hk.pkl` step depending on `go-build`, running alongside the existing `context-frontmatter` shell check.
- Add fixtures/tests: unknown key rejected; valid hand-authored file passes; generated remediation plan passes; DONE files with legacy-but-valid enums still pass (enum unchanged — G6 dropped, so no enum churn here).
- Exit criterion: a fixture with `valeu: HIGH` fails the new step; every real `.context/**/*.md` file passes; `go test ./...` and the full `hk` gate are green.

## Risks

- **Single shared schema can't model per-type requiredness.** Accepted and by design (Decision 4): the JSON-schema step catches unknown keys; the shell validator keeps per-type rules. Both run.
- **DONE files with pre-existing valid frontmatter.** Since G6 is dropped, enums are unchanged, so no DONE file becomes newly invalid. The only new rejection surface is genuinely-unknown keys, which no valid file should have. Verify by running the new step over the whole tree before wiring it as a gate.
- **New Go dependency.** `santhosh-tekuri/jsonschema` must be vetted/added to `go.mod`. Low risk (widely used, pure Go), but note it in the PR.
- **Docs-drift gate.** Touching `cmd/validate.go` and `reporter/*` maps to architecture/dev docs (observed repeatedly this session); update the mapped docs in the same PR.
- **`hk` ordering.** The Go check must run after `go-build`; declare the dependency explicitly, as the existing `skill-*` steps do.

## Verification

```bash
# Phase 1 — generated plan matches its dedicated schema (Go test)
go test ./reporter/... ./cmd/...

# Phase 2 — unknown key rejected; whole tree passes; full gate green
#   (fixture with valeu: HIGH must fail the new step)
hk check && go test ./...
```

## Open Questions

- None blocking. Schema scope (Decision 2), runner (Decision 3), and the alongside-not-replacing posture (Decision 4) are resolved. The only future question is *when* to consolidate the shell validator into the JSON-schema step — deferred to after the proving period, not part of this plan.
