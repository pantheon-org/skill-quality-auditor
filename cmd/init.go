package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const skillName = "skill-quality-auditor"

// assetSubdirs is the ordered list of subdirectories bundled alongside SKILL.md.
var assetSubdirs = []string{"references", "evals", "schemas", "templates", "requirements"}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install the skill-quality-auditor skill into agent environments",
	Long: `Install the embedded skill-quality-auditor skill into one or more agent
skill directories.

Detection behaviour:
  • Default  — installs into CWD; auto-detects agents whose harness directory
               already exists under the current working directory.
  • --global  — installs into ~ instead; auto-detects against the home directory.
  • --agent   — skip auto-detection and target the specified agent(s) explicitly.
  • --interactive — show the full agent list and let you choose interactively.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		method, _ := cmd.Flags().GetString("method")
		global, _ := cmd.Flags().GetBool("global")
		interactive, _ := cmd.Flags().GetBool("interactive")
		agentIDs, _ := cmd.Flags().GetStringArray("agent")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if method != "copy" && method != "symlink" {
			return fmt.Errorf("--method must be 'copy' or 'symlink'")
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}

		baseDir := cwd
		if global {
			baseDir = homeDir
		}

		var targets []Agent
		switch {
		case len(agentIDs) > 0:
			targets, err = resolveByIDs(agentIDs)
		case interactive:
			targets, err = resolveInteractive(cmd.InOrStdin(), out, baseDir, global)
		default:
			targets, err = resolveByHarness(baseDir, global)
		}
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			fmt.Fprintln(out, "No agent environments detected. Use --agent to specify one or --interactive to choose.")
			return nil
		}

		groups := groupTargetsByDir(targets, baseDir, global)

		if dryRun {
			canonicalPath := prettyPath(
				filepath.Join(homeDir, ".local", "share", skillName, "SKILL.md"),
				homeDir,
			)
			assetList := formatAssetList()
			for _, g := range groups {
				dir := prettyPath(g.skillDir, homeDir)
				fmt.Fprintf(out, "[dry-run] %s  (%s)\n", dir, formatAgentIDs(g.agents))
				if method == "symlink" {
					fmt.Fprintf(out, "          SKILL.md → %s (symlink)\n", canonicalPath)
				} else {
					fmt.Fprintf(out, "          SKILL.md (copy)\n")
				}
				fmt.Fprintf(out, "          assets: %s\n", assetList)
			}
			return nil
		}

		var canonical string
		if method == "symlink" {
			canonical, err = writeCanonical(homeDir)
			if err != nil {
				return fmt.Errorf("write canonical skill: %w", err)
			}
		}

		for _, g := range groups {
			skillDir := g.skillDir
			label := formatAgentIDs(g.agents)

			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				return fmt.Errorf("[%s] mkdir: %w", label, err)
			}

			dest := filepath.Join(skillDir, "SKILL.md")
			if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("[%s] remove existing file: %w", label, err)
			}

			if method == "symlink" {
				if err := os.Symlink(canonical, dest); err != nil {
					return fmt.Errorf("[%s] symlink: %w", label, err)
				}
			} else {
				if err := os.WriteFile(dest, embeddedSkill, 0o644); err != nil {
					return fmt.Errorf("[%s] write SKILL.md: %w", label, err)
				}
			}

			if err := writeAllAssets(skillDir); err != nil {
				return fmt.Errorf("[%s] write assets: %w", label, err)
			}

			fmt.Fprintf(out, "  ✓ %s → %s (%s)\n", label, prettyPath(skillDir, homeDir), method)
		}
		return nil
	},
}

// prettyPath replaces the home directory prefix with ~ for display.
func prettyPath(path, homeDir string) string {
	if homeDir == "" {
		return path
	}
	if path == homeDir {
		return "~"
	}
	if strings.HasPrefix(path, homeDir+string(filepath.Separator)) {
		return "~" + string(filepath.Separator) + path[len(homeDir)+1:]
	}
	return path
}

// targetGroup is a unique install directory together with the agents that share it.
type targetGroup struct {
	skillDir string
	agents   []Agent
}

// groupTargetsByDir groups targets by their computed install directory,
// preserving first-occurrence order and merging agents that map to the same path.
func groupTargetsByDir(targets []Agent, baseDir string, global bool) []targetGroup {
	seen := map[string]int{}
	var groups []targetGroup
	for _, a := range targets {
		dir := agentSkillDir(a, baseDir, global)
		if idx, ok := seen[dir]; ok {
			groups[idx].agents = append(groups[idx].agents, a)
		} else {
			seen[dir] = len(groups)
			groups = append(groups, targetGroup{skillDir: dir, agents: []Agent{a}})
		}
	}
	return groups
}

// formatAgentIDs formats a list of agents for display, abbreviating long lists.
func formatAgentIDs(agents []Agent) string {
	const maxShow = 3
	ids := make([]string, len(agents))
	for i, a := range agents {
		ids[i] = a.ID
	}
	if len(ids) <= maxShow {
		return strings.Join(ids, ", ")
	}
	return strings.Join(ids[:maxShow], ", ") + fmt.Sprintf(" +%d more", len(ids)-maxShow)
}

// formatAssetList returns the asset subdirectory list for display.
func formatAssetList() string {
	parts := make([]string, len(assetSubdirs))
	for i, s := range assetSubdirs {
		parts[i] = s + "/"
	}
	return strings.Join(parts, ", ")
}

// agentSkillDir returns the absolute skill install directory for an agent.
func agentSkillDir(a Agent, baseDir string, global bool) string {
	if global {
		return filepath.Join(baseDir, a.GlobalPath, skillName)
	}
	return filepath.Join(baseDir, a.ProjectPath, skillName)
}

// harnessRootDir returns the top-level harness directory for an agent
// (e.g. ".claude" for claude-code), relative to the install base.
func harnessRootDir(a Agent, global bool) string {
	p := a.ProjectPath
	if global {
		p = a.GlobalPath
	}
	parts := strings.SplitN(p, "/", 2)
	return parts[0]
}

// resolveByIDs resolves a specific list of agent IDs from the registry.
func resolveByIDs(ids []string) ([]Agent, error) {
	out := make([]Agent, 0, len(ids))
	for _, id := range ids {
		a, ok := agentByID(id)
		if !ok {
			return nil, fmt.Errorf("unknown agent %q — run 'skill-auditor init --help' for supported agents", id)
		}
		out = append(out, a)
	}
	return out, nil
}

// resolveByHarness auto-detects agents whose harness root directory exists under baseDir.
func resolveByHarness(baseDir string, global bool) ([]Agent, error) {
	var detected []Agent
	for _, a := range agentRegistry {
		harness := filepath.Join(baseDir, harnessRootDir(a, global))
		if _, err := os.Stat(harness); err == nil {
			detected = append(detected, a)
		}
	}
	return detected, nil
}

// resolveInteractive shows a numbered list of all agents and prompts the user to select.
func resolveInteractive(in io.Reader, out io.Writer, baseDir string, global bool) ([]Agent, error) {
	fmt.Fprintln(out, "Available agents (* = harness detected in target directory):")
	for i, a := range agentRegistry {
		marker := " "
		harness := filepath.Join(baseDir, harnessRootDir(a, global))
		if _, err := os.Stat(harness); err == nil {
			marker = "*"
		}
		path := a.ProjectPath
		if global {
			path = a.GlobalPath
		}
		fmt.Fprintf(out, "  %s %3d) %-20s %-30s (%s)\n", marker, i+1, a.ID, a.DisplayName, path)
	}
	fmt.Fprint(out, "\nSelect agents (comma-separated numbers, or 'all'): ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return nil, nil
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return nil, nil
	}
	if strings.ToLower(input) == "all" {
		result := make([]Agent, len(agentRegistry))
		copy(result, agentRegistry)
		return result, nil
	}

	var selected []Agent
	seen := map[string]bool{}
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(agentRegistry) {
			return nil, fmt.Errorf("invalid selection %q — enter numbers between 1 and %d", part, len(agentRegistry))
		}
		a := agentRegistry[n-1]
		if !seen[a.ID] {
			seen[a.ID] = true
			selected = append(selected, a)
		}
	}
	return selected, nil
}

// resolveTargets is the legacy entry-point used by tests.
func resolveTargets(ids []string, baseDir string, global bool) ([]Agent, error) {
	if len(ids) > 0 {
		return resolveByIDs(ids)
	}
	return resolveByHarness(baseDir, global)
}

// writeCanonical writes the embedded skill to the canonical location
// (~/.local/share/skill-quality-auditor/) and returns the SKILL.md path.
func writeCanonical(homeDir string) (string, error) {
	dir := filepath.Join(homeDir, ".local", "share", skillName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(dest, embeddedSkill, 0o644); err != nil {
		return "", err
	}
	if err := writeAllAssets(dir); err != nil {
		return "", err
	}
	return dest, nil
}

// writeAllAssets copies every embedded asset subdirectory into destDir.
func writeAllAssets(destDir string) error {
	type assetDir struct {
		fsys fs.FS
		root string
	}
	dirs := []assetDir{
		{embeddedRefs, "assets/references"},
		{embeddedEvals, "assets/evals"},
		{embeddedSchemas, "assets/schemas"},
		{embeddedTemplates, "assets/templates"},
		{embeddedRequirements, "assets/requirements"},
	}
	for _, d := range dirs {
		subdir := filepath.Base(d.root)
		if err := fs.WalkDir(d.fsys, d.root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(d.root, path)
			target := filepath.Join(destDir, subdir, rel)
			if entry.IsDir() {
				return os.MkdirAll(target, 0o755)
			}
			data, err := fs.ReadFile(d.fsys, path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0o644)
		}); err != nil {
			return err
		}
	}
	return nil
}

// writeRefs is kept for compatibility; delegates to writeAllAssets.
func writeRefs(destDir string) error {
	return writeAllAssets(destDir)
}

func init() {
	initCmd.Flags().StringArrayP("agent", "a", nil, "agent(s) to install into (default: auto-detect)")
	initCmd.Flags().BoolP("global", "g", false, "install to global skill directory (~/<agent-path>/)")
	initCmd.Flags().BoolP("interactive", "I", false, "interactively choose agents from the full registry list")
	initCmd.Flags().StringP("method", "m", "symlink", "installation method: symlink or copy")
	initCmd.Flags().BoolP("dry-run", "n", false, "print what would be done without touching disk")
	rootCmd.AddCommand(initCmd)
}
