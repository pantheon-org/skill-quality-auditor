package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
	"github.com/spf13/cobra"
)

// TrendEntry holds the computed trend for a single skill.
type TrendEntry struct {
	Skill    string `json:"skill"`
	OldDate  string `json:"old_date"`
	NewDate  string `json:"new_date"`
	OldScore int    `json:"old_score"`
	NewScore int    `json:"new_score"`
	OldGrade string `json:"old_grade"`
	NewGrade string `json:"new_grade"`
	Delta    int    `json:"delta"`
	Trend    string `json:"trend"` // "↑", "↓", "—"
}

var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Show score trends across stored audits",
	Long:  "Reads the two most recent stored audits per skill from .context/audits/ and reports score deltas.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		repoRootFlag, _ := cmd.Flags().GetString("repo-root")
		repoRoot, err := resolveRepoRoot(repoRootFlag)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		auditsRoot := filepath.Join(repoRoot, ".context", "audits")
		if !pathExists(auditsRoot) {
			return fmt.Errorf("no audits found — run 'batch ... --store' first")
		}

		entries, err := collectTrends(auditsRoot)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Fprintln(out, "No skills with at least two stored audits.")
			return nil
		}

		// sort alphabetically
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Skill < entries[j].Skill
		})

		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal trends: %w", err)
			}
			fmt.Fprintln(out, string(data))
			return nil
		}

		printTrendTable(out, entries)
		return nil
	},
}

func collectTrends(auditsRoot string) ([]TrendEntry, error) {
	auditsBySkill, err := groupAuditsBySkill(auditsRoot)
	if err != nil {
		return nil, err
	}

	var results []TrendEntry
	for skill, paths := range auditsBySkill {
		if entry, ok := buildTrendEntry(skill, paths); ok {
			results = append(results, entry)
		}
	}
	return results, nil
}

func groupAuditsBySkill(auditsRoot string) (map[string][]string, error) {
	auditsBySkill := map[string][]string{}
	err := filepath.WalkDir(auditsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "audit.json" {
			return err
		}
		rel, _ := filepath.Rel(auditsRoot, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 3 {
			return nil
		}
		skillKey := strings.Join(parts[:len(parts)-2], "/")
		auditsBySkill[skillKey] = append(auditsBySkill[skillKey], path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk audits: %w", err)
	}
	return auditsBySkill, nil
}

func buildTrendEntry(skill string, paths []string) (TrendEntry, bool) {
	if len(paths) < 2 {
		return TrendEntry{}, false
	}
	sort.Slice(paths, func(i, j int) bool {
		return dateFromAuditPath(paths[i]) < dateFromAuditPath(paths[j])
	})
	oldResult, err := loadAuditJSON(paths[len(paths)-2])
	if err != nil {
		return TrendEntry{}, false
	}
	newResult, err := loadAuditJSON(paths[len(paths)-1])
	if err != nil {
		return TrendEntry{}, false
	}
	delta := newResult.Total - oldResult.Total
	trend := trendArrow(delta)
	return TrendEntry{
		Skill:    skill,
		OldDate:  oldResult.Date,
		NewDate:  newResult.Date,
		OldScore: oldResult.Total,
		NewScore: newResult.Total,
		OldGrade: oldResult.Grade,
		NewGrade: newResult.Grade,
		Delta:    delta,
		Trend:    trend,
	}, true
}

func trendArrow(delta int) string {
	if delta > 0 {
		return "↑"
	}
	if delta < 0 {
		return "↓"
	}
	return "—"
}

func printTrendTable(out io.Writer, entries []TrendEntry) {
	// column widths
	maxSkill := 5
	for _, e := range entries {
		if len(e.Skill) > maxSkill {
			maxSkill = len(e.Skill)
		}
	}

	hdr := fmt.Sprintf("%-*s  %-10s  %-10s  %6s  %6s  %5s  %5s  %6s  %s",
		maxSkill, "Skill", "Old Date", "New Date", "Old", "New", "Old G", "New G", "Delta", "Trend")
	sep := strings.Repeat("─", len(hdr))
	fmt.Fprintln(out, hdr)
	fmt.Fprintln(out, sep)

	for _, e := range entries {
		deltaStr := fmt.Sprintf("%+d", e.Delta)
		fmt.Fprintf(out, "%-*s  %-10s  %-10s  %6d  %6d  %5s  %5s  %6s  %s\n",
			maxSkill, e.Skill,
			e.OldDate, e.NewDate,
			e.OldScore, e.NewScore,
			e.OldGrade, e.NewGrade,
			deltaStr, e.Trend)
	}
}

func dateFromAuditPath(path string) string {
	// path = .../audits/<skill>/<date>/audit.json
	dir := filepath.Dir(path)
	return filepath.Base(dir)
}

// latestAuditJSON returns the path and date of the most recent audit.json for a skill.
func latestAuditJSON(auditsBase string) (path, date string, err error) {
	entries, readErr := os.ReadDir(auditsBase)
	if readErr != nil {
		return "", "", readErr
	}

	// dates are YYYY-MM-DD — lexicographic sort gives chronological order
	var dates []string
	for _, e := range entries {
		if e.IsDir() {
			dates = append(dates, e.Name())
		}
	}
	if len(dates) == 0 {
		return "", "", fmt.Errorf("no audit dates found in %s", auditsBase)
	}
	sort.Strings(dates)
	latest := dates[len(dates)-1]
	p := filepath.Join(auditsBase, latest, "audit.json")
	if !pathExists(p) {
		return "", "", fmt.Errorf("audit.json not found at %s", p)
	}
	return p, latest, nil
}

// loadAuditJSON reads and unmarshals an audit.json file.
func loadAuditJSON(path string) (*scorer.Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r scorer.Result
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func init() {
	trendCmd.Flags().Bool("json", false, "emit JSON array output")
	trendCmd.Flags().String("repo-root", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(trendCmd)
}
