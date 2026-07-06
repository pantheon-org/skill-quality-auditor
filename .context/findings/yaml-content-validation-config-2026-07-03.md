---
title: "Finding: YAML Content Validation Config — 2026-07-03"
type: FINDING
status: ACTIVE
themes:
  - SKILL-QUALITY
related:
  - ../plans/yaml-content-validation-config-2026-07-03.md
date: 2026-07-03
value: LOW
---
# Finding: YAML Content Validation Config — 2026-07-03

> Scoper finding for externalising all hardcoded content-pattern rules (security, credential, obfuscation, URL allowlist, tool-permission checks) into a YAML config under `cmd/assets/assets/validation/`, enabling the `skill-auditor` to fail skills that match dangerous patterns.

## Summary

The shell script `scripts/validate-skill-content.sh` defines 8 pattern categories (SEC_DISABLE, SEC_PERMISSIVE, CRED_EXFIL, CRED_EXFIL_REV, OBFUSC_B64, OBFUSC_UNICODE, OBFUSC_HEX, TOOL_BROAD) plus a URL allowlist. Currently these live only in that shell script — they're not wired into `skill-auditor`'s scoring or the `validate`/`analyze` commands. The proposal is to encode them as a YAML config + JSON Schema, embedded via `//go:embed`, and expose them through a new `validate content` subcommand and/or a D-dimension scorer.

The codebase already has three places with hardcoded content-pattern rules that would consolidate into the same YAML config: `cmd/analyze.go` (trigger words, canonical sections, hedge/passive/vague lists), `scorer/d1_knowledge_delta.go` (beginner/expert signal patterns), and `scorer/d3_anti_pattern_coverage.go` (regex-based anti-pattern block detection).

## Detail

### Existing config patterns to follow

| Pattern | File | Notes |
|---------|------|-------|
| Embedded JSON config | `cmd/assets/requirements/review-report.requirements.json` | `//go:embed` in `cmd/embed.go`, loaded at init |
| Embedded YAML templates | `cmd/assets/assets/templates/*.yaml` | Same embedding pattern |
| JSON Schema validation | `cmd/assets/assets/schemas/*.schema.json` | Draft 2020-12, `additionalProperties: false` |
| Rule-based analysis | `cmd/analyze.go` → `analysis/patterns.go` | `RuleMatch` struct, hardcoded lists today |
| Validation command | `cmd/validate.go` | Accumulates errors, subcommands for artifacts/review |

### Shell script rule categories mapped to YAML config

```yaml
version: 1
dimension: "content-safety"  # or mapped into D3 (Anti-Pattern Coverage)

rules:
  # Shell script's SEC_DISABLE
  - id: "sec-disable"
    type: regex_proximity
    name: "security-disabling instructions"
    pattern: "(disable|skip|bypass|ignore|turn off|remove).{0,20}(security|auth|verification|validation|firewall|encryption|signing|protection)"
    severity: error
    weight: 5.0
    strip_code_blocks: true
    message: "Possible security-disabling instruction"

  # SEC_PERMISSIVE
  - id: "sec-permissive"
    type: regex
    pattern: "(allow all|trust all|no.verify|YOLO|permissive|0\\.0\\.0\\.0/0.*ingress)"
    severity: error
    weight: 5.0
    strip_code_blocks: true
    message: "Overly permissive instruction"

  # CRED_EXFIL
  - id: "cred-exfil"
    type: regex_proximity
    name: "credential exfiltration"
    pattern: "(curl|wget|fetch|post|send|upload|nc |netcat).{0,40}(token|key|password|credential|secret|API_KEY|AWS_SECRET|PRIVATE_KEY)"
    severity: error
    weight: 5.0
    message: "Possible credential exfiltration pattern"

  # CRED_EXFIL_REV
  - id: "cred-exfil-rev"
    type: regex_proximity
    pattern: "(token|key|password|credential|secret|API_KEY|AWS_SECRET|PRIVATE_KEY).{0,40}(curl|wget|fetch|post|send|upload|nc |netcat)"
    severity: error
    weight: 5.0
    message: "Possible credential exfiltration pattern (reversed)"

  # OBFUSC_B64
  - id: "obfusc-b64"
    type: regex
    pattern: "[A-Za-z0-9+/]{50,}={0,2}"
    severity: warning
    weight: 3.0
    strip_code_blocks: true
    message: "Possible base64-encoded content (50+ chars)"

  # OBFUSC_UNICODE
  - id: "obfusc-unicode"
    type: regex
    pattern: "\\\\u[0-9a-fA-F]{4}"
    severity: warning
    weight: 3.0
    strip_code_blocks: true
    message: "Unicode escape sequence"

  # OBFUSC_HEX
  - id: "obfusc-hex"
    type: regex
    pattern: "\\\\x[0-9a-fA-F]{2}(\\\\x[0-9a-fA-F]{2}){3,}"
    severity: warning
    weight: 3.0
    message: "Hex-encoded payload"

  # TOOL_BROAD
  - id: "tool-broad"
    type: regex
    pattern: "allowed-tools:\\s*\\*"
    severity: error
    weight: 5.0
    message: "Overly broad tool permissions (wildcard)"

url_allowlist:
  domains:
    - gitlab.com
    - docs.aws.amazon.com
    - awscli.amazonaws.com
    - console.aws.amazon.com
    - developer.hashicorp.com
    - cloud.google.com/docs
    - github.com/firecow
    - github.com/gruntwork-io
    - agentskills.io
    - conventionalcommits.org
  severity: warning
  weight: 2.0

rule_types:
  - id: "regex"
    fields: [pattern, strip_code_blocks, severity, weight, message]
  - id: "regex_proximity"
    fields: [pattern, severity, weight, message, strip_code_blocks]
  - id: "pattern_list"
    fields: [patterns, min_count, severity, weight, message, strip_code_blocks]
  - id: "trigger_frequency"
    fields: [name, min_count, severity, weight]
  - id: "jaccard_similarity"
    fields: [reference, threshold, severity, message]
  - id: "section_header"
    fields: [name, severity, weight, message]
```

