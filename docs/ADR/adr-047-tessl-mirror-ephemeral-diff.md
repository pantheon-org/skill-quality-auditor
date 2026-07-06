---
title: "ADR-047: verify the .tessl/plugins mirror via CI-only diff, don't track it"
status: accepted
date: 2026-07-06
context:
  - path: ".context/findings/tessl-mirror-drift-protection-2026-07-06.md"
  - path: ".context/findings/index-yaml-split-review-2026-07-06.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

`.context/findings/index-yaml-split-review-2026-07-06.md`'s Migration-Risk review found `.tessl/plugins/**` — the vendored, installed mirror of every helper skill this repo authors — is entirely gitignored (`.gitignore` line 11), and therefore structurally invisible to any git-based CI check. That review's own worked example (`context-index`'s mirror) had already diverged once. Run through `design-debate` (Advocate / Skeptic / Migration-Risk) to decide what, if anything, to do about it.

## Decision

1. **Do not track `.tessl/plugins/pantheon-org/**` in git.** Tracking was seriously considered (436KB/74 files, small, with real precedent in this repo for tracking generated content — `.context/audits/`) but Migration-Risk's review found tracking alone is inert: nothing verifies the tracked copy matches a fresh `tessl install` without *also* writing a new CI step. Once that CI step is required regardless, tracking adds cost (a fragile `.gitignore` negation pattern — a naive one does nothing; the correct layered form is non-obvious — plus a doubled diff on every skill-content PR) without adding protection beyond what the CI step alone provides.
2. **Instead, verify via a CI-only ephemeral diff:** a new step runs `tessl install` into a scratch location and diffs the result against `.context/plugins/pantheon-org/**`'s expected install output, failing the build on divergence. No content is committed; no `.gitignore` change is needed.
3. **Scope is `pantheon-org/**` only.** Third-party registry content (`.tessl/plugins/pantheon-ai/**` at ~174MB, `.tessl/plugins/tessl-labs/**`) is explicitly out of scope for both the rejected tracking option and the accepted diff-check option — we don't author it, and its size alone would make tracking a non-starter if it were ever proposed.
4. **Not yet implemented.** This ADR records the design decision (which approach) reached via `design-debate`; the CI step itself is a follow-up, natural input to `plan-create`.

## Consequences

- **Easier:** drift between an authored `pantheon-org` skill and its installed mirror becomes CI-catchable without any repo bloat, `.gitignore` fragility, or doubled review surface.
- **Harder:** until the CI step is actually built, the underlying gap (mirror drift is invisible) remains open — this decision picks the mechanism, it doesn't close the gap yet.
- **Binding for future work:** if someone later proposes tracking `.tessl/plugins/**` (any subset) to solve a drift concern, this ADR's reasoning — tracking is inert without a verification step, and the verification step alone is sufficient — should be addressed directly, not re-litigated from scratch.
