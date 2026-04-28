package scorer

import (
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

// nilBridge is a convenience for unit tests that don't exercise the library path.
func nilBridge() *validatorBridge { return &validatorBridge{} }

func TestD2_WhenToUse(t *testing.T) {
	content := "---\ndescription: x\n---\n## When to Use\nuse it when needed\n\n## When NOT to Use\nnot this time"
	score, _ := scoreD2(content, nilBridge())
	// 0 + 4 (when to use) + 3 (when not to) = 7
	if score != 7 {
		t.Errorf("want 7, got %d", score)
	}
}

func TestD2_MindsetHeading(t *testing.T) {
	content := "---\ndescription: x\n---\n## Mindset\nthink carefully"
	score, _ := scoreD2(content, nilBridge())
	if score != 2 {
		t.Errorf("want 2, got %d", score)
	}
}

func TestD2_PhilosophyHeading(t *testing.T) {
	content := "---\ndescription: x\n---\n## Philosophy\nthink this way"
	score, _ := scoreD2(content, nilBridge())
	if score != 2 {
		t.Errorf("want 2, got %d", score)
	}
}

func TestD2_FallbackNumberedList(t *testing.T) {
	// When bridge has no content, falls back to regex for numbered lists.
	content := "---\ndescription: x\n---\n1. step one\n2. step two\n3. step three"
	score, _ := scoreD2(content, nilBridge())
	if score < 2 {
		t.Errorf("numbered list via fallback should score ≥2, got %d", score)
	}
}

func TestD2_LibraryImperativeRatioMid(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.3, ListItemCount: 2}}
	content := "---\ndescription: x\n---\nsome content"
	score, _ := scoreD2(content, b)
	// 3 (ratio 0.25-0.39) + 1 (listItems 1-3) = 4
	if score < 3 {
		t.Errorf("want ≥3 for mid-range imperative ratio, got %d", score)
	}
}

func TestD2_LibraryImperativeRatioLow(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.15, ListItemCount: 0}}
	content := "---\ndescription: x\n---\nsome content"
	score, _ := scoreD2(content, b)
	// 2 (ratio 0.1-0.24) + 0 (no list items)
	if score < 2 {
		t.Errorf("want ≥2 for low imperative ratio, got %d", score)
	}
}

func TestD2_LibraryImperativeRatio(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.45, ListItemCount: 5}}
	content := "---\ndescription: x\n---\n## When to Use\ndo this"
	score, _ := scoreD2(content, b)
	// 4 (imperative ≥0.4) + 2 (listItems>3) + 4 (when to use) = 10
	if score != 10 {
		t.Errorf("want 10, got %d", score)
	}
}
