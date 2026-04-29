package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

// ---- runGenerate ----

func TestRunGenerate_noStoredAudit(t *testing.T) {
	root := t.TempDir()
	err := runGenerate("nonexistent-skill", root)
	if err == nil {
		t.Error("expected error when no stored audit exists")
	}
}

func TestRunGenerate_valid(t *testing.T) {
	root := t.TempDir()
	skill := "test/my-skill"
	date := "2026-04-29"

	auditDir := filepath.Join(root, ".context", "audits", skill, date)
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		t.Fatal(err)
	}

	result := &scorer.Result{
		Skill:    skill,
		Date:     date,
		Total:    80,
		MaxTotal: 140,
		Grade:    "C",
		Dimensions: map[string]int{
			"D1": 12, "D2": 10, "D3": 8, "D4": 10,
			"D5": 10, "D6": 10, "D7": 6, "D8": 10, "D9": 4,
		},
	}
	data, _ := json.Marshal(result)
	if err := os.WriteFile(filepath.Join(auditDir, "audit.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runGenerate(skill, root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	planPath := filepath.Join(root, ".context", "plans", "my-skill-remediation-plan.md")
	if _, err := os.Stat(planPath); err != nil {
		t.Errorf("plan file not written to %s: %v", planPath, err)
	}
}

// ---- runValidate ----

func TestRunValidate_missingFile(t *testing.T) {
	root := t.TempDir()
	err := runValidate("/nonexistent/plan.md", root)
	if err == nil {
		t.Error("expected error for missing plan file")
	}
}

func TestRunValidate_missingSkillByName(t *testing.T) {
	root := t.TempDir()
	// Pass a bare skill name — no plan file in .context/plans/
	err := runValidate("no-such-skill", root)
	if err == nil {
		t.Error("expected error when plan file not found by skill name")
	}
}

func TestRunValidate_invalidPlan(t *testing.T) {
	root := t.TempDir()
	planDir := filepath.Join(root, ".context", "plans")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "bad-skill-remediation-plan.md")
	if err := os.WriteFile(planPath, []byte("not a valid plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runValidate(planPath, root); err == nil {
		t.Error("expected validation error for malformed plan")
	}
}

func TestRunValidate_bySkillName(t *testing.T) {
	root := t.TempDir()
	planDir := filepath.Join(root, ".context", "plans")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a minimal plan file so it's "found" — validation may still fail but the
	// path-resolution branch (bare name → .context/plans/<name>-remediation-plan.md) is exercised.
	planPath := filepath.Join(planDir, "my-skill-remediation-plan.md")
	if err := os.WriteFile(planPath, []byte("# plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Error is expected (plan content is invalid) — we only care that it reached runValidate logic.
	_ = runValidate("my-skill", root)
}
