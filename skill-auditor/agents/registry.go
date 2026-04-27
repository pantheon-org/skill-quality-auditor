// Package agents provides the canonical registry of supported agent clients
// derived from https://github.com/vercel-labs/skills#supported-agents.
// It is the single source of truth used by both the cmd (install paths) and
// scorer (harness-dir and display-name anti-pattern detection) packages.
package agents

import (
	"path/filepath"
	"strings"
)

// Agent describes one supported agent client.
type Agent struct {
	// ID is the canonical --agent flag value (e.g. "claude-code").
	ID string
	// DisplayName is the human-readable name (e.g. "Claude Code").
	DisplayName string
	// ProjectPath is the skill dir relative to the project root.
	ProjectPath string
	// GlobalPath is the skill dir relative to the user's home directory.
	GlobalPath string
}

// Registry is the full list of supported agents.
var Registry = []Agent{
	{ID: "amp", DisplayName: "Amp", ProjectPath: ".agents/skills", GlobalPath: ".config/agents/skills"},
	{ID: "antigravity", DisplayName: "Antigravity", ProjectPath: ".agents/skills", GlobalPath: ".gemini/antigravity/skills"},
	{ID: "augment", DisplayName: "Augment", ProjectPath: ".augment/skills", GlobalPath: ".augment/skills"},
	{ID: "bob", DisplayName: "IBM Bob", ProjectPath: ".bob/skills", GlobalPath: ".bob/skills"},
	{ID: "claude-code", DisplayName: "Claude Code", ProjectPath: ".claude/skills", GlobalPath: ".claude/skills"},
	{ID: "cline", DisplayName: "Cline", ProjectPath: ".agents/skills", GlobalPath: ".agents/skills"},
	{ID: "codebuddy", DisplayName: "CodeBuddy", ProjectPath: ".codebuddy/skills", GlobalPath: ".codebuddy/skills"},
	{ID: "codex", DisplayName: "Codex", ProjectPath: ".agents/skills", GlobalPath: ".codex/skills"},
	{ID: "command-code", DisplayName: "Command Code", ProjectPath: ".commandcode/skills", GlobalPath: ".commandcode/skills"},
	{ID: "continue", DisplayName: "Continue", ProjectPath: ".continue/skills", GlobalPath: ".continue/skills"},
	{ID: "cortex", DisplayName: "Cortex Code", ProjectPath: ".cortex/skills", GlobalPath: ".snowflake/cortex/skills"},
	{ID: "crush", DisplayName: "Crush", ProjectPath: ".crush/skills", GlobalPath: ".config/crush/skills"},
	{ID: "cursor", DisplayName: "Cursor", ProjectPath: ".agents/skills", GlobalPath: ".cursor/skills"},
	{ID: "deepagents", DisplayName: "Deep Agents", ProjectPath: ".agents/skills", GlobalPath: ".deepagents/agent/skills"},
	{ID: "droid", DisplayName: "Droid", ProjectPath: ".factory/skills", GlobalPath: ".factory/skills"},
	{ID: "firebender", DisplayName: "Firebender", ProjectPath: ".agents/skills", GlobalPath: ".firebender/skills"},
	{ID: "gemini-cli", DisplayName: "Gemini CLI", ProjectPath: ".agents/skills", GlobalPath: ".gemini/skills"},
	{ID: "github-copilot", DisplayName: "GitHub Copilot", ProjectPath: ".agents/skills", GlobalPath: ".copilot/skills"},
	{ID: "goose", DisplayName: "Goose", ProjectPath: ".goose/skills", GlobalPath: ".config/goose/skills"},
	{ID: "iflow-cli", DisplayName: "iFlow CLI", ProjectPath: ".iflow/skills", GlobalPath: ".iflow/skills"},
	{ID: "junie", DisplayName: "Junie", ProjectPath: ".junie/skills", GlobalPath: ".junie/skills"},
	{ID: "kilo", DisplayName: "Kilo Code", ProjectPath: ".kilocode/skills", GlobalPath: ".kilocode/skills"},
	{ID: "kimi-cli", DisplayName: "Kimi Code CLI", ProjectPath: ".agents/skills", GlobalPath: ".config/agents/skills"},
	{ID: "kiro-cli", DisplayName: "Kiro CLI", ProjectPath: ".kiro/skills", GlobalPath: ".kiro/skills"},
	{ID: "kode", DisplayName: "Kode", ProjectPath: ".kode/skills", GlobalPath: ".kode/skills"},
	{ID: "mcpjam", DisplayName: "MCPJam", ProjectPath: ".mcpjam/skills", GlobalPath: ".mcpjam/skills"},
	{ID: "mux", DisplayName: "Mux", ProjectPath: ".mux/skills", GlobalPath: ".mux/skills"},
	{ID: "neovate", DisplayName: "Neovate", ProjectPath: ".neovate/skills", GlobalPath: ".neovate/skills"},
	{ID: "openclaw", DisplayName: "OpenClaw", ProjectPath: "skills", GlobalPath: ".openclaw/skills"},
	{ID: "opencode", DisplayName: "OpenCode", ProjectPath: ".agents/skills", GlobalPath: ".config/opencode/skills"},
	{ID: "openhands", DisplayName: "OpenHands", ProjectPath: ".openhands/skills", GlobalPath: ".openhands/skills"},
	{ID: "pi", DisplayName: "Pi", ProjectPath: ".pi/skills", GlobalPath: ".pi/skills"},
	{ID: "pochi", DisplayName: "Pochi", ProjectPath: ".pochi/skills", GlobalPath: ".pochi/skills"},
	{ID: "qoder", DisplayName: "Qoder", ProjectPath: ".qoder/skills", GlobalPath: ".qoder/skills"},
	{ID: "qwen-code", DisplayName: "Qwen Code", ProjectPath: ".qwen/skills", GlobalPath: ".qwen/skills"},
	{ID: "replit", DisplayName: "Replit", ProjectPath: ".agents/skills", GlobalPath: ".config/agents/skills"},
	{ID: "roo", DisplayName: "Roo Code", ProjectPath: ".roo/skills", GlobalPath: ".roo/skills"},
	{ID: "trae", DisplayName: "Trae", ProjectPath: ".trae/skills", GlobalPath: ".trae/skills"},
	{ID: "trae-cn", DisplayName: "Trae CN", ProjectPath: ".trae/skills", GlobalPath: ".trae-cn/skills"},
	{ID: "universal", DisplayName: "Universal", ProjectPath: ".agents/skills", GlobalPath: ".config/agents/skills"},
	{ID: "vibe", DisplayName: "Vibe", ProjectPath: ".vibe/skills", GlobalPath: ".vibe/skills"},
	{ID: "warp", DisplayName: "Warp", ProjectPath: ".agents/skills", GlobalPath: ".agents/skills"},
	{ID: "windsurf", DisplayName: "Windsurf", ProjectPath: ".windsurf/skills", GlobalPath: ".windsurf/skills"},
	{ID: "zed", DisplayName: "Zed", ProjectPath: ".zed/skills", GlobalPath: ".config/zed/skills"},
	{ID: "zencoder", DisplayName: "Zencoder", ProjectPath: ".zencoder/skills", GlobalPath: ".zencoder/skills"},
}

