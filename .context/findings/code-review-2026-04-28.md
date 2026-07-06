---
title: "Critical Code Review — skill-quality-auditor"
type: FINDING
status: DONE
date: 2026-04-28
value: HIGH
---
# Critical Code Review — skill-quality-auditor

**Date:** 2026-04-28  
**Reviewer:** Claude (claude-sonnet-4-6)  
**Scope:** All Go source files (`scorer/`, `cmd/`, `duplication/`, `reporter/`, `analysis/`, `agents/`)

**Remediation:**
- High-priority items (H1–H4) resolved 2026-04-28, commit `4f0a405`
- Medium-priority items (M1–M5) resolved 2026-04-28, commit `671df70`
- Low-priority items (L1–L4) resolved 2026-04-28, commit `8e76121`
- Coverage gaps addressed 2026-04-28, commit `b4cf31d`

All commits are on branch `fix/high-priority-code-review-findings`.

---

## Build & Test Health

| Check | Result |
|---|---|
| `go build ./...` | ✅ clean |
| `go vet ./...` | ✅ clean |
| `main.go` coverage | ❌ 0% |
| `cmd` coverage | ⚠️ 64.1% (up from 51%; zeros remain in `init`, `remediate`, `update` — require external I/O) |
| `duplication` coverage | ✅ 98.2% (up from 88%; removed unreachable dead guard) |
| `scorer` coverage | ✅ 94.3% |
| `reporter` coverage | ✅ 97.9% |
| `agents` coverage | ✅ 96% (init panic-branch untestable without reflection) |
| `analysis` coverage | ✅ 100% |
| `internal/tokenize` coverage | ✅ 100% (new package) |

---

## 🔴 High Priority

### H1 — `fileExists` duplicated across packages ✅ Fixed

`scorer/util.go:9` and `cmd/evaluate.go:132` both define an unexported `fileExists`. The `cmd` version is subtly wrong: it uses `os.Stat` without checking `IsDir`, so it returns `true` for directories. The scorer version is correct.

**Resolution:** Renamed the `cmd` version to `pathExists` across all cmd files (`evaluate.go`, `aggregate.go`, `duplication.go`, `trend.go`, `remediate.go`) and updated tests and comments. The name now accurately documents the intent (path existence, files or directories), removing the semantic confusion with scorer's files-only `fileExists`.

### H2 — Regex compiled on every call in `matchesRegexCI` ✅ Fixed

`scorer/util.go:15` calls `regexp.Compile` inside a function invoked on every scoring pass.

**Resolution:** Removed `matchesRegexCI` entirely. The two call sites in `d2_mindset_procedures.go` and one in `d3_anti_pattern.go` now use package-level `regexp.MustCompile` vars (`reD2MindsetHeader`, `reD2NumberedList`, `reD3BadGood`, `reD3AntiInstr`). Also hoisted the previously function-scoped `antiPat` compile in `scoreD3FromInstructions`. Tests updated accordingly.

### H3 — Duplicated tokenisation logic between packages ✅ Fixed

`analysis/tfidf.go` and `duplication/similarity.go` both defined identical `stopwords` maps and markdown-stripping regexes.

**Resolution:** Extracted shared logic into `internal/tokenize` (`Normalize`, `Set`, `Counts` functions plus the shared regexes and stopwords). Both `analysis` and `duplication` now import it. New package ships with 100% test coverage.

### H4 — Magic numbers in `scoreD1` scoring arithmetic ✅ Fixed

`scorer/d1_knowledge_delta.go:14-24`: the starting value (15), penalty (−2), and bonus (+1) were bare integer literals.

**Resolution:** Extracted `d1BaseScore = 15`, `d1PenaltyPerPat = 2`, `d1BonusPerPat = 1`, `d1Max = 20` as named constants. The clamp now references `d1Max` rather than a raw `20`.

---

## 🟡 Medium Priority

### M1 — Nested `append` chain allocates multiple intermediate slices ✅ Fixed

`scorer/scorer.go:37` had a quadruple-nested `append`.

**Resolution:** Pre-allocate `allDiags` with the correct capacity and append each slice sequentially.

### M2 — Inconsistent dimension function signatures ✅ Fixed

D2, D6, and D8 returned `int` only while all other dimensions returned `(int, []Diagnostic)`.

**Resolution:** All three updated to return `(int, []Diagnostic)`. D2 now emits a warning when all imperative/directive metrics are zero. D6 now emits a warning when the validator bridge is nil or no directive markers are found. D8 returns `nil` diagnostics (no new signals to surface). `scorer.go` updated to collect D2, D6, D8 diagnostics alongside the rest.

### M3 — `canonicalSkillKey` fails silently on out-of-tree paths ✅ Fixed

