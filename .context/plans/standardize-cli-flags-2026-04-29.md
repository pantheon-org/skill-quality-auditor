---
title: "Plan: Standardise CLI Flags and Output Formats"
type: plan
status: done
date: 2026-04-29
---
# Plan: Standardize CLI Flags and Output Formats

**Status: COMPLETED 2026-04-29** — All 9 PRs merged (#52, #54, #56, #57, #58, #60, #61, #63, #64).

**Goal:** Apply a consistent flag interface across all commands so users can predict what flags are
available without consulting help text.

---

## Proposed Standard

| Flag | Short | Type | Applies to | Rule |
|------|-------|------|-----------|------|
| `--json` | `-j` | bool | output-producing commands | emit JSON (default on most commands) |
| `--markdown` | `-m` | bool | output-producing commands | emit Markdown instead of JSON |
| `--store` | `-s` | bool | report-producing commands | persist output to `.context/` |
| `--repo-root` | `-r` | string | all commands that resolve paths | auto-detected if empty |
| `--dry-run` | `-n` | bool | commands that write or delete files | preview actions without side effects |
| `--json` on `evaluate` | — | — | `evaluate` only | **remove** — JSON is the default; flag is a no-op |

### Per-flag shorthand rationale

- `-j` / `-m` — first letters of the format names; universally readable
- `-s` — "store"; consistent with `--store` meaning "save to disk"
- `-r` — "root"; short and unambiguous within each command
- `-n` — Unix dry-run convention (`make -n`, `rsync -n`); `-d` avoided as it's used for `--skills-dir`
- `-d` — "directory"; reserved for `--skills-dir` on commands that have it
- `-f` — "family"; used on `aggregate`
- `-F` — "fail"; uppercase to avoid collision with `-f` on `batch` (different command, but consistent rule)
- `-t` — "target"; used on `remediate`
- `-v` — "validate"; used on `remediate`
- `-k` — "keep"; used on `prune`
- `-l` — "limit"; used on `analyze`
- `-S` — "strict"; used on `validate review` (uppercase avoids `-s`/`--store` collision)
- `-g` — "global"; already exists on `init`

### Full shorthand map

| Long flag | Short | Commands |
|-----------|-------|---------|
| `--json` | `-j` | `analyze`, `batch`, `duplication`, `trend`, `remediate`, `aggregate` |
| `--markdown` | `-m` | `evaluate`, `analyze`, `batch`, `duplication`, `trend`, `remediate` |
| `--store` | `-s` | `evaluate`, `analyze`, `batch`, `duplication`, `trend` |
| `--repo-root` | `-r` | `evaluate`, `analyze`, `batch`, `duplication`, `trend`, `remediate`, `aggregate`, `prune`, `validate` |
| `--dry-run` | `-n` | `aggregate`, `remediate`, `prune`, `init` |
| `--skills-dir` | `-d` | `aggregate`, `duplication` |
| `--family` | `-f` | `aggregate` |
| `--fail-below` | `-F` | `batch` |
| `--target-score` | `-t` | `remediate` |
| `--validate` | `-v` | `remediate` |
| `--keep` | `-k` | `prune` |
| `--limit` | `-l` | `analyze` |
| `--strict-recommended` | `-S` | `validate review` |
| `--global` | `-g` | `init` (pre-existing) |
| `--agent` | `-a` | `init` (discovered during implementation) |
| `--method` | `-m` | `init` (discovered during implementation; note: `-m` is also `--markdown` on other commands) |
| `--semantic` | `-e` | `analyze` (`-s` taken by `--store`) |
| `--patterns` | `-p` | `analyze` |
| `--pipeline` | `-P` | `analyze` |

---

## Per-Command Changes

### `evaluate`

- **Remove** `--json` flag (JSON is the default; `--markdown` already inverts it)

### `analyze`

- **Add** `--json` flag (emit JSON output — currently the default but undiscoverable)
  - Note: `--json` and `--markdown` are mutually exclusive; `--json` wins if both set (or error)

### `batch`

- **Add** `--markdown` flag (emit multi-skill summary as Markdown table)

### `duplication`

- **Add** `--markdown` flag
- **Add** `--store` flag (persist to `.context/analysis/duplication-report-YYYY-MM-DD.md`)

### `trend`

- **Add** `--markdown` flag
- **Add** `--store` flag (persist to `.context/audits/trend-YYYY-MM-DD.md`)

### `remediate`

- **Add** `--json` flag (emit plan as JSON)
- **Add** `--markdown` flag (emit plan as Markdown — this is already the natural format, so make
  it the default and have `--json` override)
- **Add** `--dry-run` flag (print plan to stdout without writing to `.context/plans/`; mirrors `aggregate`)

### `aggregate`

- **Add** `--json` flag (emit aggregation plan as JSON)
  - Note: `--dry-run` currently prints Markdown to stdout; keep that behaviour, `--json` overrides

### `prune`

- **Add** `--repo-root` flag (needed to resolve `.context/audits/`)
- **Add** `--dry-run` flag (print which audit runs would be removed without deleting anything)

### `validate`

- **Add** `--repo-root` flag (needed to resolve skill paths)

### `init`

- **Add** `--dry-run` flag (print what directories and files would be created/overwritten without touching disk)

---

## Default Output Behaviour (after changes)

| Command | Default output | `--markdown` | `--json` |
|---------|---------------|--------------|---------|
| `evaluate` | JSON | Markdown | — (removed) |
| `analyze` | JSON | Markdown | explicit no-op (self-documents the default) |
| `batch` | JSON | Markdown | ✓ (existing) |
| `duplication` | JSON | Markdown | ✓ (existing) |
| `trend` | JSON | Markdown | ✓ (existing) |
| `remediate` | Markdown | — (already default) | JSON |
| `aggregate` | Markdown (dry-run) / file | — | JSON |
| `prune` | text (table) | — | — |
| `validate` | text (pass/fail) | — | — |

> `prune` and `validate` produce structural CLI output (not domain reports), so `--json`/`--markdown`
> are out of scope for this plan.

---

## Mutual Exclusion Rule

Where both `--json` and `--markdown` exist on the same command, treat them as mutually exclusive:

- If both are passed → exit with error: `--json and --markdown are mutually exclusive`
- Enforce this with a shared helper in `cmd/` (e.g. `resolveOutputFormat`).

---

## Implementation Phases

### Phase 1 — Remove noise (no behaviour change)

- [x] Remove `--json` from `evaluate` (JSON remains the default)

### Phase 2 — Add missing `--repo-root`

- [x] `prune`: add `--repo-root`
- [x] `validate`: add `--repo-root`

### Phase 3 — Add missing `--markdown`

- [x] `batch`: add `--markdown`
- [x] `duplication`: add `--markdown`
- [x] `trend`: add `--markdown`

### Phase 4 — Add missing `--store`

- [x] `duplication`: add `--store` (path: `.context/analysis/duplication-report-YYYY-MM-DD.md`)
- [x] `trend`: add `--store` (path: `.context/audits/trend-YYYY-MM-DD.md`)

### Phase 5 — Add missing `--json` to write-oriented commands

- [x] `remediate`: add `--json` (Markdown stays default)
- [x] `aggregate`: add `--json` (Markdown stays default for dry-run)
- [x] `analyze`: add explicit `--json` flag as a self-documenting no-op

### Phase 6 — Add `--dry-run` to write/delete commands

- [x] `prune`: add `--dry-run` (list runs that would be removed, no `os.RemoveAll`)
- [x] `remediate`: add `--dry-run` (print plan to stdout, no `os.WriteFile`)
- [x] `init`: add `--dry-run` (print files that would be created/symlinked, no `os.MkdirAll`/`os.WriteFile`)

### Phase 7 — Mutual exclusion helper + enforcement

- [x] Add `resolveOutputFormat(cmd)` helper in `cmd/output.go`
- [x] Wire it into all commands that have both `--json` and `--markdown`

### Phase 8 — Add shorthands to all flags (new and existing)

- [x] Convert all `Flags().Bool/String/Int(name, ...)` calls to `Flags().BoolP/StringP/IntP(name, short, ...)` using the shorthand map above
- [x] Covers every command: `evaluate`, `analyze`, `batch`, `duplication`, `trend`, `remediate`, `aggregate`, `prune`, `validate`, `init`
- [x] `--global/-g` on `init` already correct — verify and leave

### Phase 9 — Tests and docs

- [x] Update/add unit tests for each changed command (flag registration + output format switching)
- [x] Update `README.md` command reference table

---

## Files Touched

| File | Change | PR |
|------|--------|----|
| `cmd/evaluate.go` | removed `--json`; added `-m`/`-s`/`-r` shorthands | #52 |
| `cmd/evaluate_cmd_test.go` | replaced `--json` test; added shorthand tests | #52 |
| `cmd/analyze.go` | added `--json/-j`; added all shorthands (`-e`,`-p`,`-P`,`-m`,`-s`,`-r`,`-l`) | #54 |
| `cmd/analyze_test.go` | updated flag helpers; added mutual exclusion + shorthand tests | #54 |
| `cmd/batch.go` | added `--markdown/-m`; added `-j`/`-s`/`-F`/`-r` shorthands | #56 |
| `cmd/batch_test.go` | added shorthand + mutual exclusion tests | #56 |
| `cmd/duplication.go` | added `--markdown/-m`, `--store/-s`; added `-j`/`-d`/`-r` shorthands | #57 |
| `cmd/duplication_test.go` | added store + markdown + mutual exclusion tests | #57 |
| `cmd/trend.go` | added `--markdown/-m`, `--store/-s`; added `-j`/`-r` shorthands | #57 |
| `cmd/trend_test.go` | added store + markdown + mutual exclusion tests | #57 |
| `cmd/remediate.go` | added `--json/-j`, `--dry-run/-n`; added `-t`/`-v`/`-r` shorthands | #58 |
| `cmd/remediate_test.go` | added dry-run + JSON output tests | #58 |
| `reporter/remediation_plan_generate.go` | added `RemediationPlanJSON` + JSON struct tags | #58 |
| `cmd/aggregate.go` | added `--json/-j`; added `-n`/`-f`/`-d`/`-r` shorthands | #60 |
| `cmd/aggregate_test.go` | added JSON output + skip-file tests | #60 |
| `reporter/aggregation.go` | added `AggregationPlanAsJSON` + `AggregationPlanJSON` struct | #60 |
| `cmd/prune.go` | added `--repo-root/-r`, `--dry-run/-n`; added `-k` shorthand | #61 |
| `cmd/prune_test.go` | added dry-run + repo-root tests | #61 |
| `cmd/validate.go` | added `--repo-root/-r`; added `-S` shorthand | #61 |
| `cmd/validate_test.go` | added repo-root + `-S` shorthand tests | #61 |
| `cmd/init.go` | added `--dry-run/-n`; added `-a`/`-m` shorthands (discovered: `--agent`, `--method`) | #63 |
| `cmd/init_test.go` | added dry-run + shorthand tests | #63 |
| `cmd/output.go` | new — `OutputFormat` type + `resolveOutputFormat` helper | #64 |
| `cmd/output_test.go` | new — 5 test cases for helper | #64 |
| `cmd/analyze.go` | refactored to use `resolveOutputFormat` | #64 |
| `cmd/batch.go` | refactored to use `resolveOutputFormat` | #64 |
| `cmd/duplication.go` | refactored to use `resolveOutputFormat` | #64 |
| `cmd/trend.go` | refactored to use `resolveOutputFormat` | #64 |
| `README.md` | added Flag Shorthands section + updated per-command flag listings | #64 |
