package analysis

import (
	"testing"
)

// --- DetectRequiredSections ---

func TestDetectRequiredSections_AllPresent(t *testing.T) {
	content := "# Overview\n\n## Usage\n\nSome text.\n\n## Examples\n"
	required := []string{"overview", "usage", "examples"}
	results := DetectRequiredSections(content, required)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Matched {
			t.Errorf("expected section %q to be matched", r.Rule)
		}
		if r.Score != 1.0 {
			t.Errorf("expected score 1.0 for matched section, got %f", r.Score)
		}
		if len(r.Evidence) == 0 {
			t.Errorf("expected evidence for matched section %q", r.Rule)
		}
	}
}

func TestDetectRequiredSections_SomeMissing(t *testing.T) {
	content := "# Overview\n\nSome text.\n"
	required := []string{"overview", "usage", "examples"}
	results := DetectRequiredSections(content, required)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	matched := 0
	for _, r := range results {
		if r.Matched {
			matched++
		}
	}
	if matched != 1 {
		t.Errorf("expected exactly 1 match, got %d", matched)
	}
}

func TestDetectRequiredSections_EmptyContent(t *testing.T) {
	results := DetectRequiredSections("", []string{"overview", "usage"})
	for _, r := range results {
		if r.Matched {
			t.Errorf("expected no matches for empty content, got match for %q", r.Rule)
		}
		if r.Score != 0.0 {
			t.Errorf("expected score 0.0 for empty content, got %f", r.Score)
		}
	}
}

func TestDetectRequiredSections_EmptyRequired(t *testing.T) {
	results := DetectRequiredSections("# Foo\n", []string{})
	if len(results) != 0 {
		t.Errorf("expected empty results for empty required list, got %d", len(results))
	}
}

func TestDetectRequiredSections_CaseInsensitive(t *testing.T) {
	content := "# Overview\n"
	results := DetectRequiredSections(content, []string{"Overview"})
	if len(results) != 1 || !results[0].Matched {
		t.Error("expected case-insensitive match for 'Overview'")
	}
}

// --- DetectTriggerFrequency ---

func TestDetectTriggerFrequency_AboveThreshold(t *testing.T) {
	content := "use the tool. use the tool again. use the tool once more."
	triggers := map[string]int{"use the tool": 2}
	results := DetectTriggerFrequency(content, triggers)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Matched {
		t.Error("expected match when count > minCount")
	}
	if results[0].Score != 1.0 {
		t.Errorf("expected score capped at 1.0, got %f", results[0].Score)
	}
}

func TestDetectTriggerFrequency_BelowThreshold(t *testing.T) {
	content := "use the tool once."
	triggers := map[string]int{"use the tool": 3}
	results := DetectTriggerFrequency(content, triggers)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Matched {
		t.Error("expected no match when count < minCount")
	}
	expected := 1.0 / 3.0
	if results[0].Score < expected-0.001 || results[0].Score > expected+0.001 {
		t.Errorf("expected score ~%.4f, got %f", expected, results[0].Score)
	}
}

func TestDetectTriggerFrequency_ExactThreshold(t *testing.T) {
	content := "foo bar foo bar"
	triggers := map[string]int{"foo": 2}
	results := DetectTriggerFrequency(content, triggers)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Matched {
		t.Error("expected match when count == minCount")
	}
	if results[0].Score != 1.0 {
		t.Errorf("expected score 1.0 at exact threshold, got %f", results[0].Score)
	}
}

func TestDetectTriggerFrequency_EmptyContent(t *testing.T) {
	triggers := map[string]int{"foo": 1}
	results := DetectTriggerFrequency("", triggers)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Matched {
		t.Error("expected no match for empty content")
	}
	if results[0].Score != 0.0 {
		t.Errorf("expected score 0.0 for empty content, got %f", results[0].Score)
	}
}

func TestDetectTriggerFrequency_CaseInsensitive(t *testing.T) {
	content := "FOO foo Foo"
	triggers := map[string]int{"foo": 3}
	results := DetectTriggerFrequency(content, triggers)
	if len(results) != 1 || !results[0].Matched {
		t.Error("expected case-insensitive count to reach threshold")
	}
}

func TestDetectTriggerFrequency_MultipleTriggers(t *testing.T) {
	content := "alpha beta alpha gamma"
	triggers := map[string]int{"alpha": 2, "beta": 1, "delta": 1}
	results := DetectTriggerFrequency(content, triggers)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	matched := 0
	for _, r := range results {
		if r.Matched {
			matched++
		}
	}
	if matched != 2 {
		t.Errorf("expected 2 matches (alpha, beta), got %d", matched)
	}
}

// --- DetectStructuralConformance ---

func TestDetectStructuralConformance_PerfectMatch(t *testing.T) {
	content := "# Overview\n## Usage\n## Examples\n"
	canonical := []string{"overview", "usage", "examples"}
	result := DetectStructuralConformance(content, canonical)
	if !result.Matched {
		t.Error("expected Matched=true for perfect match")
	}
	if result.Score != 1.0 {
		t.Errorf("expected Score=1.0 for perfect match, got %f", result.Score)
	}
}

func TestDetectStructuralConformance_PartialMatch(t *testing.T) {
	content := "# Overview\n## Usage\n"
	canonical := []string{"overview", "usage", "examples", "configuration"}
	result := DetectStructuralConformance(content, canonical)
	// actual={overview,usage}, canonical={overview,usage,examples,configuration}
	// intersection=2, union=4, jaccard=0.5
	if !result.Matched {
		t.Errorf("expected Matched=true for score=0.5, got false (score=%f)", result.Score)
	}
	if result.Score < 0.49 || result.Score > 0.51 {
		t.Errorf("expected score ~0.5, got %f", result.Score)
	}
}

