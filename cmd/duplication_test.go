package cmd

import (
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

// TestDuplicationCmd_nonExistentSkillsDir verifies that the duplication command
// returns an error (and does not panic) when the skills directory does not exist.
func TestDuplicationCmd_nonExistentSkillsDir(t *testing.T) {
	nonExistent := "/nonexistent/skills-dir-that-does-not-exist"
	// pathExists returns false → the RunE closure returns a formatted error.
	// We exercise that guard directly via the cobra command.
	cmd := duplicationCmd
	cmd.ResetFlags()
	// Re-register flags so the command is usable standalone.
	cmd.Flags().BoolVar(&dupJSON, "json", false, "")
	cmd.Flags().StringVar(&dupSkillsDir, "skills-dir", "", "")
	cmd.Flags().StringVar(&dupRepoRoot, "repo-root", "", "")

	// Set skills-dir flag to the non-existent path and provide a fake repo-root
	// so auto-detection is bypassed.
	dupSkillsDir = nonExistent
	dupRepoRoot = t.TempDir() // valid dir so resolveRepoRoot succeeds

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for non-existent skills directory, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message should mention 'not found', got: %v", err)
	}
}

// TestDuplicationCmd_withSkills exercises the full RunE path including report
// generation and exitCodeForPairs (the markdown text output path).
func TestDuplicationCmd_withSkills(t *testing.T) {
	skillsDir := makeSkillsDir(t, 2)
	repoRoot := t.TempDir()

	origDir, origRoot := dupSkillsDir, dupRepoRoot
	dupSkillsDir = skillsDir
	dupRepoRoot = repoRoot
	defer func() { dupSkillsDir = origDir; dupRepoRoot = origRoot }()

	_ = duplicationCmd.RunE(duplicationCmd, []string{})
}

// TestDuplicationCmd_withSkillsJSON exercises the JSON output path (line 62
// exitCodeForPairs return).
func TestDuplicationCmd_withSkillsJSON(t *testing.T) {
	skillsDir := makeSkillsDir(t, 2)
	repoRoot := t.TempDir()

	origDir, origRoot, origJSON := dupSkillsDir, dupRepoRoot, dupJSON
	dupSkillsDir = skillsDir
	dupRepoRoot = repoRoot
	dupJSON = true
	defer func() { dupSkillsDir = origDir; dupRepoRoot = origRoot; dupJSON = origJSON }()

	_ = duplicationCmd.RunE(duplicationCmd, []string{})
}

// TestDuplicationCmd_withArgs exercises the args[0] skillsDir override path.
func TestDuplicationCmd_withArgs(t *testing.T) {
	skillsDir := makeSkillsDir(t, 1)
	repoRoot := t.TempDir()

	origDir, origRoot := dupSkillsDir, dupRepoRoot
	dupSkillsDir = ""
	dupRepoRoot = repoRoot
	defer func() { dupSkillsDir = origDir; dupRepoRoot = origRoot }()

	_ = duplicationCmd.RunE(duplicationCmd, []string{skillsDir})
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
