package reporter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/analysis"
)

// CombinedAnalysis is the output of the full analysis pipeline for one skill.
type CombinedAnalysis struct {
	SkillKey    string                  `json:"skillKey"`
	Date        string                  `json:"date"`
	Keywords    []analysis.KeywordScore `json:"keywords"`
	RuleMatches []analysis.RuleMatch    `json:"ruleMatches"`
	Summary     string                  `json:"summary"`
}

// CombinedMarkdown formats a CombinedAnalysis as a markdown report.
func CombinedMarkdown(ca CombinedAnalysis) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Pattern Analysis Report — %s — %s\n\n", ca.SkillKey, ca.Date)

	fmt.Fprintf(&sb, "## Top Keywords (TF-IDF)\n\n")
	if len(ca.Keywords) == 0 {
		fmt.Fprintf(&sb, "No keywords found.\n\n")
	} else {
		fmt.Fprintf(&sb, "| Rank | Term | Score |\n")
		fmt.Fprintf(&sb, "|------|------|-------|\n")
		for i, kw := range ca.Keywords {
			fmt.Fprintf(&sb, "| %d | %s | %.4f |\n", i+1, kw.Term, kw.Score)
		}
		fmt.Fprintf(&sb, "\n")
	}

	fmt.Fprintf(&sb, "## Pattern Detection Results\n\n")
	if len(ca.RuleMatches) == 0 {
		fmt.Fprintf(&sb, "No pattern rules run.\n\n")
	} else {
		fmt.Fprintf(&sb, "| Rule | Matched | Score | Evidence |\n")
		fmt.Fprintf(&sb, "|------|---------|-------|----------|\n")
		for _, rm := range ca.RuleMatches {
			matched := "false"
			if rm.Matched {
				matched = "true"
			}
			evidence := strings.Join(rm.Evidence, ", ")
			fmt.Fprintf(&sb, "| %s | %s | %.2f | %s |\n", rm.Rule, matched, rm.Score, evidence)
		}
		fmt.Fprintf(&sb, "\n")
	}

	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "%s\n", ca.Summary)

	return sb.String()
}

// CombinedJSON serialises CombinedAnalysis as indented JSON.
func CombinedJSON(ca CombinedAnalysis) ([]byte, error) {
	return json.MarshalIndent(ca, "", "  ")
}
