---
title: "Plan: Implement Shell Script Functionality in the Go CLI"
type: PLAN
status: DONE
date: 2026-04-27
value: HIGH
---
# Plan: Implement Shell Script Functionality in the Go CLI

**Date:** 2026-04-27  
**Status:** All phases complete — 2026-06-30

## Context

The reference documents under `skill/skill-quality-auditor/references/` describe four shell/TypeScript
scripts that are assumed to exist alongside the CLI. Since the project does not ship those scripts,
their functionality must live in the Go CLI itself.

---

## Scripts to Replace

| Script | Proposed command | Purpose |
|--------|-----------------|---------|
| `scripts/detect-duplication.sh` | `skill-auditor duplication` | Pairwise similarity across all skills; outputs `duplication-report-YYYY-MM-DD.md` |
| `scripts/plan-aggregation.ts` | `skill-auditor aggregate` | Consolidation candidates within a family prefix; outputs aggregation plan |
| `scripts/generate-remediation-plan.sh` | `skill-auditor remediate` | Per-skill remediation plan with T-shirt sized tasks from audit results |
| `scripts/validate-remediation-plan.sh` | `skill-auditor remediate --validate` | Validate a plan file against the JSON schema |

---

## New CLI Commands

### 1. `skill-auditor duplication [skills-dir]`

- Walk `skills/` (or given dir), inventory all `SKILL.md` files
- Pairwise text-similarity: word-level Jaccard after stripping markdown syntax and a small
  stopword list; structural overlap on section headers as a secondary signal
- Apply thresholds: Critical >35%, High >20% (per `duplication-detection-algorithm.md`)
- Output `duplication-report-YYYY-MM-DD.md` to `.context/analysis/`
- `--json` flag for machine-readable output
- Exit code 2 if any Critical pairs found (CI gate)

### 2. `skill-auditor aggregate --family <prefix> [skills-dir]`

- Filter skills matching the family prefix
- Run size + duplication analysis on the family
- Output `aggregation-plan-<family>-YYYY-MM-DD.md` to `.context/analysis/` following the
  6-step structure from `aggregation-implementation.md`
- `--dry-run` flag: print to stdout, no file write

### 3. `skill-auditor remediate <skill> [--target-score <N>]`

- Requires a prior stored audit in `.context/audits/<skill>/`
- Reads lowest-scoring dimensions, maps them to T-shirt sized tasks per `remediation-planning.md`
- Outputs `.context/plans/<skill>-remediation-plan.md` using embedded `remediation-plan-template.yaml`
- `--validate` flag: validates an existing plan file against `remediation-plan.schema.json` (no generation)

### 4. `skill-auditor trend [skills-dir]`

- Reads all stored audits from `.context/audits/`
- Computes score delta between the two most recent runs per skill
- Outputs trend table (↑ / ↓ / —) to stdout or `--json`

---

## Implementation Order

Dependencies flow left to right — implement in this sequence:

```
duplication  →  aggregate  →  remediate  →  trend
(no deps)      (uses dup)    (uses eval)   (uses stored audits)
```

`evaluate` and `batch` are unaffected. `remediate --validate` reuses the JSON schema
already embedded via `embed.go`.

---

## Actual File Layout

```
skill-auditor/cmd/
  duplication.go          # cobra command
  aggregate.go            # cobra command
  remediate.go            # cobra command
  trend.go                # cobra command
  assets/schemas/         # embedded remediation-plan + review-report schemas
  assets/templates/       # embedded remediation-plan + review-report templates

skill-auditor/duplication/   # NOTE: lives at top level, not under scorer/
  similarity.go           # TokenSet, Jaccard, StructuralSimilarity, Similarity
  inventory.go            # SkillEntry, Inventory(dir)
  detect.go               # Pair, Detect(entries), ThresholdCritical/High constants
  similarity_test.go      # 8 unit tests

skill-auditor/reporter/
  duplication.go          # DuplicationReport(pairs, entries, date)
  aggregation.go          # AggregationPlan(family, entries, pairs, date)
  remediation_plan.go     # RemediationPlan(result, target, auditPath, date)
                          # ValidateRemediationPlan(planPath) — no external deps
```

