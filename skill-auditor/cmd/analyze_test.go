package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const analyzeTestSkillContent = `# Test Skill

## When To Use
Use this skill when you need to authenticate users via OAuth.

## Trigger
Triggered when the user says "authenticate" or "login".

## Examples
Example 1: authenticate a user with oauth token.
Example 2: refresh token after expiry.

## Anti-Patterns
Never use password in plain text. Always hash credentials.
`

func makeAnalyzeTestSkill(t *testing.T) (skillPath string, repoRoot string) {
	t.Helper()
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "skills", "domain", "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	skillPath = filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(analyzeTestSkillContent), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	return skillPath, tmp
}

// captureAnalyzeOutput resets flag globals, calls RunE, and returns stdout.
func captureAnalyzeOutput(t *testing.T, skillArg, repoRoot string, semantic, patterns, pipeline, asJSON, store bool, limit int) (string, error) {
	t.Helper()
	analyzeSemantic = semantic
	analyzePatterns = patterns
	analyzePipeline = pipeline
	analyzeJSON = asJSON
	analyzeStore = store
	analyzeRepoRoot = repoRoot
	analyzeLimit = limit

	// Capture stdout
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	err := analyzeCmd.RunE(analyzeCmd, []string{skillArg})

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		t.Fatalf("copy pipe: %v", copyErr)
	}
	return buf.String(), err
}

func TestAnalyzeCmd_missingSkill(t *testing.T) {
	_, repoRoot := makeAnalyzeTestSkill(t)
	_, err := captureAnalyzeOutput(t, "/nonexistent/path/SKILL.md", repoRoot, false, false, false, false, false, 20)
	if err == nil {
		t.Error("expected error for missing skill path")
	}
}

func TestAnalyzeCmd_semantic(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, true, false, false, false, false, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Term") || !strings.Contains(out, "Score") {
		t.Errorf("expected keyword table in semantic output, got: %s", out)
	}
}

func TestAnalyzeCmd_patterns(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, true, false, false, false, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Rule") {
		t.Errorf("expected rule table in patterns output, got: %s", out)
	}
}

func TestAnalyzeCmd_pipeline(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, false, true, false, false, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Pattern Analysis Report") {
		t.Errorf("expected Pattern Analysis Report header, got: %s", out)
	}
	if !strings.Contains(out, "Top Keywords") {
		t.Errorf("expected Top Keywords section, got: %s", out)
	}
	if !strings.Contains(out, "Pattern Detection Results") {
		t.Errorf("expected Pattern Detection Results section, got: %s", out)
	}
}

func TestAnalyzeCmd_defaultIsPipeline(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, false, false, false, false, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Pattern Analysis Report") {
		t.Errorf("expected pipeline output as default, got: %s", out)
	}
}

func TestAnalyzeCmd_jsonOutput(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, false, false, true, false, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(out), &result); jsonErr != nil {
		t.Fatalf("expected valid JSON output, got: %s, err: %v", out, jsonErr)
	}
	if _, ok := result["skillKey"]; !ok {
		t.Errorf("expected 'skillKey' field in JSON output")
	}
}

func TestAnalyzeCmd_store(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	_, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, false, false, false, true, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	analysisDir := filepath.Join(repoRoot, ".context", "analysis")
	entries, err := os.ReadDir(analysisDir)
	if err != nil {
		t.Fatalf("expected analysis dir to exist: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "pattern-report-") && strings.HasSuffix(e.Name(), ".md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected pattern-report-*.md file in %s", analysisDir)
	}
}

func TestAnalyzeCmd_storeJSON(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	_, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, false, false, true, true, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	analysisDir := filepath.Join(repoRoot, ".context", "analysis")
	entries, err := os.ReadDir(analysisDir)
	if err != nil {
		t.Fatalf("expected analysis dir to exist: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "pattern-report-") && strings.HasSuffix(e.Name(), ".json") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected pattern-report-*.json file in %s", analysisDir)
	}
}

func TestAnalyzeCmd_limitFlag(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	_, err := captureAnalyzeOutput(t, skillPath, repoRoot, true, false, false, false, false, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---- buildSummary (via pipeline output) ----

func TestBuildSummary_withKeywords(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, false, false, false, false, false, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "pattern rules matched") {
		t.Errorf("expected summary with 'pattern rules matched', got: %s", out)
	}
}

func TestAnalyzeCmd_semanticLimitZero(t *testing.T) {
	skillPath, repoRoot := makeAnalyzeTestSkill(t)
	out, err := captureAnalyzeOutput(t, skillPath, repoRoot, true, false, false, false, false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With limit=0, ExtractKeywords returns empty, only headers shown
	_ = out
}
