---
title: "Remediation Plan: adr-capture"
type: plan
status: done
date: 2026-06-30
effort: L
plan_date: "2026-06-30"
skill_name: adr-capture
source_audit: .context/audits/adr-capture/2026-06-30/Analysis.md
executive_summary:
    score:
        current: 85/140 (61%)
        target: 105/140 (75%)
    grade:
        current: F
        target: C+
    priority: Critical
    effort: L
    focus_areas:
        - 'D9: Eval Validation'
        - 'D2: Mindset + Procedures'
        - 'D6: Freedom Calibration'
    verdict: Immediate action required — significant gaps block production use
critical_issues:
    - issue: Eval Validation scores 0/20 (20 pts below max)
      dimension: 'D9: Eval Validation (0/20)'
      severity: Critical
      impact: Missing 20/20 points reduces total score by 14%
    - issue: Mindset + Procedures scores 3/15 (12 pts below max)
      dimension: 'D2: Mindset + Procedures (3/15)'
      severity: Critical
      impact: Missing 12/15 points reduces total score by 9%
    - issue: Freedom Calibration scores 5/15 (10 pts below max)
      dimension: 'D6: Freedom Calibration (5/15)'
      severity: Critical
      impact: Missing 10/15 points reduces total score by 7%
    - issue: Specification Compliance scores 10/15 (5 pts below max)
      dimension: 'D4: Specification Compliance (10/15)'
      severity: Medium
      impact: Missing 5/15 points reduces total score by 4%
    - issue: Progressive Disclosure scores 10/15 (5 pts below max)
      dimension: 'D5: Progressive Disclosure (10/15)'
      severity: Medium
      impact: Missing 5/15 points reduces total score by 4%
    - issue: Knowledge Delta scores 17/20 (3 pts below max)
      dimension: 'D1: Knowledge Delta (17/20)'
      severity: Low
      impact: Missing 3/20 points reduces total score by 2%
