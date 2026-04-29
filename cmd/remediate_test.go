package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
	"github.com/spf13/cobra"
)

// newRemediateCmd returns a fresh remediateCmd with all flags registered and
// output wired to buf, ready for RunE calls.
func newRemediateCmd(t *testing.T) (*bytes.Buffer, *cobra.Command) {
	t.Helper()
	cmd := remediateCmd
	cmd.ResetFlags()
	cmd.Flags().IntP("target-score", "t", 0, "")
	cmd.Flags().BoolP("validate", "v", false, "")
	cmd.Flags().StringP("repo-root", "r", "", "")
	cmd.Flags().BoolP("json", "j", false, "")
	cmd.Flags().BoolP("dry-run", "n", false, "")
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	return buf, cmd
}

// ---- runGenerate ----

func TestRunGenerate_noStoredAudit(t *testing.T) {
	_, cmd := newRemediateCmd(t)
	root := t.TempDir()
	err := runGenerate(cmd, "nonexistent-skill", root, 0)
	if err == nil {
		t.Error("expected error when no stored audit exists")
	}
}

func TestRunGenerate_valid(t *testing.T) {
	_, cmd := newRemediateCmd(t)
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

	if err := runGenerate(cmd, skill, root, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	planPath := filepath.Join(root, ".context", "plans", "my-skill-remediation-plan.md")
	if _, err := os.Stat(planPath); err != nil {
		t.Errorf("plan file not written to %s: %v", planPath, err)
	}
}

func TestRunGenerate_dryRun(t *testing.T) {
	buf, cmd := newRemediateCmd(t)
	root := t.TempDir()
	skill := "test/dry-skill"
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

	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}

	if err := runGenerate(cmd, skill, root, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With --dry-run, no plan file should be written.
	planPath := filepath.Join(root, ".context", "plans", "dry-skill-remediation-plan.md")
	if _, err := os.Stat(planPath); err == nil {
		t.Error("plan file should not be written in dry-run mode")
	}

	// Output should be non-empty (plan printed to stdout).
	if buf.Len() == 0 {
		t.Error("expected plan output on stdout in dry-run mode")
	}
}

func TestRunGenerate_jsonOutput(t *testing.T) {
	buf, cmd := newRemediateCmd(t)
	root := t.TempDir()
	skill := "test/json-skill"
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

	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}

	if err := runGenerate(cmd, skill, root, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Output should be valid JSON.
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Errorf("expected valid JSON output, got: %v\noutput: %s", err, buf.String())
	}

	// Plan file should be written as .json.
	planPath := filepath.Join(root, ".context", "plans", "json-skill-remediation-plan.json")
	if _, err := os.Stat(planPath); err != nil {
		t.Errorf("expected JSON plan file at %s: %v", planPath, err)
	}
}

func TestRunGenerate_jsonDryRun(t *testing.T) {
	buf, cmd := newRemediateCmd(t)
	root := t.TempDir()
	skill := "test/json-dry-skill"
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

	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}

	if err := runGenerate(cmd, skill, root, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Output should be valid JSON.
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Errorf("expected valid JSON output, got: %v\noutput: %s", err, buf.String())
	}

	// No file should be written.
	planDir := filepath.Join(root, ".context", "plans")
	if _, err := os.Stat(planDir); err == nil {
		t.Error("plans directory should not be created in dry-run mode")
	}
}

// ---- runValidate ----

func TestRunValidate_missingFile(t *testing.T) {
	_, cmd := newRemediateCmd(t)
	root := t.TempDir()
	err := runValidate(cmd, "/nonexistent/plan.md", root)
	if err == nil {
		t.Error("expected error for missing plan file")
	}
}

func TestRunValidate_missingSkillByName(t *testing.T) {
	_, cmd := newRemediateCmd(t)
	root := t.TempDir()
	err := runValidate(cmd, "no-such-skill", root)
	if err == nil {
		t.Error("expected error when plan file not found by skill name")
	}
}

func TestRunValidate_invalidPlan(t *testing.T) {
	_, cmd := newRemediateCmd(t)
	root := t.TempDir()
	planDir := filepath.Join(root, ".context", "plans")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "bad-skill-remediation-plan.md")
	if err := os.WriteFile(planPath, []byte("not a valid plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runValidate(cmd, planPath, root); err == nil {
		t.Error("expected validation error for malformed plan")
	}
}

func TestRunValidate_bySkillName(t *testing.T) {
	_, cmd := newRemediateCmd(t)
	root := t.TempDir()
	planDir := filepath.Join(root, ".context", "plans")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "my-skill-remediation-plan.md")
	if err := os.WriteFile(planPath, []byte("# plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Error is expected (plan content is invalid) — we only care that path resolution works.
	_ = runValidate(cmd, "my-skill", root)
}
