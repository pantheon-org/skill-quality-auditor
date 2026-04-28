package reporter

import (
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/duplication"
)

var testEntries = []duplication.SkillEntry{
	{Key: "bdd/bdd-gherkin", Content: "## Gherkin Syntax\nNEVER mix steps"},
	{Key: "bdd/bdd-patterns", Content: "## BDD Patterns\nAlways write scenarios"},
	{Key: "other/unrelated", Content: "## Cooking\nFlour butter sugar"},
}

var testPairs = []duplication.Pair{
	{A: "bdd/bdd-gherkin", B: "bdd/bdd-patterns", Similarity: 0.42, Severity: "Critical"},
	{A: "bdd/bdd-gherkin", B: "other/unrelated", Similarity: 0.22, Severity: "High"},
}

func TestDuplicationReport_structure(t *testing.T) {
	report := DuplicationReport(testPairs, testEntries, "2026-04-27")

	for _, want := range []string{
		"# Duplication Report",
		"2026-04-27",
		"## Summary",
		"Skills analysed: 3",
		"Pairs with >20% similarity: 2",
		"Critical (>35%): 1",
		"## High-Priority Candidates",
		"## Recommendations",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q", want)
		}
	}
}

func TestDuplicationReport_empty(t *testing.T) {
	report := DuplicationReport(nil, testEntries, "2026-04-27")
	if !strings.Contains(report, "No duplication detected") {
		t.Errorf("expected no-duplication message, got:\n%s", report)
	}
}

func TestDuplicationReport_showsPairKeys(t *testing.T) {
	report := DuplicationReport(testPairs, testEntries, "2026-04-27")
	if !strings.Contains(report, "bdd-gherkin") || !strings.Contains(report, "bdd-patterns") {
		t.Errorf("report should contain skill names, got:\n%s", report)
	}
}

func TestDuplicationReport_actionAggregate(t *testing.T) {
	report := DuplicationReport(testPairs, testEntries, "2026-04-27")
	if !strings.Contains(report, "Aggregate") {
		t.Errorf("Critical pair should suggest Aggregate, got:\n%s", report)
	}
}

func TestShortKey_withDomain(t *testing.T) {
	if got := shortKey("domain/skill-name"); got != "skill-name" {
		t.Errorf("expected 'skill-name', got %q", got)
	}
}

func TestShortKey_noDomain(t *testing.T) {
	if got := shortKey("skill-name"); got != "skill-name" {
		t.Errorf("expected 'skill-name', got %q", got)
	}
}

func TestFamilyOf_commonPrefix(t *testing.T) {
	result := familyOf("bdd/bdd-gherkin", "bdd/bdd-patterns")
	if !strings.Contains(result, "bdd") {
		t.Errorf("expected bdd family label, got %q", result)
	}
}

func TestFamilyOf_noCommonPrefix(t *testing.T) {
	result := familyOf("bdd/bdd-gherkin", "typescript/ts-basics")
	// should fall back to a slash-separated label
	if result == "" {
		t.Error("expected non-empty family label")
	}
}
