package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/analysis"
	"github.com/pantheon-org/skill-quality-auditor/duplication"
	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/spf13/cobra"
)

var canonicalSections = []string{
	"when to use", "usage", "examples", "prerequisites",
	"trigger", "anti-patterns", "output", "context",
}

var triggerWords = map[string]int{
	"example":      2,
	"trigger":      1,
	"never":        1,
	"always":       1,
	"output":       2,
	"prerequisite": 1,
}

var requiredSections = []string{"when to use", "examples", "trigger"}

var analyzeCmd = &cobra.Command{
	Use:   "analyze <skill>",
	Short: "Analyse a skill with TF-IDF keyword extraction and rule-based pattern detection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skillArg := args[0]

		repoRootFlag, _ := cmd.Flags().GetString("repo-root")
		repoRoot, err := resolveRepoRoot(repoRootFlag)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillPath := resolveSkillPath(skillArg, repoRoot)
		skillKey, err := canonicalSkillKey(skillPath, repoRoot)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(skillPath)
		if err != nil {
			return fmt.Errorf("read skill: %w", err)
		}
		content := string(data)

		skillsDir := filepath.Join(repoRoot, "skills")
		entries, _ := duplication.Inventory(skillsDir)
		corpus := make([]map[string]bool, len(entries))
		for i, e := range entries {
			corpus[i] = duplication.TokenSet(e.Content)
		}

		date := time.Now().Format("2006-01-02")

		asJSON, _ := cmd.Flags().GetBool("json")
		asMarkdown, _ := cmd.Flags().GetBool("markdown")
		if asJSON && asMarkdown {
			return fmt.Errorf("--json and --markdown are mutually exclusive")
		}
		if asMarkdown {
			asJSON = false
		} else {
			asJSON = true
		}

		semantic, _ := cmd.Flags().GetBool("semantic")
		patterns, _ := cmd.Flags().GetBool("patterns")
		store, _ := cmd.Flags().GetBool("store")
		limit, _ := cmd.Flags().GetInt("limit")
		switch {
		case semantic:
			return runSemanticOnly(cmd, content, corpus, limit)
		case patterns:
			return runPatternsOnly(cmd, content)
		default:
			return runPipeline(cmd, content, corpus, skillKey, date, limit, asJSON, store, repoRoot)
		}
	},
}

func runSemanticOnly(cmd *cobra.Command, content string, corpus []map[string]bool, limit int) error {
	out := cmd.OutOrStdout()
	keywords := analysis.ExtractKeywords(content, corpus, limit)
	fmt.Fprint(out, "| Rank | Term | Score |\n")
	fmt.Fprint(out, "|------|------|-------|\n")
	for i, kw := range keywords {
		fmt.Fprintf(out, "| %d | %s | %.4f |\n", i+1, kw.Term, kw.Score)
	}
	return nil
}

func runPatternsOnly(cmd *cobra.Command, content string) error {
	out := cmd.OutOrStdout()
	var rules []analysis.RuleMatch
	rules = append(rules, analysis.DetectRequiredSections(content, requiredSections)...)
	rules = append(rules, analysis.DetectTriggerFrequency(content, triggerWords)...)
	rules = append(rules, analysis.DetectStructuralConformance(content, canonicalSections))
	rules = append(rules, analysis.DetectAntiPatternSignals(content)...)

	fmt.Fprint(out, "| Rule | Matched | Score | Evidence |\n")
	fmt.Fprint(out, "|------|---------|-------|----------|\n")
	for _, rm := range rules {
		matched := "false"
		if rm.Matched {
			matched = "true"
		}
		evidence := strings.Join(rm.Evidence, ", ")
		fmt.Fprintf(out, "| %s | %s | %.2f | %s |\n", rm.Rule, matched, rm.Score, evidence)
	}
	return nil
}

func runPipeline(cmd *cobra.Command, content string, corpus []map[string]bool, skillKey, date string, limit int, asJSON, store bool, repoRoot string) error {
	out := cmd.OutOrStdout()
	keywords := analysis.ExtractKeywords(content, corpus, limit)

	var rules []analysis.RuleMatch
	rules = append(rules, analysis.DetectRequiredSections(content, requiredSections)...)
	rules = append(rules, analysis.DetectTriggerFrequency(content, triggerWords)...)
	rules = append(rules, analysis.DetectStructuralConformance(content, canonicalSections))
	rules = append(rules, analysis.DetectAntiPatternSignals(content)...)

	summary := buildSummary(rules, keywords)

	ca := reporter.CombinedAnalysis{
		SkillKey:    skillKey,
		Date:        date,
		Keywords:    keywords,
		RuleMatches: rules,
		Summary:     summary,
	}

	var rawBytes []byte

	if asJSON {
		data, err := reporter.CombinedJSON(ca)
		if err != nil {
			return fmt.Errorf("marshal analysis: %w", err)
		}
		rawBytes = data
		fmt.Fprintln(out, string(data))
	} else {
		output := reporter.CombinedMarkdown(ca)
		rawBytes = []byte(output)
		fmt.Fprint(out, output)
	}

	if store {
		ext := ".md"
		if asJSON {
			ext = ".json"
		}
		safeKey := strings.ReplaceAll(skillKey, "/", "-")
		outDir := filepath.Join(repoRoot, ".context", "analysis")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create analysis dir: %w", err)
		}
		outFile := filepath.Join(outDir, fmt.Sprintf("pattern-report-%s-%s%s", safeKey, date, ext))
		if err := os.WriteFile(outFile, rawBytes, 0o644); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "report written to %s\n", outFile)
	}

	return nil
}

func buildSummary(rules []analysis.RuleMatch, keywords []analysis.KeywordScore) string {
	matched := 0
	for _, r := range rules {
		if r.Matched {
			matched++
		}
	}

	topTerms := make([]string, 0, 3)
	for i, kw := range keywords {
		if i >= 3 {
			break
		}
		topTerms = append(topTerms, kw.Term)
	}

	if len(topTerms) == 0 {
		return fmt.Sprintf("%d/%d pattern rules matched.", matched, len(rules))
	}
	return fmt.Sprintf("%d/%d pattern rules matched. Top keywords: %s.", matched, len(rules), strings.Join(topTerms, ", "))
}

func init() {
	analyzeCmd.Flags().BoolP("semantic", "e", false, "run TF-IDF keyword extraction only")
	analyzeCmd.Flags().BoolP("patterns", "p", false, "run rule-based pattern detection only")
	analyzeCmd.Flags().BoolP("pipeline", "P", false, "run full pipeline (default when no flag given)")
	analyzeCmd.Flags().BoolP("json", "j", false, "emit JSON output (default)")
	analyzeCmd.Flags().BoolP("markdown", "m", false, "emit Markdown output instead of JSON")
	analyzeCmd.Flags().BoolP("store", "s", false, "write report to .context/analysis/")
	analyzeCmd.Flags().StringP("repo-root", "r", "", "repo root (auto-detected if empty)")
	analyzeCmd.Flags().IntP("limit", "l", 20, "max keywords to show")
	rootCmd.AddCommand(analyzeCmd)
}
