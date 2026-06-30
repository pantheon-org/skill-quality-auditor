---
title: "Draft Plan: Migrate to hk, markdownlint-cli2, and mise Hooks"
type: plan
status: done
date: 2026-06-29
---
# Draft plan: migrate to hk + markdownlint-cli2 + mise hooks

Status: DRAFT for review
Date: 29-06-2026
Branch: create a feature branch, for example `chore/migrate-hk-markdownlint`

## Goal

Replace the current dev tooling stack with three changes:

1. Replace **lefthook** with **hk** (https://hk.jdx.dev) as the git hook runner.
2. Replace **mdlint** (`github:swanysimon/mdlint`) with **markdownlint-cli2** (DavidAnson) as the markdown linter.
3. Use mise **enter** and **postinstall** hooks so the project bootstraps itself: `enter` runs `mise install` in an activated shell, and `postinstall` runs `hk install` after every `mise install`.

Both hk and mise are by the same author (jdx), so the integration is first-class: `mise use hk`, `HK_MISE` deep integration, and mise tasks usable as hk steps.

## Current state (what we are replacing)

| Concern | Today | Files |
| ------- | ----- | ----- |
| Hook runner | lefthook `2.1.5` | `lefthook.yml`, `mise.toml` |
| Markdown lint | `mdlint` (`github:swanysimon/mdlint`) | `mdlint.toml`, `.markdownlintignore`, `lefthook.yml`, `.github/workflows/ci.yml` |
| Bootstrap | manual `mise install` then `lefthook install` | `README.md` |

Current `lefthook.yml` hooks to preserve behaviour for:

- **pre-commit** (parallel): `go-fmt`, `go-vet`, `golangci-lint`, `mdlint`, `shellcheck`.
- **pre-push** (serial): `go-test`, `go-build`, `skill-validate`, `skill-duplication`, `skill-batch`.

Current `mdlint.toml` disables: MD013, MD032, MD051, MD055, MD058. Excludes: `cmd/assets`, `.github`, `testdata`, `CHANGELOG.md`. `.markdownlintignore` adds `.context/**`, `.claude/**`.

## Target state

### 1. mise.toml

Swap the two tools and add the postinstall hook.

```toml
[settings]
minimum_release_age = "7d"
experimental = true                 # confirm whether [hooks] needs this on the pinned mise version

[tools]
go = "1.25.5"
golangci-lint = "2.11.4"
hk = "X.Y.Z"                         # replaces lefthook — pin to current latest at implementation time
node = "latest"                     # REQUIRED: npm: backend below needs Node
"npm:markdownlint-cli2" = "A.B.C"   # replaces github:swanysimon/mdlint — pin to current latest
shellcheck = "latest"
# tessl is not available as a mise backend — install via: curl -fsSL https://get.tessl.io | sh

[env]
# HK_MISE deep integration: hk wraps each step with `mise x` so tools resolve
# without an active mise shell. Gives local/CI parity (decision 2: enabled).
HK_MISE = "1"

[hooks]
# enter: fires on entering the project dir in a `mise activate` shell.
# Installs/refreshes all deps so contributors do not run `mise install` by hand.
enter = "mise install"
# postinstall: runs after `mise install` (so it also runs as a result of the
# enter hook above); does NOT itself require activation. Installs the hk git hooks.
postinstall = "hk install"

[tasks.build]
run = "go build -o dist/skill-auditor ."
```

Decisions (resolved):

- **Pin versions (decision 1):** `hk` and `markdownlint-cli2` are pinned to explicit versions, matching the existing pinning style for go/golangci-lint. Replace the `X.Y.Z` / `A.B.C` placeholders with the current latest at implementation time.
- **`enter` hook:** runs `mise install` so an activated dev shell self-bootstraps all deps (which in turn triggers `postinstall = "hk install"`). Caveat: `enter` only fires inside a `mise activate` shell, so it is NOT a substitute for CI or for contributors who do not activate mise. `postinstall` remains the load-bearing bootstrap for those paths; the `enter` hook is the convenience layer on top. `mise install` on every `enter` is near-instant once deps are present (it no-ops), so the recurring cost is negligible.
- **`HK_MISE` (decision 2):** set to `1` in `[env]` so `hk` wraps steps with `mise x` and tools resolve without shell activation, giving local/CI parity.

### 2. hk.pkl (new file, replaces lefthook.yml)

Run `hk init` to scaffold, then edit to mirror the existing hooks. Note hk uses `{{files}}` templating (not lefthook's `{staged_files}`) and `glob` / `exclude` keys.

```pkl
amends "package://github.com/jdx/hk/releases/download/vX.Y.Z/hk@X.Y.Z#/Config.pkl"
import "package://github.com/jdx/hk/releases/download/vX.Y.Z/hk@X.Y.Z#/Builtins.pkl"

local preCommit = new Mapping<String, Step> {
    ["go-fmt"] {
        // Mirror lefthook's "**/*.go" — `*.go` may not match nested files under
        // hk's glob matcher and would silently skip scorer/, cmd/, reporter/, etc.
        glob = List("**/*.go")
        // fail if any staged Go file is not gofmt-clean
        check = "test -z \"$(gofmt -l {{files}})\""
    }
    ["go-vet"] {
        glob = List("**/*.go")
        check = "go vet ./..."
    }
    ["golangci-lint"] {
        glob = List("**/*.go")
        check = "golangci-lint run ./..."
    }
    ["markdownlint"] {
        glob = List("**/*.md")
        exclude = List("cmd/assets/**", "testdata/**", ".github/**", "CHANGELOG.md")
        check = "markdownlint-cli2 {{files}}"
        fix = "markdownlint-cli2 --fix {{files}}"
    }
    ["shellcheck"] {
        glob = List("scripts/**/*.sh")
        check = "shellcheck {{files}}"
    }
}

local prePush = new Mapping<String, Step> {
    ["go-test"]  { glob = List("**/*.go"); check = "go test ./..." }
    // Build always runs so the skill-* steps below have a fresh binary, even on a
    // push that touches only cmd/assets/** and no .go files (see ordering note).
    ["go-build"] { check = "go build -o dist/skill-auditor ." }
    ["skill-validate"]    { glob = List("cmd/assets/**"); depends = List("go-build"); check = "./dist/skill-auditor validate artifacts" }
    ["skill-duplication"] { glob = List("cmd/assets/**"); depends = List("go-build"); check = "./dist/skill-auditor duplication cmd/assets" }
    ["skill-batch"]       { glob = List("cmd/assets/**"); depends = List("go-build"); check = "./dist/skill-auditor batch ./cmd/assets --fail-below B" }
}

hooks {
    ["pre-commit"] {
        fix = true       // run fix steps (markdownlint --fix) and restage
        stash = "git"    // stash unstaged changes while fixing
        steps = preCommit
    }
    ["pre-push"] {
        steps = prePush
    }
    // enable `hk check` / `hk fix` for manual + CI use
    ["check"] { steps = preCommit }
    ["fix"]   { fix = true; steps = preCommit }
}
```

Behavioural differences to confirm:

- lefthook pre-commit ran `parallel: true`; hk parallelises steps by default, so this is preserved.
- lefthook pre-push ran `parallel: false` (serial). The `skill-*` steps depend on `go-build` producing `dist/skill-auditor` first. **Resolved:** use `depends = List("go-build")` on each skill step so hk orders them after the build regardless of parallelism. Verify `depends` semantics on the pinned hk version during implementation.
- **Binary-existence gap (closed):** `go-build` deliberately has **no `glob`**, so it runs on every pre-push. Previously (lefthook) both `go-build` (`glob: **/*.go`) and the `skill-*` steps (`glob: cmd/assets/**`) were glob-filtered, so a push touching only `cmd/assets/**` and no `.go` files would skip `go-build` and run the skill steps against a stale or missing `dist/skill-auditor`. Removing the glob on `go-build` guarantees a fresh binary. (Trade-off: build runs even on pure-markdown pushes; this is cheap and the safe default. Note `depends` alone would NOT fix this, because a glob-skipped dependency does not run.)
- Previously `mdlint check {staged_files}` only linted staged files on pre-commit but `mdlint check .` linted everything in CI. markdownlint-cli2 behaves the same way: pass `{{files}}` in the hook, and an explicit glob in CI (see below).

### 3. markdownlint-cli2 config (new file, replaces mdlint.toml + .markdownlintignore)

Create `.markdownlint-cli2.jsonc` at repo root (JSONC, not YAML — JSON is the preferred config format and markdownlint-cli2 supports `.jsonc`, which keeps inline comments). It carries both rule config and ignore globs, so it can fully replace `mdlint.toml` and `.markdownlintignore`.

```jsonc
// .markdownlint-cli2.jsonc
{
  "config": {
    // MD013 — long citation URLs / prose lines in docs exceed any sensible limit
    "MD013": false,
    // MD032 — DavidAnson: "lists should be surrounded by blank lines" (verify intent below)
    "MD032": false,
    // MD051 — DavidAnson: "link fragments should be valid" (verify intent below)
    "MD051": false,
    // MD055 — DavidAnson: "table pipe style" (verify intent below)
    "MD055": false,
    // MD058 — DavidAnson: "tables should be surrounded by blank lines" (verify intent below)
    "MD058": false
  },
  "globs": ["**/*.md"],
  "ignores": [
    "cmd/assets/**",
    "testdata/**",
    ".github/**",
    "CHANGELOG.md",
    ".context/**",
    ".claude/**",
    "node_modules/**"
  ]
}
```

Notes:

- The rule rationale comments above were rewritten to DavidAnson markdownlint's actual rule meanings (the old `mdlint.toml` comments described swanysimon/mdlint behaviour and several were inaccurate for the new linter). Verify on the first lint run whether each blanket disable is still warranted under DavidAnson semantics, or whether some rules are now auto-fixable and need no disable.
- Decide whether to keep blanket disables or scope them. Recommend keeping the same disables initially to avoid a large churn, then tightening in a follow-up.
- Delete `mdlint.toml` and `.markdownlintignore` once the new config is verified.

### 4. CI workflow (.github/workflows/ci.yml)

Replace the `mdlint` step. The workflow already uses `jdx/mise-action@v2`, so tools install from `mise.toml`.

```yaml
      - uses: jdx/mise-action@v2

      - name: markdownlint
        run: markdownlint-cli2 "**/*.md"

      - name: shellcheck
        run: shellcheck scripts/*.sh
```

- `markdownlint-cli2` reads `.markdownlint-cli2.jsonc` automatically, including `ignores`, so the explicit glob plus config replicates `mdlint check .`.
- Optional: also run `hk check` in CI for parity with local hooks instead of duplicating individual lint commands. Decide whether CI calls `hk check` (single source of truth) or keeps explicit steps. Recommend `hk check` long-term, but keep explicit steps in this first migration to limit blast radius.

### 5. README.md

Update the contributing/hooks section (around lines 446-450):

- Replace the lefthook link and instructions.
- New flow:

```bash
mise install   # installs go, node, golangci-lint, markdownlint-cli2, shellcheck, hk
               # postinstall hook runs `hk install` automatically
               # (in an activated mise shell, the enter hook runs `mise install` for you)
```

- Mention `hk run pre-commit`, `hk check`, `hk fix`, and the `HK=0 git commit` bypass escape hatch.
- Update the tool list line that currently reads "installs go, golangci-lint, mdlint, shellcheck, lefthook".

## File-by-file change checklist

| File | Action |
| ---- | ------ |
| `mise.toml` | Swap `lefthook` -> pinned `hk`, `github:swanysimon/mdlint` -> pinned `npm:markdownlint-cli2`; add `node`; add `[env] HK_MISE = "1"`; add `[hooks] enter = "mise install"` + `postinstall = "hk install"`; set `experimental = true` if needed |
| `hk.pkl` | NEW — port pre-commit + pre-push steps from `lefthook.yml` |
| `lefthook.yml` | DELETE after hk verified |
| `.markdownlint-cli2.jsonc` | NEW — port rule disables + ignores (JSONC, not YAML) |
| `mdlint.toml` | DELETE after markdownlint-cli2 verified |
| `.markdownlintignore` | DELETE (folded into `.markdownlint-cli2.jsonc` ignores) |
| `.github/workflows/ci.yml` | Replace `mdlint check .` with `markdownlint-cli2 "**/*.md"` |
| `README.md` | Update hooks/bootstrap section + tool list |

## Verification steps

1. `mise install` — confirm `hk` and `markdownlint-cli2` install and `postinstall` runs `hk install` (check `.git/hooks` or `git config --get-regexp '^hook\.hk'`).
2. `markdownlint-cli2 "**/*.md"` — confirm it passes with the same files the old `mdlint check .` passed (no new violations, ignores honoured).
3. `hk run pre-commit` — confirm all five steps run and pass on a clean tree; stage a deliberately mis-formatted Go file and a markdown lint violation to confirm they are caught / auto-fixed.
4. `hk run pre-push` — confirm build runs before skill-* steps and the binary exists; confirm `batch --fail-below B` gates correctly.
5. `go test ./...` — must pass (project rule).
6. Push the branch and confirm the CI `lint` job is green with the new markdownlint step.
7. Confirm `HK=0 git commit` bypass works.

## Resolved decisions

All prior open questions are resolved per the recommendations:

1. **Pin versions:** YES. `hk` and `markdownlint-cli2` are pinned in `mise.toml` (placeholders `X.Y.Z` / `A.B.C` to be filled with current latest at implementation), matching go/golangci-lint.
2. **`HK_MISE` deep integration:** ENABLED. Set `HK_MISE = "1"` in `mise.toml` `[env]` for local/CI parity.
3. **mise `enter` hook:** ADDED, running `mise install`. `postinstall = "hk install"` remains the load-bearing bootstrap for CI and non-activated shells; `enter` is the convenience layer (it no-ops when deps are present).
4. **CI lint source of truth:** explicit `markdownlint-cli2` / `shellcheck` steps in this migration to limit blast radius. Switching CI to `hk check` as the single source of truth is a deferred follow-up.
5. **markdownlint rule disables:** KEEP the same five (MD013, MD032, MD051, MD055, MD058) initially to avoid churn; re-evaluate/tighten under DavidAnson semantics in a follow-up (rule meanings already re-checked in the config comments).
6. **pre-push ordering:** RESOLVED via `depends = List("go-build")` on each skill-* step, plus removing the `glob` from `go-build` so the binary always exists (see Behavioural differences). Verify `depends` semantics on the pinned hk version during implementation.

Remaining items to verify at implementation time (not blocking the design): exact pinned versions; that `[hooks]` works without `experimental` on the chosen mise version; and `depends` semantics on the chosen hk version.

## Risks / notes

- hk.pkl uses the Pkl config language. By default hk uses its built-in pklr evaluator, so the pkl CLI is not required (`HK_PKL_BACKEND=pkl` only if we want the real pkl CLI).
- `hk install` is global-vs-local: this plan uses per-repo `hk install` via postinstall. If a contributor has hk installed globally (`hk install --global`), `hk init` alone is enough; the postinstall `hk install` is still safe and idempotent.
- The `npm:` backend for markdownlint-cli2 requires Node. RESOLVED: `node` is added to `[tools]`, so CI (`jdx/mise-action@v2`) and local installs both provide it. Without this, the CI tool install would fail.
- CI noise (cosmetic): once `hk` is in `mise.toml`, `jdx/mise-action@v2` installs it and the `postinstall` hook runs `hk install` on the CI runner, installing git hooks that never fire in CI. Harmless; noted so it does not surprise anyone reading CI logs.
- Do not bump `cmd/assets/tile.json` version manually (release-please owns it) — unaffected by this change but noted.
