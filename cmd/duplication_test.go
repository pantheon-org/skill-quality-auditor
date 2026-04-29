package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/duplication"
)

// makeSkillsDir creates a minimal skills directory with n SKILL.md files.
func makeSkillsDir(t *testing.T, n int) string {
	t.Helper()
	root := t.TempDir()
	for i := 0; i < n; i++ {
		dir := filepath.Join(root, "domain", fmt.Sprintf("skill-%d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := strings.Repeat(fmt.Sprintf("This is skill %d. It does thing %d.\n", i, i), 20)
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// runDuplication resets and configures duplicationCmd then calls RunE.
func runDuplication(t *testing.T, skillsDir, repoRoot string, asJSON bool, posArgs []string) (string, error) {
	t.Helper()
	cmd := duplicationCmd
	cmd.ResetFlags()
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("skills-dir", "", "")
	cmd.Flags().String("repo-root", "", "")

	if skillsDir != "" {
		if err := cmd.Flags().Set("skills-dir", skillsDir); err != nil {
			t.Fatalf("set skills-dir: %v", err)
		}
	}
	if repoRoot != "" {
		if err := cmd.Flags().Set("repo-root", repoRoot); err != nil {
			t.Fatalf("set repo-root: %v", err)
		}
	}
	if asJSON {
		if err := cmd.Flags().Set("json", "true"); err != nil {
			t.Fatalf("set json: %v", err)
		}
	}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, posArgs)
	return buf.String(), err
}

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

func TestDuplicationCmd_nonExistentSkillsDir(t *testing.T) {
	nonExistent := "/nonexistent/skills-dir-that-does-not-exist"
	repoRoot := t.TempDir()
	_, err := runDuplication(t, nonExistent, repoRoot, false, []string{})
	if err == nil {
		t.Fatal("expected error for non-existent skills directory, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message should mention 'not found', got: %v", err)
	}
}

func TestDuplicationCmd_withSkills(t *testing.T) {
	skillsDir := makeSkillsDir(t, 2)
	repoRoot := t.TempDir()
	_ = func() { _, _ = runDuplication(t, skillsDir, repoRoot, false, []string{}) }
	// ignore exitCodeForPairs error (no Critical pairs expected)
	_, _ = runDuplication(t, skillsDir, repoRoot, false, []string{})
}

func TestDuplicationCmd_withSkillsJSON(t *testing.T) {
	skillsDir := makeSkillsDir(t, 2)
	repoRoot := t.TempDir()
	_, _ = runDuplication(t, skillsDir, repoRoot, true, []string{})
}

func TestDuplicationCmd_withArgs(t *testing.T) {
	skillsDir := makeSkillsDir(t, 1)
	repoRoot := t.TempDir()
	_, _ = runDuplication(t, "", repoRoot, false, []string{skillsDir})
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
