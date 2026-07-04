---
title: "Plan: User-Configurable Scoring Pattern Overrides"
type: plan
status: done
date: 2026-07-03
related:
  - ../../internal/patternconfig/patternconfig.go
  - ../../cmd/root.go
  - ../../cmd/embed.go
  - ../../docs/ADR/adr-028-scoring-pattern-config.md
  - ../../docs/reference/d1-knowledge-delta.md
  - ../../docs/reference/d6-freedom-calibration.md
  - yaml-content-validation-config-2026-07-03.md
---
# Plan: User-Configurable Scoring Pattern Overrides

## Goal

Let a user of a pre-built `skill-auditor` binary (install.sh, `go install`, mise) override the D1/D6/analysis-quality scoring pattern lists without forking and recompiling. Today `internal/patternconfig.Init` only reads `cmd/assets/assets/config/scoring-patterns.yaml` from the binary's embedded filesystem (`embed.FS`) — there is no `-c/--config` flag, no user config-directory lookup, and no code path that can read an arbitrary OS path at all. This plan adds a disk-based loader, a precedence chain, and a CLI flag, while preserving the existing "a bad config must degrade, never crash" guarantee.

## Background

ADR-028 (accepted, PR #114) externalised the pattern lists from hardcoded Go into the embedded YAML file, but scoped user-level overrides out entirely — Phase 1 (done) covers embedding, Phase 2 (unimplemented) covers new content-safety pattern categories. Neither phase decided anything about per-user customisation. `internal/patternconfig.Init(fs embed.FS, path string)` reads via `embed.FS.ReadFile`, which can only resolve paths baked into the binary at compile time — pointing it at a real disk path requires a second, `os.ReadFile`-based load path, not just a different `path` argument.

## Amendments (2026-07-03 plan review)

A 3-reviewer plan-review pass (Technical/Strategic/Risk, Claude Sonnet 5) found one factual error and two internal contradictions in the original draft, plus several gaps. Resolved directly in the sections below; see each Phase/Decision for the fix. Summary of what changed:

- **Factual error corrected:** `os.UserConfigDir()` does **not** return `~/.config` cross-platform — only on Linux (via `$XDG_CONFIG_HOME`, falling back to `~/.config`). On macOS it returns `~/Library/Application Support`; on Windows, `%AppData%`. The plan previously stated "`~/.config`... via `os.UserConfigDir()`" as if that were universal. Fixed in Scope, Decisions, and Phase 2/3 below.
- **Contradiction resolved — flag scope:** Phase 2 already assumed a persistent `-c` flag while Open Questions listed it as undecided. Resolved: persistent (see Decisions). Removed from Open Questions.
- **Contradiction resolved — merge semantics:** Open Questions said this "should be settled before Phase 1 starts," but Phase 1's tasks already assumed whole-file-replace. Resolved: whole-file-replace, matching existing `validate()` behaviour (see Decisions). Removed from Open Questions.
- **New decision added:** `skill-auditor eval` (including the CI structural eval gate) always scores against the embedded/hardcoded config, ignoring both the `-c` flag and the default config-directory path — closes an eval-reproducibility risk the original draft left as an open question with real CI-correctness implications.
- **New tasks added to Phase 2:** an early audit of existing `PreRunE`/`PersistentPreRunE` hooks (Cobra only runs the closest one in the command tree — the new hook could be silently shadowed for exactly the CI-gate commands that matter most), a test-isolation task (flag reset / `t.Setenv`, since ~15 `cmd/*_test.go` files drive commands directly with no reset today), and a `--no-user-config` CI escape hatch.
- **New tasks added to Phase 1:** explicit handling for symlink/directory/permission-denied `os.ReadFile` outcomes, and a distinct error message for "structurally valid YAML but missing some pattern groups" vs. a plain parse error.
- **Open Questions trimmed** to genuine unresolved judgement calls (debug visibility, ADR form). A forward-compatibility note on ADR-028 Phase 2 was added under Out of scope rather than left as an open question, since it isn't blocking.

## Amendments (2026-07-03, second pass — user directives)

- **ADR-032 framing resolved.** Checked `adr-capture`'s supersession reference against how ADR-031 actually handled the same "extends ADR-028" situation: ADR-031 stayed fully standalone (`status: accepted`, no `superseded_by`), linked only via `context:` to its source finding — because it doesn't reverse anything ADR-028 decided. ADR-032 follows the same pattern (see Decisions). Removed from Open Questions.
- **Auto-generation on first run added.** If no `-c` flag and no config resolves from any tier, the tool now writes the active (embedded/hardcoded) config out to the default config-directory path before proceeding, so subsequent runs have a real file to edit rather than needing to know the path/format upfront.
- **CWD lookup tier added.** A `./scoring-patterns.yaml` in the current working directory is now checked as an opportunistic, non-auto-created tier between the `-c` flag and the default config-directory path — see Decisions for exact precedence and rationale.
- Two design points (exact default filename convention, CWD-vs-flag precedence ordering) were proposed with a "Recommended" default rather than confirmed by the user before this edit — flagged inline below for correction if wrong.

## Decisions

These were judgement calls the original draft left open or got wrong; resolved here so Phase 1 can start without redoing work.

- **Merge semantics: whole-file replace, not partial merge.** A user config must define all pattern groups, exactly like the existing embedded-config `validate()` requires today. Partial per-group override is explicitly deferred — it's a bigger design change (merge logic, precedence *per group* instead of per file) than this plan's scope justifies. Revisit only if a real user request surfaces.
- **`-c/--config` is a persistent flag on `rootCmd`.** Pattern config affects every scoring command (`evaluate`, `batch`, `analyze`, `duplication`), not one, so it belongs at the root rather than duplicated per-subcommand.
- **`skill-auditor eval` always uses the embedded/hardcoded config, never the `-c` flag or the default config-directory path.** Eval scenarios (`evals/summary.json`, the CI structural eval gate) must be reproducible across machines and CI runners; letting a local override silently change eval scores would break that. The precedence function only applies to scoring commands, not `eval`.
- **Default config directory: use `os.UserConfigDir()` as-is, not a hand-rolled XDG-only path.** It's the Go-idiomatic, already-cross-platform-correct choice — this repo already ships Windows/macOS/Linux binaries (release workflow builds for all three). The actual resolved directory differs per OS and must be documented as such, not summarized as "`~/.config`":
  - Linux: `$XDG_CONFIG_HOME/skill-quality-auditor/scoring-patterns.yaml`, or `~/.config/skill-quality-auditor/scoring-patterns.yaml` if `$XDG_CONFIG_HOME` is unset
  - macOS: `~/Library/Application Support/skill-quality-auditor/scoring-patterns.yaml`
  - Windows: `%AppData%\skill-quality-auditor\scoring-patterns.yaml`
- **ADR form: a new, standalone ADR (ADR-032) that references and extends ADR-028 — not a supersession, not an in-place amendment.** Confirmed against `adr-capture`'s own supersession reference and how ADR-031 handled the identical situation: ADR-031 stayed `status: accepted` on its own, no `superseded_by`, linking back only via `context:`. ADR-032 does the same: `status: proposed` initially (default convention), `context:` pointing at this plan, and — since this one more directly extends ADR-028's scope than ADR-031 did — also listing `docs/ADR/adr-028-scoring-pattern-config.md` in `context:` for discoverability. Confirmed 032 is the next free number as of this plan's date.
- **Default config-directory filename:** `os.UserConfigDir()/skill-quality-auditor/scoring-patterns.yaml` (directory named after the repo, matching how the plan already refers to it, leaving room for other config files later). *Proposed default, not explicitly confirmed — correct if a different convention is wanted.*
- **First-run auto-generation.** If no `-c` flag is given, no `./scoring-patterns.yaml` exists in the CWD, and no file exists yet at the default config-directory path, the tool writes the currently-active config (embedded, or hardcoded if even the embedded load failed) out to that default path as YAML before continuing — so the very first run produces a real, editable file instead of requiring the user to know the schema upfront. Every subsequent run then finds that file already in place at tier 3 (see the updated precedence order below). Explicit `-c` and a CWD file are treated as "the user already provided a config" and skip auto-generation entirely — only the true last-resort case materialises anything.
- **Auto-generation is best-effort, never fatal.** If the default config directory can't be created or written (permission error, read-only filesystem — plausible in some containerised CI images), the tool warns once to stderr and proceeds with the in-memory embedded/hardcoded config for that run, exactly as it does today. Writing a config file is a convenience, not a requirement for scoring to work.
- **CWD lookup: `./scoring-patterns.yaml` in the current working directory, checked between the `-c` flag and the default config-directory path, and never auto-created.** Rationale for the name: matches the embedded file's own basename rather than inventing a new dotfile convention. Rationale for the position: `-c` is the most explicit signal (a user named that exact file) so it still wins; a project-local CWD file is the next most specific signal (this project wants this override) and should beat a machine-wide personal default; the auto-generated home-directory file remains the catch-all editable fallback. *Both the filename and this precedence position are proposed defaults, not explicitly confirmed by the user — correct if wrong.*
- **`--no-user-config` also suppresses auto-generation.** If a user explicitly opts out of user config with this flag, the tool must not write one to disk either — it goes straight to embedded/hardcoded, exactly like `eval` does unconditionally.

## Scope

### In scope

- A disk-based config loader in `internal/patternconfig`, sharing validation logic with the existing embedded loader
- A `-c/--config` persistent flag on the root command
- An opportunistic `./scoring-patterns.yaml` lookup in the current working directory
- A default config-directory lookup path via `os.UserConfigDir()` (see Decisions for the real per-OS paths), auto-generated on first run if nothing else resolves
- A defined 5-tier precedence order (`-c` flag > CWD file > default path > embedded > hardcoded), with `eval` excluded from it entirely (see Decisions)
- A `--no-user-config` escape hatch to force embedded/hardcoded-only resolution and skip auto-generation
- Documentation updates to the doc set already touched by the recent scoring-patterns doc-drift fix
- A new ADR (ADR-032) capturing the decision

### Out of scope

- Phase 2 content-safety pattern categories (SEC_DISABLE, CRED_EXFIL, etc.) — tracked separately under ADR-028
- Per-skill or per-directory config overrides (this plan is user/machine-scoped only, one config per invocation)
- Partial per-group config merge (see Decisions — deferred)
- A config-management subcommand (`skill-auditor config init`, `config edit`, etc.) beyond the debug-visibility option raised in Open Questions
- Remote/URL-sourced config (S3, HTTP, etc.)
- **Forward-compatibility note (not blocking):** ADR-028 Phase 2 will eventually add new pattern groups to the embedded config. Combined with the whole-file-replace decision above, an old user config written before that lands will fail `validate()`'s "every group non-empty" check after upgrading — surfacing today's existing generic error, not a crash. No extra engineering needed now; noted so it isn't a surprise later.

## Phases

### Phase 1: Disk-based config loader

**Exit criterion:** `internal/patternconfig` can load and validate a `scoring-patterns.yaml` from an arbitrary OS path, fully unit-tested, with zero behavioural change when no user config is present.

Tasks:

- Extract the shared "unmarshal + `validate()`" logic currently inlined in `Init` into a private helper (e.g. `parseAndValidate(data []byte) (Config, error)`), so the embedded and disk loaders share one code path instead of two copies of the same validation logic.
- Add a disk loader, e.g. `LoadFromPath(path string) (Config, bool, error)`, using `os.ReadFile`. Distinguish outcomes precisely:
  - file absent (`ok=false, err=nil` — not configured, not an error)
  - file present but a directory, a permission error, or a symlink to a missing target (`ok=false, err=<detail>` — treat these as configuration errors, not "absent", since something exists at the path and silently ignoring it would hide a real user mistake)
  - file present but fails YAML parsing (`ok=false, err=<parse detail>`)
  - file present, parses, but is missing one or more required pattern groups (`ok=false, err=<"missing groups: X, Y">` — distinct message from a parse error, so a user editing an existing file down to a partial override gets a message that points at *which* groups are missing, not a generic parse failure)
  - file present and fully valid (`ok=true, err=nil`)
- Wire the disk loader into `active` the same way `Init` does today (mutex-guarded swap), without changing `Get()`'s public signature.
- Add a `WriteDefault(path string, cfg Config) error` helper that marshals a `Config` back to YAML (`yaml.Marshal`, matching the embedded file's structure exactly so a generated file round-trips through the same loader) and writes it via `os.MkdirAll` + `os.WriteFile` — the first-run auto-generation primitive. Failure here (permission denied, read-only filesystem) must return an error the caller can log-and-ignore, never panic.
- Unit tests: valid disk config overrides the prior active config; missing file is silently skipped; a directory/permission-error/malformed-YAML/missing-groups path each leave the prior active config untouched and return their distinct error; empty pattern group is rejected by the existing `validate()` reused unchanged; `WriteDefault` round-trips (write then `LoadFromPath` the same file back and compare) and surfaces (not swallows) a write failure to its caller.
- Add fixtures under `internal/patternconfig/testdata/` for the disk-path tests, reusing `valid.yaml` / `malformed.yaml` / `empty-group.yaml` where the content already fits, plus a new `partial-groups.yaml` fixture for the missing-groups case.

### Phase 2: Precedence wiring + CLI flag

**Exit criterion:** `skill-auditor -c path/to/config.yaml <command>` visibly changes scoring output; a `./scoring-patterns.yaml` in the CWD is honoured when no flag is given; the default config-directory path is honoured when neither of those exist, and gets auto-generated on a true first run; `skill-auditor eval` is provably unaffected by all of it; `--no-user-config` forces embedded/hardcoded-only behaviour and suppresses auto-generation; existing embedded/hardcoded behaviour is provably unchanged when nothing resolves; `hk check && go test ./...` green.

Tasks:

- **Audit `cmd/*.go` for existing `PreRunE`/`PersistentPreRunE` hooks before writing the new one.** Cobra only runs the *closest* `PersistentPreRunE` in the command tree — if `eval`, `batch`, or `duplication` (the CI-gate commands) already define their own `PreRunE`, adding one on `rootCmd` alone would silently never fire for that subtree. This must be confirmed clear (or worked around) before the next task, not discovered after.
- Add a persistent string flag `-c/--config` on `rootCmd` (default `""`), and a boolean `--no-user-config` flag that skips the `-c` value, the CWD lookup, the default config-directory lookup, and auto-generation — forcing embedded/hardcoded-only resolution, the CI escape hatch for a stray file on a runner.
- Move pattern-config initialisation out of `cmd/root.go`'s package `init()` into a `PersistentPreRunE` on `rootCmd` (or an equivalent early hook in `Execute()`). This is a required structural change: the `-c` flag's value is only known after cobra parses `os.Args`, which happens strictly after package `init()` has already run.
- Add a test-isolation helper (flag reset between test runs, `t.Setenv` for the config-directory environment variable and for CWD via `t.Chdir`) before wiring the flag into any command — the ~15 existing `cmd/*_test.go` files drive commands via direct `RunE`/`Execute()` calls with no flag-reset step today, so a persistent flag (and a CWD-sensitive lookup) risks state leaking across tests without this.
- Implement one precedence function, e.g. `resolveConfig(flagPath string, noUserConfig bool)`, applied in this order (skipped entirely, going straight to tier 4, when `noUserConfig` is true, and never invoked at all for the `eval` command per the Decisions above):
  1. **Explicit `-c` flag** — if set, load it via the Phase 1 disk loader. A missing or malformed file here is a **hard error** (the user asked for this file by name).
  2. **CWD file** (`./scoring-patterns.yaml`, relative to the process's working directory at invocation time) — if present, load it via the same disk loader. Missing is silently skipped; malformed **warns and falls through** to the next tier. Never auto-created.
  3. **Default config-directory path** (`os.UserConfigDir()/skill-quality-auditor/scoring-patterns.yaml` — see Decisions for the real per-OS value) — if present, load it via the same disk loader; malformed **warns and falls through**. If absent, call `WriteDefault` with the config that tier 4/5 would otherwise produce, then proceed with that config for this run (best-effort — a write failure here also just warns and falls through, per Decisions).
  4. **Embedded config** — the existing `Init(embeddedConfig, ...)` call, unchanged.
  5. **Hardcoded `defaultConfig`** — already the final fallback inside `internal/patternconfig`; no change needed.
- Tests: flag path wins over CWD file wins over default path; CWD file used when flag absent and file exists; default path auto-generated only when neither flag nor CWD file resolved anything, and used as-is on the next run without rewriting it; `--no-user-config` skips all three user-facing tiers and never writes a file; `eval` never picks up any user source even when all three are present; embedded/hardcoded behaviour is bit-for-bit unchanged when nothing resolves and auto-generation itself fails (regression guard); an explicit `-c` pointing at a missing or malformed file surfaces a clear, actionable error and a non-zero exit — unlike the silent skip for the two opportunistic tiers.

### Phase 3: Documentation + decision record

**Exit criterion:** docs and README accurately describe the override mechanism (including the real per-OS default path and the `eval` exclusion), `scripts/check-docs-drift.sh` and `hk check` are clean, and the decision is captured as ADR-032.

Tasks:

- Update the "Signal word configuration" notes already added to `docs/reference/d1-knowledge-delta.md` and `d6-freedom-calibration.md` (this session's doc-drift fix) to mention the `-c` flag and the real per-OS default path, not just the embedded YAML.
- Update `docs/development/setup.md`, `docs/development/adding-a-scorer.md`, and README.md's flags/usage section with the new `-c/--config` and `--no-user-config` flags, the 5-tier precedence order (flag > CWD file > default path (auto-generated) > embedded > hardcoded), and the note that `eval` is exempt.
- Check whether install.sh/mise-related quickstart docs exist and mention config customisation anywhere; if so, add a pointer there too — the Goal specifically targets binary-install users, who are the least likely to find a `docs/development/` page.
- Run the `adr-capture` skill to record ADR-032, referencing and extending ADR-028 per the Decisions section above (not an in-place amendment).
- Run `scripts/check-docs-drift.sh` and `hk check` before merging, to confirm this doc set doesn't trip the drift heuristic added earlier this session.

## Risks

- Moving config initialisation from package `init()` to `PersistentPreRunE` changes startup timing — Phase 2's first task is now an explicit audit of existing `PreRunE`/`PersistentPreRunE` hooks specifically to catch this before it ships silently broken. Code paths that call scorer/analysis functions directly in tests (bypassing `cmd.Execute()`) are unaffected either way, since they never trigger `cmd/root.go`'s `init()` today either.
- A default config-directory lookup can silently change scoring output for a user who forgot they left a stale config file in place, or for a CI runner with a stray file in its cached image. Mitigated by `--no-user-config` (Phase 2) and the debug-visibility option in Open Questions below.
- Asymmetric error handling (hard-fail on explicit `-c`, soft-fail on the opportunistic default path) is easy to get backwards; Phase 2's test list explicitly covers both directions to guard against that.
- Phase 2 introduces the first-ever hard-fail path in a subsystem whose own doc comment says "scoring must never fail because of a bad config" (the explicit `-c` case). Worth a CHANGELOG note when this ships, since it's a behavioural change in kind, not just degree, even though it only fires when a user explicitly opts in via `-c`.
- Auto-generation is a new filesystem side effect on what was previously a read-only CLI for these commands. Mitigated by making the write best-effort (a failed write just warns and continues), but it's still worth confirming in Phase 2 testing that no sandboxed/restricted environment treats a stderr warning as a hard failure.
- CWD-relative lookup means the same command run from two different directories can silently score differently. This is the intended behaviour (a project-local override), but worth calling out in Phase 3 docs explicitly so it isn't mistaken for nondeterminism.

## Verification

```bash
go test ./internal/patternconfig/... ./cmd/...
go test ./...
hk check
scripts/check-docs-drift.sh
```

Manual checks:

- Delete any local default-path config, run `skill-auditor evaluate <skill>` with no flag and no CWD file, and confirm a new file is written at the real per-OS default path (see Decisions) — the first-run auto-generation path.
- Run the same command again and confirm the file is reused as-is, not rewritten.
- Place a `./scoring-patterns.yaml` in a scratch directory, `cd` into it, and confirm `skill-auditor evaluate <skill>` picks it up over the default-path config from the previous checks.
- Run `skill-auditor -c /tmp/custom-patterns.yaml evaluate <skill>` from that same directory and confirm the `-c` flag wins over the CWD file.
- Point `-c` at a nonexistent file and confirm a clear, non-zero-exit error (not a silent fallback).
- Run `skill-auditor --no-user-config evaluate <skill>` with a CWD file, a default-path config, and `-c` all set, and confirm none of them are picked up and no file is written.
- Run `skill-auditor eval ./cmd/assets` with a default-path config present and confirm the eval score is unaffected (matches a run with no config present), and that no auto-generation is attempted.

## Open Questions

- **Debug visibility:** should there be a `skill-auditor config path` subcommand or `--print-config-source` flag so a user can see which of the five tiers actually loaded? Raised but not decided — worth doing given the silent-fallthrough tiers above, even with `--no-user-config` available as a blunter mitigation.
- **Default filename and CWD-vs-flag precedence are proposed, not confirmed.** This amendment applied "Recommended" defaults (`scoring-patterns.yaml` as both the CWD and default-directory filename; `-c` > CWD > default path) without an explicit confirmation round — flag if either should be different before Phase 1 starts.
