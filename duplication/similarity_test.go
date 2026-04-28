package duplication

import (
	"testing"
)

func TestJaccard_identical(t *testing.T) {
	a := TokenSet("alpha beta gamma delta")
	b := TokenSet("alpha beta gamma delta")
	if got := Jaccard(a, b); got != 1.0 {
		t.Fatalf("expected 1.0, got %f", got)
	}
}

func TestJaccard_disjoint(t *testing.T) {
	a := TokenSet("alpha beta gamma")
	b := TokenSet("delta epsilon zeta")
	if got := Jaccard(a, b); got != 0.0 {
		t.Fatalf("expected 0.0, got %f", got)
	}
}

func TestJaccard_partial(t *testing.T) {
	a := TokenSet("alpha beta gamma delta")
	b := TokenSet("alpha beta epsilon zeta")
	got := Jaccard(a, b)
	// intersection=2, union=6 → 0.333...
	if got < 0.33 || got > 0.34 {
		t.Fatalf("expected ~0.333, got %f", got)
	}
}

func TestTokenSet_stripsMarkdown(t *testing.T) {
	text := "## Introduction\n- **alpha** beta\n```go\ngamma\n```"
	tokens := TokenSet(text)
	for _, bad := range []string{"##", "**", "```", "-"} {
		if tokens[bad] {
			t.Errorf("stopword/markdown token %q should be stripped", bad)
		}
	}
	if !tokens["alpha"] || !tokens["beta"] || !tokens["gamma"] {
		t.Error("expected content tokens to be present")
	}
}

func TestSectionHeaders(t *testing.T) {
	text := "# Title\n## Section One\n### Subsection\n## Section Two"
	headers := SectionHeaders(text)
	if len(headers) != 4 {
		t.Fatalf("expected 4 headers, got %d", len(headers))
	}
}

func TestSimilarity_identical(t *testing.T) {
	text := "## Anti-patterns\nNEVER use global state in production\nalpha beta gamma"
	if got := Similarity(text, text); got < 0.99 {
		t.Fatalf("identical texts should score near 1.0, got %f", got)
	}
}

func TestSimilarity_unrelated(t *testing.T) {
	a := "## Cooking\nRecipes flour butter sugar oven bake"
	b := "## Networking\nTCP UDP packets latency throughput routing"
	if got := Similarity(a, b); got > 0.15 {
		t.Fatalf("unrelated texts should score low, got %f", got)
	}
}

func TestJaccard_bothEmpty(t *testing.T) {
	// Both empty sets → returns 1.0 per the implementation
	if got := Jaccard(map[string]bool{}, map[string]bool{}); got != 1.0 {
		t.Fatalf("both empty sets should return 1.0, got %f", got)
	}
}

func TestJaccard_oneEmpty(t *testing.T) {
	a := TokenSet("alpha beta gamma")
	if got := Jaccard(a, map[string]bool{}); got != 0.0 {
		t.Fatalf("one empty set should return 0.0, got %f", got)
	}
}

func TestDetect_highButNotCritical(t *testing.T) {
	// Two skills with ~25% similarity (above High threshold, below Critical)
	aContent := "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda"
	bContent := "alpha beta gamma delta zeta eta theta iota kappa lambda unique-word-1 unique-word-2 unique-word-3 unique-word-4"
	entries := []SkillEntry{
		{Key: "a/skill-a", Content: aContent},
		{Key: "a/skill-b", Content: bContent},
	}
	pairs := Detect(entries)
	// If similarity is above ThresholdHigh, check severity assignment
	for _, p := range pairs {
		if p.Similarity >= ThresholdCritical && p.Severity != "Critical" {
			t.Errorf("expected Critical severity for sim=%.2f, got %s", p.Similarity, p.Severity)
		}
		if p.Similarity < ThresholdCritical && p.Severity != "High" {
			t.Errorf("expected High severity for sim=%.2f, got %s", p.Similarity, p.Severity)
		}
	}
}

func TestDetect_sortedDescending(t *testing.T) {
	// Three entries with different similarities; result should be sorted descending
	hi := "alpha beta gamma delta epsilon zeta eta theta"
	mid := "alpha beta gamma zeta eta theta unique-x unique-y unique-z"
	lo := "alpha beta completely-different-words here nothing shared"
	entries := []SkillEntry{
		{Key: "a/low", Content: lo},
		{Key: "a/high", Content: hi},
		{Key: "a/mid", Content: mid},
	}
	pairs := Detect(entries)
	for i := 1; i < len(pairs); i++ {
		if pairs[i].Similarity > pairs[i-1].Similarity {
			t.Errorf("pairs not sorted descending: [%d]=%.2f > [%d]=%.2f", i, pairs[i].Similarity, i-1, pairs[i-1].Similarity)
		}
	}
}

func TestDetect_returnsHighPairs(t *testing.T) {
	content := "## Anti-patterns\nNEVER use global state production gotcha pitfall\nalpha beta gamma delta epsilon"
	entries := []SkillEntry{
		{Key: "a/foo", Content: content},
		{Key: "a/bar", Content: content},
		{Key: "a/baz", Content: "completely different content about unrelated topic xyz"},
	}
	pairs := Detect(entries)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Severity != "Critical" {
		t.Errorf("identical content should be Critical, got %s", pairs[0].Severity)
	}
}
