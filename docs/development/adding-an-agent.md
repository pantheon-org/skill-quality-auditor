# Adding an init target agent

This guide walks through adding support for a new AI agent harness to the `init` command.

## Steps

### 1. Add the agent definition

Edit `agents/registry.go` and add an entry to the `Registry` slice:

```go
{
    ID:          "my-agent",
    DisplayName: "My Agent",
    ProjectPath: ".my-agent/skills/",  // relative to project root
    GlobalPath:  ".my-agent/skills/",  // relative to home directory
},
```

The `Agent` struct fields:

| Field | Description |
|-------|-------------|
| `ID` | Machine-readable identifier (kebab-case) |
| `DisplayName` | Human-readable name for `--interactive` mode |
| `ProjectPath` | Install path relative to project root |
| `GlobalPath` | Install path relative to home directory |

### 2. Add a test case

Add a test case to `agents/registry_test.go`:

```go
func TestRegistry_MyAgent(t *testing.T) {
    a := ByID("my-agent")
    if a == nil {
        t.Fatal("my-agent not found in registry")
    }
    if a.DisplayName != "My Agent" {
        t.Errorf("expected 'My Agent', got %q", a.DisplayName)
    }
}
```

### 3. Run tests

```go
go test ./agents/...
```

No changes to `cmd/init.go` are needed — the registry drives everything
via `ByID`, `HarnessDirs`, and `DisplayNames`.

## How `init` auto-detects agents

The `resolveByHarness` function checks whether each agent's harness root
directory exists under the install target:

```go
func resolveByHarness(baseDir string, global bool) []Agent {
    // For each agent in the registry:
    //   root = first path component of ProjectPath (e.g., ".cursor")
    //   check if root exists under baseDir
    //   if yes, include the agent
}
```

For example, if `.cursor/` exists in the project, Cursor is auto-detected.
If `~/.claude/` exists, Claude Code is auto-detected (with `--global`).

## Harness directory convention

The harness root is always the first path component of `ProjectPath`/`GlobalPath`:

```text
ProjectPath: ".cursor/rules/" → harness root: ".cursor"
GlobalPath:  ".cursor/"       → harness root: ".cursor"
```

This convention means `HarnessDirs()` returns the set of unique dot-directory
names (e.g., `.claude`, `.cursor`, `.agents`, `.windsurf`).
