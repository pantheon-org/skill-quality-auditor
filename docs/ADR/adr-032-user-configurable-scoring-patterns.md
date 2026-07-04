---
title: "ADR-032: User-configurable scoring pattern overrides"
status: accepted
date: 2026-07-03
context:
  - path: ".context/plans/user-configurable-scoring-patterns-2026-07-03.md"
  - path: "docs/ADR/adr-028-scoring-pattern-config.md"
---

**Status:** Accepted
**Date:** 2026-07-03

## Context

ADR-028 externalised the D1/D6/analysis-quality scoring pattern lists from hardcoded Go into an embedded YAML file (`cmd/assets/assets/config/scoring-patterns.yaml`), loaded once at startup via `internal/patternconfig.Init(embeddedConfig, ...)`. That closed the "requires a recompile" problem for maintainers editing the source tree, but did nothing for anyone running a pre-built `skill-auditor` binary (install.sh, `go install`, mise) — `Init`'s signature only reads from the binary's embedded filesystem, with no `-c/--config` flag, no config-directory lookup, and no code path that can read an arbitrary OS path at all. ADR-028 explicitly scoped user-level overrides out of both its Phase 1 (done) and Phase 2 (unimplemented) — this ADR covers the gap it left.

## Decision

1. **Five-tier precedence chain**, highest to lowest: explicit `-c/--config` flag → an opportunistic `./scoring-patterns.yaml` in the current working directory → a default path under `os.UserConfigDir()/skill-quality-auditor/scoring-patterns.yaml` → the existing embedded config → the existing hardcoded Go defaults. `--no-user-config` skips the first three tiers entirely.
2. **`skill-auditor eval` is exempt from the entire chain.** It always scores against the embedded/hardcoded config, so `evals/summary.json` and the CI structural eval gate stay reproducible across machines regardless of any local override.
3. **Merge semantics: whole-file replace, not partial per-group merge.** A user config must define every pattern group, exactly like the existing embedded-config `validate()` already requires. Partial merge is deferred — it's a materially bigger design change (merge logic, per-group precedence) than this decision's scope justifies.
4. **First-run auto-generation.** If neither the `-c` flag nor a CWD file resolves anything, the tool writes the currently-active config out to the default config-directory path as YAML before proceeding, so a first-time user gets a real, editable file instead of needing to know the schema upfront. This write is best-effort: a failure (permission denied, read-only filesystem) warns and falls through to the embedded/hardcoded config for that run, never a hard failure.
5. **`os.UserConfigDir()` is used as-is, not a hand-rolled XDG-only path**, since this repo already ships Windows/macOS/Linux binaries and the stdlib function is already correct per-OS: `$XDG_CONFIG_HOME` (or `~/.config` fallback) on Linux, `~/Library/Application Support` on macOS, `%AppData%` on Windows.
6. **An explicit `-c` pointing at a missing or malformed file is a hard error**, unlike the two opportunistic tiers (CWD, default path) which warn-and-fall-through on malformed content and silently skip on absence — the user named that file directly, so silently ignoring a problem with it would be worse than failing loudly.
7. **This ADR stays standalone; it does not supersede ADR-028.** It extends ADR-028's scope rather than reversing anything it decided, following the same pattern ADR-031 used for the same relationship (`status: accepted`, no `superseded_by`, linked only via `context:`).

## Consequences

- **Easier:** users of pre-built binaries can tune scoring patterns without forking and recompiling, closing the gap ADR-028 left open.
- **Easier:** CI and local eval runs remain reproducible, since `eval` never sees a config override.
- **Harder:** `internal/patternconfig` gains its first filesystem-write side effect (auto-generation) and its first hard-fail path (explicit `-c` with a bad file) — both are new failure modes for a subsystem whose original design principle was "a bad config must degrade, never crash." Both are scoped narrowly (write failures are best-effort; the hard-fail only fires when a user explicitly opts in via `-c`) to preserve that principle everywhere else.
- **Harder:** pattern-config initialisation must move from `cmd/root.go`'s package `init()` to a `PersistentPreRunE` on `rootCmd`, since the `-c` flag's value is only known after cobra parses arguments — a structural change to CLI startup ordering that must be checked against any existing `PreRunE`/`PersistentPreRunE` hooks on `eval`/`batch`/`duplication` (Cobra only runs the closest one in the command tree).
- **Deferred, not decided:** whether a `skill-auditor config path` / `--print-config-source` debug flag should exist to show which tier actually loaded. Left open in the source plan.
