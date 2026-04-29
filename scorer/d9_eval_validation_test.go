package scorer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempSKILL writes a minimal SKILL.md with the given body to a new temp directory
// and returns the full path to the file.
func writeTempSKILL(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte("# Skill\n\n"+body+"\n"), 0o644); err != nil {
		t.Fatalf("writeTempSKILL: %v", err)
	}
	return path
}

func TestD9_NoEvalsDir(t *testing.T) {
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	score, diags := scoreD9(filepath.Join(t.TempDir(), "nonexistent"), skillPath)
	if score != 0 {
		t.Errorf("want 0, got %d", score)
	}
	if len(diags) != 1 || diags[0].Dimension != "D9" {
		t.Errorf("expected D9 warning, got %v", diags)
	}
}

func TestD9_FullScore(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	writeTestFile(t, filepath.Join(evalsDir, "instructions.json"),
		`{"instructions":[{"type":"a"},{"type":"b"}]}`)
	writeTestFile(t, filepath.Join(evalsDir, "summary.json"),
		`{"instructions_coverage":{"coverage_percentage":85}}`)
	for i := 1; i <= 3; i++ {
		dir := filepath.Join(evalsDir, "scenario-"+string(rune('0'+i)))
		writeTestFile(t, filepath.Join(dir, "task.md"), "# Task")
		writeTestFile(t, filepath.Join(dir, "capability.txt"), "cap")
		writeTestFile(t, filepath.Join(dir, "criteria.json"),
			`{"checklist":[{"description":"x","max_score":60},{"description":"y","max_score":40}]}`)
	}
	score, _ := scoreD9(evalsDir, skillPath)
	// New scoring: evals/ dir (3) + instructions.json (3) + summary ≥80% (5) + ≥3 valid scenarios (2) = 13
	// No mutation coverage (skillPath doesn't exist) → +0
	if score != 13 {
		t.Errorf("want 13, got %d", score)
	}
}

func TestD9_LowCoverageWarning(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	writeTestFile(t, filepath.Join(evalsDir, "summary.json"),
		`{"instructions_coverage":{"coverage_percentage":72}}`)
	_, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected D9 warning for low coverage")
	}
}

func TestD9_InvalidInstructionsJSON(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	writeTestFile(t, filepath.Join(evalsDir, "instructions.json"), "not json")
	_, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "error" {
			found = true
		}
	}
	if !found {
		t.Error("expected D9 error for invalid instructions.json")
	}
}

func TestD9_InvalidSummaryJSON(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	writeTestFile(t, filepath.Join(evalsDir, "summary.json"), "{bad}")
	_, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "error" {
			found = true
		}
	}
	if !found {
		t.Error("expected D9 error for invalid summary.json")
	}
}

func TestD9_ScenarioMissingFiles(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	// scenario dir with only task.md — missing criteria.json and capability.txt
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "task.md"), "# Task")
	_, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected D9 warning for incomplete scenario")
	}
}

func TestD9_CriteriaNotSummingTo100(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "task.md"), "# Task")
	writeTestFile(t, filepath.Join(dir, "capability.txt"), "cap")
	writeTestFile(t, filepath.Join(dir, "criteria.json"),
		`{"checklist":[{"description":"x","max_score":60},{"description":"y","max_score":30}]}`)
	score, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected D9 warning for criteria not summing to 100")
	}
	// 0 valid scenarios → +0 scenario points
	if score > 10 {
		t.Errorf("expected no scenario bonus, got total %d", score)
	}
}

func TestD9_InvalidCriteriaJSON(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "task.md"), "# Task")
	writeTestFile(t, filepath.Join(dir, "capability.txt"), "cap")
	writeTestFile(t, filepath.Join(dir, "criteria.json"), "{bad json}")
	_, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "error" {
			found = true
		}
	}
	if !found {
		t.Error("expected D9 error for invalid criteria.json")
	}
}

func TestD9_FlatScenarioFormatWarning(t *testing.T) {
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	writeTestFile(t, filepath.Join(evalsDir, "scenario-01.md"), "# Scenario 1")
	writeTestFile(t, filepath.Join(evalsDir, "scenario-02.md"), "# Scenario 2")
	score, diags := scoreD9(evalsDir, skillPath)
	found := false
	for _, d := range diags {
		if d.Dimension == "D9" && d.severity == "warning" && strings.Contains(d.Message, "flat scenario") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected D9 flat-format warning, got diags: %v", diags)
	}
	// evals/ exists → +3 (new scoring: evals dir present, no instructions.json = 3 pts base)
	if score != 3 {
		t.Errorf("want score 3, got %d", score)
	}
}

