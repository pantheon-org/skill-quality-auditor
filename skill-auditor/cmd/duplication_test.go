package cmd

import (
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/duplication"
)

func TestExitCodeForPairs_empty(t *testing.T) {
	if err := exitCodeForPairs(nil); err != nil {
		t.Errorf("expected nil for empty pairs, got %v", err)
	}
	if err := exitCodeForPairs([]duplication.Pair{}); err != nil {
		t.Errorf("expected nil for empty pairs slice, got %v", err)
	}
}

func TestExitCodeForPairs_noCritical(t *testing.T) {
	pairs := []duplication.Pair{
		{A: "a/skill-a", B: "b/skill-b", Similarity: 0.22, Severity: "High"},
		{A: "a/skill-a", B: "c/skill-c", Similarity: 0.15, Severity: "High"},
	}
	if err := exitCodeForPairs(pairs); err != nil {
		t.Errorf("expected nil when no Critical pairs, got %v", err)
	}
}

func TestExitCodeForPairs_withCritical(t *testing.T) {
	pairs := []duplication.Pair{
		{A: "a/skill-a", B: "b/skill-b", Similarity: 0.40, Severity: "Critical"},
	}
	if err := exitCodeForPairs(pairs); err == nil {
		t.Error("expected error for Critical pair")
	}
}

func TestExitCodeForPairs_criticalAmongMany(t *testing.T) {
	pairs := []duplication.Pair{
		{A: "a/skill-a", B: "b/skill-b", Similarity: 0.22, Severity: "High"},
		{A: "a/skill-a", B: "c/skill-c", Similarity: 0.41, Severity: "Critical"},
		{A: "b/skill-b", B: "c/skill-c", Similarity: 0.18, Severity: "High"},
	}
	if err := exitCodeForPairs(pairs); err == nil {
		t.Error("expected error when at least one Critical pair exists")
	}
}
