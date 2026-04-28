package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/scorer"
	"github.com/spf13/cobra"
)

var (
	trendJSON     bool
	trendRepoRoot string
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
		repoRoot, err := resolveRepoRoot(trendRepoRoot)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		auditsRoot := filepath.Join(repoRoot, ".context", "audits")
		if !fileExists(auditsRoot) {
			return fmt.Errorf("no audits found — run 'batch ... --store' first")
		}

		entries, err := collectTrends(auditsRoot)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No skills with at least two stored audits.")
			return nil
		}

		// sort alphabetically
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Skill < entries[j].Skill
		})

		if trendJSON {
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal trends: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		printTrendTable(entries)
		return nil
	},
}

func collectTrends(auditsRoot string) ([]TrendEntry, error) {
	// Walk one level down: auditsRoot/<domain>/<skill>/ or auditsRoot/<skill>/
	// The actual layout is auditsRoot/<skillKey>/<date>/audit.json
	// where skillKey may contain path separators (domain/skill).
	// We walk two levels: first level = domain (or flat skill), second = skill.

	var results []TrendEntry

	// Find all audit.json files and group by skill key.
	auditsBySkill := map[string][]string{}
	err := filepath.WalkDir(auditsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "audit.json" {
			return err
		}
		rel, _ := filepath.Rel(auditsRoot, path)
		// rel = <skillKey>/<date>/audit.json
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 3 {
			return nil
		}
		// skill key is everything except last two components (date + filename)
		skillKey := strings.Join(parts[:len(parts)-2], "/")
		auditsBySkill[skillKey] = append(auditsBySkill[skillKey], path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk audits: %w", err)
	}

	for skill, paths := range auditsBySkill {
		if len(paths) < 2 {
			continue
		}
		// sort by date component (second-to-last dir)
		sort.Slice(paths, func(i, j int) bool {
			return dateFromAuditPath(paths[i]) < dateFromAuditPath(paths[j])
		})

		oldPath := paths[len(paths)-2]
		newPath := paths[len(paths)-1]

		oldResult, err := loadAuditJSON(oldPath)
		if err != nil {
			continue
		}
		newResult, err := loadAuditJSON(newPath)
		if err != nil {
			continue
		}

		delta := newResult.Total - oldResult.Total
		trend := "—"
		if delta > 0 {
			trend = "↑"
		} else if delta < 0 {
			trend = "↓"
		}

		results = append(results, TrendEntry{
			Skill:    skill,
			OldDate:  oldResult.Date,
			NewDate:  newResult.Date,
			OldScore: oldResult.Total,
			NewScore: newResult.Total,
			OldGrade: oldResult.Grade,
			NewGrade: newResult.Grade,
			Delta:    delta,
			Trend:    trend,
		})
	}
	return results, nil
}

func printTrendTable(entries []TrendEntry) {
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
	fmt.Println(hdr)
	fmt.Println(sep)

	for _, e := range entries {
		deltaStr := fmt.Sprintf("%+d", e.Delta)
		fmt.Printf("%-*s  %-10s  %-10s  %6d  %6d  %5s  %5s  %6s  %s\n",
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
	if !fileExists(p) {
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
	trendCmd.Flags().BoolVar(&trendJSON, "json", false, "emit JSON array output")
	trendCmd.Flags().StringVar(&trendRepoRoot, "repo-root", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(trendCmd)
}
