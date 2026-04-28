package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeSkillsDir creates a minimal skills/ tree under root and returns its path.
func makeSkillsDir(t *testing.T, root string) string {
	t.Helper()
	d := filepath.Join(root, "skills")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

// addSkill creates a skills/<name>/ directory; SKILL.md is written only if content != "".
func addSkill(t *testing.T, skillsDir, name, skillMDContent string) string {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if skillMDContent != "" {
		writeFile(t, filepath.Join(dir, "SKILL.md"), skillMDContent)
	}
	return dir
}

// --------------------------------------------------------------------------
// lintCmd RunE via direct function invocation
// --------------------------------------------------------------------------

func runLint(t *testing.T, skillsDir string) error {
	t.Helper()
	// Build a fake repo root with a go.mod so resolveRepoRoot works.
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	// Symlink (or copy) the skills dir inside the fake root so the command
	// can resolve it. Easiest: just pass an absolute path via args.
	cmd := lintCmd
	return cmd.RunE(cmd, []string{skillsDir})
}

func TestLint_noSkillsDir(t *testing.T) {
	// When skills/ doesn't exist the command should succeed with 0 issues.
	tmp := t.TempDir()
	nonExistent := filepath.Join(tmp, "skills")
	if err := runLint(t, nonExistent); err != nil {
		t.Errorf("expected nil for absent skills dir, got: %v", err)
	}
}

func TestLint_cleanSkill(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	addSkill(t, sd, "my-skill", "---\nname: my-skill\n---\n# Hello\n")
	if err := runLint(t, sd); err != nil {
		t.Errorf("expected clean skill to pass, got: %v", err)
	}
}

func TestLint_missingSkillMD(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	addSkill(t, sd, "bad-skill", "") // no SKILL.md
	err := runLint(t, sd)
	if err == nil {
		t.Error("expected error for missing SKILL.md")
	}
	if !strings.Contains(err.Error(), "lint failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLint_noFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	addSkill(t, sd, "no-fm", "# No frontmatter here\n")
	if err := runLint(t, sd); err == nil {
		t.Error("expected error for SKILL.md without frontmatter")
	}
}

func TestLint_badShebang(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	dir := addSkill(t, sd, "shebang-skill", "---\nname: shebang-skill\n---\n")
	scriptsDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(scriptsDir, "run.sh"), "#!/bin/sh\necho hi\n")
	if err := runLint(t, sd); err == nil {
		t.Error("expected error for bad shebang")
	}
}

func TestLint_goodShebang(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	dir := addSkill(t, sd, "shebang-skill", "---\nname: shebang-skill\n---\n")
	scriptsDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(scriptsDir, "run.sh"), "#!/usr/bin/env sh\necho hi\n")
	if err := runLint(t, sd); err != nil {
		t.Errorf("expected clean skill with good shebang, got: %v", err)
	}
}

func TestLint_nonShFileInScriptsIgnored(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	dir := addSkill(t, sd, "py-skill", "---\nname: py-skill\n---\n")
	scriptsDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// .py files are not checked by lint (only .sh)
	writeFile(t, filepath.Join(scriptsDir, "run.py"), "#!/usr/bin/env python3\n")
	if err := runLint(t, sd); err != nil {
		t.Errorf("expected .py file to be ignored by lint, got: %v", err)
	}
}

func TestLint_multipleIssues(t *testing.T) {
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	addSkill(t, sd, "missing", "")                      // no SKILL.md
	addSkill(t, sd, "no-fm", "# No frontmatter\n")      // no frontmatter
	addSkill(t, sd, "clean", "---\nname: clean\n---\n") // clean
	err := runLint(t, sd)
	if err == nil {
		t.Error("expected error for multiple issues")
	}
}

func TestLint_fileInSkillsDirIgnored(t *testing.T) {
	// Non-directory entries at the skills/ level should be silently skipped.
	tmp := t.TempDir()
	sd := makeSkillsDir(t, tmp)
	writeFile(t, filepath.Join(sd, "README.md"), "# readme\n")
	addSkill(t, sd, "real-skill", "---\nname: real-skill\n---\n")
	if err := runLint(t, sd); err != nil {
		t.Errorf("file at skills/ root should be ignored, got: %v", err)
	}
}
