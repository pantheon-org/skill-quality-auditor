---
title: "ADR-028: Externalise hardcoded scoring patterns to YAML config"
status: proposed
date: 2026-07-03
context:
  - path: .context/findings/yaml-content-validation-config-2026-07-03.md
  - path: .context/plans/yaml-content-validation-config-2026-07-03.md
---

**Status:** Proposed
**Date:** 2026-07-03

## Context

Six hardcoded pattern lists (beginner/expert signals in D1, hedge/vague/passive words in analysis, when-not-to-use patterns in D6) are embedded across four Go files, making them invisible to reviewers, untestable as configuration, and requiring a recompile to update.

A proposed content-safety scanning layer (SEC_DISABLE, SEC_PERMISSIVE, CRED_EXFIL, OBFUSC_*, TOOL_BROAD categories) provides 8 additional regex patterns that map conceptually to D3 (Anti-Pattern Coverage) and D6 (Freedom Calibration) — supported by academic references (Reflect-Guard, Automated Red-Teaming, APD, Prompt Attack Detection).

## Decision

1. **Phase 1 — Externalise existing patterns.** Move the 6 real hardcoded lists into `cmd/assets/assets/config/scoring-patterns.yaml`, loaded at init via `//go:embed` through a new `internal/patternconfig/` package. This covers `analysis/patterns.go`, `scorer/d1_knowledge_delta.go`, and `scorer/d6_freedom_calibration.go`.

2. **Config format is YAML only.** No TOML fallback. YAML v3 is already available as an indirect dependency. JSON Schema validation lives at `cmd/assets/assets/schemas/scoring-patterns.schema.json`.

3. **Phase 2 — Content-safety patterns.** The 8 regex categories are adopted as a follow-up, wired into D3 (SEC_DISABLE, CRED_EXFIL, OBFUSC_*) and D6 (SEC_PERMISSIVE, TOOL_BROAD) with the `strip_code_blocks: true` flag using `scorer/dimensions.go`'s existing `removeCodeBlocks` utility.

4. **Config path convention.** Config files go to `cmd/assets/assets/config/`, schemas to `cmd/assets/assets/schemas/` — consistent with existing patterns.

5. **Duplicate `stripCodeBlocks`/`removeCodeBlocks` implementation shall be consolidated** into `internal/patternconfig/` as a prerequisite step.

## Consequences

- **Easier:** Pattern changes no longer require Go recompilation. Reviewers can inspect patterns in a single YAML file.
- **Easier:** Content-safety patterns get clear dimensional mapping (D3/D6) with academic backing.
- **Easier:** New pattern additions follow a documented process (add to YAML + JSON Schema validation).
- **Harder:** Init-time config loading must be wired into `main.go`/`cmd/root.go` — a new dependency for the startup path.
- **Harder:** The `analysis/` package currently has no init or config loading; it imports from stdlib only. Adding an `internal/` import breaks that constraint.
- **Risk:** OBFUSC_B64 pattern `[A-Za-z0-9+/]{50,}={0,2}` will match legitimate base64 strings (e.g., data URIs, test fixtures). Phase 2 must pair it with context heuristics to avoid false positives.
