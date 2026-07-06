---
title: "Plan: Externalise Hardcoded Scoring Patterns to YAML Config"
type: PLAN
status: DONE
date: 2026-07-03
value: MEDIUM
related:
  - ../findings/yaml-content-validation-config-2026-07-03.md
  - ../../analysis/patterns.go
  - ../../scorer/dimensions.go
  - ../../scorer/d1_knowledge_delta.go
  - ../../scorer/d3_anti_pattern_coverage.go
  - ../../scorer/d6_freedom_calibration.go
  - ../../scorer/d7_pattern_recognition.go
  - ../../cmd/analyze.go
  - ../../cmd/validate.go
  - ../../cmd/embed.go
---
# Plan: Externalise Hardcoded Scoring Patterns to YAML Config

## Goal

Move the ~12 hardcoded pattern lists spread across `analysis/patterns.go`, `scorer/d1_knowledge_delta.go`, `scorer/d3_anti_pattern_coverage.go`, and `scorer/d6_freedom_calibration.go` into a single embedded YAML config under `cmd/assets/assets/config/`, loaded at init and consumed by each pattern's detection function — making patterns auditable, testable, and changeable without recompilation.

## Critical Review of the Finding

The originating finding (`.context/findings/yaml-content-validation-config-2026-07-03.md`) proposes externalising hardcoded content-pattern rules into a YAML config. The pattern categories it defines (SEC_DISABLE, SEC_PERMISSIVE, CRED_EXFIL, etc.) are conceptually valid and DO map to existing D3 and D6 dimensions — but the finding contains several factual errors and design issues corrected here.

### 1. `scripts/validate-skill-content.sh` never existed in this repo

No shell script named `validate-skill-content.sh` has ever existed in this repo's git history. However, the 8 pattern categories it describes are a useful framework that maps well to existing dimensions:

| Category | Concept | Dimension | Rationale |
|----------|---------|-----------|-----------|
| SEC_DISABLE | Security-disabling instructions | **D3** (Anti-Pattern Coverage) | Instructing an agent to disable security is a critical anti-pattern |
| SEC_PERMISSIVE | Overly permissive instructions | **D6** (Freedom Calibration) | Calibrating permissiveness is core to D6 scoring |
| CRED_EXFIL / CRED_EXFIL_REV | Credential exfiltration | **D3** (Anti-Pattern Coverage) | Exfiltrating credentials is a critical anti-pattern |
| OBFUSC_B64 / UNICODE / HEX | Obfuscated content | **D3** (Anti-Pattern Coverage) | Obfuscation is an anti-pattern signal |
| TOOL_BROAD | Wildcard tool permissions | **D6** (Freedom Calibration) | Unconstrained tool access relates to calibration |
| URL allowlist | Approved external domains | — | Policy/convention check, not a dimension |

The script may have existed in a sister repo or the user's broader workflow. The pattern concepts are sound — this plan adopts them as a **Phase 2 extension** to D3 and D6.

### 2. No URL allowlists exist in the codebase

The finding claims `cmd/analyze.go` contains URL allowlists. It does not — it has section header lists and trigger word maps. No URL allowlist exists in any file. The `url_allowlist` YAML section in the finding is speculative but could form a future `validate content` subcommand.

### 3. Three pattern domains, not one

The finding conflates three domains under a single "content-safety" umbrella. They should be split:

| Location | Pattern Type | Domain | Phase |
|----------|-------------|--------|-------|
| `analysis/patterns.go` | hedge/vague/passive word lists | Quality analysis signals | **1** |
| `scorer/d1_knowledge_delta.go` | beginner/expert register patterns | Scoring heuristic (D1) | **1** |
| `scorer/d3_anti_pattern_coverage.go` | anti-pattern structural regexes | Scoring logic (D3) | — (regex-coupled) |
| `scorer/d6_freedom_calibration.go` | when-not-to-use patterns | Scoring logic (D6) | **1** |
| Content-safety categories (SEC_*, CRED_*, OBFUSC_*, TOOL_*) | Security scanning | D3 / D6 extension | **2** |

