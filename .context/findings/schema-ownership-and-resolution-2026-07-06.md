---
title: "Finding: schema resolution is hardcoded and the remediation-plan schema is duplicated (one copy stale)"
type: FINDING
status: ACTIVE
date: 2026-07-06
value: MEDIUM
themes:
  - GOVERNANCE
  - SKILL-QUALITY
related:
  - ../plans/governance-enum-and-schema-hardening-2026-07-06.md
  - ../findings/github-action-packaging-2026-07-04.md
  - ../../cmd/validate_context.go
  - ../../cmd/assets/assets/schemas/remediation-plan.schema.json
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/remediation-plan.schema.json
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
  - ../../docs/ADR/adr-050-uppercase-frontmatter-enums.md
---

# Finding: schema resolution is hardcoded and the remediation-plan schema is duplicated (one copy stale)

> The G5 `validate context` command (PR #215/#216) now takes a path for *what*
> to validate, but still loads the schemas from two paths hardcoded relative to
> the repo root. That is fine for this repo, but it couples the tool to this
> repo's layout, and it exposed a second problem: there are now **two**
> remediation-plan schemas — a fresh one authored for G5 and a pre-existing
> embedded tile copy that is stale. This finding records both and the options,
> without deciding.

## Summary

Two related schema-hygiene issues surfaced while resolving the `validate context`
path-hardcoding feedback:

1. **Schema location is hardcoded, not resolved.** `validate_context.go` reads
   `context-frontmatter.schema.json` and `remediation-plan.schema.json` from fixed
   paths under `.context/plugins/pantheon-org/context-mgmt/context-file/assets/schemas/`,
   relative to the repo root. The *target* is now path-configurable but the
   *contract* is not, so the validator only works where those exact paths exist.

2. **The remediation-plan schema is duplicated, and the copies disagree.** G5
   added `.context/plugins/.../context-file/assets/schemas/remediation-plan.schema.json`
   (current: UPPER_CASE enums, `value`/`themes`, `additionalProperties:false`),
   but `cmd/assets/assets/schemas/remediation-plan.schema.json` already existed as
   an embedded tile asset and is **stale**: lowercase `type: "plan"` / `status`
   enums (pre-ADR-050) and no `value`/`themes` (pre-G4). Two schemas for the same
   artifact, one wrong.

## Detail

### Co-location is the right model — this is not "embed everything"

A schema is a skill asset. The repo already co-locates them: `plan-scaffold.schema.json`
with plan-create, `review-report.schema.json` with plan-review, and
`context-frontmatter.schema.json` with the **context-file** skill. So the frontmatter
contract correctly travels *with* the context-file skill — install that skill into
another project and its schema comes along. The fix for issue 1 is therefore **not**
to bake schemas into the binary; it is to **resolve the schema relative to a location
you can point at** (e.g. a `--schema-dir` flag defaulting to today's path), so the
validator finds the co-located schema wherever the skill is installed. This also
pairs with `github-action-packaging-2026-07-04.md` (external reuse of the auditor).

### The two schemas have different owners

- **`context-frontmatter.schema.json`** is owned by the **context-file skill** — a
  shared contract for all `.context/` file types. Co-location with that skill is
  correct; the validator should resolve it there.
- **`remediation-plan.schema.json`** describes output the **Go binary itself
  generates** (`reporter/remediation_plan_generate.go`). That shape is the
  *auditor's* asset, not a skill's. The embedded `cmd/assets/...` copy is the
  natural source of truth (the binary owns both the generator and its schema) — but
  it is the stale one, and G5 read from a second hand-authored copy instead of
  fixing it.

## Options (to weigh in a plan, not decide here)

1. **Schema resolution:** add a `--schema-dir` flag (default = current path) so the
   validator resolves co-located schemas wherever the skill is installed; or a
   convention-based lookup relative to the validated tree.
2. **Remediation-plan de-duplication:** make one schema authoritative. Likely the
   embedded `cmd/assets/...` copy (the binary owns the generated shape) — fix its
   staleness (UPPER_CASE, add `value`/`themes`, `additionalProperties:false`) and
   have the validator read it via `go:embed`, deleting the G5-authored duplicate.
   Alternatively keep the context-file copy authoritative and delete/redirect the
   tile asset.
3. **Do nothing yet:** both copies work for this repo today (the validator uses the
   correct fresh copy; the stale embedded copy is not enforced against anything).
   Defer until external packaging (`github-action-packaging`) makes portability real.

## Next Steps

Fold into a plan alongside `github-action-packaging-2026-07-04.md` if/when external
reuse of the auditor is pursued — schema resolution and de-duplication are
prerequisites for the tool working outside this repo. Until then this is a tracked
hygiene item, not urgent: the live `validate context` path uses the correct schema.
