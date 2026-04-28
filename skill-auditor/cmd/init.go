package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const skillName = "skill-quality-auditor"

var (
	initAgents []string
	initGlobal bool
	initMethod string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install the skill-quality-auditor skill into agent environments",
	Long: `Install the embedded skill-quality-auditor SKILL.md into one or more agent
skill directories. When no --agent flag is given, all agents whose global skill
directory already exists on this machine are targeted automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if initMethod != "copy" && initMethod != "symlink" {
			return fmt.Errorf("--method must be 'copy' or 'symlink'")
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}

		targets, err := resolveTargets(initAgents, homeDir, initGlobal)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			fmt.Println("No agent environments detected. Use --agent to specify one explicitly.")
			return nil
		}

		var canonical string
		if initMethod == "symlink" {
			canonical, err = writeCanonical(homeDir)
			if err != nil {
				return fmt.Errorf("write canonical skill: %w", err)
			}
		}

		for _, a := range targets {
			skillDir := filepath.Join(a.SkillDir(homeDir, initGlobal), skillName)
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				return fmt.Errorf("[%s] mkdir: %w", a.ID, err)
			}

			dest := filepath.Join(skillDir, "SKILL.md")
			if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("[%s] remove existing file: %w", a.ID, err)
			}

			if initMethod == "symlink" {
				if err := os.Symlink(canonical, dest); err != nil {
					return fmt.Errorf("[%s] symlink: %w", a.ID, err)
				}
			} else {
				if err := os.WriteFile(dest, embeddedSkill, 0o644); err != nil {
					return fmt.Errorf("[%s] write SKILL.md: %w", a.ID, err)
				}
			}

			if err := writeRefs(skillDir); err != nil {
				return fmt.Errorf("[%s] write references: %w", a.ID, err)
			}

			fmt.Printf("  ✓ %s → %s (%s)\n", a.ID, skillDir, initMethod)
		}
		return nil
	},
}

// resolveTargets returns the agents to install into.
// If ids is non-empty, each id must exist in the registry.
// Otherwise, auto-detect by probing global skill dirs.
func resolveTargets(ids []string, homeDir string, global bool) ([]Agent, error) {
	if len(ids) > 0 {
		agents := make([]Agent, 0, len(ids))
		for _, id := range ids {
			a, ok := agentByID(id)
			if !ok {
				return nil, fmt.Errorf("unknown agent %q — run 'skill-auditor init --help' for supported agents", id)
			}
			agents = append(agents, a)
		}
		return agents, nil
	}

	// auto-detect: include any agent whose global skill dir exists
	var detected []Agent
	for _, a := range agentRegistry {
		dir := filepath.Join(homeDir, a.GlobalPath)
		if _, err := os.Stat(dir); err == nil {
			detected = append(detected, a)
		}
	}
	return detected, nil
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
	if err := writeRefs(dir); err != nil {
		return "", err
	}
	return dest, nil
}

// writeRefs extracts the embedded references directory into destDir/references/.
func writeRefs(destDir string) error {
	return fs.WalkDir(embeddedRefs, "assets/references", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// path is like "assets/references/foo.md" — strip "assets/references" prefix
		rel, _ := filepath.Rel("assets/references", path)
		target := filepath.Join(destDir, "references", rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := embeddedRefs.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func init() {
	initCmd.Flags().StringArrayVar(&initAgents, "agent", nil, "agent(s) to install into (default: auto-detect)")
	initCmd.Flags().BoolVarP(&initGlobal, "global", "g", false, "install to global skill directory (~/<agent>/skills/)")
	initCmd.Flags().StringVar(&initMethod, "method", "symlink", "installation method: symlink or copy")
	rootCmd.AddCommand(initCmd)
}