**Deviation from plan:** `duplication/` was placed at `skill-auditor/duplication/` (sibling of `scorer/`) rather than `skill-auditor/scorer/duplication/`, keeping similarity logic independent of the D1–D9 scorer interface.

---

## Decisions

| # | Question | Decision |
|---|----------|----------|
| 1 | `trend` command scope | Include in this iteration |
| 2 | Similarity algorithm | Word-level Jaccard + markdown stripping + stopwords; no TF-IDF |
| 3 | `remediate --validate` | Required from day one; use `encoding/json` + manual schema walk, no external deps |

---

## Phase 2 — Remaining Script Gaps (2026-04-28)

The 17 scripts in `skill/skill-quality-auditor/scripts/` were audited against the CLI.
Eight scripts are already fully covered (evaluate, batch, duplication, aggregate, remediate, trend).
Nine remain. Of those, four are high/medium priority for Go porting; five are low-value or
require external tooling (Python/vector DBs) and are explicitly out-of-scope.

### Gap table

| Script | Status | CLI target |
|--------|--------|-----------|
| `validate-skill-artifacts.sh` | **Done** — `cmd/validate.go` | `skill-auditor validate artifacts [paths...]` |
| `validate-review-format.sh` | **Done** — `cmd/validate.go` | `skill-auditor validate review <file> [--strict-recommended]` |
| `check-consistency.sh` | **Done** — `cmd/lint.go` | `skill-auditor lint [skills-dir]` |
| `prune-audits.sh` | **Done** — `cmd/prune.go` | `skill-auditor prune [--keep N]` |
| `detect-duplication-enhanced.sh` | Out of scope | Basic Jaccard already covers the core case |
| `semantic-analysis.sh` | **Done** — `cmd/analyze.go --semantic` | `skill-auditor analyze <skill> --semantic` |
| `ml-pattern-detection.sh` | **Done** — `cmd/analyze.go --patterns` | `skill-auditor analyze <skill> --patterns` |
| `pattern-recognition-pipeline.sh` | **Done** — `cmd/analyze.go --pipeline` | `skill-auditor analyze <skill> [--pipeline]` |
| `tessl-compliance-check.sh` | Deferred | Future `skill-auditor compliance` command |

### New commands (Phase 2)

#### `skill-auditor validate artifacts [paths...]`

Ports `validate-skill-artifacts.sh`. Walks `skills/` or accepts explicit file paths.

- **`check_file`** routes by subdirectory context:
  - `assets/templates/` — YAML validity (best-effort; no external dep required)
  - `assets/schemas/` — must use `.schema.json` extension, valid JSON, must contain `"$schema"` from json-schema.org
  - `scripts/` — shebang check per type (`.sh` → `#!/usr/bin/env sh` or `#!/usr/bin/env bash` + `# shell: bash`; `.py` → `#!/usr/bin/env python3`; `.ts` → bun shebang; `.js` → node shebang)
- **`check_skill_dir`** for each `skills/<domain>/<name>/` dir:
  - `assets/` subdirs must be one of `templates`, `schemas`, `requirements`, `examples`
  - No YAML files directly under `assets/` (must go in `assets/templates/`)
  - `SKILL.md` ≤ 500 lines
  - Frontmatter `name:` must match directory name
  - No `../` path references outside fenced code blocks
- Exit code 1 on any error; prints `ERROR:` prefix per violation

#### `skill-auditor validate review <file> [--strict-recommended]`

Ports `validate-review-format.sh`. Reads `review-report.requirements.json` from embedded assets.

- Extracts H1 title, H2 headings, YAML frontmatter from report file
- Checks required title prefix, required/recommended frontmatter keys, required metadata labels
- Checks required H2 groups exist and appear in the mandated order
- Checks recommended H2 groups (warnings by default; errors with `--strict-recommended`)
- Checks required/recommended dimension labels and commands are present
- Exit code 1 on errors; warnings to stderr only

