package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/pantheon-org/skill-quality-auditor/scorer"
	"github.com/spf13/cobra"
)

func init() {
	var flags struct {
		json      bool
		store     bool
		failBelow string
		repoRoot  string
	}

	cmd := &cobra.Command{
		Use:   "batch <skill1> [skill2 ...]",
		Short: "Evaluate multiple skills",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepoRoot(flags.repoRoot)
			if err != nil {
				return fmt.Errorf("cannot determine repo root: %w", err)
			}

			type entry struct {
				arg    string
				result *scorer.Result
				err    error
			}

			var storeErrors []string
			entries := make([]entry, len(args))
			for i, arg := range args {
				skillPath := resolveSkillPath(arg, repoRoot)
				result, err := scorer.Score(cmd.Context(), skillPath)
				if err != nil {
					entries[i] = entry{arg: arg, err: err}
					continue
				}
				entries[i] = entry{arg: arg, result: result}

				if flags.store {
					if storeErr := reporter.Store(repoRoot, arg, result); storeErr != nil {
						fmt.Fprintf(os.Stderr, "warn: store %s: %v\n", arg, storeErr)
						storeErrors = append(storeErrors, fmt.Sprintf("%s: %v", arg, storeErr))
					}
				}
			}

			if flags.json {
				results := make([]*scorer.Result, 0, len(entries))
				for _, e := range entries {
					if e.result != nil {
						results = append(results, e.result)
					}
				}
				data, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal results: %w", err)
				}
				fmt.Println(string(data))
			} else {
				sort.Slice(entries, func(i, j int) bool {
					if entries[i].result == nil {
						return false
					}
					if entries[j].result == nil {
						return true
					}
					return entries[i].result.Total > entries[j].result.Total
				})

				maxLen := 0
				for _, e := range entries {
					if len(e.arg) > maxLen {
						maxLen = len(e.arg)
					}
				}
				if maxLen < 40 {
					maxLen = 40
				}

				totalScore := 0
				successCount := 0
				for _, e := range entries {
					if e.err != nil {
						fmt.Printf("%-*s  ERROR: %v\n", maxLen, e.arg, e.err)
						continue
					}
					fmt.Printf("%-*s  %-2s (%d/%d)\n", maxLen, e.arg, e.result.Grade, e.result.Total, e.result.MaxTotal)
					totalScore += e.result.Total
					successCount++
				}

				sep := strings.Repeat("─", maxLen+20)
				fmt.Println(sep)
				avg := 0
				if successCount > 0 {
					avg = totalScore / successCount
				}
				fmt.Printf("Total: %d skill(s)  Average: %d/140\n", len(entries), avg)
			}

			if flags.failBelow != "" {
				threshold, ok := scorer.GradeRank[flags.failBelow]
				if !ok {
					return fmt.Errorf("unknown grade %q for --fail-below", flags.failBelow)
				}
				for _, e := range entries {
					if e.result == nil {
						continue
					}
					if scorer.GradeRank[e.result.Grade] < threshold {
						return fmt.Errorf("skill %s scored %s, below threshold %s", e.arg, e.result.Grade, flags.failBelow)
					}
				}
			}

			if len(storeErrors) > 0 {
				return fmt.Errorf("store failed for %d skill(s): %s", len(storeErrors), strings.Join(storeErrors, "; "))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.json, "json", false, "emit JSON array output")
	cmd.Flags().BoolVar(&flags.store, "store", false, "persist each result to .context/audits/")
	cmd.Flags().StringVar(&flags.failBelow, "fail-below", "", "exit 1 if any skill scores below this grade (e.g. B+)")
	cmd.Flags().StringVar(&flags.repoRoot, "repo-root", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(cmd)
}