// ByID returns the Agent for the given id, or false if not found.
func ByID(id string) (Agent, bool) {
	for _, a := range Registry {
		if a.ID == id {
			return a, true
		}
	}
	return Agent{}, false
}

// SkillDir returns the absolute path to the skill install directory.
func (a Agent) SkillDir(homeDir string, global bool) string {
	if global {
		return filepath.Join(homeDir, a.GlobalPath)
	}
	return a.ProjectPath
}

// HarnessDirs returns the unique set of dot-directory names used by all agents
// (e.g. ".claude", ".cursor"). Used by the scorer to detect harness-specific paths.
func HarnessDirs() []string {
	seen := map[string]bool{}
	var dirs []string
	for _, a := range Registry {
		// GlobalPath is like ".claude/skills" — we want ".claude"
		parts := strings.SplitN(a.GlobalPath, "/", 2)
		if len(parts) > 0 && strings.HasPrefix(parts[0], ".") {
			if !seen[parts[0]] {
				seen[parts[0]] = true
				dirs = append(dirs, parts[0])
			}
		}
	}
	return dirs
}

// DisplayNames returns the lowercase display names of all agents.
// Used by the scorer to detect agent-specific references in skill content.
func DisplayNames() []string {
	names := make([]string, len(Registry))
	for i, a := range Registry {
		names[i] = strings.ToLower(a.DisplayName)
	}
	return names
}