#### `skill-auditor lint [skills-dir]`

Ports `check-consistency.sh`. Scans `skills/` (or given dir).

- Each skill dir must contain `SKILL.md`
- `SKILL.md` must contain a frontmatter block (`---`)
- Scripts under `scripts/` must use `#!/usr/bin/env sh` shebang
- Prints `MISSING_SKILL`, `NO_FRONTMATTER`, `BAD_SHEBANG` tags per issue
- Exits with the issue count (0 = clean)

#### `skill-auditor prune [--keep N]`

Ports `prune-audits.sh`. Cleans `.context/audits/` keeping the N most recent date-dirs per skill.

- `--keep` default: 5
- Walks each `<skill>/` subdirectory, sorts date dirs descending, removes beyond keep threshold
- Preserves `latest` symlinks
- Prints `Keeping:` / `Removing:` per directory; summary at end

### New asset

`review-report.requirements.json` copied from `skill/skill-quality-auditor/assets/requirements/` to
`skill-auditor/cmd/assets/requirements/` and embedded via `embeddedRequirements` in `embed.go`.

### Implementation order

```
validate artifacts  →  validate review  →  lint  →  prune
(no deps)              (needs requirements asset)   (no deps)  (no deps)
```

### Phase 2 delivery (2026-04-28)

All four commands implemented, tested, and committed.

| Commit | Description |
|--------|-------------|
| `edf7846` | feat: add validate, lint, and prune commands |
| `1611717` | test: comprehensive tests for validate, lint, and prune |

**Test coverage:** 122 tests passing across cmd package. New functions at 86–100% coverage.
Remaining gaps are dead-code paths (empty `recommended_metadata_labels` slice in requirements JSON)
or pre-existing uncovered commands (remediate, trend, init) outside Phase 2 scope.

**TDD note:** tests were added immediately after implementation in the same session; future work
on this repo must follow TDD discipline — write tests alongside every new Go file.

---

## Phase 3 — Semantic Analysis, Pattern Detection, Pipeline (2026-04-28)

The three scripts previously marked "Out of scope" (requiring external NLP/ML tooling) are now
ported as pure-Go, stdlib-only commands under `skill-auditor analyze`.

### New packages

| Package | Files | Purpose |
|---------|-------|---------|
| `skill-auditor/analysis/` | `tfidf.go`, `patterns.go` | TF-IDF keyword extractor + rule-based pattern detectors |
| `skill-auditor/reporter/` | `combined_analysis.go` | CombinedAnalysis struct + markdown/JSON formatters |
| `skill-auditor/cmd/` | `analyze.go` | `analyze` cobra command |

### New command

```
skill-auditor analyze <skill> [--semantic] [--patterns] [--pipeline] [--json] [--store] [--limit N]
```

- `--semantic`: TF-IDF keyword extraction (ports `semantic-analysis.sh`)
- `--patterns`: rule-based pattern detection — section presence, trigger frequency, structural conformance, anti-pattern signals (ports `ml-pattern-detection.sh`)
- `--pipeline` / no flag: full pipeline, combined report to `.context/analysis/` (ports `pattern-recognition-pipeline.sh`)

### Phase 3 delivery (2026-04-28)

All three commands implemented via parallel worktrees A+B (independent), then C (dependent).

| Commit | Description |
|--------|-------------|
| `a80182d` | feat: add rule-based pattern detectors to analysis package |
| `4d31560` | feat: add TF-IDF keyword extractor to analysis package |
| `f7f60bb` | feat: add analyze command and combined analysis reporter |

**Test coverage:** analysis 100%, reporter 97.9%, cmd 57.0% (pre-existing untested commands account for the gap).

---

## Phase 4 — Command Consolidation (completed)

Cross-command analysis revealed duplication between existing commands that should be resolved
before adding further functionality.

