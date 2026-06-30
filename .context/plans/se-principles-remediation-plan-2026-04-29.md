---
title: "SE Principles Remediation Plan"
type: plan
status: done
date: 2026-04-29
related:
  - ../analysis/code-review-se-principles.md
---

# SE Principles Remediation Plan

> Source: `.context/analysis/code-review-se-principles.md`
> Date: 2026-04-29
> **Status: ALL PHASES COMPLETE** — branch `refactor/phase3-cmd-io-writer`

---

## Phase 1 — Quick Wins ✅ DONE

**Branch:** `refactor/phase3-cmd-io-writer` (commit `317898f`)

### 1.1 Extract scoring threshold literals to `scorer/thresholds.go` ✅

`scorer/thresholds.go` created with unexported constants:

```go
const (
    d3StrongMarkersHigh = 8
    d3StrongMarkersMid  = 4
    d5TokenCompact      = 800
    d5TokenModerate     = 1200
    d5TokenVerbose      = 1600
    d9CoverageMin       = 80
    d8BlocksHigh        = 5
    d8BlocksMid         = 2
)
```

All inline literals in `d3_anti_pattern.go`, `d4_specification.go`,
`d5_progressive_disclosure.go`, `d8_practical_usability.go`,
`d9_eval_validation.go` replaced with these constants.

### 1.2 Replace hand-rolled YAML parser with `yaml.v3` ✅

`extractFrontmatterField` now delegates to `parseFrontmatter(content)` which
uses `yaml.Unmarshal` into a typed `skillFrontmatter` struct. The old string-
scanning implementation is gone.

**Note:** `extractFrontmatterField` itself was retained as a thin wrapper for
call-site compatibility (used in `scorer/dimensions.go`). The hand-rolled
parser was replaced; the function name was kept.

### 1.3 Document `duplication.Detect` complexity; add corpus size guard ✅

`duplication/detect.go` now exports:

```go
// MaxDetectEntries caps the corpus size fed to Detect. O(n²) comparisons become
// expensive beyond a few hundred entries; entries beyond this cap are silently dropped.
MaxDetectEntries = 500
```

`Detect` truncates `entries` to `MaxDetectEntries` at entry. Function-level
`// O(n²) pairwise comparison` comment added.

---

## Phase 2 — Scorer OCP: Dimension Registry ✅ DONE

**Branch:** `refactor/phase3-cmd-io-writer` (commit `317898f`)

### 2.1 `dimensionFn` type and registry in `scorer/dimensions.go` ✅

```go
type dimensionFn func(content, skillDir string, b *validatorBridge) (int, []Diagnostic)

type dimensionEntry struct {
    Dimension
    fn dimensionFn
}
```

### 2.2 `scorer.Score()` loops over registry ✅

`ScoreFromContent` builds a local `registry []dimensionEntry` slice and loops:

```go
for i, entry := range registry {
    s, diags := entry.fn(content, skillDir, bridge)
    scores[i] = s
    total += s
    allDiags = append(allDiags, diags...)
}
```

**Implementation note:** The registry uses inline closure adapters (not a
package-level `var registry`) because D1, D5, D9 require extra parameters
captured from the enclosing scope (e.g. `evalsDir`, metadata side-effects).
The `scoreD[0-9]` function names still appear inside those closures — this is
intentional. The OCP goal (no hard-coded dispatch chain in `Score()`) is met.

### 2.3 Verification ✅

`go test ./scorer/... -race` passes. `scorer.Score()` contains no `if d == "D1"` branches.

---

## Phase 3 — cmd/ Globals → `io.Writer` injection ✅ DONE

**Branch:** `refactor/phase3-cmd-io-writer` (commit `6a47f4a`)

### 3.1 `NewRootCmd(out io.Writer)` in `cmd/root.go` ✅

Factory function added. `main.go` compatibility preserved via `Execute()`.

### 3.2 All command files converted ✅

Files converted: `analyze.go`, `remediate.go`, `trend.go`, `duplication.go`,
`aggregate.go`, `prune.go`, `init.go`, `validate.go`.

**Approach diverged from plan:** Rather than per-command factory functions
(`newEvaluateCmd(out)`), the implementation kept the existing `cobra.Command`
vars and eliminated package-level flag variables. Flags are registered in
`init()` without `Var` binding; `RunE` reads them via
`cmd.Flags().GetBool("flag-name")`. Output goes through `cmd.OutOrStdout()`.

Tests use `cmd.ResetFlags()` + `cmd.Flags().Set(name, value)` + `cmd.SetOut(buf)` —
no `os.Pipe()` or global mutation.

**`MarkFlagRequired("family")` removed from aggregate:** Moved guard into
`RunE` as `if family == "" { return fmt.Errorf(...) }` to allow tests to reach
`RunE` cleanly.

**`.golangci.yml` added:** `errcheck` exclusions for `fmt.Fprint*` (fire-and-
forget CLI writes to `io.Writer`).

### 3.3 Verification ✅

`go test ./cmd/... -race` exits 0. Build and evaluate round-trip confirmed.

---

## Phase 4 — reporter/ SRP Split ✅ DONE

**Branch:** `refactor/phase3-cmd-io-writer` (commit `3a757a2`)

### 4.1 File split ✅

`reporter/remediation_plan.go` (591 lines, two concerns) replaced by:

| File | Single mandate |
|------|---------------|
| `remediation_plan_generate.go` | YAML structs + `RemediationPlan()` + all generation helpers |
| `remediation_plan_validate.go` | `ValidateRemediationPlan()` + frontmatter parsing + validation regexes |

### 4.2 Single-mandate doc comments ✅

All eight `reporter/*.go` files now open with a one-line mandate comment:

| File | Mandate |
|------|---------|
| `reporter.go` | result formatting: `scorer.Result` → human-readable text |
| `store.go` | audit persistence: write to `.context/audits/` |
| `analysis.go` | pattern analysis persistence: write to `.context/analysis/` |
| `aggregation.go` | aggregation plan formatting |
| `combined_analysis.go` | `CombinedAnalysis` struct and its serialisers |
| `duplication.go` | duplication report formatting |
| `remediation.go` | simple remediation: prioritised action plan from `scorer.Result` |
| `remediation_plan_generate.go` | schema-compliant YAML-frontmatter plan generation |
| `remediation_plan_validate.go` | plan validation against schema constraints |

**Implementation corrections during Phase 4:**
- `gapEffort` threshold adjusted: `>= 15 → L`, `>= 5 → M`, else `S`
- `gapTime` returns human strings (`"3+ hours"`, `"1-2 hours"`, `"30 min"`)
- `planVerdict` strings updated to contain `"Immediate"` / `"Priority"` substrings per tests
- `planSkillName` now strips trailing `.md` file components
- `fmt.Sscanf` replaced with `gapHours()` helper to satisfy `errcheck`

### 4.3 Verification ✅

`go test ./reporter/... -race` exits 0. All 591-line concerns now testable in isolation.

---

## Execution Order

```
Phase 1  ──►  Phase 2  ──►  Phase 3
                              │
Phase 4 ◄─────────────────────┘  (can start after Phase 1, independent of 2–3)
```

All phases completed on branch `refactor/phase3-cmd-io-writer`.
Next step: open PR against `main`.
