package scorer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func TestD4_AgentRefNoFalsePositive(t *testing.T) {
	// "example" and "sample" contain "amp" as a substring but must not trigger
	// the agent-reference warning after switching to word-boundary matching.
	cases := []string{
		"See the examples section for details.",
		"Here is a sample workflow.",
		"A champion approach to scoring.",
	}
	for _, content := range cases {
		full := "---\ndescription: does something useful\n---\n# Skill\n" + content
		_, diags := scoreD4(full, t.TempDir(), nilBridge())
		for _, d := range diags {
			if d.Dimension == "D4" && d.severity == "warning" && d.Message != "" {
				if len(d.Message) > 30 && d.Message[len(d.Message)-3:] == "amp" {
					t.Errorf("false positive: %q triggered amp agent warning", content)
				}
			}
		}
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

// --- Mutation Resistance tests ---

func TestD4_MutationResistance_ZeroCredit(t *testing.T) {
	// No hard constraints, no conditional branches, no exclusions → 0 pts
	content := "---\ndescription: does something useful\n---\n# Skill\nDo the thing. Be helpful."
	score1, _ := scoreD4(content, t.TempDir(), nilBridge())
	// Verify scoreSpecificationMutationResistance returns 0 for this content
	if got := scoreSpecificationMutationResistance(content); got != 0 {
		t.Errorf("want 0 mutation resistance pts, got %d", got)
	}
	_ = score1
}

func TestD4_MutationResistance_OneOfThree_HardConstraint(t *testing.T) {
	// MUST + ≥4-word verb phrase → 1 pt (int(1.5) = 1)
	content := "---\ndescription: does something useful\n---\n# Skill\nMUST validate all user inputs before processing them."
	if got := scoreSpecificationMutationResistance(content); got != 1 {
		t.Errorf("want 1 pt for hard constraint only, got %d", got)
	}
}

func TestD4_MutationResistance_OneOfThree_Conditional(t *testing.T) {
	// if + ≥2-word noun phrase → 1 pt (int(1.5) = 1)
	content := "---\ndescription: does something useful\n---\n# Skill\nif the user provides a file path, validate it first."
	if got := scoreSpecificationMutationResistance(content); got != 1 {
		t.Errorf("want 1 pt for conditional branch only, got %d", got)
	}
}

func TestD4_MutationResistance_OneOfThree_Exclusion(t *testing.T) {
	// does not + ≥6-word sentence → 1 pt
	content := "---\ndescription: does something useful\n---\n# Skill\nThis skill does not handle authentication or session management."
	if got := scoreSpecificationMutationResistance(content); got != 1 {
		t.Errorf("want 1 pt for exclusion only, got %d", got)
	}
}

func TestD4_MutationResistance_TwoOfThree_HardConstraintAndConditional(t *testing.T) {
	// MUST + if → 1.5 + 1.5 = 3 pts
	content := "---\ndescription: does something useful\n---\n# Skill\nMUST validate all user inputs before processing them.\nif the user provides a file path, validate it first."
	if got := scoreSpecificationMutationResistance(content); got != 3 {
		t.Errorf("want 3 pts for hard constraint + conditional, got %d", got)
	}
}

func TestD4_MutationResistance_TwoOfThree_HardConstraintAndExclusion(t *testing.T) {
	// MUST + does not → 1.5 + 1.0 = 2 pts (int(2.5) = 2)
	content := "---\ndescription: does something useful\n---\n# Skill\nMUST validate all user inputs before processing them.\nThis skill does not handle authentication or session management."
	if got := scoreSpecificationMutationResistance(content); got != 2 {
		t.Errorf("want 2 pts for hard constraint + exclusion, got %d", got)
	}
}

func TestD4_MutationResistance_FullCredit(t *testing.T) {
	// MUST + if + does not → 1.5 + 1.5 + 1.0 = 4 pts
	content := "---\ndescription: does something useful\n---\n# Skill\n" +
		"MUST validate all user inputs before processing them.\n" +
		"if the user provides a file path, validate it first.\n" +
		"This skill does not handle authentication or session management."
	if got := scoreSpecificationMutationResistance(content); got != 4 {
		t.Errorf("want 4 pts for all three criteria, got %d", got)
	}
}

func TestD4_MutationResistance_CodeBlockFalsePositiveGuard(t *testing.T) {
	// Keywords inside code blocks must NOT earn credit
	content := "---\ndescription: does something useful\n---\n# Skill\n```\nMUST validate all user inputs before processing them.\nif the user provides a file path, validate it first.\nThis skill does not handle authentication or session management.\n```"
	if got := scoreSpecificationMutationResistance(content); got != 0 {
		t.Errorf("want 0 pts when keywords are inside code blocks, got %d", got)
	}
}

func TestD4_MutationResistance_NonSpecificMUST_NoCredit(t *testing.T) {
	// "MUST follow best practices" is explicitly excluded (vague phrase)
	content := "---\ndescription: does something useful\n---\n# Skill\nMUST follow best practices when writing code."
	if got := scoreSpecificationMutationResistance(content); got != 0 {
		t.Errorf("want 0 pts for non-specific MUST phrase, got %d", got)
	}
}

func TestD4_MutationResistance_NonSpecificIf_NoCredit(t *testing.T) {
	// "if needed, proceed" — comma after first word prevents 2-word noun phrase match
	content := "---\ndescription: does something useful\n---\n# Skill\nif needed, proceed with caution."
	if got := scoreSpecificationMutationResistance(content); got != 0 {
		t.Errorf("want 0 pts for non-specific if phrase, got %d", got)
	}
}

func TestD4_E2E_VagueSkill_ScoresBelow12(t *testing.T) {
	// A skill with no specificity markers should score ≤ 11
	content := "---\ndescription: does something useful\n---\n# Skill\nDo the thing. Be helpful. Work well."
	score, _ := scoreD4(content, t.TempDir(), nilBridge())
	if score > 11 {
		t.Errorf("vague skill: want score <= 11, got %d", score)
	}
}

func TestD4_E2E_MutationResistantSkill_ScoresAtLeast13(t *testing.T) {
	// A well-specified skill with all three criteria should score ≥ 13
	content := "---\ndescription: Validates and sanitizes user inputs before processing to prevent injection attacks and data corruption in production systems\n---\n# Skill\n" +
		"MUST validate all user inputs before processing them.\n" +
		"if the user provides a file path, validate it first.\n" +
		"This skill does not handle authentication or session management.\n" +
		"See references/guide.md for details."
	score, _ := scoreD4(content, t.TempDir(), nilBridge())
	if score < 13 {
		t.Errorf("mutation-resistant skill: want score >= 13, got %d", score)
	}
}

func TestD4_LooseScripts_Warning(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"run.sh", "deploy.py", "build.ts", "helper.js"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	_, diags := scoreD4(content, tmpDir, nilBridge())
	warnCount := 0
	for _, d := range diags {
		if d.Dimension == "D4" && d.severity == "warning" && strings.Contains(d.Message, "should live in scripts/") {
			warnCount++
		}
	}
	if warnCount != 4 {
		t.Errorf("expected 4 loose-script warnings, got %d", warnCount)
	}
}

func TestD4_LooseScripts_SKILLmdNotFlagged(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	_, diags := scoreD4(content, tmpDir, nilBridge())
	for _, d := range diags {
		if d.Dimension == "D4" && strings.Contains(d.Message, "SKILL.md") {
			t.Error("SKILL.md should not be flagged as a loose script")
		}
	}
}

func TestD4_LooseScripts_ScriptsDirNotFlagged(t *testing.T) {
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"run.sh", "deploy.py"} {
		if err := os.WriteFile(filepath.Join(scriptsDir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	_, diags := scoreD4(content, tmpDir, nilBridge())
	for _, d := range diags {
		if d.Dimension == "D4" && strings.Contains(d.Message, "should live in scripts/") {
			t.Error("scripts inside scripts/ dir should not be flagged")
			break
		}
	}
}

func TestD4_LooseScripts_NoFalsePositive(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"README.md", "notes.txt", "config.yaml", "template.json"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	_, diags := scoreD4(content, tmpDir, nilBridge())
	for _, d := range diags {
		if d.Dimension == "D4" && strings.Contains(d.Message, "should live in scripts/") {
			t.Errorf("unexpected loose-script warning for non-script file: %s", d.Message)
			break
		}
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

func TestD4_ArtifactTrio_FullBonus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "schemas", "thing.schema.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "templates", "thing-template.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "validate.sh"), []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	score, _ := scoreD4(content, tmpDir, nilBridge())
	// baseline 8 + trio bonus 1 = 9
	if score < 9 {
		t.Errorf("expected >= 9 with artifact trio, got %d", score)
	}
}

func TestD4_ArtifactTrio_NonShellScript(t *testing.T) {
	for _, script := range []string{"validate.py", "validate.ts", "validate.js"} {
		t.Run(script, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "schemas"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "templates"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "assets", "schemas", "thing.schema.json"), []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "assets", "templates", "thing-template.yaml"), []byte("key: val"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "scripts", script), []byte("content"), 0o644); err != nil {
				t.Fatal(err)
			}
			content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
			s1, _ := scoreD4(content, t.TempDir(), nilBridge())
			s2, _ := scoreD4(content, tmpDir, nilBridge())
			if s2 <= s1 {
				t.Errorf("expected bonus with %s: without=%d with=%d", script, s1, s2)
			}
		})
	}
}

