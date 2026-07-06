---
title: "Plan: Close 3 live .context/ADR governance-gate gaps (adr-index freshness, undocumented-decisions false-negative, remediation-plan value)"
type: PLAN
status: DONE
date: 2026-07-06
effort: M
value: MEDIUM
themes:
  - GOVERNANCE
  - SKILL-QUALITY
related:
  - ../known-issues/adr-index-check-existence-only-2026-07-06.md
  - ../known-issues/adr-undocumented-index-yaml-false-negative-2026-07-06.md
  - ../known-issues/remediation-plan-missing-value-2026-07-06.md
  - ../known-issues/known-issues-lack-enforcement-2026-07-06.md
  - governance-enum-and-schema-hardening-2026-07-06.md
  - ../../docs/ADR/adr-052-governance-known-issue-triage.md
  - ../../hk.pkl
  - ../../reporter/remediation_plan_generate.go
---

**Effort:** M. Three independent, individually-small fixes — two shell CI gates and one Go struct field — but each needs a test and one (G4) interacts with the docs-drift PR gate, so the total is a solid M rather than S. No ADR-level decisions; no cross-gap dependencies.

**Review status:** Reviewed by a 3-reviewer `plan-review` (Sonnet Technical + Strategic, Haiku Risk) on 2026-07-06. The review's convergent finding — that the LOW-value, high-blast-radius effort-enum and schema-validator work was hard-gating the valuable G4 fix — was resolved by splitting those two gaps into a separate deferred plan (see `governance-enum-and-schema-hardening-2026-07-06.md`). This plan now holds only the three quick, genuinely-valuable gaps. Editorial fixes from the review are folded in.

## Goal

Close the three *live, low-cost* enforcement gaps in the `.context/` and ADR governance tooling so the CI gates catch what they claim to: a stale ADR index fails CI, the undocumented-decisions gate stops silently skipping files that merely mention `index.yaml`, and generated remediation plans satisfy the `value` contract they currently violate. See the three `known-issues/` entries in `related`. Two further GOVERNANCE known-issues (`known-issues-lack-enforcement` — a settled deferral, Decision 1; and the effort-enum + real-schema-validator work) are out of scope here; the latter lives in the split-out deferred plan.

## Scope

**In scope (3 gaps):**

- **G2 — adr-index freshness.** Add a `--check` mode to `regenerate-adr-index.sh` (regenerate in memory, diff against the committed `docs/ADR/index.yaml`, exit non-zero with a "run to regenerate" message on mismatch) and repoint `hk.pkl`'s `adr-index.check` at it, replacing the `test -f` existence-only check. Mirrors the existing `regenerate-context-index.sh --check`.
- **G3 — undocumented-decisions false negative.** Replace the bare `if "index.yaml" in content: continue` full-body substring skip in `check-undocumented-decisions.sh` with a precise rule (path-based allowlist, or a self-referential-mention test) so a real decision file that merely mentions `index.yaml` in prose is no longer skipped.
- **G4 — remediation plans carry `value`.** Add a `Value` field to `remPlanFrontmatter` and populate it in `buildRemediationFrontmatter` with a static default of `MEDIUM` (Decision 3), so generated plans pass `validate-context-frontmatter.sh`.

**Out of scope:**

