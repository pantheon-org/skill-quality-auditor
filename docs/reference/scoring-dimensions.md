# Scoring dimensions

The framework evaluates skills across 9 dimensions (D1–D9), totalling 140 points.

## Dimension table

| ID | Dimension | Max | Source file | Key |
|----|-----------|-----|-------------|-----|
| D1 | Knowledge Delta | 20 | `scorer/d1_knowledge_delta.go` | `knowledgeDelta` |
| D2 | Mindset & Procedures | 15 | `scorer/d2_mindset_procedures.go` | `mindsetProcedures` |
| D3 | Anti-Pattern Coverage | 15 | `scorer/d3_anti_pattern_coverage.go` | `antiPatternQuality` |
| D4 | Specification Compliance | 15 | `scorer/d4_specification.go` | `specificationCompliance` |
| D5 | Progressive Disclosure | 15 | `scorer/d5_progressive_disclosure.go` | `progressiveDisclosure` |
| D6 | Freedom Calibration | 15 | `scorer/d6_freedom_calibration.go` | `freedomCalibration` |
| D7 | Pattern Recognition | 10 | `scorer/d7_pattern_recognition.go` | `patternRecognition` |
| D8 | Practical Usability | 15 | `scorer/d8_practical_usability.go` | `practicalUsability` |
| D9 | Eval Validation | 20 | `scorer/d9_eval_validation.go` | `evalValidation` |

## Per-dimension details

| | |
|---|---|
| [D1: Knowledge Delta](d1-knowledge-delta.md) | Expert-only knowledge, knowledge types, redundancy scoring |
| [D2: Mindset & Procedures](d2-mindset-procedures.md) | Role adoption, PKO structural model, procedure quality |
| [D3: Anti-Pattern Coverage](d3-anti-pattern-coverage.md) | Negative guidance, SYMPTOM/CONSEQUENCE format |
| [D4: Specification Compliance](d4-specification-compliance.md) | Behaviour precision, mutation resistance, edge cases |
| [D5: Progressive Disclosure](d5-progressive-disclosure.md) | Information hierarchy, line/token limits, negative-condition scoring |
| [D6: Freedom Calibration](d6-freedom-calibration.md) | Allow/require/disallow balance, action weighting |
| [D7: Pattern Recognition](d7-pattern-recognition.md) | Pattern descriptions, discriminativeness signal |
| [D8: Practical Usability](d8-practical-usability.md) | Outcome linkage, triggers, examples |
| [D9: Eval Validation](d9-eval-validation.md) | Scenario quality, mutation coverage, criteria weighting |

## Grade bands

| Grade | Range | Grade | Range |
|-------|-------|-------|-------|
| A+ | 133–140 | C+ | 105–111 |
| A | 126–132 | C | 98–104 |
| B+ | 119–125 | D | 91–97 |
| B | 112–118 | F | 0–90 |

## Diagnostics

Each dimension scorer returns diagnostics classified by severity:

| Severity | Constructor | Purpose |
|----------|-------------|---------|
| error | `errDiag(dim, msg)` | Issues that reduce the score |
| warning | `warnDiag(dim, msg)` | Issues to note, minor impact |
| hint | `hintDiag(dim, msg)` | Suggestions for improvement |

Diagnostics are surfaced in audit reports and remediation plans.

## Source files

| File | Purpose |
|------|---------|
| `scorer/dimensions.go` | AllDimensions slice, Diagnostic, Result types |
| `scorer/grades.go` | GradeRank map, Grade() function |
| `scorer/thresholds.go` | Centralised rubric cut-points |
| `scorer/d1_knowledge_delta.go` | D1 scorer |
| `scorer/d2_mindset_procedures.go` | D2 scorer |
| `scorer/d3_anti_pattern_coverage.go` | D3 scorer |
| `scorer/d4_specification.go` | D4 scorer |
| `scorer/d5_progressive_disclosure.go` | D5 scorer |
| `scorer/d6_freedom_calibration.go` | D6 scorer |
| `scorer/d7_pattern_recognition.go` | D7 scorer |
| `scorer/d8_practical_usability.go` | D8 scorer |
| `scorer/d9_eval_validation.go` | D9 scorer |
