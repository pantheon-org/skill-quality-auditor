package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint [skills-dir]",
	Short: "Check basic consistency across skill directories",
	Long: `Scan each skill directory and report structural issues.

Checks per skill:
  - SKILL.md exists
  - SKILL.md contains a frontmatter block (---)
  - scripts/*.sh files use #!/usr/bin/env sh shebang

Issue tags: MISSING_SKILL, NO_FRONTMATTER, BAD_SHEBANG

Exit code equals the number of issues found (0 = clean).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot("")
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillsDir := filepath.Join(repoRoot, "skills")
		if len(args) == 1 {
			skillsDir = args[0]
			if !filepath.IsAbs(skillsDir) {
				skillsDir = filepath.Join(repoRoot, skillsDir)
			}
		}

		if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
			fmt.Printf("Consistency check complete: 0 issue(s) found\n")
			return nil
		}

		issues := 0

		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			return fmt.Errorf("cannot read skills dir: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillName := entry.Name()
			skillDir := filepath.Join(skillsDir, skillName)

			skillMD := filepath.Join(skillDir, "SKILL.md")
			if _, err := os.Stat(skillMD); os.IsNotExist(err) {
				fmt.Printf("MISSING_SKILL: %s/SKILL.md\n", skillName)
				issues++
				continue
			}

			data, err := os.ReadFile(skillMD)
			if err == nil {
				if !strings.Contains(string(data), "---") {
					fmt.Printf("NO_FRONTMATTER: %s/SKILL.md\n", skillName)
					issues++
				}
			}

			scriptsDir := filepath.Join(skillDir, "scripts")
			if info, err := os.Stat(scriptsDir); err == nil && info.IsDir() {
				scriptEntries, _ := os.ReadDir(scriptsDir)
				for _, se := range scriptEntries {
					if se.IsDir() || !strings.HasSuffix(se.Name(), ".sh") {
						continue
					}
					scriptPath := filepath.Join(scriptsDir, se.Name())
					line := firstLine(scriptPath)
					if line != "#!/usr/bin/env sh" {
						fmt.Printf("BAD_SHEBANG: %s\n", scriptPath)
						issues++
					}
				}
			}
		}

		fmt.Printf("Consistency check complete: %d issue(s) found\n", issues)
		if issues > 0 {
			return fmt.Errorf("lint failed (%d issue(s))", issues)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
