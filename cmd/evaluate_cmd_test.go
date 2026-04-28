package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runEvaluate(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"evaluate"}, args...))
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	return buf.String(), err
}

// makeEvalSkill creates a temp repo with <fixture> SKILL.md placed under
// skills/<domain>/<name>/ and returns (repoRoot, "domain/name").
func makeEvalSkill(t *testing.T, fixture, domain, name string) (repoRoot, key string) {
	t.Helper()
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "skills", domain, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(fixturesBase, fixture, "SKILL.md"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), src, 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	return tmp, domain + "/" + name
}

func TestEvaluateCmd_fullSkill(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-full", "domain", "full")
	_, err := runEvaluate(t, "--repo-root", root, key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEvaluateCmd_minimalSkill(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-minimal", "domain", "minimal")
	_, err := runEvaluate(t, "--repo-root", root, key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEvaluateCmd_jsonFlag(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-full", "domain", "json-skill")
	_, err := runEvaluate(t, "--repo-root", root, "--json", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluateCmd_nonexistentSkill(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	_, err := runEvaluate(t, "--repo-root", tmp, "domain/no-such-skill")
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

func TestEvaluateCmd_storeFlag(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-minimal", "domain", "stored-skill")
	_, err := runEvaluate(t, "--repo-root", root, "--store", key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	auditGlob := filepath.Join(root, ".context", "audits", "domain", "stored-skill", "*", "audit.json")
	matches, globErr := filepath.Glob(auditGlob)
	if globErr != nil || len(matches) == 0 {
		t.Errorf("expected audit.json to be stored, glob matched: %v (err: %v)", matches, globErr)
	}
}

func TestEvaluateCmd_skillPathNotUnderSkillsDir(t *testing.T) {
	// A path that exists but is not under <repoRoot>/skills/ should error at canonicalSkillKey
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	skillFile := filepath.Join(tmp, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# Title\n"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	_, err := runEvaluate(t, "--repo-root", tmp, skillFile)
	if err == nil {
		t.Error("expected error when skill path is not under skills/")
	}
	if err != nil && !strings.Contains(err.Error(), "not under") {
		t.Errorf("expected 'not under' in error, got: %v", err)
	}
}
