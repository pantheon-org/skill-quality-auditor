package scorer

import (
	"path/filepath"
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

// makeD7Bridge builds a bridge with the given description-length message and rawContent.
func makeD7Bridge(msg, raw string) *validatorBridge {
	return &validatorBridge{
		Structure: &types.Report{
			Results: []types.Result{
				{Category: "Frontmatter", Level: types.Pass, Message: msg},
			},
		},
		rawContent: raw,
	}
}

// longRawContent returns a minimal SKILL.md frontmatter with a description of n x-characters.
func longRawContent(n int) string {
	desc := make([]byte, n)
	for i := range desc {
		desc[i] = 'x'
	}
	return "---\ndescription: " + string(desc) + "\n---\n"
}

// hasDiagSeverity reports whether any diagnostic in diags has the given severity.
func hasDiagSeverity(diags []Diagnostic, severity string) bool {
	for _, d := range diags {
		if d.Severity() == severity {
			return true
		}
	}
	return false
}

// hasDiagD7Severity reports whether any D7 diagnostic has the given severity.
func hasDiagD7Severity(diags []Diagnostic, severity string) bool {
	for _, d := range diags {
		if d.Dimension == "D7" && d.Severity() == severity {
			return true
		}
	}
	return false
}

func TestD7_VeryLongDescription(t *testing.T) {
	// descLen > 200 → score 10; no anchors → WARN discriminativeness diagnostic
	b := makeD7Bridge("description: (210 chars)", longRawContent(210))
	score, diags := scoreD7(b)
	if score != 10 {
		t.Errorf("want 10, got %d", score)
	}
	if !hasDiagSeverity(diags, "warning") {
		t.Errorf("expected WARN discriminativeness diagnostic for plain long description, got %v", diags)
	}
}

func TestD7_MediumDescription(t *testing.T) {
	// 120 < descLen <= 200 → score 9; no anchors → WARN discriminativeness diagnostic
	b := makeD7Bridge("description: (150 chars)", longRawContent(150))
	score, diags := scoreD7(b)
	if score != 9 {
		t.Errorf("want 9, got %d", score)
	}
	if !hasDiagSeverity(diags, "warning") {
		t.Errorf("expected WARN discriminativeness diagnostic for medium description, got %v", diags)
	}
}

func TestD7_ShortDescription(t *testing.T) {
	// 60 < descLen <= 120 → score 8; no anchors → WARN discriminativeness diagnostic
	b := makeD7Bridge("description: (80 chars)", longRawContent(80))
	score, diags := scoreD7(b)
	if score != 8 {
		t.Errorf("want 8, got %d", score)
	}
	if !hasDiagSeverity(diags, "warning") {
		t.Errorf("expected WARN discriminativeness diagnostic for short description, got %v", diags)
	}
}

func TestD7_VeryShortDescription(t *testing.T) {
	// descLen <= 30 → score 6 + length warning; no anchors → discriminativeness WARN too
	b := makeD7Bridge("description: (20 chars)", longRawContent(20))
	score, diags := scoreD7(b)
	if score != 6 {
		t.Errorf("want 6, got %d", score)
	}
	if len(diags) == 0 {
		t.Error("expected at least one D7 warning for very short description")
	}
}

func TestD7_FallbackNoBridge(t *testing.T) {
	// nilBridge() has no Structure → fallback 6 + warning
	score, diags := scoreD7(nilBridge())
	if score != 6 {
		t.Errorf("want 6, got %d", score)
	}
	if len(diags) == 0 {
		t.Error("expected D7 warning when bridge unavailable")
	}
}

func TestD7_DescLen31to60(t *testing.T) {
	// 30 < descLen <= 60 → score 6; no anchors → discriminativeness WARN
	b := makeD7Bridge("description: (45 chars)", longRawContent(45))
	score, diags := scoreD7(b)
	if score != 6 {
		t.Errorf("want 6, got %d", score)
	}
	// The 31-60 band emits no length-specific warning but discriminativeness WARN is expected.
	if !hasDiagSeverity(diags, "warning") {
		t.Errorf("expected discriminativeness WARN for 45-char no-anchor description, got %v", diags)
	}
}

// TestD7_DiscriminativenessWarning: long description with no anchors → score 10, WARN diagnostic.
func TestD7_DiscriminativenessWarning(t *testing.T) {
	b := makeD7Bridge("description: (210 chars)", longRawContent(210))
	score, diags := scoreD7(b)
	if score != 10 {
		t.Errorf("want score 10, got %d", score)
	}
	if !hasDiagD7Severity(diags, "warning") {
		t.Errorf("expected D7 WARN discriminativeness diagnostic, got %v", diags)
	}
}

// TestD7_DiscriminativenessInfo: long description with both anchors → score 10, hint diagnostic.
func TestD7_DiscriminativenessInfo(t *testing.T) {
	// negative anchor: "does not apply" + workflow anchor: "commit"
	desc := "Analyzes pull request diffs and detects style issues in each commit. Does not apply to auto-generated migration files or vendored dependencies. Use before merging to catch regressions early in your pipeline."
	raw := "---\ndescription: " + desc + "\n---\n"
	b := makeD7Bridge("description: (210 chars)", raw)
	score, diags := scoreD7(b)
	if score != 10 {
		t.Errorf("want score 10, got %d", score)
	}
	if !hasDiagD7Severity(diags, "hint") {
		t.Errorf("expected D7 hint discriminativeness diagnostic, got %v", diags)
	}
}

// TestD7_DiscriminativenessNeutral: long description with only one anchor → score 10, no discriminativeness diagnostic.
func TestD7_DiscriminativenessNeutral(t *testing.T) {
	// only workflow anchor (commit), no negative anchor
	desc := "Analyzes pull request diffs and detects style issues in each commit. Checks for common formatting violations and reports them back to the author for review before merging into the main branch here."
	raw := "---\ndescription: " + desc + "\n---\n"
	b := makeD7Bridge("description: (210 chars)", raw)
	score, diags := scoreD7(b)
	if score != 10 {
		t.Errorf("want score 10, got %d", score)
	}
	for _, d := range diags {
		if d.Dimension == "D7" && (d.Severity() == "warning" || d.Severity() == "hint") {
			t.Errorf("expected no D7 discriminativeness diagnostic for single-anchor description, got %v", d)
		}
	}
}

// TestD7_ExistingFixturesNoRegression verifies existing fixture D7 scores are not lower
// than the baseline captured before the discriminativeness change.
// Baseline D7 scores: skill-full → 10, skill-minimal → 6.
func TestD7_ExistingFixturesNoRegression(t *testing.T) {
	const fixturesDir = "../testdata/fixtures"
	cases := []struct {
		fixture  string
		minScore int
	}{
		{"skill-full", 10},
		{"skill-minimal", 6},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			skillPath := filepath.Join(fixturesDir, tc.fixture, "SKILL.md")
			result, err := Score(t.Context(), skillPath)
			if err != nil {
				t.Fatalf("Score() error: %v", err)
			}
			d7Score := result.Dimensions["patternRecognition"]
			if d7Score < tc.minScore {
				t.Errorf("D7 regression: %s got %d, expected >= %d", tc.fixture, d7Score, tc.minScore)
			}
		})
	}
}

