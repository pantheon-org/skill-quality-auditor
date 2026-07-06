---
title: "Software Engineering Principles Review"
type: ANALYSIS
status: DONE
date: 2026-04-29
related:
  - ../plans/se-principles-remediation-plan-2026-04-29.md
---
# Software Engineering Principles Review

> Codebase: `skill-quality-auditor` (Go CLI)
> Date: 2026-04-29
> Scope: All packages — `scorer/`, `reporter/`, `duplication/`, `cmd/`

---

## Executive Summary

The codebase is well-structured at the package level and has strong test coverage. The domain model is clear and self-consistent. The main structural debt falls into three areas: **package-level mutable globals in `cmd/`** (the most serious issue), **no interface seams between layers**, and **a `reporter` package carrying too many concerns**.

Severity scale: 🔴 Critical · 🟠 High · 🟡 Medium · 🟢 Low

---

## 1. SOLID Analysis

### 1.1 Single Responsibility Principle (SRP)

#### 🟠 `reporter` package — five concerns in one package

The `reporter` package currently handles:
1. Text formatting (`reporter.go`, `Format()`)
2. JSON serialisation (`CombinedJSON`)
3. Aggregation plan generation (`aggregation_plan.go`)
4. Remediation plan generation + schema validation (`remediation_plan.go`, `remediation.go`)
5. Trend/combined analysis (`analysis.go`)

Each of these has a separate reason to change (output format, YAML schema, aggregation logic, trend algorithm). The package will absorb every future output feature.

**Refactor signal:** Split into focused sub-packages or files with a single exported surface:
`reporter/format`, `reporter/aggregate`, `reporter/remediate`, `reporter/trend`.

#### 🟡 `validatorBridge` — data ferry vs. behaviour host

`validatorBridge` wraps two external library results (`ContentReport`, `types.Report`) and adds three query methods (`skillMDTokens`, `descriptionLen`, `hasInternalLinkWarning`). It is not a cohesive object — it is a struct used to avoid passing the same two values to every dimension function.

This is an acceptable pragmatic choice, but it should be an unexported implementation detail (it already is) and should not grow further.

---

### 1.2 Open/Closed Principle (OCP)

#### 🟠 `scorer.go` — adding D10 requires editing the core orchestrator

`scorer.go` calls all nine dimensions explicitly:

```go
d1, _ := scoreD1(content, skillDir, b)
d2, _ := scoreD2(content, skillDir, b)
// ... d3–d9
return &Result{..., Dimensions: dimensionScores(d1, d2, ..., d9)}, nil
```

Adding a new dimension (D10) forces a modification to the `Score()` function and `dimensionScores()`. The open/closed violation is real: the extension point for new scoring dimensions is the centre of the module, not its edge.

**Option A (minimal):** Register dimensions in a slice in `dimensions.go` and iterate:

```go
type DimensionFn func(content, skillDir string, b *validatorBridge) (int, []Diagnostic)

var dimensionFns = []struct{
    d    Dimension
    fn   DimensionFn
}{
    {D1, scoreD1}, {D2, scoreD2}, ...
}
```

`Score()` then loops. Adding D10 becomes: add an entry to the slice.

**Option B (interface):** Define a `DimensionScorer` interface and register implementations — heavier but needed if scorers gain state.

---

### 1.3 Liskov Substitution Principle (LSP)

No inheritance hierarchies exist in the codebase (Go composition model). LSP is not violated. ✅

---

### 1.4 Interface Segregation Principle (ISP)

No interfaces are defined in the codebase at all (see DIP below). ISP cannot be violated when there are no interfaces, but the absence itself is the issue.

---

### 1.5 Dependency Inversion Principle (DIP)

#### 🔴 `cmd/` depends directly on `scorer` and `reporter` concrete packages

Every command (`evaluate`, `batch`, `analyze`, etc.) imports and calls `scorer.Score()` and `reporter.Format()` directly. There is no interface boundary between the CLI layer and the domain layer.

```
cmd/evaluate.go  →  scorer.Score()        (concrete call)
cmd/evaluate.go  →  reporter.Format()     (concrete call)
cmd/evaluate.go  →  reporter.Store()      (concrete call)
```

This is not a critical problem for a CLI (CLIs are naturally thin), but it means:
- The `cmd` package cannot be tested without running the full scorer pipeline.
- Swapping output format or scoring strategy requires touching command code.

For the evaluate/batch commands, at minimum the `reporter` output should accept an `io.Writer` so tests don't need `os.Pipe()` + `os.Stdout = w` gymnastics.

---

## 2. Anti-Pattern Analysis

### 2.1 🔴 Global mutable state in `cmd/` (worst issue in the codebase)

Every command stores its flags in package-level `var` declarations:

```go
// cmd/evaluate.go
var (
    evalJSON     bool
    evalStore    bool
    evalRepoRoot string
)
```

This pattern is replicated across every command file (`aggregate`, `analyze`, `duplication`, `remediate`, `trend`, `init`, `batch`, `prune`). The consequences:

1. **Tests cannot run in parallel.** Every test helper (`captureAnalyzeOutput`, `TestAggregateCmd_*`, etc.) saves globals, mutates them, defers restore, and pipes stdout. This is thread-unsafe:

   ```go
   origFamily := aggFamily   // race window
   aggFamily = "test"
   defer func() { aggFamily = origFamily }()
   ```

   A single `t.Parallel()` call in any of these tests will introduce flaky races.

