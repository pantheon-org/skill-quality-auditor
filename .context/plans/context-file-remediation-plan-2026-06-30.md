---
title: "Remediation Plan: context-file"
type: PLAN
status: DONE
date: 2026-06-30
value: MEDIUM
effort: L
plan_date: "2026-06-30"
skill_name: context-file
source_audit: .context/audits/context-file/2026-06-30/Analysis.md
executive_summary:
    score:
        current: 91/140 (65%)
        target: 111/140 (79%)
    grade:
        current: D
        target: C+
    priority: Critical
    effort: L
    focus_areas:
        - 'D9: Eval Validation'
        - 'D6: Freedom Calibration'
        - 'D2: Mindset + Procedures'
    verdict: Immediate action required — significant gaps block production use
critical_issues:
    - issue: Eval Validation scores 0/20 (20 pts below max)
      dimension: 'D9: Eval Validation (0/20)'
      severity: Critical
      impact: Missing 20/20 points reduces total score by 14%
    - issue: Freedom Calibration scores 6/15 (9 pts below max)
      dimension: 'D6: Freedom Calibration (6/15)'
      severity: High
      impact: Missing 9/15 points reduces total score by 6%
    - issue: Mindset + Procedures scores 8/15 (7 pts below max)
      dimension: 'D2: Mindset + Procedures (8/15)'
      severity: High
      impact: Missing 7/15 points reduces total score by 5%
    - issue: Specification Compliance scores 10/15 (5 pts below max)
      dimension: 'D4: Specification Compliance (10/15)'
      severity: Medium
      impact: Missing 5/15 points reduces total score by 4%
    - issue: Knowledge Delta scores 17/20 (3 pts below max)
      dimension: 'D1: Knowledge Delta (17/20)'
      severity: Low
      impact: Missing 3/20 points reduces total score by 2%
    - issue: Progressive Disclosure scores 12/15 (3 pts below max)
      dimension: 'D5: Progressive Disclosure (12/15)'
      severity: Low
      impact: Missing 3/15 points reduces total score by 2%
    - issue: Practical Usability scores 13/15 (2 pts below max)
      dimension: 'D8: Practical Usability (13/15)'
      severity: Low
      impact: Missing 2/15 points reduces total score by 1%
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
      dimension: 'D6: Freedom Calibration'
      priority: High
      target: Reach 15/15
      steps:
        - step: "2.1"
          title: Balance prescriptive language (NEVER/ALWAYS) with
          description: Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).
    - phase: 3
      dimension: 'D2: Mindset + Procedures'
      priority: High
      target: Reach 15/15
      steps:
        - step: "3.1"
          title: Add a `## Mindset` or
          description: Add a `## Mindset` or `## Philosophy` section.
        - step: "3.2"
          title: Use numbered procedure lists
          description: Use numbered procedure lists.
        - step: "3.3"
          title: Add `## When to Use`
          description: Add `## When to Use` and `## When NOT to Use` sections.
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
      dimension: 'D1: Knowledge Delta'
      priority: Low
      target: Reach 20/20
      steps:
        - step: "5.1"
          title: 'Add expert-signal keywords: NEVER, ALWAYS'
          description: 'Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern.'
        - step: "5.2"
          title: Remove beginner-oriented patterns (npm install
          description: Remove beginner-oriented patterns (npm install, getting started, hello world).
    - phase: 6
      dimension: 'D5: Progressive Disclosure'
      priority: Low
      target: Reach 15/15
      steps:
        - step: "6.1"
          title: Add a `references/` directory with
          description: Add a `references/` directory with focused deep-dive `.md` files.
        - step: "6.2"
          title: Keep `SKILL.md` under 150 lines
          description: Keep `SKILL.md` under 150 lines to maximise the score.
    - phase: 7
      dimension: 'D8: Practical Usability'
      priority: Low
      target: Reach 15/15
      steps:
        - step: "7.1"
          title: Add more fenced code blocks
          description: Add more fenced code blocks (aim for >5 pairs).
        - step: "7.2"
          title: Include `./` or `bun run`
          description: Include `./` or `bun run` commands.
        - step: "7.3"
          title: Use language-tagged fences (```bash, ```typescript)
          description: Use language-tagged fences (```bash, ```typescript).
verification_commands:
    - cd skill-auditor && go build -o skill-auditor . && ./skill-auditor evaluate context-file --store
    - ./skill-auditor evaluate context-file --json | jq '.grade'
success_criteria:
    - criterion: Total score target
      measurement: Score >= 111/140
    - criterion: Grade improvement
      measurement: '>= C+ (from D)'
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
      effort: S
      time: 30 min
    - phase: Phase 6
      effort: S
      time: 30 min
    - phase: Phase 7
      effort: S
      time: 30 min
    - phase: Total
      effort: L
      time: 12h
dependencies:
    - Completed audit stored in .context/audits/
rollback_plan: git checkout HEAD -- skills/context-file/SKILL.md
notes:
    rating: 6/10
    assessment: Audit reported 7 warning(s). Review before publishing.
---

# Remediation Plan — context-file

**Generated:** 2026-06-30  
**Current:** D (91/140)  
**Target:** C+ (111/140)

## Executive Summary

| Field | Current | Target |
|-------|---------|--------|
| Score | 91/140 (65%) | 111/140 (79%) |
| Grade | D | C+ |
| Priority | Critical | — |

## Critical Issues

| Issue | Dimension | Severity | Impact |
|-------|-----------|----------|--------|
| Eval Validation scores 0/20 (20 pts below max) | D9: Eval Validation (0/20) | Critical | Missing 20/20 points reduces total score by 14% |
| Freedom Calibration scores 6/15 (9 pts below max) | D6: Freedom Calibration (6/15) | High | Missing 9/15 points reduces total score by 6% |
| Mindset + Procedures scores 8/15 (7 pts below max) | D2: Mindset + Procedures (8/15) | High | Missing 7/15 points reduces total score by 5% |
| Specification Compliance scores 10/15 (5 pts below max) | D4: Specification Compliance (10/15) | Medium | Missing 5/15 points reduces total score by 4% |
| Knowledge Delta scores 17/20 (3 pts below max) | D1: Knowledge Delta (17/20) | Low | Missing 3/20 points reduces total score by 2% |
| Progressive Disclosure scores 12/15 (3 pts below max) | D5: Progressive Disclosure (12/15) | Low | Missing 3/15 points reduces total score by 2% |
| Practical Usability scores 13/15 (2 pts below max) | D8: Practical Usability (13/15) | Low | Missing 2/15 points reduces total score by 1% |

## Remediation Phases

### Phase 1

**Dimension:** D9: Eval Validation  
**Target:** Reach 20/20  
**Priority:** Critical

- **Create an `evals/` directory with** (`1.1`): Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

### Phase 2

**Dimension:** D6: Freedom Calibration  
**Target:** Reach 15/15  
**Priority:** High

- **Balance prescriptive language (NEVER/ALWAYS) with** (`2.1`): Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Phase 3

**Dimension:** D2: Mindset + Procedures  
**Target:** Reach 15/15  
**Priority:** High

- **Add a `## Mindset` or** (`3.1`): Add a `## Mindset` or `## Philosophy` section.
- **Use numbered procedure lists** (`3.2`): Use numbered procedure lists.
- **Add `## When to Use`** (`3.3`): Add `## When to Use` and `## When NOT to Use` sections.

### Phase 4

**Dimension:** D4: Specification Compliance  
**Target:** Reach 15/15  
**Priority:** Medium

- **Expand the `description` frontmatter to** (`4.1`): Expand the `description` frontmatter to >100 characters.
- **Ensure no harness-specific paths, agent** (`4.2`): Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Phase 5

**Dimension:** D1: Knowledge Delta  
**Target:** Reach 20/20  
**Priority:** Low

- **Add expert-signal keywords: NEVER, ALWAYS** (`5.1`): Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern.
- **Remove beginner-oriented patterns (npm install** (`5.2`): Remove beginner-oriented patterns (npm install, getting started, hello world).

### Phase 6

**Dimension:** D5: Progressive Disclosure  
**Target:** Reach 15/15  
**Priority:** Low

- **Add a `references/` directory with** (`6.1`): Add a `references/` directory with focused deep-dive `.md` files.
- **Keep `SKILL.md` under 150 lines** (`6.2`): Keep `SKILL.md` under 150 lines to maximise the score.

### Phase 7

**Dimension:** D8: Practical Usability  
**Target:** Reach 15/15  
**Priority:** Low

- **Add more fenced code blocks** (`7.1`): Add more fenced code blocks (aim for >5 pairs).
- **Include `./` or `bun run`** (`7.2`): Include `./` or `bun run` commands.
- **Use language-tagged fences (```bash, ```typescript)** (`7.3`): Use language-tagged fences (```bash, ```typescript).

## Verification Commands

```bash
cd skill-auditor && go build -o skill-auditor . && ./skill-auditor evaluate context-file --store
```

```bash
./skill-auditor evaluate context-file --json | jq '.grade'
```

## Success Criteria

- [ ] Total score target: Score >= 111/140
- [ ] Grade improvement: >= C+ (from D)
- [ ] No critical diagnostics: >= 0 Critical issues resolved
- [ ] All phase steps completed: >= all phases complete

