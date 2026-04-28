package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// setupLintRoot creates a temp repo root with go.mod + skills/ dir, chdirs into
// it for the duration of the test, and returns the skills dir path.
func setupLintRoot(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	sd := filepath.Join(tmp, "skills")
	if err := os.MkdirAll(sd, 0o755); err != nil {
		t.Fatal(err)
	}
	return sd
}

// TestLint_deprecatedAlias verifies that lint delegates to validate artifacts
// and reports MISSING_SKILL for a skill dir without SKILL.md.
func TestLint_deprecatedAlias(t *testing.T) {
	sd := setupLintRoot(t)
	if err := os.MkdirAll(filepath.Join(sd, "domain", "bad-skill"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := lintCmd.RunE(lintCmd, nil); err == nil {
		t.Error("expected error from deprecated lint alias when SKILL.md is missing")
	}
}

// TestLint_deprecatedAlias_clean verifies the alias passes for a valid skill.
func TestLint_deprecatedAlias_clean(t *testing.T) {
	sd := setupLintRoot(t)
	skillDir := filepath.Join(sd, "domain", "good-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(skillDir, "SKILL.md"), "---\nname: good-skill\n---\n# Hello\n")
	if err := lintCmd.RunE(lintCmd, nil); err != nil {
		t.Errorf("expected clean skill to pass via lint alias, got: %v", err)
	}
}
