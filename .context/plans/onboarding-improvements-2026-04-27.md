---
title: "Onboarding Improvements Plan"
type: PLAN
status: DONE
date: 2026-04-27
value: MEDIUM
---
# Onboarding Improvements Plan

**Goal:** Lower the barrier to install for new users and provide a first-class upgrade path for existing ones.

---

## Current state

| What exists | Where |
|---|---|
| GoReleaser v2 | `.goreleaser.yaml` — builds `linux/darwin × amd64/arm64` tarballs |
| GitHub release workflow | `.github/workflows/release.yml` — triggered on `v*` tags |
| release-please | `.github/workflows/release-please.yml` — automates tag + changelog |
| No install script | — |
| No Homebrew tap | — |
| No mise/asdf plugin | — |
| No `update` CLI command | — |

---

## Phase 1 — install.sh (new users, quick win)

**Deliverable:** `scripts/install.sh` — a curl-pipe script hosted from GitHub Releases.

Behaviour:
1. Detect OS (`uname -s`) and arch (`uname -m`).
2. Resolve latest release tag via GitHub API (`/repos/{owner}/{repo}/releases/latest`).
3. Download the matching `skill-auditor_<os>_<arch>.tar.gz` from the release assets.
4. Verify `checksums.txt` with `sha256sum` / `shasum`.
5. Move `skill-auditor` binary to `/usr/local/bin` (or `$INSTALL_DIR` if set).
6. Print version confirmation.

Usage after merge:
```bash
curl -fsSL https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/scripts/install.sh | sh
```

CI gate: add a `scripts/install.sh` shellcheck step to `ci.yml`.

---

## Phase 2 — Homebrew tap (new users, macOS/Linux power users)

**Deliverable:** A `homebrew-tap` repository at `github.com/pantheon-org/homebrew-tap` (new repo, separate from this one) with a `Formula/skill-auditor.rb` formula.

GoReleaser integration — add to `.goreleaser.yaml`:
```yaml
brews:
  - name: skill-auditor
    repository:
      owner: pantheon-org
      name: homebrew-tap
    homepage: "https://github.com/pantheon-org/skill-quality-auditor"
    description: "Score AI skills against a 9-dimension quality framework"
    install: |
      bin.install "skill-auditor"
```

GoReleaser will push the updated formula automatically on each tag, provided `HOMEBREW_TAP_GITHUB_TOKEN` secret is set in the repo.

Usage after merge:
```bash
brew tap pantheon-org/tap
brew install skill-auditor
```

Update path for Homebrew users:
```bash
brew upgrade skill-auditor
```

---

## Phase 3 — mise plugin (new users, mise/asdf ecosystem)

**Deliverable:** `mise.toml` backend entry so `mise install skill-auditor` resolves to GitHub Releases.

Two sub-options (pick one):

| Option | Effort | Notes |
|---|---|---|
| A — `ubi` backend (no new repo) | S | `mise use ubi:pantheon-org/skill-quality-auditor` — relies on `ubi` resolving the GH release binary |
| B — dedicated `asdf-skill-auditor` plugin repo | M | Full `bin/install`, `bin/list-all`, `bin/latest-stable` scripts; publishable to asdf-community |

Recommended: **Option A** first (zero infra), promote to Option B once usage warrants it.

Add to `mise.toml`:
```toml
[tools]
"ubi:pantheon-org/skill-quality-auditor" = "latest"
```

Document in README under a `mise` tab.

---

## Phase 4 — `skill-auditor update` command (existing users)

**Deliverable:** New `update` cobra command at `cmd/update.go`.

Behaviour:
1. Fetch latest release tag from GitHub API (unauthenticated, public endpoint).
2. Compare against `main.version` (injected at build time by GoReleaser ldflags).
3. If up-to-date: print `skill-auditor vX.Y.Z is already the latest` and exit 0.
4. If behind: download the matching tarball for the current OS/arch, verify checksum, replace the running binary via a temp-file swap, and print `Updated to vX.Y.Z`.
5. `--check` flag: report available version without installing (CI-friendly).
6. `--version-target vX.Y.Z` flag: pin to a specific release.

This command is only useful when installed via `install.sh` (direct binary). Homebrew and mise users should use their own update mechanisms — document this in the command help text.

Tests (`cmd/update_test.go`):
- Mock HTTP responses for latest-release lookup.
- Assert correct OS/arch selection.
- Assert checksum mismatch returns non-zero exit.
- Assert `--check` exits 0 without touching the binary.

---

## Phase 5 — README + docs refresh

Update `README.md` to add an **Install** section with four tabs:

| Method | Audience |
|---|---|
| `install.sh` | Any POSIX shell user |
| Homebrew | macOS / Linux Homebrew |
| mise | mise / asdf users |
| Go install | Go developers |

Also add an **Updating** section pointing each method to its upgrade command.

---

## Sequencing and sizing

| Phase | T-shirt size | Dependency |
|---|---|---|
| 1 — install.sh | S | none |
| 4 — `update` command | M | Phase 1 (needs install.sh to be useful) |
| 2 — Homebrew tap | M | Requires new `homebrew-tap` repo + secret |
| 3 — mise plugin | S (Option A) | none |
| 5 — README | S | Phases 1–4 done |

Suggested order: **1 → 3 → 4 → 2 → 5**

---

## Out of scope

- Windows installer / `choco` formula — defer until there is demand.
- Auto-update on startup (check silently in background) — too invasive for a CLI audit tool.
- Signed binaries / notarization — revisit if Homebrew tap submission requires it.
