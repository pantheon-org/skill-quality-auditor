---
title: "ADR-057: Adopt CalVer YYYY.0M.PATCH with monthly patch reset"
status: proposed
date: 2026-07-06
context:
  - path: docs/ADR/adr-023-version-sync-policy.md
  - path: docs/ADR/adr-020-multi-method-distribution.md
  - path: docs/ADR/adr-008-homebrew-distribution.md
---

**Status:** Proposed
**Date:** 2026-07-06

## Context

The project currently versions with SemVer (`X.Y.Z`, e.g. `0.25.0`), driven by release-please
from Conventional Commits. We want to move to a calendar-based scheme so a version tells you
*when* it shipped at a glance.

The chosen target is **`YYYY.0M.PATCH`** — four-digit year, **zero-padded** two-digit month,
and a PATCH that **resets to 0 at the start of each month**:

```text
2026.01.0
2026.07.0
2026.07.1
2026.08.0
2027.01.0
```

This scheme is deliberate and it is **not valid SemVer**. The SemVer spec forbids leading
zeros in numeric identifiers, so `2026.07.0` is rejected by strict SemVer parsers (`2026.7.0`
would be the valid form). That single property drives every consequence below, because the
whole distribution and release chain was built on the assumption that tags are SemVer:
ADR-020 (multi-method distribution), ADR-008 (Homebrew tap), ADR-023 (binary/tile.json version
sync), and ADR-053 (release build/publish).

## Decision

Adopt **`YYYY.0M.PATCH`** as the version format, PATCH resetting per month. Git tags are
prefixed `v` as today (`v2026.07.0`).

The trade-off was made with eyes open: the zero-padded month is preferred for readability and
consistency over strict SemVer compatibility, and the mitigations below are accepted as the
cost of that choice.

## Consequences

This is a breaking change to the release machinery. It **must not** be rolled out until each
item below is resolved. This ADR is `proposed` until an implementation plan lands.

- **release-please does not natively produce CalVer with a monthly reset.** Its version
  computation is SemVer-based. Options to investigate: driving each version explicitly via
  `Release-As:` commit footers (loses automation), a `versioning` strategy override, or
  replacing release-please's version step with a small custom action that computes
  `YYYY.0M.PATCH` (reading the last tag, resetting PATCH when the month rolls over). This is
  the largest unknown and needs a spike before commitment.
- **SemVer-based install tooling may reject or mis-order the tags.** The `ubi` backend used by
  mise, and version comparison in mise itself, sort by SemVer. Leading-zero months are invalid
  identifiers; "latest" resolution and upgrade ordering must be verified for each. If they
  break, an alternative resolution path (explicit version pinning, or a redirect asset) is
  required.
- **Homebrew tap (ADR-008).** Homebrew uses its own version tokenizer rather than strict
  SemVer, and generally tolerates `2026.07.0`, but upgrade ordering across the month/year
  boundary must be confirmed on the tap.
- **tile.json `$.version` and the Tessl registry (ADR-023).** release-please syncs
  `$.version` in `cmd/assets/tile.json`. If the Tessl registry validates SemVer, `2026.07.0`
  may be rejected at publish time. Confirm the registry's version rules before switching.
- **goreleaser (ADR-053).** `{{.Version}}` derives from the tag; SemVer template helpers
  (`{{.Major}}` etc.) may fail on a non-SemVer tag. The build must be exercised end-to-end on
  a CalVer tag.
- **`skill-auditor update` self-update.** Any version comparison in the self-updater must be
  updated to compare CalVer strings correctly (string/tuple comparison, not SemVer parsing).
- **Ordering within the scheme is preserved** for human reading and simple tuple comparison:
  `2026.07.0 < 2026.08.0 < 2027.01.0`. The risk is entirely in third-party tools that assume
  strict SemVer, not in the scheme's own monotonicity.
- **One-way door.** Published tags are immutable (ADR-053). Once CalVer tags exist, reverting
  to SemVer would require versions that sort *above* `2026.xx` forever, so the first CalVer
  release is effectively irreversible. Prove the chain works on a pre-release/test tag first.

## Follow-up

A `.context/` implementation plan should be created to spike the release-please CalVer
mechanism and verify each install path against a throwaway CalVer tag before the first real
CalVer release cuts. This ADR records the direction and the constraints; it does not authorise
the switch until that plan is green.