- **G1 — known-issues enforcement.** A `do_not_proceed_for_now` design-debate verdict, not a defect (self-assigned severity, creation-date can't detect neglect, non-deterministic wall-clock gate, trivially silenceable). Stays `ACTIVE` as a tracked deferral (Decision 1).
- **G5 + G6 — real JSON-schema validator and effort full-word enum.** Both are `value: LOW` with real blast radius; split into `governance-enum-and-schema-hardening-2026-07-06.md` (Decision 2) so this MEDIUM plan is not gated on LOW work.
- Redesigning the remediation-plan document format beyond adding `value`.

## Decisions

1. **Exclude G1 (known-issues enforcement).** A settled `do_not_proceed_for_now` deferral, not a bug; actioning it would re-litigate a concluded design-debate. Left `ACTIVE` as tracked.
2. **Split G5 + G6 into a separate deferred LOW-value plan** (`governance-enum-and-schema-hardening-2026-07-06.md`). The `plan-review` Strategic and Risk reviewers independently found that bundling two `value: LOW` gaps into a MEDIUM plan hid the LOW majority-effort tail, and that G6 hard-gated the valuable S-effort G4 with no fallback. Splitting keeps this plan's three gaps quick, valuable, and independently shippable, and lets each plan sort correctly under the read protocol.
3. **G4 assigns `value: MEDIUM` as a static default, not a derived grade.** Deriving `value` from the score gap (a `gapValue()` mirroring `gapEffort()`) is a plausible refinement but invents a second heuristic for no proven benefit; a static `MEDIUM` is the minimal fix that makes generated plans valid. Note the derive-from-gap option in the generator so a future change can revisit it. Revisit trigger: if remediation-plan `value` grades are observed to be systematically wrong, add `gapValue()`.
4. **G2 and G3 ship together in one PR** — they are different files in the same directory (`adr-capture/scripts/`), so zero mutual merge-conflict risk, and both are pure shell-gate fixes. **G4 ships in its own PR** (Go + reporter tests + a docs-drift-triggered doc touch). Two small PRs, not one mixed one.

## Phases

### Wave A — shell CI gate fixes (G2, G3) — one PR

- **G2:** add `--check` to `regenerate-adr-index.sh` mirroring `regenerate-context-index.sh`'s mode (regenerate in memory, diff committed index, exit non-zero + "run to regenerate" on mismatch); repoint `hk.pkl` `adr-index.check`; keep `adr-index.fix` writing. Add `test-regenerate-adr-index.sh` (following the existing `test-merge-status-sync.sh` shell-test precedent in the same directory) asserting a hand-edited/stale `docs/ADR/index.yaml` fails `--check` and a fresh one passes.
- **G3:** replace the substring skip in `check-undocumented-decisions.sh` with a precise rule; add a fixture proving a decision file that mentions `index.yaml` in prose is still scanned, while a genuinely index-about file (by path) is still exempted.
- Run `tessl install` in the same change (both scripts are in a plugin bundle mirrored to `.tessl`).
- Exit criterion: a stale ADR index fails `hk check`; a decision file mentioning `index.yaml` is no longer silently skipped; `test-regenerate-adr-index.sh` passes; mirror-drift clean; full gate green.

### Wave B — remediation generator (G4) — own PR

- Add `Value string` to `remPlanFrontmatter`; populate `Value: "MEDIUM"` in `buildRemediationFrontmatter` (Decision 3); emit it in the frontmatter block adjacent to `effort`.
- Update/extend the reporter tests that assert on generated frontmatter.
- Because this changes `reporter/remediation_plan_generate.go`, the docs-drift PR gate will likely flag a related doc (e.g. `docs/development/skills-and-rules.md` or `docs/reference/*remediation*`) — update the relevant doc in the same PR (as done on #208), or justify in the PR description if genuinely unwarranted.
- Exit criterion: a freshly generated remediation plan (`./dist/skill-auditor remediate <stored-skill>`) passes `validate-context-frontmatter.sh`; `go test ./...` green; docs-drift gate satisfied.

## Risks

- **G3 over/under-correction.** Swapping the substring skip for a filename/path rule could either still over-skip or newly fail to exempt genuinely index-about files. Mitigated by the Wave A fixture covering both directions (prose-mention scanned; index-about-by-path exempted).
- **G2 stricter gate surfaces pre-existing staleness.** Turning on freshness may fail on an already-stale committed `docs/ADR/index.yaml`. Mitigated by regenerating and committing the index as part of Wave A.
- **G4 docs-drift gate.** The Go change will likely trip the docs-drift PR gate (observed on #208). Mitigated by updating the mapped doc in the same PR.
- **Mirror drift** on the two edited plugin scripts — run `tessl install` in the Wave A change; the gate's drift check covers it.
- **index.yaml regeneration conflicts** if Wave A and B land close together (seen on #200/#201, #206/#207). Mitigated by rebasing the second PR and regenerating once.

## Verification

```bash
# Wave A — stale ADR index now fails; undocumented-decisions no longer over-skips
bash .context/plugins/pantheon-org/governance/adr-capture/scripts/regenerate-adr-index.sh --check   # must exist; fails on stale
bash .context/plugins/pantheon-org/governance/adr-capture/scripts/test-regenerate-adr-index.sh       # new fixture test
hk check

# Wave B — generated remediation plan validates; tests green
go test ./...
# ./dist/skill-auditor remediate <stored-skill> && \
#   .context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh <generated-plan>
```

## Open Questions

- **G3 precise-rule form:** path-based allowlist (exempt only files under a named path) vs a self-referential-mention test (exempt only when the sole `index.yaml` mention is inside a code fence)? (Leaning: path-based allowlist — simplest and deterministic; resolve at implementation with the fixture proving both directions.)
- None blocking. The genuinely open design questions (effort vocabulary, schema-validator runner, remediation-plan schema scope) moved with G5/G6 to the split-out plan.
