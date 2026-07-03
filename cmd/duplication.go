package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/duplication"
	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/spf13/cobra"
)

var duplicationCmd = &cobra.Command{
	Use:   "duplication [skills-dir]",
	Short: "Detect duplicate or overlapping skills",
	Long:  "Performs pairwise word-level Jaccard similarity across all SKILL.md files and reports pairs above the High (20%) and Critical (35%) thresholds.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
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

		entries, err := duplication.Inventory(skillsDir)
		if err != nil {
			return fmt.Errorf("inventory skills: %w", err)
		}
		if len(entries) == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "no SKILL.md files found")
			return nil
		}

		pairs := duplication.Detect(entries)
		date := time.Now().Format("2006-01-02")

		format, err := resolveOutputFormat(cmd, OutputFormatMarkdown)
		if err != nil {
			return err
		}
		asJSON := format == OutputFormatJSON
		store, _ := cmd.Flags().GetBool("store")

		var rawBytes []byte
		if asJSON {
			data, err := json.MarshalIndent(pairs, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal pairs: %w", err)
			}
			rawBytes = data
			fmt.Fprintln(out, string(data))
		} else {
			report := reporter.DuplicationReport(pairs, entries, date)
			rawBytes = []byte(report)
			fmt.Fprint(out, report)
		}

		if store {
			ext := ".md"
			if asJSON {
				ext = ".json"
			}
			outDir := filepath.Join(repoRoot, ".context", "analysis")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return fmt.Errorf("create analysis dir: %w", err)
			}
			outFile := filepath.Join(outDir, fmt.Sprintf("duplication-report-%s%s", date, ext))
			if err := os.WriteFile(outFile, rawBytes, 0o644); err != nil {
				return fmt.Errorf("write report: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "report written to %s\n", outFile)
		}

		return exitCodeForPairs(pairs)
	},
}

// criticalDuplicationError signals that a Critical-severity duplication pair
// was found; Execute() checks ExitCode() to distinguish this from a generic
// command error in CI pipeline logic.
type criticalDuplicationError struct{}

func (criticalDuplicationError) Error() string { return "critical duplication detected" }
func (criticalDuplicationError) ExitCode() int { return 2 }

func exitCodeForPairs(pairs []duplication.Pair) error {
	for _, p := range pairs {
		if p.Severity == "Critical" {
			return criticalDuplicationError{}
		}
	}
	return nil
}

func init() {
	duplicationCmd.Flags().BoolP("json", "j", false, "emit JSON array output")
	duplicationCmd.Flags().BoolP("markdown", "m", false, "emit Markdown output (default)")
	duplicationCmd.Flags().BoolP("store", "s", false, "persist report to .context/analysis/")
	duplicationCmd.Flags().StringP("skills-dir", "d", "", "skills directory (default: <repo-root>/skills)")
	duplicationCmd.Flags().StringP("repo-root", "r", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(duplicationCmd)
}
