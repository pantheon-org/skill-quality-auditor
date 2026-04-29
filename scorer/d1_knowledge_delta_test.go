package scorer

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Demonstration Concreteness sub-criterion tests (direct) ---

func TestScoreDemonstrationConcreteness_Zero(t *testing.T) {
	// No code fence, no →, no output section, no expert-signal patterns
	content := "---\ndescription: x\n---\n# Skill\nValidate inputs before processing."
	if got := scoreDemonstrationConcreteness(content); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestScoreDemonstrationConcreteness_OneExpertSignal(t *testing.T) {
	// Expert-register NEVER but no fence/arrow/output section → 1 pt
	content := "---\ndescription: x\n---\n# Skill\nNEVER skip validation."
	if got := scoreDemonstrationConcreteness(content); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

func TestScoreDemonstrationConcreteness_TwoCodeFence(t *testing.T) {
	// Code fence present, no output section → 2 pts
	content := "---\ndescription: x\n---\n# Skill\n```go\nfoo()\n```"
	if got := scoreDemonstrationConcreteness(content); got != 2 {
		t.Errorf("want 2, got %d", got)
	}
}

func TestScoreDemonstrationConcreteness_TwoArrow(t *testing.T) {
	// → arrow present, no output section → 2 pts
	content := "---\ndescription: x\n---\n# Skill\ninput → output"
	if got := scoreDemonstrationConcreteness(content); got != 2 {
		t.Errorf("want 2, got %d", got)
	}
}

func TestScoreDemonstrationConcreteness_ThreeCodeFenceAndOutput(t *testing.T) {
	// Code fence (signal 1) + "Output:" marker (signal 2) → 3 pts
	content := "---\ndescription: x\n---\n# Skill\n```go\nfoo()\n```\nOutput:\nbar"
	if got := scoreDemonstrationConcreteness(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestScoreDemonstrationConcreteness_ThreeArrowAndResult(t *testing.T) {
	// → arrow (signal 1) + "Result:" marker (signal 2) → 3 pts
	content := "---\ndescription: x\n---\n# Skill\ninput → output\nResult:\nthe final value"
	if got := scoreDemonstrationConcreteness(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestScoreDemonstrationConcreteness_OutputAloneIsZero(t *testing.T) {
	// Output section alone (signal 2) without signal 1, no expert signal → 0 pts
	content := "---\ndescription: x\n---\n# Skill\nDo something.\nOutput:\nresult"
	if got := scoreDemonstrationConcreteness(content); got != 0 {
		t.Errorf("want 0 (output section alone, no signal 1 or expert patterns), got %d", got)
	}
}

// --- Integration tests via scoreD1 ---

func TestD1_ConcretenessZero(t *testing.T) {
	// Plain prose, no signals → base(12) + concreteness(0) = 12
	content := "---\ndescription: x\n---\n# Skill\nValidate inputs before processing."
	score, _ := scoreD1(content, t.TempDir())
	if score != 12 {
		t.Errorf("want 12, got %d", score)
	}
}

func TestD1_ConcretenessOne(t *testing.T) {
	// NEVER → expert-signal bonus(+1) + concreteness(1) = 12 + 1 + 1 = 14
	content := "---\ndescription: x\n---\n# Skill\nNEVER skip validation."
	score, _ := scoreD1(content, t.TempDir())
	if score != 14 {
		t.Errorf("want 14, got %d", score)
	}
}

func TestD1_ConcretenessTwo(t *testing.T) {
	// Code fence only → base(12) + concreteness(2) = 14
	content := "---\ndescription: x\n---\n# Skill\nUse this pattern:\n```go\nfoo()\n```\nThat's it."
	score, _ := scoreD1(content, t.TempDir())
	if score != 14 {
		t.Errorf("want 14, got %d", score)
	}
}

func TestD1_ConcretenessThree(t *testing.T) {
	// Code fence + Output: → base(12) + concreteness(3) = 15
	content := "---\ndescription: x\n---\n# Skill\nUse this pattern:\n```go\nfoo()\n```\nOutput:\nbar"
	score, _ := scoreD1(content, t.TempDir())
	if score != 15 {
		t.Errorf("want 15, got %d", score)
	}
}

// --- Regression tests (base 15→12, expected values updated) ---

func TestD1_Penalties(t *testing.T) {
	content := "---\ndescription: something\n---\n# Skill\ngetting started with npm install"
	score, _ := scoreD1(content, t.TempDir())
	// base=12, -2 for "getting started", -2 for "npm install", concreteness=0 → 8
	if score != 8 {
		t.Errorf("want 8, got %d", score)
	}
}

func TestD1_Rewards(t *testing.T) {
	content := "---\ndescription: something\n---\n# Skill\nNEVER do this. ALWAYS validate. anti-pattern here. production gotcha pitfall."
	score, _ := scoreD1(content, t.TempDir())
	// base=12, +6 expert signals, concreteness=1 (expert signal, no fence/output) = 19
	if score != 19 {
		t.Errorf("want 19, got %d", score)
	}
}

func TestD1_ExpertRatioBonus(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 8 of 10 instructions are "new knowledge" or "preference" → ratio 80% → +2
	instrJSON := `{"instructions":[` +
		`{"why_given":"new knowledge"},{"why_given":"new knowledge"},` +
		`{"why_given":"new knowledge"},{"why_given":"new knowledge"},` +
		`{"why_given":"preference"},{"why_given":"preference"},` +
		`{"why_given":"preference"},{"why_given":"preference"},` +
		`{"why_given":"other"},{"why_given":"other"}` +
		`]}`
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte(instrJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill\nsome content"
	score, diags := scoreD1(content, tmpDir)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
	// base 12 + 2 (expert ratio ≥70%) + concreteness 0 = 14
	if score != 14 {
		t.Errorf("want 14, got %d", score)
	}
}

func TestD1_ExpertRatioPenalty(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 2 of 10 are expert → ratio 20% → -2
	instrJSON := `{"instructions":[` +
		`{"why_given":"new knowledge"},{"why_given":"other"},` +
		`{"why_given":"other"},{"why_given":"other"},{"why_given":"other"},` +
		`{"why_given":"other"},{"why_given":"other"},{"why_given":"other"},` +
		`{"why_given":"other"},{"why_given":"other"}` +
		`]}`
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte(instrJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill\nsome content"
	score, _ := scoreD1(content, tmpDir)
	// base 12 - 2 (ratio < 30%) + concreteness 0 = 10
	if score != 10 {
		t.Errorf("want 10, got %d", score)
	}
}

func TestD1_ExpertRatioNeutral(t *testing.T) {
	// 5 of 10 expert → ratio 50% → neither bonus nor penalty (returns 0)
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	instrJSON := `{"instructions":[` +
		`{"why_given":"new knowledge"},{"why_given":"new knowledge"},` +
		`{"why_given":"preference"},{"why_given":"preference"},` +
		`{"why_given":"preference"},{"why_given":"other"},` +
		`{"why_given":"other"},{"why_given":"other"},` +
		`{"why_given":"other"},{"why_given":"other"}` +
		`]}`
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte(instrJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill\nsome content"
	score, diags := scoreD1(content, tmpDir)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
	// base 12, neutral ratio → no delta, concreteness 0 = 12
	if score != 12 {
		t.Errorf("want 12 (neutral ratio), got %d", score)
	}
}

func TestD1_InstructionsParseError(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill"
	_, diags := scoreD1(content, tmpDir)
	if len(diags) != 1 || diags[0].severity != "error" || diags[0].Dimension != "D1" {
		t.Errorf("expected D1 error diagnostic, got %v", diags)
	}
}