func TestD4_ArtifactTrio_MissingScript_NoBonus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "schemas", "thing.schema.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "templates", "thing-template.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	s1, _ := scoreD4(content, t.TempDir(), nilBridge())
	s2, _ := scoreD4(content, tmpDir, nilBridge())
	if s2 > s1 {
		t.Errorf("expected no bonus without scripts/: without=%d with=%d", s1, s2)
	}
}

func TestD4_ArtifactTrio_MissingSchemas_NoBonus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "templates", "thing-template.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "validate.sh"), []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	s1, _ := scoreD4(content, t.TempDir(), nilBridge())
	s2, _ := scoreD4(content, tmpDir, nilBridge())
	if s2 > s1 {
		t.Errorf("expected no bonus without schemas/: without=%d with=%d", s1, s2)
	}
}

func TestD4_ArtifactTrio_MissingTemplates_NoBonus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "schemas", "thing.schema.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "validate.sh"), []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	s1, _ := scoreD4(content, t.TempDir(), nilBridge())
	s2, _ := scoreD4(content, tmpDir, nilBridge())
	if s2 > s1 {
		t.Errorf("expected no bonus without templates/: without=%d with=%d", s1, s2)
	}
}

func TestD4_ArtifactTrio_EmptyDirs_NoBonus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	// directories exist but are empty
	content := "---\ndescription: does something useful\n---\n# Skill\ncontent"
	s1, _ := scoreD4(content, t.TempDir(), nilBridge())
	s2, _ := scoreD4(content, tmpDir, nilBridge())
	if s2 > s1 {
		t.Errorf("expected no bonus with empty directories: without=%d with=%d", s1, s2)
	}
}

func TestD4_E2E_FullArtifactTrioSkill_ScoresAboveBaseline(t *testing.T) {
	// A skill with all three artifact directories should score >= a baseline skill
	baseContent := "---\ndescription: Validates and sanitizes user inputs before processing to prevent injection attacks and data corruption in production systems\n---\n# Skill\nSee references/guide.md for details."
	baseScore, _ := scoreD4(baseContent, t.TempDir(), nilBridge())

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets", "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "schemas", "thing.schema.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "templates", "thing-template.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "scripts", "validate.sh"), []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	trioScore, _ := scoreD4(baseContent, tmpDir, nilBridge())

	if trioScore <= baseScore {
		t.Errorf("expected trio artifact skill to score >= baseline: baseline=%d trio=%d", baseScore, trioScore)
	}
}
