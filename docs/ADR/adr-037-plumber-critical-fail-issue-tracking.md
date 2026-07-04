---
title: "ADR-037: Plumber CI/CD gate — fail on Critical, track lower severities as issues"
status: proposed
date: 2026-07-04
context:
  - path: "docs/ADR/adr-036-plumber-cicd-security-advisory-workflow.md"
  - path: ".context/findings/plumber-cicd-security-2026-07-04.md"
  - path: ".context/plans/plumber-advisory-workflow-2026-07-04.md"
---

**Status:** Proposed
**Date:** 2026-07-04

## Context

`docs/ADR/adr-036-plumber-cicd-security-advisory-workflow.md` recorded an advisory-only, `continue-on-error: true` design for integrating `getplumber/plumber` into this repo's CI, mirroring the Tessl review's proving-period pattern, and left three items open: whether to commit a trimmed `.plumber.yaml` now or later, whether this repo's regulated-org posture (assumed at the time to be PLG's) required a Security/Compliance check-in before adding an unvetted third-party Action, and who would own reading advisory findings with what graduation trigger.

The user has since resolved all three directly: (1) `.plumber.yaml` needs configuring for this repo regardless of timing, so do it now; (2) this repository is confirmed not in PLG's regulated scope, so no compliance check-in applies; (3) the integration should hard-fail the pipeline on Critical-severity findings and file every lower-severity finding as a tracked GitHub issue, rather than run purely advisory. Point 3 reverses ADR-036's central decision (non-blocking for everything) rather than extending it, so per this repo's ADR-immutability rule, ADR-036 is superseded rather than edited.

## Decision

1. **Critical-severity findings fail the pipeline outright; High, Medium, and Low findings are filed as tracked GitHub issues instead of blocking.** This replaces ADR-036's advisory-only, continue-on-error design in full.
2. **Plumber's own exit code is not used to gate the job.** It reflects an overall A-E score against `--threshold`, not per-finding severity, so it cannot express "block on Critical only." Instead, the workflow runs `plumber analyze --output results.json` with `continue-on-error: true` on that step, then a separate, non-tolerant step parses the JSON and fails the job only when a Critical finding is present.
3. **Every High/Medium/Low finding is filed as a deduplicated GitHub issue**, via a custom step (`if: always()`) added specifically for this integration — Plumber has no native issue-filing output, so this is bespoke automation, not a documented feature of the tool.
4. **`.plumber.yaml` is trimmed to its `github:` controls block and committed in the same phase that adds the workflow.** This repo has no GitLab CI, so the `gitlab:` block is dropped rather than carried as dead configuration.
5. **`permissions: { contents: read, issues: write }`** — `issues: write` is new relative to ADR-036's `contents: read`-only grant, required by the issue-filing step.
6. **No PLG-specific compliance routing applies to this decision.** This repository is confirmed out of PLG's regulated scope; the Security/Compliance check-in ADR-036 flagged as an open question is resolved as not applicable here.
7. **Score-push stays disabled** (`score-push: false`) and the `getplumber/plumber` action reference stays pinned by commit SHA — both carried over unchanged from ADR-036's reasoning.
8. **Before this change ships, the current repo state is checked for pre-existing Critical findings** (via a dry run against `main`). If any exist among the repo's known pinning/permissions gaps, fixing them is pulled forward into this same change, since the new gate would otherwise block its own introduction.

## Consequences

- **Easier:** the repo gets a self-enforcing, permanent gate on the severity that matters (Critical) from day one, with no proving period, no advisory-only ambiguity, and no reliance on someone remembering to read log output — the failure mode ADR-036's own reviewers flagged as its biggest weakness.
- **Easier:** lower-severity findings are not lost — they become tracked, triageable GitHub issues instead of either blocking merges or sitting silently in CI logs.
- **Harder:** this design depends on implementation details not yet verified against Plumber's real output — the exact JSON severity field name/values, and whether Plumber has any native baseline/suppression mechanism for pre-existing findings. Both are recorded as open questions in the accompanying plan and must be confirmed before the gate-parsing script is finalized.
- **Harder:** the issue-filing step is bespoke, unsupported-by-upstream automation with a dedup key this repo must design and maintain itself — a bug here either spams the repo with duplicate issues or silently stops tracking new ones.
- **Harder:** `issues: write` is a new permission surface for this repo's CI that did not exist before; it should be scoped as narrowly as the workflow allows.