`cmd/evaluate.go` — when `skillPath` was outside `<repoRoot>/skills/`, `TrimPrefix` was a no-op.

**Resolution:** `canonicalSkillKey` now returns `(string, error)` and errors explicitly when the skill path is not under `<repoRoot>/skills/`. Updated callers in `evaluate.go` and `analyze.go`. Test updated to assert the error rather than accept any non-empty string.

### M4 — Map iteration in `store.go` is non-deterministic ✅ Fixed

`reporter/store.go` iterated `map[string][]byte` with randomised order.

**Resolution:** Replaced with an ordered slice of `{name, data}` structs — write order is now deterministic: `audit.json`, `Analysis.md`, `Remediation.md`.

### M5 — `interface{}` should be `any` (Go 1.18+) ✅ Fixed

**Resolution:** Replaced all `interface{}` occurrences with `any` in `scorer/d9_eval_validation.go`, `scorer/d3_anti_pattern.go`, and both packages' test files.

---

## 🟢 Low / Style

### ✅ L1 — Package-level Cobra flag vars leak state between tests

`cmd/evaluate.go` and `cmd/batch.go` store flag values in package-level `var`. This means test runs share flag state. The idiomatic fix is to scope flags to the command's `RunE` closure or use an options struct.

**Resolved:** Both commands refactored to local `flags` struct inside `init()`. Commit `8e76121`.

### ✅ L2 — No `context.Context` threading

No scorer or reporter function accepts `context.Context`. This prevents cancellation of `orchestrate.RunContentAnalysis` (an external package call that could block). Low priority for a CLI, but a ceiling on future testability.

**Resolved:** `scorer.Score` and `scorer.ScoreFromContent` now accept `context.Context` as first argument; CLI callers pass `cmd.Context()`, tests pass `t.Context()`. Commit `8e76121`.

### ✅ L3 — Negative TF-IDF scores possible in `analysis/tfidf.go`

`analysis/tfidf.go:86`: IDF is `log(N / (1+df))`. When `df == N` (term appears in every document), this returns a small negative number. Terms common to all documents should score 0, not negative. Change to `math.Max(0, math.Log(...))` or use the `log(1 + N/(1+df))` smoothed variant.

**Resolved:** Changed to `math.Max(0, math.Log(...))`. Added `TestExtractKeywords_IDFNeverNegative` test. Commit `8e76121`.

### ✅ L4 — No deduplication guard on `agents.Registry`

`agents/registry.go`: the global `Registry` slice has no init-time check for duplicate IDs. `ByID` silently returns the first match. An `init()` validation or constructor pattern would catch accidental duplicates at startup.

**Resolved:** Added `init()` panic guard that checks for duplicate IDs at startup. Commit `8e76121`.

---

## ✅ Coverage Gap — `cmd/` at 51% → 64%

Added tests for all tractable gaps (commit `b4cf31d`):

- `trend` helpers: `trendArrow`, `buildTrendEntry`, `groupAuditsBySkill`, `collectTrends`, `printTrendTable` (all were 0%)
- `batch` command RunE: single skill, multiple skills, JSON flag, `--fail-below` pass/fail, unknown grade
- `evaluate` command RunE: full/minimal skill, JSON flag, `--store` (verifies audit.json written), out-of-tree path error
- `validate` review helpers: `checkReviewFrontmatter`, `checkReviewMetadataLabels`, `checkReviewCommands` — both required and recommended branches
- `duplication.ShortKey` (was 0%), `Inventory` unreadable-file path, removed unreachable `union==0` guard from `Jaccard`
- `agentByID` wrapper (was 0%)

Remaining zeros (`cmd/init`, `cmd/remediate`, `cmd/update`, `cmd/root:Execute`) require external I/O (filesystem agent install paths, subprocess calls, GitHub API + binary download) and are not practical to unit-test without mocking infrastructure.

---

## Remediation Priority Order

1. ~~**H3** — deduplicate tokenisation~~ ✅ Done
2. ~~**H1** — fix and consolidate `fileExists`~~ ✅ Done
3. ~~**H4** — name D1 scoring constants~~ ✅ Done
4. ~~**H2** — pre-compile regexes~~ ✅ Done
5. ~~**M1** — flatten nested append chain~~ ✅ Done
6. ~~**M2** — align dimension function signatures~~ ✅ Done
7. ~~**M3** — guard `canonicalSkillKey` against out-of-tree paths~~ ✅ Done
8. ~~**M4** — deterministic write order in store.go~~ ✅ Done
9. ~~**M5** — replace `interface{}` with `any`~~ ✅ Done
10. Coverage: add tests for the `cmd` commands listed above (L-level, deferred)
