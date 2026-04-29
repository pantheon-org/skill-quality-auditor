package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune old audit directories, keeping only the most recent ones",
	Long: `Remove older audit runs from .context/audits/, keeping the N most recent
date-stamped directories per skill (default: 5).

The 'latest' symlink inside each skill's audit directory is preserved.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		keep, _ := cmd.Flags().GetInt("keep")
		repoRoot, err := resolveRepoRoot("")
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		auditRoot := filepath.Join(repoRoot, ".context", "audits")
		if _, err := os.Stat(auditRoot); os.IsNotExist(err) {
			fmt.Fprintf(out, "No audit directory found: %s\n", auditRoot)
			return nil
		}

		fmt.Fprintln(out, "=== Audit Pruning ===")
		fmt.Fprintf(out, "Keeping last %d audit(s) per skill\n\n", keep)

		skillDirs, err := os.ReadDir(auditRoot)
		if err != nil {
			return fmt.Errorf("cannot read audit root: %w", err)
		}

		totalKept, totalRemoved := 0, 0

		for _, sd := range skillDirs {
			if !sd.IsDir() {
				continue
			}
			skillAuditDir := filepath.Join(auditRoot, sd.Name())

			dateDirs, err := os.ReadDir(skillAuditDir)
			if err != nil {
				continue
			}

			var dirs []string
			for _, d := range dateDirs {
				if !d.IsDir() || d.Name() == "latest" {
					continue
				}
				dirs = append(dirs, d.Name())
			}

			// Sort descending (newest first) — date-stamped dirs sort lexicographically.
			sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

			for i, name := range dirs {
				fullPath := filepath.Join(skillAuditDir, name)
				if i < keep {
					fmt.Fprintf(out, "Keeping:  %s\n", fullPath)
					totalKept++
				} else {
					fmt.Fprintf(out, "Removing: %s\n", fullPath)
					if err := os.RemoveAll(fullPath); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: could not remove %s: %v\n", fullPath, err)
					}
					totalRemoved++
				}
			}

			// Report latest symlink target if it exists.
			latestPath := filepath.Join(skillAuditDir, "latest")
			if target, err := os.Readlink(latestPath); err == nil {
				fmt.Fprintf(out, "Latest symlink (%s) → %s\n", sd.Name(), target)
			}
		}

		fmt.Fprintf(out, "\nDone. Kept %d, removed %d audit run(s).\n", totalKept, totalRemoved)
		return nil
	},
}

func init() {
	pruneCmd.Flags().Int("keep", 5, "number of audit runs to keep per skill")
	rootCmd.AddCommand(pruneCmd)
}
