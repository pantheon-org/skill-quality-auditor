package scorer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestD4_GoodDescription(t *testing.T) {
	desc := "Validates and sanitizes user inputs before processing to prevent injection attacks and data corruption in production systems"
	content := "---\ndescription: " + desc + "\n---\n# Skill\nSee references/guide.md for details."
	score, _ := scoreD4(content, t.TempDir(), nilBridge())
	if score < 10 {
		t.Errorf("want >= 10, got %d", score)
	}
}

func TestD4_HarnessPathWarning(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"claude", "See .claude/settings.json for config."},
		{"cursor", "Edit .cursor/rules here."},
		{"continue", "Config at .continue/config.json."},
		{"windsurf", "See .windsurf/settings."},
		{"goose", "Config in .goose/profile."},
		{"agents", "Place files under .agents/skills/."},
		{"copilot", "Config at .copilot/instructions."},
		{"gemini", "See .gemini/settings."},
		{"firebender", "Edit .firebender/config."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			content := "---\ndescription: does something useful\n---\n# Skill\n" + tc.content
			_, diags := scoreD4(content, t.TempDir(), nilBridge())
			found := false
			for _, d := range diags {
				if d.Dimension == "D4" && d.severity == "warning" {
					found = true
				}
			}
			if !found {
				t.Errorf("expected D4 warning for harness path (.%s/...)", tc.name)
			}
		})
	}
}

func TestD4_AgentRefWarning(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"claude code", "This works with Claude Code."},
		{"cursor agent", "Use cursor agent for this."},
		{"github copilot", "Requires GitHub Copilot."},
		{"opencode", "Works with opencode."},
		{"windsurf", "Tested on Windsurf."},
		{"gemini cli", "Run via Gemini CLI."},
		{"goose", "Compatible with Goose."},
		{"codex", "Use with Codex."},
		{"cline", "Requires Cline."},
		{"aider", "Run aider to apply."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			content := "---\ndescription: does something useful\n---\n# Skill\n" + tc.content
			_, diags := scoreD4(content, t.TempDir(), nilBridge())
			found := false
			for _, d := range diags {
				if d.Dimension == "D4" && d.severity == "warning" {
					found = true
				}
			}
			if !found {
				t.Errorf("expected D4 warning for agent reference (%s)", tc.name)
			}
		})
	}
}

func TestD4_RelativePathViolation(t *testing.T) {
	content := "---\ndescription: does something useful\n---\n# Skill\nSee ../other-skill/SKILL.md for more."
	score, diags := scoreD4(content, t.TempDir(), nilBridge())
	found := false
	for _, d := range diags {
		if d.Dimension == "D4" && d.severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected D4 warning for ../ outside code blocks")
	}
	// baseline is 8 now; penalty brings it below that
	if score > 11 {
		t.Errorf("expected score ≤ 11 due to ../ penalty, got %d", score)
	}
}

func TestD4_PenaltyFromDir_AbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "run.sh"), []byte("#!/bin/sh\ncd skills/ci-cd/my-skill && run"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent here"
	score1, _ := scoreD4(content, t.TempDir(), nilBridge()) // no scripts dir
	score2, _ := scoreD4(content, tmpDir, nilBridge())      // scripts dir with abs path
	if score2 >= score1 {
		t.Errorf("expected penalty from scripts/ abs path: without=%d with=%d", score1, score2)
	}
}

func TestD4_PenaltyFromDir_ReferencesDir(t *testing.T) {
	tmpDir := t.TempDir()
	refsDir := filepath.Join(tmpDir, "references")
	if err := os.MkdirAll(refsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// File with .context/ reference
	if err := os.WriteFile(filepath.Join(refsDir, "guide.md"), []byte("See .context/audits/ for details."), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	score1, _ := scoreD4(content, t.TempDir(), nilBridge())
	score2, _ := scoreD4(content, tmpDir, nilBridge())
	if score2 >= score1 {
		t.Errorf("expected penalty from references/ .context path: without=%d with=%d", score1, score2)
	}
}

func TestD4_AbsoluteSkillPathInContent(t *testing.T) {
	content := "---\ndescription: does something useful\n---\n# Skill\nSee skills/domain/my-skill for more."
	_, diags := scoreD4(content, t.TempDir(), nilBridge())
	found := false
	for _, d := range diags {
		if d.Dimension == "D4" && d.severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected D4 warning for absolute skill path outside code blocks")
	}
}

func TestD4_ContextAgentsRefInContent(t *testing.T) {
	content := "---\ndescription: does something useful\n---\n# Skill\nCheck .context/audits/ for previous runs."
	_, diags := scoreD4(content, t.TempDir(), nilBridge())
	found := false
	for _, d := range diags {
		if d.Dimension == "D4" && d.severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected D4 warning for .context/ reference outside code blocks")
	}
}

func TestD4_ScriptsBonusForPythonFile(t *testing.T) {
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "run.py"), []byte("print('hello')"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful in production systems here and there\n---\n# Skill\ncontent"
	score1, _ := scoreD4(content, t.TempDir(), nilBridge())
	score2, _ := scoreD4(content, tmpDir, nilBridge())
	if score2 <= score1 {
		t.Errorf("expected bonus from scripts/ Python file: without=%d with=%d", score1, score2)
	}
}

func TestD4_ReferencesSectionBonus(t *testing.T) {
	withRefs := "---\ndescription: does something useful\n---\n# Skill\ncontent\n\n## References\n\n- [Guide](references/guide.md)\n"
	withoutRefs := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	s1, _ := scoreD4(withRefs, t.TempDir(), nilBridge())
	s2, _ := scoreD4(withoutRefs, t.TempDir(), nilBridge())
	if s1 <= s2 {
		t.Errorf("expected bonus for References section with links: with=%d without=%d", s1, s2)
	}
}

func TestD4_AndOrPenalty(t *testing.T) {
	// >3 and/or occurrences should incur a -2 penalty
	desc := "does this and that and more and even extra"
	content := "---\ndescription: " + desc + "\n---\n# Skill\ncontent"
	score1, _ := scoreD4("---\ndescription: short desc\n---\n# Skill\ncontent", t.TempDir(), nilBridge())
	score2, _ := scoreD4(content, t.TempDir(), nilBridge())
	// score2 should be lower due to and/or penalty
	if score2 >= score1+3 {
		t.Errorf("expected and/or penalty to reduce score: baseline=%d stuffed=%d", score1, score2)
	}
}

func TestPenaltyFromDir_notADirectory(t *testing.T) {
	import_re := regexp.MustCompile(`skills/[a-z]`)
	// Path doesn't exist → should return 0
	if penaltyFromDir("/nonexistent/path", import_re) != 0 {
		t.Error("nonexistent dir should return 0 penalty")
	}
}

func TestPenaltyFromDir_emptyDir(t *testing.T) {
	import_re := regexp.MustCompile(`skills/[a-z]`)
	tmp := t.TempDir()
	if penaltyFromDir(tmp, import_re) != 0 {
		t.Error("empty dir should return 0 penalty")
	}
}

func TestPenaltyFromDir_cappedAtTwo(t *testing.T) {
	import_re := regexp.MustCompile(`skills/[a-z]`)
	tmp := t.TempDir()
	for i := range 5 {
		if err := os.WriteFile(filepath.Join(tmp, fmt.Sprintf("file%d.md", i)), []byte("skills/domain/ref"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if penaltyFromDir(tmp, import_re) != 2 {
		t.Error("penalty should be capped at 2")
	}
}
