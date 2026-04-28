package reporter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/analysis"
)

func sampleCombinedAnalysis() CombinedAnalysis {
	return CombinedAnalysis{
		SkillKey: "domain/my-skill",
		Date:     "2026-04-28",
		Keywords: []analysis.KeywordScore{
			{Term: "authentication", TF: 0.05, IDF: 2.1, Score: 0.105},
			{Term: "oauth", TF: 0.03, IDF: 2.5, Score: 0.075},
			{Term: "token", TF: 0.04, IDF: 1.8, Score: 0.072},
		},
		RuleMatches: []analysis.RuleMatch{
			{Rule: "required-section:when to use", Matched: true, Score: 1.0, Evidence: []string{"when to use"}},
			{Rule: "required-section:examples", Matched: false, Score: 0.0, Evidence: []string{}},
			{Rule: "anti-pattern:hedge-language", Matched: true, Score: 1.5, Evidence: []string{"maybe", "possibly"}},
		},
		Summary: "2/3 pattern rules matched. Top keywords: authentication, oauth, token.",
	}
}

func TestCombinedMarkdown_sectionHeaders(t *testing.T) {
	ca := sampleCombinedAnalysis()
	out := CombinedMarkdown(ca)

	for _, hdr := range []string{
		"# Pattern Analysis Report",
		"## Top Keywords (TF-IDF)",
		"## Pattern Detection Results",
		"## Summary",
	} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected section header %q in output", hdr)
		}
	}
}

func TestCombinedMarkdown_titleContainsSkillKeyAndDate(t *testing.T) {
	ca := sampleCombinedAnalysis()
	out := CombinedMarkdown(ca)
	if !strings.Contains(out, "domain/my-skill") {
		t.Error("expected skillKey in title")
	}
	if !strings.Contains(out, "2026-04-28") {
		t.Error("expected date in title")
	}
}

func TestCombinedMarkdown_keywordTableRows(t *testing.T) {
	ca := sampleCombinedAnalysis()
	out := CombinedMarkdown(ca)

	for _, term := range []string{"authentication", "oauth", "token"} {
		if !strings.Contains(out, term) {
			t.Errorf("expected keyword %q in output", term)
		}
	}
	if !strings.Contains(out, "| Rank | Term | Score |") {
		t.Error("expected keyword table header")
	}
}

func TestCombinedMarkdown_ruleMatchRows(t *testing.T) {
	ca := sampleCombinedAnalysis()
	out := CombinedMarkdown(ca)

	for _, rule := range []string{"required-section:when to use", "required-section:examples", "anti-pattern:hedge-language"} {
		if !strings.Contains(out, rule) {
			t.Errorf("expected rule %q in output", rule)
		}
	}
	if !strings.Contains(out, "| Rule | Matched | Score | Evidence |") {
		t.Error("expected rule match table header")
	}
}

func TestCombinedMarkdown_summaryText(t *testing.T) {
	ca := sampleCombinedAnalysis()
	out := CombinedMarkdown(ca)
	if !strings.Contains(out, ca.Summary) {
		t.Errorf("expected summary %q in output", ca.Summary)
	}
}

func TestCombinedMarkdown_emptyKeywords(t *testing.T) {
	ca := sampleCombinedAnalysis()
	ca.Keywords = nil
	out := CombinedMarkdown(ca)
	if !strings.Contains(out, "No keywords found") {
		t.Error("expected 'No keywords found' for empty keywords")
	}
}

func TestCombinedMarkdown_emptyRuleMatches(t *testing.T) {
	ca := sampleCombinedAnalysis()
	ca.RuleMatches = nil
	out := CombinedMarkdown(ca)
	if !strings.Contains(out, "No pattern rules run") {
		t.Error("expected 'No pattern rules run' for empty rule matches")
	}
}

func TestCombinedMarkdown_noPanic(t *testing.T) {
	ca := CombinedAnalysis{}
	out := CombinedMarkdown(ca)
	if out == "" {
		t.Error("expected non-empty output even for zero-value CombinedAnalysis")
	}
}

func TestCombinedJSON_roundTrip(t *testing.T) {
	ca := sampleCombinedAnalysis()
	data, err := CombinedJSON(ca)
	if err != nil {
		t.Fatalf("CombinedJSON error: %v", err)
	}

	var got CombinedAnalysis
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got.SkillKey != ca.SkillKey {
		t.Errorf("skillKey: got %q, want %q", got.SkillKey, ca.SkillKey)
	}
	if got.Date != ca.Date {
		t.Errorf("date: got %q, want %q", got.Date, ca.Date)
	}
	if len(got.Keywords) != len(ca.Keywords) {
		t.Errorf("keywords length: got %d, want %d", len(got.Keywords), len(ca.Keywords))
	}
	if len(got.RuleMatches) != len(ca.RuleMatches) {
		t.Errorf("ruleMatches length: got %d, want %d", len(got.RuleMatches), len(ca.RuleMatches))
	}
}

func TestCombinedJSON_keysPresent(t *testing.T) {
	ca := sampleCombinedAnalysis()
	data, err := CombinedJSON(ca)
	if err != nil {
		t.Fatalf("CombinedJSON error: %v", err)
	}
	s := string(data)
	for _, key := range []string{"skillKey", "date", "keywords", "ruleMatches", "summary"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Errorf("expected JSON key %q in output", key)
		}
	}
}

func TestCombinedJSON_emptySlices(t *testing.T) {
	ca := CombinedAnalysis{SkillKey: "x", Date: "2026-04-28"}
	data, err := CombinedJSON(ca)
	if err != nil {
		t.Fatalf("CombinedJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}
