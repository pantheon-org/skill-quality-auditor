package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/duplication"
	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/reporter"
	"github.com/spf13/cobra"
)

var (
	aggFamily    string
	aggDryRun    bool
	aggRepoRoot  string
	aggSkillsDir string
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate --family <prefix> [skills-dir]",
	Short: "Generate an aggregation plan for a skill family",
	Long:  "Analyses all skills whose key starts with the given family prefix and produces an aggregation plan.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if aggFamily == "" {
			return fmt.Errorf("--family is required")
		}

		repoRoot, err := resolveRepoRoot(aggRepoRoot)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillsDir := aggSkillsDir
		if len(args) == 1 {
			skillsDir = args[0]
		}
		if skillsDir == "" {
			skillsDir = filepath.Join(repoRoot, "skills")
		}

		if !fileExists(skillsDir) {
			return fmt.Errorf("skills directory not found: %s", skillsDir)
		}

		all, err := duplication.Inventory(skillsDir)
		if err != nil {
			return fmt.Errorf("inventory skills: %w", err)
		}

		// filter to family
		var family []duplication.SkillEntry
		for _, e := range all {
			name := skillBaseName(e.Key)
			if strings.HasPrefix(name, aggFamily) {
				family = append(family, e)
			}
		}

		if len(family) == 0 {
			return fmt.Errorf("no skills found with family prefix %q", aggFamily)
		}

		pairs := duplication.Detect(family)
		date := time.Now().Format("2006-01-02")
		plan := reporter.AggregationPlan(aggFamily, family, pairs, date)

		if aggDryRun {
			fmt.Print(plan)
			return nil
		}

		outDir := filepath.Join(repoRoot, ".context", "analysis")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create analysis dir: %w", err)
		}
		outFile := filepath.Join(outDir, fmt.Sprintf("aggregation-plan-%s-%s.md", aggFamily, date))
		if err := os.WriteFile(outFile, []byte(plan), 0o644); err != nil {
			return fmt.Errorf("write plan: %w", err)
		}
		fmt.Print(plan)
		fmt.Fprintf(os.Stderr, "plan written to %s\n", outFile)
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
	aggregateCmd.Flags().StringVar(&aggFamily, "family", "", "skill family prefix (e.g. bdd, typescript)")
	aggregateCmd.Flags().BoolVar(&aggDryRun, "dry-run", false, "print plan to stdout without writing to disk")
	aggregateCmd.Flags().StringVar(&aggSkillsDir, "skills-dir", "", "skills directory (default: <repo-root>/skills)")
	aggregateCmd.Flags().StringVar(&aggRepoRoot, "repo-root", "", "repo root (auto-detected if empty)")
	if err := aggregateCmd.MarkFlagRequired("family"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(aggregateCmd)
}