func TestD9_CountValidScenariosWrapper(t *testing.T) {
	evalsDir := t.TempDir()
	for i := 1; i <= 2; i++ {
		dir := filepath.Join(evalsDir, "scenario-"+string(rune('0'+i)))
		writeTestFile(t, filepath.Join(dir, "task.md"), "# Task")
		writeTestFile(t, filepath.Join(dir, "capability.txt"), "cap")
		writeTestFile(t, filepath.Join(dir, "criteria.json"),
			`{"checklist":[{"max_score":100}]}`)
	}
	if n := countValidScenarios(evalsDir); n != 2 {
		t.Errorf("want 2, got %d", n)
	}
}

func TestParseCoveragePercentage(t *testing.T) {
	cases := []struct {
		input any
		want  int
	}{
		{float64(85), 85},
		{int(72), 72},
		{"90%", 90},
		{"75.5%", 75},
		{"", -1},
		{nil, -1},
		{"abc", -1},
		{true, -1},
	}
	for _, tc := range cases {
		got := parseCoveragePercentage(tc.input)
		if got != tc.want {
			t.Errorf("parseCoveragePercentage(%v): want %d, got %d", tc.input, tc.want, got)
		}
	}
}

func TestScoreD9Instructions_EmptyArray(t *testing.T) {
	// Valid JSON but empty instructions array — data is still non-empty so returns 3.
	evalsDir := t.TempDir()
	writeTestFile(t, filepath.Join(evalsDir, "instructions.json"), `{"instructions":[]}`)
	delta, diags := scoreD9Instructions(evalsDir)
	if len(diags) != 0 {
		t.Errorf("expected no diags, got %v", diags)
	}
	if delta != 3 {
		t.Errorf("want 3 (non-empty file), got %d", delta)
	}
}

func TestScoreD9Summary_NilCoverageNonEmptyFile(t *testing.T) {
	// summary.json exists, valid JSON, but coverage_percentage is absent → returns 3.
	evalsDir := t.TempDir()
	writeTestFile(t, filepath.Join(evalsDir, "summary.json"), `{"instructions_coverage":{}}`)
	delta, diags := scoreD9Summary(evalsDir)
	if len(diags) != 0 {
		t.Errorf("expected no diags, got %v", diags)
	}
	if delta != 3 {
		t.Errorf("want 3 for non-empty file with nil coverage, got %d", delta)
	}
}

// --- scoreMutationCoverage tests ---

func TestScoreMutationCoverage_FullCoverage(t *testing.T) {
	// SKILL.md has two MUST constraints; both are referenced in criteria descriptions.
	// Expected: ≥ 80% coverage → 5 pts.
	skillPath := writeTempSKILL(t, "MUST validate input before processing.\nMUST log all errors.")
	evalsDir := t.TempDir()
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "criteria.json"),
		`{"checklist":[{"description":"validates input before processing","max_score":50},{"description":"logs all errors","max_score":50}]}`)
	score, diags := scoreMutationCoverage(skillPath, evalsDir)
	if score != 5 {
		t.Errorf("want 5, got %d (diags: %v)", score, diags)
	}
}

func TestScoreMutationCoverage_PartialCoverage(t *testing.T) {
	// SKILL.md has two MUST constraints; only one is referenced in criteria.
	// Expected: 50% coverage → 3 pts.
	skillPath := writeTempSKILL(t, "MUST validate input.\nMUST purge stale tokens.")
	evalsDir := t.TempDir()
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "criteria.json"),
		`{"checklist":[{"description":"validates input","max_score":100}]}`)
	score, diags := scoreMutationCoverage(skillPath, evalsDir)
	if score != 3 {
		t.Errorf("want 3, got %d (diags: %v)", score, diags)
	}
}

func TestScoreMutationCoverage_ZeroCoverage(t *testing.T) {
	// SKILL.md has MUST constraints; no criteria item matches any of them.
	// Expected: 0 pts.
	skillPath := writeTempSKILL(t, "MUST validate input.\nNEVER expose secrets.")
	evalsDir := t.TempDir()
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "criteria.json"),
		`{"checklist":[{"description":"output is formatted correctly","max_score":100}]}`)
	score, diags := scoreMutationCoverage(skillPath, evalsDir)
	if score != 0 {
		t.Errorf("want 0, got %d (diags: %v)", score, diags)
	}
}

func TestScoreMutationCoverage_CriterionFieldFallback(t *testing.T) {
	// criteria.json uses "criterion" (not "description") — must still count for mutation coverage.
	skillPath := writeTempSKILL(t, "NEVER skip the baseline comparison.")
	evalsDir := t.TempDir()
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "criteria.json"),
		`{"checklist":[{"criterion":"always compare against baseline","max_score":100}]}`)
	score, _ := scoreMutationCoverage(skillPath, evalsDir)
	if score < 1 {
		t.Errorf("want ≥1 pt when criterion field covers a NEVER statement, got %d", score)
	}
}

