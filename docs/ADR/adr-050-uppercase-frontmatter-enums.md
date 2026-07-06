---
title: "ADR-050: .context frontmatter enum values are UPPER_CASE"
status: accepted
date: 2026-07-06
context:
  - path: "docs/ADR/adr-049-context-value-frontmatter-field.md"
  - path: ".context/plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

The `.context/` frontmatter contract used mixed-case enum values: `type` and
`status` were lowercase (`plan`, `draft`), `severity` and `value` were lowercase
(`high`, `critical`), while `effort` was already upper (`S`/`M`/`L`/`TBD`). ADR-049
codified `value` with lowercase enums, consistent with the fields around it. The
casing was inconsistent across the contract and we chose to standardise it.

This ADR amends the enum **casing** established across ADR-046 (`severity`,
`known-issue`), the earlier `effort`/`status`/`type` conventions, and ADR-049
(`value`). It does **not** revisit any of their substantive decisions — the
three-axes model, the value rubric, the read protocol, and per-type requiredness
from ADR-049 all stand unchanged. Only the surface form of the values changes.

## Decision

1. **All `.context/` frontmatter enum values are UPPER_CASE:**
   - `type`: `PLAN` | `FINDING` | `ANALYSIS` | `INSTRUCTION` | `AUDIT` | `KNOWN_ISSUE`
   - `status`: `DRAFT` | `ACTIVE` | `DONE` | `SUPERSEDED`
   - `severity`: `CRITICAL` | `HIGH` | `MEDIUM` | `LOW`
   - `value`: `HIGH` | `MEDIUM` | `LOW`
   - `effort` was already `S` | `M` | `L` | `TBD` — unchanged.

2. **`known-issue` becomes `KNOWN_ISSUE`** (SCREAMING_SNAKE_CASE). The directory
   stays `known-issues/`; the generator maps the `KNOWN_ISSUE` type value to that
   directory group. The enum value and directory name intentionally differ in case
   and separator.

3. **ADR status vocabulary is out of scope and stays lowercase.** ADRs use
   `proposed` | `accepted` | `deprecated` | `superseded`, a separate vocabulary
   with its own tooling (`merge-status-sync.sh`, the ADR schema). It is not changed
   here. Where a script handles both (e.g. `merge-status-sync-lib.sh`), only the
   plan-status branch was uppercased; the ADR-status branch was left as-is.

4. **The migration is mechanical and total.** The schema enums, the validator's
   per-type conditionals, the index generator's `type_group_key` and
   `severity_rank`, `check-plan-drift.sh`, the plan-status branch of
   `merge-status-sync.sh`, the Go remediation-plan generator, every existing
   `.context/**/*.md` frontmatter value, the authoring skills and their templates,
   and the fixture tests were all converted in one PR.

## Consequences

- **Easier:** one casing convention across the whole frontmatter contract; enum
  values are visually distinct from surrounding prose and directory names.
- **Breaking, one-time:** every `.context/` file changed, so the migration had to
  land atomically with the validator, generator, and Go changes, or validation
  would fail mid-migration. Done as a single follow-up PR after the value-signal
  work (ADR-049) had merged, to keep that PR focused and green.
- **Case-sensitive going forward:** `validate-context-frontmatter.sh` rejects
  lowercase values (its enum checks derive from the schema). A file authored with
  `status: active` now fails validation with a clear message.
- **Existing ADRs are unchanged.** Per the ADR immutability rule, historical ADRs
  (including ADR-049 and ADR-046) keep their original lowercase examples as
  point-in-time records; this ADR is the forward-looking casing authority.
- **Not addressed:** ADR-status casing (deliberately out of scope) and the
  separate observation that Go-generated remediation plans do not yet carry a
  `value` field (tracked independently, not a casing concern).