### Merge candidates

#### 1. `lint` → `validate artifacts` (High priority — true duplication)

`lint` is a strict subset of `validate artifacts`. Both walk `skills/`, check SKILL.md existence,
check frontmatter presence, and check script shebangs. `validate artifacts` is stricter on every
axis (name-matches-dir, no `../` refs, per-type shebang rules, schema/template checks).

The only thing `lint` adds that `validate artifacts` lacks is the `MISSING_SKILL` tag for a
skill dir with no `SKILL.md`. That check should be absorbed into `validate artifacts`.

**Proposed action:**
- Add `MISSING_SKILL` detection to `validate artifacts` walk
- Deprecate `lint`; keep it as a thin alias (`lint` → `validate artifacts`) for one release,
  then remove `cmd/lint.go`
- Update CI references from `skill-auditor lint` to `skill-auditor validate artifacts`

**Basis:** `lint` adds no unique value; having two commands that check the same properties
creates confusion about which to use in CI pipelines.

#### 2. `evaluate` vs `batch` (Low priority — ergonomic overlap only)

`batch` with one argument is functionally identical to `evaluate`. Differences are cosmetic:
`evaluate` returns a per-dimension report; `batch` returns a sorted summary table with an
optional `--fail-below` grade gate.

**Proposed action:** Keep both commands. Document that `batch` accepts a single skill and that
`--fail-below` is the idiomatic CI gate. No code change needed.

**Basis:** The conceptual distinction (detailed inspect vs bulk CI gate) justifies separate
commands. Merging would make `evaluate` flags more complex without user benefit.

#### 3. `analyze` flag semantics (No action needed)

`--pipeline` / no flag runs both semantic + pattern analysis. `--semantic` and `--patterns`
exist for targeted use. The design is correct; the only risk is that bare `analyze <skill>`
silently runs the full pipeline — consider adding a short header line clarifying mode in output.

### Implementation order

```
validate artifacts (add MISSING_SKILL)  →  deprecate lint  →  remove lint
```

### Phase 4 delivery (2026-04-28)

| Commit | Change | File | Detail |
|--------|--------|------|--------|
| `1906a74` | `MISSING_SKILL` check | `cmd/validate.go` | Added `else` branch in `walkSkillDirs` — emits `ERROR: MISSING_SKILL: <rel>/SKILL.md` |
| `1906a74` | Deprecation alias | `cmd/lint.go` | Replaced full implementation with cobra `Deprecated` field + delegate to `validateArtifactsCmd.RunE(cmd, nil)` |
| `1906a74` | New test | `cmd/validate_test.go` | `TestWalkSkillsDir_missingSkillMD` |
| `1906a74` | Rewritten tests | `cmd/lint_test.go` | Two alias tests replacing the eight original lint-logic tests (logic now covered by validate tests) |
| `c35a566` | Remove lint entirely | `cmd/lint.go` deleted | Deprecation window elapsed; `cmd/lint.go` and its test removed from the tree |

### Decisions

| # | Question | Decision |
|---|----------|----------|
| 1 | Keep `lint` as alias? | One release only (1906a74), then removed entirely (c35a566) |
| 2 | Merge `evaluate` into `batch`? | No — different output format serves different use cases |
| 3 | Pass skills-dir arg through alias? | No — `validate artifacts` takes file paths, not a dir; alias calls with `nil` args (repo-root auto-detected) |

---

## Phase 5 — Asset Hygiene: Shell Ref Replacement & Eval Format Migration (in progress)

Branch: `fix/assets-shell-refs-and-evals`

### Motivation

Two classes of stale references remained in embedded assets after Phases 1–4:

1. **Shell script references** — several `references/*.md` documents still cited `scripts/detect-duplication.sh`, `scripts/plan-aggregation.ts`, etc. rather than the Go CLI commands that replaced them.
2. **Old eval format** — evals used the retired multi-file layout (`scenario-N/task.md` + `criteria.json` + `capability.txt` + `instructions.json` + `summary.json`). The canonical format is flat `scenario-NN.md` files per the `eval-scenario-format` reference.

