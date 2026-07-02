# Init, update, prune

Lifecycle commands for installation, self-update, and cleanup.

## Init

The `init` command installs the skill-quality-auditor skill into agent harness
directories (Claude Code, Cursor, Cline, GitHub Copilot, etc.).

### Agent detection

```go
// Resolution order:
// 1. Auto-detect: check each agent's harness directory exists
// 2. By ID: --agent <id> (e.g., "claude-code")
// 3. Interactive: --interactive → numbered list prompt
```

### Installation targets

- Default (project): installs into CWD
- `--global`: installs into `~`
- Each agent has `ProjectPath` (relative) and `GlobalPath` (home-relative)

### Installation methods

| Method | Flag | Behaviour |
|--------|------|-----------|
| Symlink | `--method=symlink` (default) | Writes canonical SKILL.md to `~/.local/share/skill-quality-auditor/`, symlinks from each agent dir |
| Copy | `--method=copy` | Copies SKILL.md + all assets directly to each agent dir |

### Assets installed

All embedded assets are copied alongside SKILL.md:

- `references/` — scoring rubrics, anti-pattern docs, thresholds
- `evals/` — evaluation scenarios
- `schemas/` — remediation plan JSON schemas
- `templates/` — template files
- `requirements/` — requirement definitions

### Agent registry

The `agents/registry.go` file maintains a registry of 44 supported agents
with their harness directory paths. Key types:

```go
type Agent struct {
    ID          string   // e.g., "claude-code"
    DisplayName string   // e.g., "Claude Code"
    ProjectPath string   // e.g., ".claude/skills/"
    GlobalPath  string   // e.g., ".claude/skills/"
}
```

## Update

The `update` command performs a self-update by downloading the latest release
from GitHub.

### Update pipeline

```text
update [--check] [--version-target <tag>]
  │
  ├── determine target version (latest or specific tag)
  ├── if --check: print current vs latest, exit
  │
  ├── fetch release from GitHub API:
  │     /repos/pantheon-org/skill-quality-auditor/releases/latest
  │     or /repos/.../releases/tags/<tag>
  │
  ├── select tarball for GOOS/GOARCH
  ├── download archive + checksums.txt
  ├── verify SHA256 checksum
  ├── extract binary from tar.gz
  └── atomically replace running binary:
        write temp in same dir → rename
        (respects symlinks via EvalSymlinks)
```

## Prune

The `prune` command removes old audit directories, keeping N most recent
per skill.

### Prune pipeline

```text
prune [--keep N] [--dry-run]
  │
  ├── read .context/audits/<skill>/ directories
  ├── for each skill:
  │     ├── list date-stamped subdirectories
  │     ├── sort descending (lexicographic YYYY-MM-DD)
  │     ├── keep top N (default: 5)
  │     └── os.RemoveAll() for the rest
  │
  ├── skip "latest" symlink directories
  └── report removed paths and latest symlink targets
```

## Source files

| File | Purpose |
|------|---------|
| `cmd/init.go` | Init command, agent resolution, asset installation |
| `cmd/update.go` | Self-update, GitHub release download, checksum verify |
| `cmd/prune.go` | Audit directory cleanup |
| `agents/registry.go` | Agent struct and registry |
| `cmd/agents.go` | Agent type re-export for cobra |
| `cmd/embed.go` | Embedded asset declarations |