remediation_phases:
    - phase: 1
      dimension: 'D9: Eval Validation'
      priority: Critical
      target: Reach 20/20
      steps:
        - step: "1.1"
          title: Create an `evals/` directory with
          description: Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.
    - phase: 2
      dimension: 'D2: Mindset + Procedures'
      priority: Critical
      target: Reach 15/15
      steps:
        - step: "2.1"
          title: Add a `## Mindset` or
          description: Add a `## Mindset` or `## Philosophy` section.
        - step: "2.2"
          title: Use numbered procedure lists
          description: Use numbered procedure lists.
        - step: "2.3"
          title: Add `## When to Use`
          description: Add `## When to Use` and `## When NOT to Use` sections.
    - phase: 3
      dimension: 'D6: Freedom Calibration'
      priority: Critical
      target: Reach 15/15
      steps:
        - step: "3.1"
          title: Balance prescriptive language (NEVER/ALWAYS) with
          description: Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).
    - phase: 4
      dimension: 'D4: Specification Compliance'
      priority: Medium
      target: Reach 15/15
      steps:
        - step: "4.1"
          title: Expand the `description` frontmatter to
          description: Expand the `description` frontmatter to >100 characters.
        - step: "4.2"
          title: Ensure no harness-specific paths, agent
          description: Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.
    - phase: 5
      dimension: 'D5: Progressive Disclosure'
      priority: Medium
      target: Reach 15/15
      steps:
        - step: "5.1"
          title: Add a `references/` directory with
          description: Add a `references/` directory with focused deep-dive `.md` files.
        - step: "5.2"
          title: Keep `SKILL.md` under 150 lines
          description: Keep `SKILL.md` under 150 lines to maximise the score.
    - phase: 6
      dimension: 'D1: Knowledge Delta'
      priority: Low
      target: Reach 20/20
      steps:
        - step: "6.1"
          title: 'Add expert-signal keywords: NEVER, ALWAYS'
          description: 'Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern.'
        - step: "6.2"
          title: Remove beginner-oriented patterns (npm install
          description: Remove beginner-oriented patterns (npm install, getting started, hello world).
verification_commands:
    - cd skill-auditor && go build -o skill-auditor . && ./skill-auditor evaluate adr-capture --store
    - ./skill-auditor evaluate adr-capture --json | jq '.grade'
success_criteria:
    - criterion: Total score target
      measurement: Score >= 105/140
    - criterion: Grade improvement
      measurement: '>= C+ (from F)'
    - criterion: No critical diagnostics
      measurement: '>= 0 Critical issues resolved'
    - criterion: All phase steps completed
      measurement: '>= all phases complete'
effort_estimates:
    - phase: Phase 1
      effort: L
      time: 3+ hours
    - phase: Phase 2
      effort: M
      time: 1-2 hours
    - phase: Phase 3
      effort: M
      time: 1-2 hours
    - phase: Phase 4
      effort: M
      time: 1-2 hours
    - phase: Phase 5
      effort: M
      time: 1-2 hours
    - phase: Phase 6
      effort: S
      time: 30 min
    - phase: Total
      effort: L
      time: 12h
dependencies:
    - Completed audit stored in .context/audits/
rollback_plan: git checkout HEAD -- skills/adr-capture/SKILL.md
notes:
    rating: 6/10
    assessment: Audit reported 9 warning(s). Review before publishing.
---

# Remediation Plan — adr-capture

**Generated:** 2026-06-30  
**Current:** F (85/140)  
**Target:** C+ (105/140)

## Executive Summary

| Field | Current | Target |
|-------|---------|--------|
| Score | 85/140 (61%) | 105/140 (75%) |
| Grade | F | C+ |
| Priority | Critical | — |

## Critical Issues

| Issue | Dimension | Severity | Impact |
|-------|-----------|----------|--------|
| Eval Validation scores 0/20 (20 pts below max) | D9: Eval Validation (0/20) | Critical | Missing 20/20 points reduces total score by 14% |
| Mindset + Procedures scores 3/15 (12 pts below max) | D2: Mindset + Procedures (3/15) | Critical | Missing 12/15 points reduces total score by 9% |
| Freedom Calibration scores 5/15 (10 pts below max) | D6: Freedom Calibration (5/15) | Critical | Missing 10/15 points reduces total score by 7% |
| Specification Compliance scores 10/15 (5 pts below max) | D4: Specification Compliance (10/15) | Medium | Missing 5/15 points reduces total score by 4% |
| Progressive Disclosure scores 10/15 (5 pts below max) | D5: Progressive Disclosure (10/15) | Medium | Missing 5/15 points reduces total score by 4% |
| Knowledge Delta scores 17/20 (3 pts below max) | D1: Knowledge Delta (17/20) | Low | Missing 3/20 points reduces total score by 2% |

## Remediation Phases

### Phase 1

**Dimension:** D9: Eval Validation  
**Target:** Reach 20/20  
**Priority:** Critical

- **Create an `evals/` directory with** (`1.1`): Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

### Phase 2

**Dimension:** D2: Mindset + Procedures  
**Target:** Reach 15/15  
**Priority:** Critical

- **Add a `## Mindset` or** (`2.1`): Add a `## Mindset` or `## Philosophy` section.
- **Use numbered procedure lists** (`2.2`): Use numbered procedure lists.
- **Add `## When to Use`** (`2.3`): Add `## When to Use` and `## When NOT to Use` sections.

### Phase 3

**Dimension:** D6: Freedom Calibration  
**Target:** Reach 15/15  
**Priority:** Critical

- **Balance prescriptive language (NEVER/ALWAYS) with** (`3.1`): Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Phase 4

**Dimension:** D4: Specification Compliance  
**Target:** Reach 15/15  
**Priority:** Medium

- **Expand the `description` frontmatter to** (`4.1`): Expand the `description` frontmatter to >100 characters.
- **Ensure no harness-specific paths, agent** (`4.2`): Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Phase 5

**Dimension:** D5: Progressive Disclosure  
**Target:** Reach 15/15  
**Priority:** Medium

- **Add a `references/` directory with** (`5.1`): Add a `references/` directory with focused deep-dive `.md` files.
- **Keep `SKILL.md` under 150 lines** (`5.2`): Keep `SKILL.md` under 150 lines to maximise the score.

### Phase 6

**Dimension:** D1: Knowledge Delta  
**Target:** Reach 20/20  
**Priority:** Low

- **Add expert-signal keywords: NEVER, ALWAYS** (`6.1`): Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern.
- **Remove beginner-oriented patterns (npm install** (`6.2`): Remove beginner-oriented patterns (npm install, getting started, hello world).

## Verification Commands

```bash
cd skill-auditor && go build -o skill-auditor . && ./skill-auditor evaluate adr-capture --store
```

```bash
./skill-auditor evaluate adr-capture --json | jq '.grade'
```

## Success Criteria

- [ ] Total score target: Score >= 105/140
- [ ] Grade improvement: >= C+ (from F)
- [ ] No critical diagnostics: >= 0 Critical issues resolved
- [ ] All phase steps completed: >= all phases complete

