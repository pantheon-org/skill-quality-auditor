package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/duplication"
	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/spf13/cobra"
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate --family <prefix> [skills-dir]",
	Short: "Generate an aggregation plan for a skill family",
	Long:  "Analyses all skills whose key starts with the given family prefix and produces an aggregation plan.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		family, _ := cmd.Flags().GetString("family")
		if family == "" {
			return fmt.Errorf("--family is required")
		}

		repoRootFlag, _ := cmd.Flags().GetString("repo-root")
		repoRoot, err := resolveRepoRoot(repoRootFlag)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillsDirFlag, _ := cmd.Flags().GetString("skills-dir")
		skillsDir := skillsDirFlag
		if len(args) == 1 {
			skillsDir = args[0]
		}
		if skillsDir == "" {
			skillsDir = filepath.Join(repoRoot, "skills")
		}

		if !pathExists(skillsDir) {
			return fmt.Errorf("skills directory not found: %s", skillsDir)
		}

		all, err := duplication.Inventory(skillsDir)
		if err != nil {
			return fmt.Errorf("inventory skills: %w", err)
		}

		// filter to family
		var familyEntries []duplication.SkillEntry
		for _, e := range all {
			name := skillBaseName(e.Key)
			if strings.HasPrefix(name, family) {
				familyEntries = append(familyEntries, e)
			}
		}

		if len(familyEntries) == 0 {
			return fmt.Errorf("no skills found with family prefix %q", family)
		}

		pairs := duplication.Detect(familyEntries)
		date := time.Now().Format("2006-01-02")
		plan := reporter.AggregationPlan(family, familyEntries, pairs, date)

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Fprint(out, plan)
			return nil
		}

		outDir := filepath.Join(repoRoot, ".context", "analysis")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create analysis dir: %w", err)
		}
		outFile := filepath.Join(outDir, fmt.Sprintf("aggregation-plan-%s-%s.md", family, date))
		if err := os.WriteFile(outFile, []byte(plan), 0o644); err != nil {
			return fmt.Errorf("write plan: %w", err)
		}
		fmt.Fprint(out, plan)
		fmt.Fprintf(cmd.ErrOrStderr(), "plan written to %s\n", outFile)
		return nil
	},
}

// skillBaseName returns the last path component of a skill key (after the last /).
func skillBaseName(key string) string {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return key
}

func init() {
	aggregateCmd.Flags().String("family", "", "skill family prefix (e.g. bdd, typescript)")
	aggregateCmd.Flags().Bool("dry-run", false, "print plan to stdout without writing to disk")
	aggregateCmd.Flags().String("skills-dir", "", "skills directory (default: <repo-root>/skills)")
	aggregateCmd.Flags().String("repo-root", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(aggregateCmd)
}
