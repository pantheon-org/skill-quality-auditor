package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/analysis"
	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/duplication"
	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/reporter"
	"github.com/spf13/cobra"
)

var (
	analyzeJSON     bool
	analyzeStore    bool
	analyzeSemantic bool
	analyzePatterns bool
	analyzePipeline bool
	analyzeRepoRoot string
	analyzeLimit    int
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

		repoRoot, err := resolveRepoRoot(analyzeRepoRoot)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillPath := resolveSkillPath(skillArg, repoRoot)
		skillKey := canonicalSkillKey(skillPath, repoRoot)

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

		switch {
		case analyzeSemantic:
			return runSemanticOnly(cmd, content, corpus, analyzeLimit)
		case analyzePatterns:
			return runPatternsOnly(cmd, content)
		default:
			return runPipeline(cmd, content, corpus, skillKey, date, analyzeLimit, analyzeJSON, analyzeStore, repoRoot)
		}
	},
}

func runSemanticOnly(_ *cobra.Command, content string, corpus []map[string]bool, limit int) error {
	keywords := analysis.ExtractKeywords(content, corpus, limit)
	fmt.Print("| Rank | Term | Score |\n")
	fmt.Print("|------|------|-------|\n")
	for i, kw := range keywords {
		fmt.Printf("| %d | %s | %.4f |\n", i+1, kw.Term, kw.Score)
	}
	return nil
}

func runPatternsOnly(_ *cobra.Command, content string) error {
	var rules []analysis.RuleMatch
	rules = append(rules, analysis.DetectRequiredSections(content, requiredSections)...)
	rules = append(rules, analysis.DetectTriggerFrequency(content, triggerWords)...)
	rules = append(rules, analysis.DetectStructuralConformance(content, canonicalSections))
	rules = append(rules, analysis.DetectAntiPatternSignals(content)...)

	fmt.Print("| Rule | Matched | Score | Evidence |\n")
	fmt.Print("|------|---------|-------|----------|\n")
	for _, rm := range rules {
		matched := "false"
		if rm.Matched {
			matched = "true"
		}
		evidence := strings.Join(rm.Evidence, ", ")
		fmt.Printf("| %s | %s | %.2f | %s |\n", rm.Rule, matched, rm.Score, evidence)
	}
	return nil
}

func runPipeline(_ *cobra.Command, content string, corpus []map[string]bool, skillKey, date string, limit int, asJSON, store bool, repoRoot string) error {
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
		fmt.Println(string(data))
	} else {
		output := reporter.CombinedMarkdown(ca)
		rawBytes = []byte(output)
		fmt.Print(output)
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
		fmt.Fprintf(os.Stderr, "report written to %s\n", outFile)
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
	analyzeCmd.Flags().BoolVar(&analyzeSemantic, "semantic", false, "run TF-IDF keyword extraction only")
	analyzeCmd.Flags().BoolVar(&analyzePatterns, "patterns", false, "run rule-based pattern detection only")
	analyzeCmd.Flags().BoolVar(&analyzePipeline, "pipeline", false, "run full pipeline (default when no flag given)")
	analyzeCmd.Flags().BoolVar(&analyzeJSON, "json", false, "emit JSON output")
	analyzeCmd.Flags().BoolVar(&analyzeStore, "store", false, "write report to .context/analysis/")
	analyzeCmd.Flags().StringVar(&analyzeRepoRoot, "repo-root", "", "repo root (auto-detected if empty)")
	analyzeCmd.Flags().IntVar(&analyzeLimit, "limit", 20, "max keywords to show")
	rootCmd.AddCommand(analyzeCmd)
}
