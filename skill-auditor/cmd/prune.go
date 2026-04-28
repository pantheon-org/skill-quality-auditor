package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var pruneKeep int

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune old audit directories, keeping only the most recent ones",
	Long: `Remove older audit runs from .context/audits/, keeping the N most recent
date-stamped directories per skill (default: 5).

The 'latest' symlink inside each skill's audit directory is preserved.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot("")
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		auditRoot := filepath.Join(repoRoot, ".context", "audits")
		if _, err := os.Stat(auditRoot); os.IsNotExist(err) {
			fmt.Printf("No audit directory found: %s\n", auditRoot)
			return nil
		}

		fmt.Println("=== Audit Pruning ===")
		fmt.Printf("Keeping last %d audit(s) per skill\n\n", pruneKeep)

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
				if i < pruneKeep {
					fmt.Printf("Keeping:  %s\n", fullPath)
					totalKept++
				} else {
					fmt.Printf("Removing: %s\n", fullPath)
					if err := os.RemoveAll(fullPath); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: could not remove %s: %v\n", fullPath, err)
					}
					totalRemoved++
				}
			}

			// Report latest symlink target if it exists.
			latestPath := filepath.Join(skillAuditDir, "latest")
			if target, err := os.Readlink(latestPath); err == nil {
				fmt.Printf("Latest symlink (%s) → %s\n", sd.Name(), target)
			}
		}

		fmt.Printf("\nDone. Kept %d, removed %d audit run(s).\n", totalKept, totalRemoved)
		return nil
	},
}

func init() {
	pruneCmd.Flags().IntVar(&pruneKeep, "keep", 5, "number of audit runs to keep per skill")
	rootCmd.AddCommand(pruneCmd)
}