### 4. TOML support is unnecessary

The finding proposes JSON → YAML → TOML fallback with a new dependency (`github.com/BurntSushi/toml`). JSON is stdlib. YAML v3 is already an indirect dependency. TOML adds a direct dependency for marginal benefit — YAML is already available and more human-writable than JSON. The plan adopts YAML only.

### 5. The `rule_types` section mixes metadata with data

The finding embeds type system definitions in the config YAML. The Go type system already constrains valid rule types. These definitions belong in the JSON Schema or Go code, not in the config. Removed.

### 6. Config path inconsistency

The finding puts the schema at `cmd/assets/assets/schemas/` (correct) but the config at `cmd/assets/assets/validation/` (no such directory). Following existing conventions: config → `cmd/assets/assets/config/`, schema → `cmd/assets/assets/schemas/`.

### 7. Two `stripCodeBlocks` implementations exist (verified)

`scorer/dimensions.go:183` has `removeCodeBlocks` and `analysis/patterns.go:47` has `stripCodeBlocks`. Both strip fenced code blocks identically but live in separate packages. This is relevant because all Phase 2 content-safety patterns require code-block stripping (`strip_code_blocks: true`). The duplicate should be consolidated as part of Phase 1 — move to `internal/patternconfig/` or a shared utility.

### 8. All 8 content-safety regex patterns compile (verified)

Each of the 8 regex patterns from the finding was tested with `regexp.Compile` and compiles successfully in Go. The OBFUSC_B64 pattern `[A-Za-z0-9+/]{50,}={0,2}` is the most aggressive — it will match any long base64-like string including legitimate line noise. Production use should pair it with a minimum-threshold context check (e.g., exclude hex/noise that is not surrounded by instructions or comments).

The finding puts the schema at `cmd/assets/assets/schemas/` (correct) but the config at `cmd/assets/assets/validation/` (no such directory). Following existing conventions: config → `cmd/assets/assets/config/`, schema → `cmd/assets/assets/schemas/`.

### Valid kernel preserved

Hardcoded pattern lists are a real maintenance concern. There are 6 lists across 3 files for Phase 1, plus the content-safety categories as Phase 2:
- `analysis/patterns.go`: `hedgeWords`, `vagueWords`, `passivePatterns` (3 lists)
- `scorer/d1_knowledge_delta.go`: beginner patterns (7 items), expert patterns (6 items) (2 lists)
- `scorer/d6_freedom_calibration.go`: `whenNotToUsePatterns` (1 list)
- `cmd/analyze.go`: `canonicalSections`, `triggerWords`, `requiredSections` (3 config-like structures — lower priority)
- Content-safety (Phase 2): SEC_DISABLE, SEC_PERMISSIVE, CRED_EXFIL, OBFUSC_B64, OBFUSC_UNICODE, OBFUSC_HEX, TOOL_BROAD patterns

## Scope

### Phase 1 (this plan)

- **Externalise** 6 scoring-pattern lists from `analysis/patterns.go`, `scorer/d1_knowledge_delta.go`, and `scorer/d6_freedom_calibration.go` into `cmd/assets/assets/config/scoring-patterns.yaml`
- **Load** via `//go:embed` in `cmd/embed.go` at init
- **Wire** into existing detection functions — patterns become the default, no CLI flag needed
- **Validate** the config against a JSON Schema at `cmd/assets/assets/schemas/scoring-patterns.schema.json`
- **Unit tests** for the config loader and pattern resolution

### Phase 2 (follow-up)

- Add content-safety rule categories (SEC_DISABLE, SEC_PERMISSIVE, CRED_EXFIL, OBFUSC_*, TOOL_BROAD) mapped to D3 and D6 scoring
- Add URL allowlist check if a use case emerges
- Add `validate content` subcommand to expose content-safety checks independently
- Consider per-skill allowfile mechanism for `skip_rules` overrides

### Out of scope (deferred)

