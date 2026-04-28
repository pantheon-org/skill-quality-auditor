package reporter

import (
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/duplication"
)

var aggEntries = []duplication.SkillEntry{
	{Key: "bdd/bdd-gherkin", Content: strings.Repeat("line\n", 300)},
	{Key: "bdd/bdd-patterns", Content: strings.Repeat("line\n", 250)},
	{Key: "bdd/bdd-principles", Content: strings.Repeat("line\n", 200)},
}

var aggPairs = []duplication.Pair{
	{A: "bdd/bdd-gherkin", B: "bdd/bdd-patterns", Similarity: 0.40, Severity: "Critical"},
	{A: "bdd/bdd-gherkin", B: "bdd/bdd-principles", Similarity: 0.38, Severity: "Critical"},
	{A: "bdd/bdd-patterns", B: "bdd/bdd-principles", Similarity: 0.22, Severity: "High"},
}

func TestAggregationPlan_structure(t *testing.T) {
	plan := AggregationPlan("bdd", aggEntries, aggPairs, "2026-04-27")

	for _, want := range []string{
		"# Aggregation Plan",
		"bdd",
		"2026-04-27",
		"## Family Analysis",
		"## Decision",
	} {
		if !strings.Contains(plan, want) {
			t.Errorf("plan missing %q", want)
		}
	}
}

func TestAggregationPlan_aggregateDecision(t *testing.T) {
	plan := AggregationPlan("bdd", aggEntries, aggPairs, "2026-04-27")
	// high duplication + 3 skills → should recommend AGGREGATE
	if !strings.Contains(plan, "AGGREGATE") {
		t.Errorf("expected AGGREGATE decision, got:\n%s", plan)
	}
	if !strings.Contains(plan, "6-Step Aggregation Process") {
		t.Errorf("expected 6-step process section, got:\n%s", plan)
	}
}

func TestAggregationPlan_monitorDecision(t *testing.T) {
	single := []duplication.SkillEntry{{Key: "t/skill-a", Content: "unique content about alpha"}}
	plan := AggregationPlan("skill", single, nil, "2026-04-27")
	if !strings.Contains(plan, "MONITOR") {
		t.Errorf("single skill with no duplication should MONITOR, got:\n%s", plan)
	}
}

func TestAggregationPlan_listsSkills(t *testing.T) {
	plan := AggregationPlan("bdd", aggEntries, aggPairs, "2026-04-27")
	for _, e := range aggEntries {
		name := shortKey(e.Key)
		if !strings.Contains(plan, name) {
			t.Errorf("plan missing skill %q", name)
		}
	}
}

func TestAggregationPlan_effort(t *testing.T) {
	plan := AggregationPlan("bdd", aggEntries, aggPairs, "2026-04-27")
	if !strings.Contains(plan, "## Effort Estimate") {
		t.Errorf("plan missing effort estimate section")
	}
}

func TestAggregationDecision_thresholds(t *testing.T) {
	cases := []struct {
		count   int
		lines   int
		avgSim  float64
		wantKey string
	}{
		{6, 2500, 0.40, "AGGREGATE"},
		{2, 500, 0.10, "MONITOR"},
		{4, 1500, 0.25, "AGGREGATE"}, // count ≥3 + sim ≥ ThresholdHigh
	}
	for _, c := range cases {
		decision, _ := aggregationDecision(c.count, c.lines, c.avgSim)
		if !strings.Contains(decision, c.wantKey) {
			t.Errorf("count=%d lines=%d sim=%.2f: expected %s, got %s",
				c.count, c.lines, c.avgSim, c.wantKey, decision)
		}
	}
}

func TestAggregationEffort_sizes(t *testing.T) {
	if aggregationEffort(2, 500) != "S" {
		t.Error("small family should be S")
	}
	if aggregationEffort(5, 2000) != "M" {
		t.Error("medium family should be M")
	}
	if aggregationEffort(10, 5000) != "L" {
		t.Error("large family should be L")
	}
}

func TestAggregationEffort_boundaries(t *testing.T) {
	// Boundary: count=3 and lines=1000 is still S
	if aggregationEffort(3, 1000) != "S" {
		t.Error("count=3,lines=1000 should be S")
	}
	// Boundary: count=4 tips into M
	if aggregationEffort(4, 1000) != "M" {
		t.Error("count=4,lines=1000 should be M")
	}
	// Boundary: lines=1001 tips into M
	if aggregationEffort(3, 1001) != "M" {
		t.Error("count=3,lines=1001 should be M")
	}
	// Boundary: count=6 and lines=3000 is still M
	if aggregationEffort(6, 3000) != "M" {
		t.Error("count=6,lines=3000 should be M")
	}
	// One over tips to L
	if aggregationEffort(7, 3000) != "L" {
		t.Error("count=7,lines=3000 should be L")
	}
}

func TestAggregationTime_smallCount(t *testing.T) {
	// count=1 → ceil(0.5) = 1 → "1 hour"
	if aggregationTime(1) != "1 hour" {
		t.Errorf("count=1 should be '1 hour', got %q", aggregationTime(1))
	}
	// count=2 → ceil(1.0) = 1 → "1 hour"
	if aggregationTime(2) != "1 hour" {
		t.Errorf("count=2 should be '1 hour', got %q", aggregationTime(2))
	}
	// count=3 → ceil(1.5) = 2 → "2 hours"
	if aggregationTime(3) != "2 hours" {
		t.Errorf("count=3 should be '2 hours', got %q", aggregationTime(3))
	}
}

func TestAggregationDecision_considerBranch(t *testing.T) {
	// count<3, no high sim, but lines>2000 → single reason → CONSIDER
	decision, rationale := aggregationDecision(2, 2500, 0.10)
	if decision != "CONSIDER" {
		t.Errorf("expected CONSIDER, got %q (rationale: %s)", decision, rationale)
	}
}

func TestAggregationDecision_familySizeOnly(t *testing.T) {
	// count>=3, low sim, lines<=2000 → single reason from family size → CONSIDER
	decision, _ := aggregationDecision(3, 500, 0.05)
	if decision != "CONSIDER" {
		t.Errorf("expected CONSIDER for count=3 low-sim, got %q", decision)
	}
}

func TestAggregationDecision_criticalSim(t *testing.T) {
	// avgSim >= ThresholdCritical alone → reasons has 1 item, but it meets
	// the critical-sim shortcut in the AGGREGATE check.
	decision, _ := aggregationDecision(2, 500, 0.40)
	if decision != "AGGREGATE ✅" {
		t.Errorf("expected AGGREGATE for critical sim, got %q", decision)
	}
}

func TestAggregationPlan_monitorNoSteps(t *testing.T) {
	// MONITOR decision should not include the 6-step section.
	single := []duplication.SkillEntry{{Key: "t/skill-a", Content: "unique content"}}
	plan := AggregationPlan("skill", single, nil, "2026-04-27")
	if strings.Contains(plan, "6-Step Aggregation Process") {
		t.Error("MONITOR plan should not contain 6-step process")
	}
}