func TestDetectStructuralConformance_NoOverlap(t *testing.T) {
	content := "# Foo\n## Bar\n"
	canonical := []string{"overview", "usage", "examples"}
	result := DetectStructuralConformance(content, canonical)
	if result.Matched {
		t.Error("expected Matched=false for no overlap")
	}
	if result.Score != 0.0 {
		t.Errorf("expected score 0.0 for no overlap, got %f", result.Score)
	}
}

func TestDetectStructuralConformance_EmptyContent(t *testing.T) {
	canonical := []string{"overview", "usage"}
	result := DetectStructuralConformance("", canonical)
	if result.Matched {
		t.Error("expected Matched=false for empty content")
	}
	if result.Score != 0.0 {
		t.Errorf("expected score 0.0 for empty content, got %f", result.Score)
	}
}

func TestDetectStructuralConformance_EmptyCanonical(t *testing.T) {
	content := "# Foo\n"
	result := DetectStructuralConformance(content, []string{})
	// actual={foo}, canonical={}, intersection=0, union=1, score=0
	if result.Matched {
		t.Error("expected Matched=false when canonical is empty")
	}
}

func TestDetectStructuralConformance_BothEmpty(t *testing.T) {
	result := DetectStructuralConformance("", []string{})
	if result.Score != 0.0 {
		t.Errorf("expected score 0.0 when both empty, got %f", result.Score)
	}
}

// --- DetectAntiPatternSignals ---

func TestDetectAntiPatternSignals_HedgeWords(t *testing.T) {
	content := "Maybe you should try this. Perhaps it will work.\n"
	results := DetectAntiPatternSignals(content)
	hedge := findRule(results, "anti-pattern:hedge-language")
	if hedge == nil {
		t.Fatal("expected anti-pattern:hedge-language rule in results")
	}
	if !hedge.Matched {
		t.Errorf("expected hedge rule to fire (score=%f, evidence=%v)", hedge.Score, hedge.Evidence)
	}
}

func TestDetectAntiPatternSignals_VagueWords(t *testing.T) {
	content := "Handle appropriately and as needed for each case.\n"
	results := DetectAntiPatternSignals(content)
	vague := findRule(results, "anti-pattern:vague-instructions")
	if vague == nil {
		t.Fatal("expected anti-pattern:vague-instructions rule in results")
	}
	if !vague.Matched {
		t.Errorf("expected vague rule to fire (score=%f, evidence=%v)", vague.Score, vague.Evidence)
	}
}

func TestDetectAntiPatternSignals_PassiveVoice(t *testing.T) {
	content := "The task is done. The file was created. The function can be used.\n"
	results := DetectAntiPatternSignals(content)
	passive := findRule(results, "anti-pattern:passive-voice")
	if passive == nil {
		t.Fatal("expected anti-pattern:passive-voice rule in results")
	}
	if !passive.Matched {
		t.Errorf("expected passive rule to fire (score=%f, evidence=%v)", passive.Score, passive.Evidence)
	}
}

func TestDetectAntiPatternSignals_CleanContent(t *testing.T) {
	content := "# Good Skill\n\nAlways do X. Never do Y. Use Z for all operations.\n"
	results := DetectAntiPatternSignals(content)
	for _, r := range results {
		if r.Matched {
			t.Errorf("expected no anti-pattern matches for clean content, but %q matched", r.Rule)
		}
	}
}

func TestDetectAntiPatternSignals_CodeBlocksExcluded(t *testing.T) {
	// hedge words inside a code block should NOT be counted
	content := "Clean instructions here.\n\n```\nmaybe = true\nperhaps = false\n```\n"
	results := DetectAntiPatternSignals(content)
	hedge := findRule(results, "anti-pattern:hedge-language")
	if hedge == nil {
		t.Fatal("expected anti-pattern:hedge-language rule in results")
	}
	if hedge.Matched {
		t.Errorf("expected hedge rule NOT to fire when words are inside code blocks (evidence=%v)", hedge.Evidence)
	}
}

func TestDetectAntiPatternSignals_ReturnsThreeRules(t *testing.T) {
	results := DetectAntiPatternSignals("some content")
	if len(results) != 3 {
		t.Errorf("expected 3 anti-pattern rules, got %d", len(results))
	}
}

func TestDetectAntiPatternSignals_ScoreCapsAtThreshold(t *testing.T) {
	// many hedge words — score should be >= 1.0 (uncapped for anti-patterns, just threshold-based)
	content := "Maybe perhaps feel free you might possibly maybe perhaps feel free you might.\n"
	results := DetectAntiPatternSignals(content)
	hedge := findRule(results, "anti-pattern:hedge-language")
	if hedge == nil {
		t.Fatal("expected anti-pattern:hedge-language rule")
	}
	if !hedge.Matched {
		t.Error("expected hedge to match with many occurrences")
	}
	if hedge.Score < 1.0 {
		t.Errorf("expected score >= 1.0 for many hedge words, got %f", hedge.Score)
	}
}

func findRule(results []RuleMatch, rule string) *RuleMatch {
	for i := range results {
		if results[i].Rule == rule {
			return &results[i]
		}
	}
	return nil
}