- Externalising D3's regex-based structural patterns (tied to block-parsing logic, not simple word lists)
- Externalising D7 anchor patterns (only 2 items, low maintenance burden)
- The `cmd/analyze.go` section/trigger/required lists (control flow, not scoring)
- TOML support
- URL allowlisting (no existing feature)

## Steps

### 1. Create `cmd/assets/assets/config/scoring-patterns.yaml`

Define dimension-keyed pattern groups:

```yaml
version: 1
patterns:
  d1_knowledge_delta:
    beginner_signals:
      - "npm install"
      - "yarn add"
      - "pip install"
      - "getting started"
      - "introduction"
      - "basic syntax"
      - "hello world"
    expert_signals:
      - "anti-pattern"
      - "NEVER"
      - "ALWAYS"
      - "production"
      - "gotcha"
      - "pitfall"
  analysis_quality:
    hedge_words:
      - "maybe"
      - "perhaps"
      - "might want to"
      - "could be"
      - "feel free"
      - "you might"
      - "possibly"
    vague_words:
      - "do something"
      - "handle appropriately"
      - "as needed"
      - "when necessary"
      - "if applicable"
    passive_patterns:
      - "is done"
      - "was created"
      - "can be used"
      - "is used"
      - "are used"
      - "is called"
      - "was called"
  d6_freedom_calibration:
    when_not_to_use:
      - "when not to use"
      - "do not use"
      - "not intended for"
      - "outside the scope"
      - "avoid using"
```

### 2. Create `cmd/assets/assets/schemas/scoring-patterns.schema.json`

Draft 2020-12 schema with `additionalProperties: false`, matching the structure above. Each pattern group validates that values are non-empty string arrays.

### 3. Add embed directives to `cmd/embed.go`

```go
//go:embed assets/assets/config
var embeddedConfig embed.FS //nolint:unused // reserved for pattern config
```

### 4. Create `internal/patternconfig/` package

Integration note: each scorer function follows the pattern `func scoreD{N}(content, skillDir string, b *validatorBridge) (int, []Diagnostic)`. The `internal/patternconfig` package provides a `Get()` singleton accessible from both `scorer/` and `analysis/` (both can import `internal/`). The `analysis/patterns.go` `stripCodeBlocks` function should be consolidated here alongside `scorer/dimensions.go`'s `removeCodeBlocks` — one canonical implementation.

Small loader (~80 LoC):
- `Init(fs embed.FS, path string)` — called at startup, populates a package-level `*Config` singleton
- `Get() *Config` — returns the loaded config (panics if not initialised)
- `Config` struct with `version int` and `Patterns map[string]PatternGroup`
- `PatternGroup` with `[]string` fields matching each group above
- Init-time validation: reject unknown group keys, non-empty string checks, regex compilation

Tests: verify round-trip load, version mismatch rejection, unknown key rejection, fallback to hardcoded defaults on load failure.

### 5. Wire into `analysis/patterns.go`

Replace `var hedgeWords = []string{...}` with a call to a central `GetConfig()` singleton loaded at init. The existing `DetectAntiPatternSignals` function receives its word lists as parameters (or a `*Config` reference) instead of package-level vars.

Keep backward compatibility: the existing `var` declarations become package-level defaults, overridden when the config loader initialises.

### 6. Wire into `scorer/d1_knowledge_delta.go`

Replace the two inline `[]string{...}` in `scoreD1()` with config lookups. The beginner-signal penalty loop and expert-signal bonus loop reference `cfg.D1.BeginnerSignals` / `cfg.D1.ExpertSignals` instead of inline slices.

### 7. Wire into `scorer/d6_freedom_calibration.go`

Replace the `var whenNotToUsePatterns` initialization with a config lookup.

### 8. Init-time wiring in `main.go` or `cmd/root.go`

Call `patternconfig.Init(embeddedConfig, "assets/assets/config/scoring-patterns.yaml")` at startup, so all consumers find a populated singleton. Fail hard on config load error — misconfigured patterns should surface at startup, not silently during scoring.

### 9. Preserve existing scorer tests

