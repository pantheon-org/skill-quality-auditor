package scorer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func bridgeWithContent(strong, weak int, specificity float64) *validatorBridge {
	return &validatorBridge{Content: &types.ContentReport{
		StrongMarkers:          strong,
		WeakMarkers:            weak,
		InstructionSpecificity: specificity,
	}}
}

// ---------------------------------------------------------------------------
// scoreCalibrationBalance
// ---------------------------------------------------------------------------

func TestD6_CalibrationBalance_Balanced(t *testing.T) {
	// specificity 0.6 → balanced range [0.3, 0.8] → 5 pts
	b := bridgeWithContent(3, 2, 0.6)
	if got := scoreCalibrationBalance(b); got != 5 {
		t.Errorf("want 5, got %d", got)
	}
}

func TestD6_CalibrationBalance_MarginalLow(t *testing.T) {
	// specificity 0.25 → marginal range [0.2, 0.3) → 3 pts
	b := bridgeWithContent(1, 3, 0.25)
	if got := scoreCalibrationBalance(b); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_CalibrationBalance_MarginalHigh(t *testing.T) {
	// specificity 0.85 → marginal range (0.8, 0.9] → 3 pts
	b := bridgeWithContent(5, 1, 0.85)
	if got := scoreCalibrationBalance(b); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_CalibrationBalance_OutsideRange(t *testing.T) {
	// specificity 0.95 → outside all ranges → 1 pt
	b := bridgeWithContent(9, 1, 0.95)
	if got := scoreCalibrationBalance(b); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

func TestD6_CalibrationBalance_ZeroMarkers(t *testing.T) {
	// zero markers → 0 pts
	b := bridgeWithContent(0, 0, 1.0)
	if got := scoreCalibrationBalance(b); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestD6_CalibrationBalance_NilContent(t *testing.T) {
	if got := scoreCalibrationBalance(nilBridge()); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// scoreWhenNotToUse
// ---------------------------------------------------------------------------

func TestD6_WhenNotToUse_Present_SectionHeading(t *testing.T) {
	content := "## When NOT to Use\n\n- Do not apply for internal constants\n"
	if got := scoreWhenNotToUse(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_WhenNotToUse_Present_Phrase(t *testing.T) {
	content := "This skill is not intended for use with compile-time values.\n"
	if got := scoreWhenNotToUse(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_WhenNotToUse_Present_OutsideScope(t *testing.T) {
	content := "Outside the scope of this skill: anything unrelated to HTTP.\n"
	if got := scoreWhenNotToUse(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_WhenNotToUse_Present_AvoidUsing(t *testing.T) {
	content := "Avoid using this skill for internal microservice calls.\n"
	if got := scoreWhenNotToUse(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_WhenNotToUse_Present_DoNotUse(t *testing.T) {
	content := "Do not use this when the data is already validated.\n"
	if got := scoreWhenNotToUse(content); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_WhenNotToUse_Absent(t *testing.T) {
	content := "## When to Use\n\n- Processing user data\n"
	if got := scoreWhenNotToUse(content); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestD6_WhenNotToUse_Empty(t *testing.T) {
	if got := scoreWhenNotToUse(""); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// scorePermissiveImperativeRatio
// ---------------------------------------------------------------------------

func TestD6_PermissiveImperativeRatio_Balanced(t *testing.T) {
	// ratio 0.5 → [0.3, 0.7] → 3 pts
	b := &validatorBridge{Content: &types.ContentReport{
		StrongMarkers:   3,
		WeakMarkers:     2,
		ImperativeRatio: 0.5,
	}}
	if got := scorePermissiveImperativeRatio(b); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_PermissiveImperativeRatio_Marginal(t *testing.T) {
	// ratio 0.25 → [0.2, 0.3) → 2 pts
	b := &validatorBridge{Content: &types.ContentReport{
		StrongMarkers:   2,
		WeakMarkers:     1,
		ImperativeRatio: 0.25,
	}}
	if got := scorePermissiveImperativeRatio(b); got != 2 {
		t.Errorf("want 2, got %d", got)
	}
}

func TestD6_PermissiveImperativeRatio_Outside(t *testing.T) {
	// ratio 0.9 → outside all ranges → 1 pt
	b := &validatorBridge{Content: &types.ContentReport{
		StrongMarkers:   5,
		WeakMarkers:     1,
		ImperativeRatio: 0.9,
	}}
	if got := scorePermissiveImperativeRatio(b); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

func TestD6_PermissiveImperativeRatio_ZeroMarkers(t *testing.T) {
	b := bridgeWithContent(0, 0, 0.0)
	if got := scorePermissiveImperativeRatio(b); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestD6_PermissiveImperativeRatio_NilContent(t *testing.T) {
	if got := scorePermissiveImperativeRatio(nilBridge()); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// scoreConstraintTypology
// ---------------------------------------------------------------------------

func TestD6_ConstraintTypology_BothTypes_Full(t *testing.T) {
	// ≥2 hard and ≥2 soft uppercase markers → 4 pts
	content := "MUST validate inputs. NEVER skip validation. PREFER lightweight schemas. AVOID inline validation."
	b := bridgeWithContent(2, 2, 0.5)
	if got := scoreConstraintTypology(content, b); got != 4 {
		t.Errorf("want 4, got %d", got)
	}
}

func TestD6_ConstraintTypology_BothTypes_Minimal(t *testing.T) {
	// exactly 1 hard and 1 soft → 3 pts
	content := "MUST validate. PREFER lightweight approach."
	b := bridgeWithContent(1, 1, 0.5)
	if got := scoreConstraintTypology(content, b); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

func TestD6_ConstraintTypology_HardOnly(t *testing.T) {
	// hard markers only, no soft → 1 pt
	content := "MUST validate. NEVER skip. ALWAYS check."
	b := bridgeWithContent(3, 0, 1.0)
	if got := scoreConstraintTypology(content, b); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

func TestD6_ConstraintTypology_SoftOnly(t *testing.T) {
	// soft markers only, no hard → 1 pt
	content := "PREFER lightweight schemas. AVOID inline logic. TYPICALLY use schema-first."
	b := bridgeWithContent(0, 3, 0.0)
	if got := scoreConstraintTypology(content, b); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

func TestD6_ConstraintTypology_NoMarkers_FallbackBridge(t *testing.T) {
	// no ALLCAPS markers in content, but bridge shows imperativeRatio > 0 and weakMarkers > 0 → 2 pts
	content := "You should validate inputs. Consider using schemas for consistency."
	b := &validatorBridge{Content: &types.ContentReport{
		ImperativeRatio: 0.5,
		WeakMarkers:     2,
	}}
	if got := scoreConstraintTypology(content, b); got != 2 {
		t.Errorf("want 2, got %d", got)
	}
}

func TestD6_ConstraintTypology_None(t *testing.T) {
	// no constraint language at all → 0 pts
	content := "This skill describes validation workflows."
	b := bridgeWithContent(0, 0, 0.0)
	if got := scoreConstraintTypology(content, b); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestD6_ConstraintTypology_NilContent(t *testing.T) {
	content := "MUST do something."
	if got := scoreConstraintTypology(content, nilBridge()); got != 1 {
		t.Errorf("want 1 (hard only, nil bridge), got %d", got)
	}
}

// ---------------------------------------------------------------------------
// scoreD6 — new signature (content string, b *validatorBridge)
// ---------------------------------------------------------------------------

func TestD6_NilContent_NewSignature(t *testing.T) {
	if score, _ := scoreD6("", nilBridge()); score != 0 {
		t.Errorf("want 0 when bridge has no content, got %d", score)
	}
}

func TestD6_ZeroMarkers_NewSignature(t *testing.T) {
	b := bridgeWithContent(0, 0, 1.0)
	if score, _ := scoreD6("", b); score != 0 {
		t.Errorf("want 0 when no markers, got %d", score)
	}
}

func TestD6_HighScore_WellCalibrated(t *testing.T) {
	// balanced specificity (0.6) + when-not-to-use section + ≥2 hard and ≥2 soft markers
	// + imperativeRatio in balanced range → ≥13
	content := "MUST validate. NEVER skip. PREFER lightweight schema. AVOID inline logic.\n\n## When NOT to Use\n\nDo not apply for compile-time constants."
	b := &validatorBridge{Content: &types.ContentReport{
		StrongMarkers:          3,
		WeakMarkers:            2,
		InstructionSpecificity: 0.6,
		ImperativeRatio:        0.5, // balanced → 3 pts
	}}
	score, _ := scoreD6(content, b)
	// expected: calibrationBalance=5 + whenNotToUse=3 + imperativeRatio=3 + typology=4 = 15
	if score < 13 {
		t.Errorf("want ≥13 for well-calibrated skill, got %d", score)
	}
}

func TestD6_HardOnly_CappedScore(t *testing.T) {
	// hard-only constraints, no soft defaults, no when-not-to-use → ≤9
	content := "MUST validate. NEVER skip. ALWAYS check boundaries."
	b := bridgeWithContent(3, 0, 1.0)
	score, _ := scoreD6(content, b)
	if score > 9 {
		t.Errorf("want ≤9 for hard-only skill, got %d", score)
	}
}

// ---------------------------------------------------------------------------
// fixture integration tests
// ---------------------------------------------------------------------------

func TestD6_SkillFullFixture(t *testing.T) {
	fixtureDir := filepath.Join("..", "testdata", "fixtures", "skill-full")
	skillPath := filepath.Join(fixtureDir, "SKILL.md")
	contentBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Skipf("fixture not found: %v", err)
	}
	content := string(contentBytes)
	b := newValidatorBridge(fixtureDir)
	score, _ := scoreD6(content, b)
	// skill-full has been updated to include soft defaults → must score ≥13
	if score < 13 {
		t.Errorf("skill-full fixture: want ≥13, got %d", score)
	}
}

func TestD6_SkillMinimalFixture(t *testing.T) {
	fixtureDir := filepath.Join("..", "testdata", "fixtures", "skill-minimal")
	skillPath := filepath.Join(fixtureDir, "SKILL.md")
	contentBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Skipf("fixture not found: %v", err)
	}
	content := string(contentBytes)
	b := newValidatorBridge(fixtureDir)
	score, _ := scoreD6(content, b)
	if score != 0 {
		t.Errorf("skill-minimal fixture: want 0, got %d", score)
	}
}
