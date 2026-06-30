---
title: "Go Code Review Findings — 2026-04-28"
type: finding
status: done
date: 2026-04-28
---
# Go Code Review Findings — 2026-04-28

Critical review of `skill-auditor/` packages: `scorer/`, `duplication/`, `reporter/`, `cmd/`.

Remediated 2026-04-28. All tests pass (`go test ./...` — 5 packages, 0 failures, 1 expected skip).

---

## Critical

### 1. `shortKey()` undefined in `reporter/aggregation.go:72` ✅ Fixed

`aggregation.go` called `shortKey(e.Key)` but the function only existed in `duplication/detect.go`
as an unexported symbol — a compile error.

**Resolution:** Exported as `duplication.ShortKey` in `duplication/detect.go`. Updated
`reporter/aggregation.go:72` to call `duplication.ShortKey(e.Key)`.

---

### 2. Silent error swallow in `cmd/init.go:60` ✅ Fixed

`_ = os.Remove(dest)` dropped removal errors, producing confusing downstream failures.

**Resolution:** Replaced with:
```go
if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
    return fmt.Errorf("[%s] remove existing file: %w", a.ID, err)
}
```

---

## High

### 3. Hand-rolled O(n²) bubble sort in `duplication/detect.go:39-45` ✅ Fixed

Custom bubble sort replaced with:
```go
sort.Slice(pairs, func(i, j int) bool { return pairs[i].Similarity > pairs[j].Similarity })
```

---

### 4. Discarded `MarkFlagRequired` error in `cmd/aggregate.go:104` ✅ Fixed

`_ = aggregateCmd.MarkFlagRequired("family")` replaced with the standard cobra init-time
panic pattern:
```go
if err := aggregateCmd.MarkFlagRequired("family"); err != nil {
    panic(err)
}
```

---

### 5. Dangling `.tmp` files in `reporter/store.go:39` ✅ Fixed

Cleanup error on rename failure now logs to stderr without masking the rename error:
```go
if removeErr := os.Remove(tmp); removeErr != nil && !os.IsNotExist(removeErr) {
    fmt.Fprintf(os.Stderr, "store: cleanup temp file %s: %v\n", tmp, removeErr)
}
```

---

### 6. No error-path test coverage ✅ Fixed

Added error-path tests to:
- `cmd/evaluate_test.go` — missing `SKILL.md`, non-existent path
- `cmd/duplication_test.go` — non-existent skills directory
- `reporter/store_test.go` — unwritable destination
- `reporter/remediation_plan_test.go` — `targetScore` ≤ current score (skipped pending Finding 7)

---

### 7. No bounds validation in `reporter/remediation_plan.go:77-86` ✅ Fixed

Added guard before the default-calculation block:
```go
if targetScore > 0 && targetScore <= r.Total {
    return "", fmt.Errorf("targetScore %d must exceed current score %d", targetScore, r.Total)
}
```
The previously skipped test in Finding 6 now passes with this guard in place.

---

## Medium

### 8. Full inventory loaded into memory — `duplication/inventory.go:17-18` ✅ Documented

Eager loading is acceptable at current scale. Added doc comment:
```
// All file content is loaded eagerly; for repos with >500 skills, consider lazy loading.
```
Deferred as a future optimisation.

---

## Low

### 9. Dead function `latestAuditJSON()` in `cmd/trend.go:183` ❌ Finding Incorrect

**Correction:** `latestAuditJSON` is actively called from `cmd/remediate.go:45` and tested
in `cmd/evaluate_test.go:235,249,262`. It is not dead code. Original finding was wrong.

---

### 10. Ambiguous exit code in `cmd/batch.go` ✅ Fixed

Added `var storeErrors []string` accumulator. After output, returns an aggregated error:
```go
if len(storeErrors) > 0 {
    return fmt.Errorf("store failed for %d skill(s): %s", len(storeErrors), strings.Join(storeErrors, "; "))
}
```

---

## Final Status

| # | Severity | Finding | Status |
|---|----------|---------|--------|
| 1 | Critical | `shortKey` compile error | ✅ Fixed |
| 2 | Critical | Silent `os.Remove` in init.go | ✅ Fixed |
| 3 | High | Bubble sort in detect.go | ✅ Fixed |
| 4 | High | Discarded MarkFlagRequired error | ✅ Fixed |
| 5 | High | Dangling .tmp files in store.go | ✅ Fixed |
| 6 | High | Missing error-path tests | ✅ Fixed |
| 7 | High | No targetScore validation | ✅ Fixed |
| 8 | Medium | Eager inventory loading | ✅ Documented |
| 9 | Low | Dead `latestAuditJSON` function | ❌ Finding incorrect — function is used |
| 10 | Low | Ambiguous batch exit code | ✅ Fixed |
