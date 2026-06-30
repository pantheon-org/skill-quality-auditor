---
title: "ADR-017: Add mutation coverage scoring to D9 eval validation"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d9-eval-validation-2026-04-29.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D9 (Eval Validation) scorer checked eval scenario structural integrity (criteria sum to 100, task.md/capability.txt presence, instructions coverage) but had no signal for how thoroughly the skill's content was exercised by its evals. A skill could have perfect structural evals that barely touch the skill's actual instructions.

## Decision

Add a Mutation Score requirement (5 pts) that computes what fraction of the skill's actionable statements are covered by at least one eval criterion. Also add:
- **Adversarial scenario bonus** (diagnostic-only, 3 pts) — rewards scenarios that test edge cases the skill does not explicitly authorise
- **Independent authoring bonus** (diagnostic-only, 2 pts) — rewards scenarios written by a different author than the skill (detected via `git log --follow` authorship comparison)

Change `scoreD9` signature to `scoreD9(evalsDir, skillPath string)`. Reweight existing components to free 5 pts for the mutation score.

## Consequences

- D9 now measures eval thoroughness, not just structural validity
- Mutation score prevents "empty" evals that exist but test nothing
- Bonus points incentivise adversarial testing and independent scenario authoring
- Reweighting preserves D9 max of 20
- Highest-priority dimension improvement — executed first
