# Evaluate flow

The `evaluate` command is the primary entry point for scoring a single skill.

## Resolution chain

```text
User input (path or key)
    │
    ▼
resolveRepoRoot(flag)
    │   Uses `--repo-root` flag or walks up from CWD
    │   looking for .git/ or go.mod
    ▼
resolveSkillPath(arg, repoRoot)
    │   ├── Absolute path / ./ / ../ prefix → used directly
    │   ├── Bare "domain/skill-name"        → <repoRoot>/skills/domain/skill-name/SKILL.md
    │   └── Appends SKILL.md if not present
    ▼
canonicalSkillKey(path, root)
        Strips <repoRoot>/skills/ prefix and SKILL.md suffix
        → "domain/skill-name"
```

**Source:** `cmd/evaluate.go`

## Core scoring pipeline

```text
scorer.Score(ctx, skillPath)
  │
  ├── ioutil.ReadFile(skillPath)         → skill content (string)
  ├── resolveEvalsDir(skillPath)         → evals directory
  │
  └── scorer.ScoreFromContent(ctx, skillPath, content, evalsDir)
        │
        ├── newValidatorBridge(skillDir)
        │     ├── orchestrate.RunContentAnalysis(skillDir)
        │     └── structure.Validate(skillDir, options)
        │     → ContentReport + Structure (cached)
        │
        ├── builds dimensionEntry registry (D1–D9)
        │
        ├── for each entry:
        │     ├── entry.fn(content, skillDir, bridge) → (score, []Diagnostic)
        │     ├── accumulate total += score
        │     └── collect diagnostics
        │
        └── returns Result{
              Skill, Date, Total, MaxTotal, Grade,
              Lines, HasReferences, ReferenceCount,
              Errors, Warnings,
              ErrorDetails, WarningDetails,
              Dimensions map[string]int
            }
```

**Source:** `scorer/scorer.go`

## Validator bridge

The `validatorBridge` wraps the external `github.com/agent-ecosystem/skill-validator` library:

- `newValidatorBridge(skillDir)` runs content analysis and structure validation
- Results are cached and accessed lazily by dimension scorers
- Key accessors: `skillMDTokens()`, `descriptionLen()`, `rawDescription()`, `hasInternalLinkWarning()`

**Source:** `scorer/validator_bridge.go`

## Dimension execution order

| # | Scorer | Source | Max | Key inputs |
|---|--------|--------|-----|------------|
| D1 | Knowledge Delta | `d1_knowledge_delta.go` | 20 | content, skillDir |
| D2 | Mindset & Procedures | `d2_mindset_procedures.go` | 15 | content, bridge |
| D3 | Anti-Pattern Coverage | `d3_anti_pattern_coverage.go` | 15 | content, skillDir, bridge |
| D4 | Specification Compliance | `d4_specification.go` | 15 | content, skillDir, bridge |
| D5 | Progressive Disclosure | `d5_progressive_disclosure.go` | 15 | content, skillDir, bridge |
| D6 | Freedom Calibration | `d6_freedom_calibration.go` | 15 | content, bridge |
| D7 | Pattern Recognition | `d7_pattern_recognition.go` | 10 | bridge |
| D8 | Practical Usability | `d8_practical_usability.go` | 15 | content, bridge |
| D9 | Eval Validation | `d9_eval_validation.go` | 20 | evalsDir, skillPath |

Each scorer returns `(int, []Diagnostic)` — a numeric score and a slice of error/warning/hint diagnostics.

**Source:** `scorer/dimensions.go` (registry), individual `scorer/dN_*.go` files

## Grade bands

| Grade | Range  | Grade | Range  |
|-------|--------|-------|--------|
| A+    | 133–140 | C+   | 105–111 |
| A     | 126–132 | C    | 98–104  |
| B+    | 119–125 | D    | 91–97   |
| B     | 112–118 | F    | 0–90    |

Max total: 140pts.

**Source:** `scorer/grades.go`

## Output

- Default: JSON to stdout
- `--markdown` flag: human-readable text
- `--store` flag: persists to `.context/audits/<skill-key>/<date>/`
  - `audit.json` — raw `scorer.Result`
  - `Analysis.md` — rendered markdown report
  - `Remediation.md` — simple remediation advice

**Source:** `cmd/evaluate.go`, `reporter/store.go`