### Changes (commit `67db3af`)

| Area | Before | After |
|------|--------|-------|
| `references/*.md` (in both `skill/skill-quality-auditor/references/` and `skill-auditor/cmd/assets/references/`) | `bun run scripts/…`, `bash scripts/…` invocations | `skill-auditor <command>` equivalents |
| `evals/scenario-N/` subdirs (task.md + criteria.json + capability.txt) | Old multi-file layout × 5 scenarios | Removed |
| `evals/instructions.json` + `evals/summary.json` | Meta-artifacts from retired framework | Removed |
| `evals/scenario-NN.md` flat files | Did not exist | Added × 5 scenarios (01–05), in both `skill/skill-quality-auditor/evals/` and `skill-auditor/cmd/assets/evals/` |

### Scope

- No Go source changes; pure asset and documentation update.
- Both canonical (`skill/skill-quality-auditor/`) and embedded copy (`skill-auditor/cmd/assets/`) updated in lockstep per CLAUDE.md rules.

### Phase 5 delivery

Merged via `3560e4b`.

---

## Phase 6 — Skill Consolidation & CI Quality Gate (in progress)

Branch: `feat/phase6-skill-consolidation`

### Motivation

After Phase 5, `skill/skill-quality-auditor/` was a redundant directory: every asset it contained
was already mirrored under `skill-auditor/cmd/assets/`, and every script it contained had a Go CLI
equivalent. The dual-location model created a maintenance burden (two copies to keep in sync) and
confused the tile manifest, which required `tile.json` to sit inside the `skill/` subtree.

Additionally, the deferred `tessl-compliance-check.sh` → `skill-auditor compliance` command was
re-evaluated. Tessl registry compliance is a publish-gate concern, not a skill quality concern.
It belongs in CI, not in the CLI.

### Changes

| Area | Change |
|------|--------|
| `skill/skill-quality-auditor/` | **Deleted** — all assets already in `skill-auditor/cmd/assets/` |
| `tile.json` | Moved to repo root; `skills.path` updated to `skill-auditor/cmd/assets/SKILL.md` |
| `.github/workflows/skill-quality.yml` | New workflow: build → test → validate artifacts → duplication → batch --fail-below B → tessl eval run |
| `lefthook.yml` | `pre-push` extended with `skill-validate`, `skill-duplication`, `skill-batch` commands scoped to `skill-auditor/cmd/assets/**` changes; `go-build` updated to write `dist/skill-auditor` |
| `CLAUDE.md` | Mirror-copy rules removed; all asset paths updated to `skill-auditor/cmd/assets/`; Tessl tile workflow updated |

### Won't-do: `skill-auditor compliance`

The planned `compliance` command (porting `tessl-compliance-check.sh`) is explicitly **not implemented**.
Rationale: Tessl registry compliance is enforced by the GH workflow (`tessl eval run`) and the
`TESSL_TOKEN`-gated publish step. Duplicating that logic in the CLI adds no value for local development
and couples the auditor to Tessl internals.

### Phase 6 addendum — build artifact path (`dist/`)

After Phase 6 merged, the build output path was consolidated to `dist/skill-auditor` at the repo root,
aligning local builds with the release workflow (which already wrote to `../dist/`).

| File | Change |
|------|--------|
| `.gitignore` (root, new) | Added; ignores `dist/` |
| `skill-auditor/.gitignore` | Removed `bin/` entry (now covered by root `.gitignore`) |
| `lefthook.yml` | `go-build` writes `../dist/skill-auditor`; `skill-validate/duplication/batch` run via `./dist/skill-auditor` |
| `.github/workflows/skill-quality.yml` | Build step writes `../dist/skill-auditor`; subsequent steps use `./dist/skill-auditor` |
| `CLAUDE.md` / `AGENTS.md` | All binary references updated from `bin/` to `dist/` |
