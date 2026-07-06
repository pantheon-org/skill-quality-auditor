---
title: "Known Issue: generated remediation plans omit value and fail the required-value check"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-06
severity: MEDIUM
value: MEDIUM
related:
  - ../../reporter/remediation_plan_generate.go
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh
  - ../../docs/ADR/adr-049-context-value-frontmatter-field.md
---

`skill-auditor remediate` writes a plan to `.context/plans/` with frontmatter built
from `remPlanFrontmatter` in `reporter/remediation_plan_generate.go`. That struct
emits `title`, `type` (`PLAN`), `status` (`DRAFT`), `date`, `effort`, and several
tool-specific fields — but **no `value`**. Since ADR-049 (Phase 4) made `value`
required for `type: PLAN` while `status` is `DRAFT`/`ACTIVE`, a freshly generated
remediation plan now fails `validate-context-frontmatter.sh`:

    type: PLAN with status: DRAFT must set 'value' (HIGH/MEDIUM/LOW ...)

## Why it matters

The remediation generator's own comment claims its output "matches the standard
.context/ frontmatter schema, so a freshly generated plan is picked up by
context-index/frontmatter validation without hand-patching." That is no longer true:
the pre-commit gate rejects the file until a human adds `value` by hand, breaking the
zero-touch generate → commit flow the tool was built for.

## Discovered

While auditing the blast radius of the UPPER_CASE migration (ADR-050) — inspecting
`remPlanFrontmatter` surfaced that it carries no `value` field, which the Phase-4
requiredness check (merged in #204) now demands.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Add a `Value string` field to `remPlanFrontmatter` and populate it in
`generateRemediationPlan` — e.g. derive it from the score gap the same way `effort`
is derived from `gapEffort()`, or default to `MEDIUM`. Alternatively, if
tool-generated plans should be exempt from the requiredness check, that exemption
must be made explicit in `validate-context-frontmatter.sh` rather than left implicit.
Note the separate additional-properties gap (tracked in
`schema-not-enforced-by-real-validator-2026-07-06.md`): `remPlanFrontmatter` also
emits keys like `skill_name` that `additionalProperties: false` would reject if a
real JSON-schema validator ran.
