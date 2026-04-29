package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/pantheon-org/skill-quality-auditor/scorer"
	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch <skill1> [skill2 ...]",
	Short: "Evaluate multiple skills",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")
		asMarkdown, _ := cmd.Flags().GetBool("markdown")
		if asJSON && asMarkdown {
			return fmt.Errorf("--json and --markdown are mutually exclusive")
		}

		repoRootFlag, _ := cmd.Flags().GetString("repo-root")
		repoRoot, err := resolveRepoRoot(repoRootFlag)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		store, _ := cmd.Flags().GetBool("store")
		failBelow, _ := cmd.Flags().GetString("fail-below")

		out := cmd.OutOrStdout()

		var storeErrors []string
		entries := make([]batchEntry, len(args))
		for i, arg := range args {
			skillPath := resolveSkillPath(arg, repoRoot)
			result, err := scorer.Score(cmd.Context(), skillPath)
			if err != nil {
				entries[i] = batchEntry{arg: arg, err: err}
				continue
			}
			entries[i] = batchEntry{arg: arg, result: result}

			if store {
				if storeErr := reporter.Store(repoRoot, arg, result); storeErr != nil {
					fmt.Fprintf(os.Stderr, "warn: store %s: %v\n", arg, storeErr)
					storeErrors = append(storeErrors, fmt.Sprintf("%s: %v", arg, storeErr))
				}
			}
		}

		if asJSON {
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
			fmt.Fprintln(out, string(data))
		} else if asMarkdown {
			printBatchMarkdown(out, entries)
		} else {
			printBatchText(out, entries)
		}

		if failBelow != "" {
			threshold, ok := scorer.GradeRank[failBelow]
			if !ok {
				return fmt.Errorf("unknown grade %q for --fail-below", failBelow)
			}
			for _, e := range entries {
				if e.result == nil {
					continue
				}
				if scorer.GradeRank[e.result.Grade] < threshold {
					return fmt.Errorf("skill %s scored %s, below threshold %s", e.arg, e.result.Grade, failBelow)
				}
			}
		}

		if len(storeErrors) > 0 {
			return fmt.Errorf("store failed for %d skill(s): %s", len(storeErrors), strings.Join(storeErrors, "; "))
		}

		return nil
	},
}

type batchEntry struct {
	arg    string
	result *scorer.Result
	err    error
}

func sortBatchEntries(entries []batchEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].result == nil {
			return false
		}
		if entries[j].result == nil {
			return true
		}
		return entries[i].result.Total > entries[j].result.Total
	})
}

func printBatchMarkdown(out io.Writer, entries []batchEntry) {
	sortBatchEntries(entries)
	fmt.Fprintln(out, "| Skill | Grade | Score |")
	fmt.Fprintln(out, "| --- | --- | --- |")
	totalScore := 0
	successCount := 0
	for _, e := range entries {
		if e.err != nil {
			fmt.Fprintf(out, "| %s | ERROR | %v |\n", e.arg, e.err)
			continue
		}
		fmt.Fprintf(out, "| %s | %s | %d/%d |\n", e.arg, e.result.Grade, e.result.Total, e.result.MaxTotal)
		totalScore += e.result.Total
		successCount++
	}
	avg := 0
	if successCount > 0 {
		avg = totalScore / successCount
	}
	fmt.Fprintf(out, "\n**Total:** %d skill(s) | **Average:** %d/140\n", len(entries), avg)
}

func printBatchText(out io.Writer, entries []batchEntry) {
	sortBatchEntries(entries)
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
			fmt.Fprintf(out, "%-*s  ERROR: %v\n", maxLen, e.arg, e.err)
			continue
		}
		fmt.Fprintf(out, "%-*s  %-2s (%d/%d)\n", maxLen, e.arg, e.result.Grade, e.result.Total, e.result.MaxTotal)
		totalScore += e.result.Total
		successCount++
	}

	sep := strings.Repeat("─", maxLen+20)
	fmt.Fprintln(out, sep)
	avg := 0
	if successCount > 0 {
		avg = totalScore / successCount
	}
	fmt.Fprintf(out, "Total: %d skill(s)  Average: %d/140\n", len(entries), avg)
}

func init() {
	batchCmd.Flags().BoolP("json", "j", false, "emit JSON array output")
	batchCmd.Flags().BoolP("markdown", "m", false, "emit Markdown table output")
	batchCmd.Flags().BoolP("store", "s", false, "persist each result to .context/audits/")
	batchCmd.Flags().StringP("fail-below", "F", "", "exit 1 if any skill scores below this grade (e.g. B+)")
	batchCmd.Flags().StringP("repo-root", "r", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(batchCmd)
}