func TestScoreMutationCoverage_MissingSKILLMd(t *testing.T) {
	// SKILL.md does not exist — should return 0 pts, no error diagnostic.
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	evalsDir := t.TempDir()
	score, diags := scoreMutationCoverage(skillPath, evalsDir)
	if score != 0 {
		t.Errorf("want 0, got %d", score)
	}
	for _, d := range diags {
		if d.severity == "error" {
			t.Errorf("unexpected error diag: %v", d)
		}
	}
}

func TestScoreMutationCoverage_NoInstructions(t *testing.T) {
	// SKILL.md exists but has no MUST/NEVER/ALWAYS lines.
	// Expected: 0 pts (no constraints to cover).
	skillPath := writeTempSKILL(t, "Use this skill to do things.")
	evalsDir := t.TempDir()
	score, _ := scoreMutationCoverage(skillPath, evalsDir)
	if score != 0 {
		t.Errorf("want 0, got %d", score)
	}
}

// --- scoreIndependentAuthoring fallback tier tests ---

func TestScoreIndependentAuthoring_NonGitRepo(t *testing.T) {
	// Outside a git repo: git log will fail → return 0 pts, no error.
	evalsDir := t.TempDir()
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	writeTestFile(t, skillPath, "# Skill\n")
	writeTestFile(t, filepath.Join(evalsDir, "scenario-1", "task.md"), "# Task")
	score, diags := scoreIndependentAuthoring(evalsDir, skillPath)
	if score != 0 {
		t.Errorf("want 0 for non-git repo, got %d", score)
	}
	for _, d := range diags {
		if d.severity == "error" {
			t.Errorf("unexpected error diag in non-git repo: %v", d)
		}
	}
}

func TestScoreIndependentAuthoring_MtimeSameWindow(t *testing.T) {
	// Both files have the same mtime (within 1 hour) → hint diagnostic emitted, 0 pts returned.
	evalsDir := t.TempDir()
	skillDir := t.TempDir()
	skillPath := filepath.Join(skillDir, "SKILL.md")
	writeTestFile(t, skillPath, "# Skill\n")
	writeTestFile(t, filepath.Join(evalsDir, "scenario-1", "task.md"), "# Task")
	// Both created just now → same window → 0 pts to score (diagnostic only).
	score, _ := scoreIndependentAuthoring(evalsDir, skillPath)
	// Must return 0 (bonus is diagnostic-only, not added to score).
	if score != 0 {
		t.Errorf("want 0 (bonus is diagnostic-only), got %d", score)
	}
}

func TestScoreIndependentAuthoring_MissingFiles(t *testing.T) {
	// Both skillPath and evalsDir point to non-existent files → stat error → 0 pts, no error diag.
	evalsDir := filepath.Join(t.TempDir(), "nonexistent")
	skillPath := filepath.Join(t.TempDir(), "SKILL.md")
	score, diags := scoreIndependentAuthoring(evalsDir, skillPath)
	if score != 0 {
		t.Errorf("want 0, got %d", score)
	}
	for _, d := range diags {
		if d.severity == "error" {
			t.Errorf("unexpected error diag: %v", d)
		}
	}
}

// --- scoreAdversarialScenario tests ---

func TestScoreAdversarialScenario_DiagnosticOnly(t *testing.T) {
	// Adversarial bonus must never add to score (diagnostic-only).
	evalsDir := t.TempDir()
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "task.md"),
		"# Task\nHandle the error case when input is invalid.")
	score, diags := scoreAdversarialScenario(evalsDir)
	if score != 0 {
		t.Errorf("adversarial score must be 0 (diagnostic-only), got %d", score)
	}
	// Should emit at least one hint diagnostic when adversarial content found.
	found := false
	for _, d := range diags {
		if d.severity == "hint" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected hint diagnostic for adversarial scenario, got %v", diags)
	}
}

func TestScoreAdversarialScenario_HappyPathOnly(t *testing.T) {
	// All happy-path tasks → no adversarial hint emitted (or hint with 0 bonus).
	evalsDir := t.TempDir()
	dir := filepath.Join(evalsDir, "scenario-1")
	writeTestFile(t, filepath.Join(dir, "task.md"), "# Task\nGenerate a summary of the document.")
	score, _ := scoreAdversarialScenario(evalsDir)
	if score != 0 {
		t.Errorf("want 0, got %d", score)
	}
}

// writeTestFile creates parent directories and writes content to path.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