### What would need to change

| Area | Change | Effort |
|------|--------|--------|
| New package `validation/` | Config loader (JSON/YAML/TOML), rule engine, URL checker, result types | Medium (250-350 LoC) |
| JSON Schema | `cmd/assets/assets/schemas/content-validation-rules.schema.json` | Small (80 LoC) |
| Config file | `cmd/assets/assets/validation/content-validation-rules.{json,yaml,toml}` | Small (100 LoC) |
| `cmd/embed.go` | Add `//go:embed` for new file(s) | Tiny (1 line) |
| TOML dependency | Add `github.com/BurntSushi/toml` to `go.mod` | Tiny |
| `cmd/validate.go` | Add `validate content` subcommand | Medium (100 LoC) |
| `scorer/d3_anti_pattern_coverage.go` | Optionally wire rule engine into D3 scoring | Medium (50 LoC) |
| `analysis/patterns.go` | Optionally replace hardcoded lists with YAML config load | Medium (50 LoC) |
| `cmd/analyze.go` | Optionally load lists from YAML instead of var blocks | Small (30 LoC) |
| JSON Schema validation for config | Wire into `validate` command or init-time validation | Small (30 LoC) |
| Per-skill `.content-check-allow` | Port to a field in the YAML config or keep as `.content-check-allow` files | Small |

### Total estimated effort

**Core:** new `validation/` package + CLI integration ≈ 450-650 LoC new Go code, ~200 LoC config/schema. The shell script's URL check and per-skill allow file mechanism would be preserved or ported.

### Key design decisions to make

1. **Config format precedence** — JSON first, YAML fallback, TOML last. Loader tries `content-validation-rules.json` → `.yaml` → `.toml` and uses the first found. Minimises dependencies (JSON is stdlib, YAML is already in `go.sum` via gopkg.in/yaml.v3, TOML adds `github.com/BurntSushi/toml`).

2. **Standalone `validate content` subcommand, or wire into scoring?** — The shell script exits non-zero on matches. The tool could either (a) add a new `validate content` subcommand, (b) integrate into the `analyze` command, or (c) contribute to an existing dimension (likely D3 — Anti-Pattern Coverage). All three are compatible; (a) is lowest risk, (c) is deepest integration.

3. **Fail mode** — Does a match block evaluation (hard fail), reduce score proportionally, or just warn? The shell script exits with code 1. For the Go CLI, the `--fail-below` flag on `batch` could gate on content-safety violations.

4. **Per-skill allowlist** — The shell script uses `.content-check-allow` files with pattern IDs. Port to a config field? Keep as sidecar files? A `skip_rules: [sec-disable]` array in the rule config or a per-skill override file would centralise it.

5. **URL allowlist scope** — Project-specific or generic? The current allowlist excludes internal/private domains. A generic config should ship with only public domains; projects extend via their own config file.

## Recommended Action

1. Make the above design decisions before implementing.
2. If proceeding: create `cmd/assets/assets/validation/content-validation-rules.schema.json`, then `content-validation-rules.yaml`, then the `validation/` package.
3. Wire into `validate content` subcommand as first consumer; evaluate D3 integration as follow-up.
4. Create an ADR if the decisions are binding (e.g., "all content-pattern rules move to external config, hardcoded lists removed").