// TestD7_WithAnchorsFixture verifies the skill-d7-with-anchors fixture emits an hint diagnostic.
func TestD7_WithAnchorsFixture(t *testing.T) {
	const fixturesDir = "../testdata/fixtures"
	skillPath := filepath.Join(fixturesDir, "skill-d7-with-anchors", "SKILL.md")
	result, err := Score(t.Context(), skillPath)
	if err != nil {
		t.Fatalf("Score() error: %v", err)
	}
	found := false
	for _, d := range result.WarningDetails {
		if d.Dimension == "D7" && d.Severity() == "hint" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected D7 hint diagnostic for skill-d7-with-anchors; WarningDetails=%v", result.WarningDetails)
	}
}

// TestD7_NoAnchorsFixture verifies the skill-d7-no-anchors fixture emits a WARN diagnostic.
func TestD7_NoAnchorsFixture(t *testing.T) {
	const fixturesDir = "../testdata/fixtures"
	skillPath := filepath.Join(fixturesDir, "skill-d7-no-anchors", "SKILL.md")
	result, err := Score(t.Context(), skillPath)
	if err != nil {
		t.Fatalf("Score() error: %v", err)
	}
	found := false
	for _, d := range result.WarningDetails {
		if d.Dimension == "D7" && d.Severity() == "warning" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected D7 WARN diagnostic for skill-d7-no-anchors; WarningDetails=%v", result.WarningDetails)
	}
}
