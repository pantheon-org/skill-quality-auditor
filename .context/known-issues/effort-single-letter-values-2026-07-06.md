---
title: "Known Issue: effort's single-letter values (S/M/L) violate the full-word enum convention"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-06
severity: LOW
value: LOW
themes:
  - GOVERNANCE
related:
  - ../../docs/ADR/adr-050-uppercase-frontmatter-enums.md
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
  - ../instructions/ways-of-working.md
---

ADR-050 standardised every `.context/` frontmatter enum on full UPPER_CASE words:
`type` (`PLAN`/`FINDING`/…), `status` (`DRAFT`/`ACTIVE`/…), `severity`
(`CRITICAL`/`HIGH`/…), and `value` (`HIGH`/`MEDIUM`/`LOW`). The `effort` field was
left as its pre-existing T-shirt sizes — `S` / `M` / `L` / `TBD` — and is now the
lone outlier: single-letter abbreviations where the rest of the contract uses
spelled-out words.

## Why it matters

Single-letter values are unacceptable under the convention ADR-050 establishes.
`S`/`M`/`L` are terse to the point of ambiguity (is `S` "small" or "short"?), they
read inconsistently beside `HIGH`/`MEDIUM`/`LOW` in the same index entry, and they
force a reader to know the T-shirt-size mapping rather than read the value directly.
The whole point of the UPPER_CASE migration was one legible convention across the
contract; `effort` breaks it.

## Discovered

Immediately after the ADR-050 UPPER_CASE migration, while reviewing the resulting
frontmatter — `effort: L` sitting next to `value: HIGH` makes the inconsistency
obvious.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Rename the `effort` enum to full words: `S → SMALL`, `M → MEDIUM`, `L → LARGE`,
and `TBD → UNKNOWN` (or keep `TBD` as an established, non-single-letter term). This
is a migration of the same shape as ADR-050 and should be scoped the same way:
schema enum, `validate-context-frontmatter.sh`, the `plan-scaffold` schema/template,
`gapEffort()` in `reporter/remediation_plan_generate.go` (which returns `S`/`M`/`L`),
`check-plan-drift.sh` and any effort-sizing prose, every existing `effort:` value in
`.context/**/*.md`, and the authoring-skill docs. Note the collision risk: `MEDIUM`
would then be a legal value for both `effort` and `severity`/`value`, so per-field
enums must stay distinct in the schema even though they share a token. Would warrant
its own ADR (or an amendment to ADR-050).