Existing tests in `scorer/d1_knowledge_delta_test.go`, `scorer/d6_freedom_calibration_test.go`, and `analysis/patterns_test.go` use the current hardcoded lists. The default-loaded config should produce identical output. Add a test that verifies the loaded config matches the hardcoded defaults.

### 10. Validate with existing CI

```bash
go test ./internal/patternconfig/...   # new package
go test ./analysis/...                  # verify unchanged behaviour
go test ./scorer/...                    # verify unchanged behaviour
go test ./...                           # full suite
```

## Sequencing & risk summary

| Risk | Mitigation |
|------|------------|
| Config loader fails at startup, blocking all commands | Fail-open approach: if config fails to load, log a warning and fall back to hardcoded defaults. Hard-fail only in `validate` or `analyze` commands. |
| YAML indentation mistake breaks all pattern matching | JSON Schema validation is wired in CI (`validate artifacts`), not at runtime. Runtime loading validates structure via the Go struct. |
| Scorer tests diverge from config defaults | Step 9 adds a test asserting config content matches the original hardcoded lists. Any change to the config that changes scoring behaviour must also update this test. |
| Implicit knowledge that patterns come from config (not code) | Add a comment at each replacement site pointing to the config file. No public API change — consumers call the same detection functions. |

## Academic References

The following papers support the content-safety pattern-matching approach proposed for Phase 2:

| Paper | Relevance | Mapping |
|-------|-----------|---------|
| **Reflect-Guard** (arXiv:2605.24834, May 2026) — Demonstrates that "standard pattern-matching approaches" detect obfuscated adversarial prompts, validating our regex-based OBFUSC category approach | Obfuscation detection patterns (OBFUSC_B64, OBFUSC_UNICODE, OBFUSC_HEX) are a well-established safety signal in the literature | Phase 2: D3 anti-pattern coverage |
| **Automated Adversarial Red-Teaming** (arXiv:2512.20677, EACL 2026) — Proposes a detection pipeline covering "data exfiltration" and "inappropriate tool use" as threat categories, directly matching our CRED_EXFIL and TOOL_BROAD pattern concepts | Credential exfiltration and wildcard tool permissions are established threat categories in LLM safety evaluation | Phase 2: D3 (CRED_EXFIL) + D6 (TOOL_BROAD) |
| **Adversarial Prompt Disentanglement** (arXiv:2605.27823, AAAI 2026) — Uses graph-based semantic decomposition and spectral analysis for malicious pattern detection in prompts | Pattern-based classification of prompt intent is a first-class defense technique | Phase 2: general approach |
| **Prompt Attack Detection with LLM-as-a-Judge** (arXiv:2603.25176, Mar 2026) — Compares rule-based classifiers vs LLM-as-a-judge for guardrail enforcement, deployed in production | Validates that rule-based pattern matching (our approach) is complementary to LLM-based judges — each catches cases the other misses | Phase 2: justification for hybrid approach |

## Open Questions

- **Should `cmd/analyze.go`'s `canonicalSections`, `triggerWords`, and `requiredSections` be in the same config?** These are control-flow parameters (defining what "good" looks like for analysis), not scoring heuristics. Deferred — they can move later if the pattern proves useful.
- **Should D3's `reD3BadGood`, `reD3AntiInstr` regexes be in config?** Unlike word lists, these are regex patterns that couple to Go's `regexp` API. Externalising them as strings adds complexity (regex compilation at init, Go vs PCRE differences). Deferred.
- **Is the YAML config in the right location?** `cmd/assets/assets/config/` follows the existing `cmd/assets/assets/schemas/` and `cmd/assets/assets/templates/` pattern. Alternative: put config under `cmd/assets/references/` alongside the scoring rubrics. Prefer `config/` for machine-consumed data vs `references/` for human docs.
- **Should `stripCodeBlocks` be consolidated during Phase 1?** Two identical implementations exist (`scorer/dimensions.go`, `analysis/patterns.go`). Moving both to `internal/patternconfig` avoids future confusion — but adds scope to a package whose primary job is config loading. Recommend doing it as a prerequisite commit before the config work.
