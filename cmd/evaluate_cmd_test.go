package cmd

import (
	"bytes"
	"encoding/json"
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

func TestEvaluateCmd_defaultJSONOutput(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-full", "domain", "json-skill")
	out, err := runEvaluate(t, "--repo-root", root, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(out), &result); jsonErr != nil {
		t.Fatalf("expected valid JSON output by default, got: %s, err: %v", out, jsonErr)
	}
}

func TestEvaluateCmd_markdownFlag(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-full", "domain", "md-skill")
	out, err := runEvaluate(t, "--repo-root", root, "--markdown", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("expected non-empty markdown output")
	}
	// Markdown output should not be valid JSON
	var result map[string]interface{}
	if json.Unmarshal([]byte(out), &result) == nil {
		t.Error("expected markdown output, not JSON")
	}
}

func TestEvaluateCmd_markdownShorthand(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-full", "domain", "md-short-skill")
	out, err := runEvaluate(t, "--repo-root", root, "-m", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("expected non-empty markdown output with -m shorthand")
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

func TestEvaluateCmd_storeFlagShorthand(t *testing.T) {
	root, key := makeEvalSkill(t, "skill-minimal", "domain", "stored-short-skill")
	_, err := runEvaluate(t, "-r", root, "-s", key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	auditGlob := filepath.Join(root, ".context", "audits", "domain", "stored-short-skill", "*", "audit.json")
	matches, globErr := filepath.Glob(auditGlob)
	if globErr != nil || len(matches) == 0 {
		t.Errorf("expected audit.json to be stored with -s shorthand, glob matched: %v (err: %v)", matches, globErr)
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
