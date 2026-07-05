# Remediation flow

The `remediate` command generates or validates structured remediation plans
based on stored audit results.

## Modes

### Generate mode

```text
remediate <skill> [--target-score N] [--dry-run]
  │
  ├── resolve skill path → locate most recent stored audit
  │     └── latestAuditJSON(.context/audits/<skill>/) → most recent <date>/audit.json
  │
  ├── load audit JSON → *scorer.Result
  │
  ├── determine target score (default: min(current + 20, 140))
  │
  ├── build remediation frontmatter:
  │     ├── execution summary (current → target)
  │     ├── score range (current, target, max)
  │     ├── grade range (current_grade, target_grade)
  │     └── critical issues
  │
  ├── build gaps (dimensions sorted by max-score descending)
  │     ├── current score, max, gap
  │     ├── associated diagnostics
  │     └── generic advice per dimension
  │
  ├── build phases:
  │     ├── Phase 1: Critical fixes
  │     ├── Phase 2: Core improvements
  │     └── Phase 3: Stretch goals
  │     └── Each phase has steps with verification commands
  │
  ├── build effort estimates:
  │     ├── overall effort (S/M/L)
  │     ├── total steps
  │     └── time estimate per effort level
  │
  ├── render:
  │     ├── Markdown (.context/plans/<skill>-remediation-plan-<date>.md)
  │     └── JSON (--json flag)
  │
  └── if --dry-run: stdout only
```

### Validate mode

```text
remediate <skill> --validate
  │
  ├── resolve plan path (direct path or glob in .context/plans/)
  │
  ├── reporter.ValidateRemediationPlan(planPath)
  │     ├── extract YAML frontmatter from `---` block
  │     └── validate against regex patterns and allowed values:
  │           ├── plan_date: YYYY-MM-DD
  │           ├── skill_name: kebab-case
  │           ├── source_audit: .context/audits/.../*.md path
  │           ├── score pattern: NNN/140 (NN%)
  │           ├── valid grades: A+, A, B+, B, C+, C, D, F
  │           ├── valid priorities: critical, high, medium, low
  │           ├── valid severities: critical, major, minor, info
  │           ├── valid efforts: S, M, L
  │           ├── step pattern: ## N.
  │           └── notes rating: N/10
  │
  └── returns list of validation errors (empty = valid)
```

## Plan schema

The remediation plan follows a JSON schema defined at
`cmd/assets/schemas/remediation-plan.schema.json`. Key structural types:

```text
remPlanFrontmatter (YAML frontmatter)
  ├── title, type ("plan"), status ("draft"), date, effort (S/M/L/TBD)
  │     └── matches the standard .context/ frontmatter schema, so a freshly
  │         generated plan is picked up by context-index/frontmatter
  │         validation without hand-patching
  ├── plan_date, skill_name, source_audit
  ├── execution_summary (current_score, target_score)
  ├── score_range (min, max, current, target)
  ├── grade_range (current_grade, target_grade)
  ├── critical_issues → []remCritical{description, dimension}
  └── phases → []remPhase{
        phase, title, objective,
        steps → []remStep{
          step, action, details,
          verification → string (shell command),
          code → remCode{language, content},
          effort → remEffort{size, ...},
          success_criteria → []remSuccessCriterion,
          notes → remNotes{rating, text}
        }
      }
```

## Source files

| File | Purpose |
|------|---------|
| `cmd/remediate.go` | Command entry, generate/validate dispatch |
| `reporter/remediation.go` | Simple plan (legacy) |
| `reporter/remediation_plan_generate.go` | Schema-compliant plan generation |
| `reporter/remediation_plan_validate.go` | YAML frontmatter validation |