2. **Tests require OS-level pipe replacement** (`os.Stdout = w`) to capture output — a fragile and non-composable approach.

**Fix:** Replace global flag vars with a struct per command and accept `io.Writer` for output:

```go
type evaluateOptions struct {
    asJSON   bool
    store    bool
    repoRoot string
    out      io.Writer
}

func newEvaluateCmd(out io.Writer) *cobra.Command {
    opts := &evaluateOptions{out: out}
    cmd := &cobra.Command{...}
    cmd.Flags().BoolVar(&opts.asJSON, "json", false, "")
    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        return runEvaluate(args[0], opts)
    }
    return cmd
}
```

This is the canonical Cobra pattern (used by `kubectl`, `gh`, etc.).

---

### 2.2 🟠 Magic numbers scattered through scorer dimension files

Scoring thresholds are inline literals with no named constant:

| File | Magic value | Meaning |
|------|-------------|---------|
| `d5_progressive_disclosure.go` | `800`, `1200`, `1600` | Token thresholds |
| `d9_eval_validation.go` | `80` | Coverage % threshold |
| `d3_anti_pattern.go` | `8`, `4` | Strong-marker thresholds |
| `d4_specification.go` | `17`, `8` | Max score, base score |
| `d8_practical_usability.go` | `5`, `2`, `4` | Code block scoring steps |

These are the scoring rubric in code form — they will change when the rubric evolves, and a reviewer cannot distinguish an intentional threshold from an accidental literal.

**Fix:** Collect all threshold constants into `scorer/thresholds.go`:

```go
const (
    D5TokenCompact  = 800
    D5TokenModerate = 1200
    D5TokenVerbose  = 1600
    D9CoverageMin   = 80
    // ...
)
```

---

### 2.3 🟡 `os.Stdout` written directly in command `RunE` closures

All commands write to `os.Stdout` via `fmt.Printf` or `fmt.Fprintln`. Because there is no `io.Writer` injection, tests require:

```go
r, w, _ := os.Pipe()
old := os.Stdout
os.Stdout = w
// ... run command ...
os.Stdout = old
```

This is the "Humble Object" pattern in reverse — the testable logic is entangled with I/O side effects. The fix is the same as §2.1: inject `io.Writer` through command options.

---

### 2.4 🟡 `duplication.Detect` is O(n²) with no guard

```go
for i := 0; i < len(entries); i++ {
    for j := i + 1; j < len(entries); j++ {
        sim := Similarity(entries[i].Content, entries[j].Content)
        // ...
    }
}
```

For n=50 skills this is 1225 similarity comparisons, each computing token sets. At n=200 it becomes 19900 comparisons. There is no size guard or early exit.

This is not a correctness issue today (the skill corpus is small) but violates "don't optimise before measuring" in the opposite direction — the algorithm should at least document its complexity and accept a `limit` parameter or a cap on corpus size.

---

### 2.5 🟢 `extractFrontmatterField` reimplements YAML parsing

`scorer/util.go` has a hand-rolled YAML frontmatter parser that scans lines for `field: value` patterns. The project already depends on `gopkg.in/yaml.v3` (used in `reporter/`). The hand-rolled parser handles quoted values and multi-field frontmatter but will silently fail on multi-line values, block scalars, and anchors.

**Fix:** Use `yaml.v3` consistently. Define a minimal struct, unmarshal only the frontmatter block.

---

## 3. Architectural Observations

### 3.1 Dependency direction is correct ✅

```
cmd → scorer, reporter, duplication
reporter → scorer
scorer → (external: skill-validator)
duplication → (stdlib only)
```

No circular dependencies. The inner packages (`scorer`, `duplication`) do not import the outer layers. This is clean.

### 3.2 `cmd/root.go` — no explicit dependency wiring

`root.go` adds all subcommands in `init()`. There is no explicit wiring point (a `main` component / composition root) where dependencies are assembled. This is standard for small CLIs but means the global var pattern (§2.1) is the only available injection mechanism today.

A `NewRootCmd(out io.Writer) *cobra.Command` factory would be the composition root and would unlock both testability and `io.Writer` injection.

### 3.3 `scorer/dimensions.go` as the single source of truth ✅

The recent refactor (PR #27) introducing `Dimension` as the canonical type is architecturally correct. `AllDimensions` slice + `dimLabelToCode` lookup centralises the metadata that was previously spread across `scorer` and `reporter`. This is the right direction — continue expanding it (add max scores, thresholds) to allow the scorer loop pattern from §1.2.

---

## 4. Priority Remediation Plan

| # | Issue | Severity | Effort |
|---|-------|----------|--------|
| 1 | Replace global flag vars with per-command option structs + `io.Writer` injection | 🔴 | L |
| 2 | Extract scoring threshold literals to named constants in `thresholds.go` | 🟠 | S |
| 3 | Register dimension scorers in a slice so `Score()` can iterate (OCP fix) | 🟠 | M |
| 4 | Split `reporter` into focused sub-packages | 🟠 | M |
| 5 | Replace `extractFrontmatterField` with `yaml.v3` unmarshal | 🟡 | S |
| 6 | Document `Detect` O(n²) complexity; add corpus size guard or cap | 🟡 | S |

Items 2 and 5 are safe, independent, and achievable in a single PR. Items 1 + 3 are the architectural changes that would unlock parallel tests and clean extension points.
