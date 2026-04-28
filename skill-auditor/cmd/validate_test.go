package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// --------------------------------------------------------------------------
// firstLine
// --------------------------------------------------------------------------

func TestFirstLine_returnsFirstLine(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "f.sh")
	writeFile(t, p, "#!/usr/bin/env sh\necho hello\n")
	if got := firstLine(p); got != "#!/usr/bin/env sh" {
		t.Errorf("got %q, want #!/usr/bin/env sh", got)
	}
}

func TestFirstLine_emptyFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "empty.sh")
	writeFile(t, p, "")
	if got := firstLine(p); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFirstLine_missingFile(t *testing.T) {
	if got := firstLine("/nonexistent/file.sh"); got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

// --------------------------------------------------------------------------
// fileContains
// --------------------------------------------------------------------------

func TestFileContains_found(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "f.sh")
	writeFile(t, p, "#!/usr/bin/env bash\n# shell: bash\n")
	if !fileContains(p, "# shell: bash") {
		t.Error("expected fileContains to return true")
	}
}

func TestFileContains_notFound(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "f.sh")
	writeFile(t, p, "#!/usr/bin/env sh\n")
	if fileContains(p, "# shell: bash") {
		t.Error("expected fileContains to return false")
	}
}

func TestFileContains_missingFile(t *testing.T) {
	if fileContains("/nonexistent/file.sh", "anything") {
		t.Error("expected false for missing file")
	}
}

// --------------------------------------------------------------------------
// Markdown helpers
// --------------------------------------------------------------------------

func TestExtractH1_present(t *testing.T) {
	content := "# My Title\n## Section\n"
	if got := extractH1(content); got != "My Title" {
		t.Errorf("got %q, want 'My Title'", got)
	}
}

func TestExtractH1_missing(t *testing.T) {
	if got := extractH1("## Only H2\n"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestExtractH2Headings(t *testing.T) {
	content := "# Title\n## Alpha\n## Beta\nsome text\n## Gamma\n"
	got := extractH2Headings(content)
	want := []string{"Alpha", "Beta", "Gamma"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExtractFrontmatter_present(t *testing.T) {
	content := "---\nfoo: bar\nbaz: qux\n---\n# Title\n"
	fm := extractFrontmatter(content)
	if !strings.Contains(fm, "foo: bar") {
		t.Errorf("expected frontmatter to contain 'foo: bar', got %q", fm)
	}
}

func TestExtractFrontmatter_missing(t *testing.T) {
	if got := extractFrontmatter("# No frontmatter\n"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestHasFrontmatterKey_found(t *testing.T) {
	fm := "review_date: 2026-04-28\nreviewer: alice\n"
	if !hasFrontmatterKey(fm, "review_date") {
		t.Error("expected hasFrontmatterKey to return true")
	}
}

func TestHasFrontmatterKey_notFound(t *testing.T) {
	fm := "reviewer: alice\n"
	if hasFrontmatterKey(fm, "review_date") {
		t.Error("expected hasFrontmatterKey to return false")
	}
}

func TestH2GroupPresent_found(t *testing.T) {
	h2s := []string{"Executive Summary", "Dimension Scores"}
	if !h2GroupPresent([]string{"Summary", "Executive Summary"}, h2s) {
		t.Error("expected group to be found via alternate name")
	}
}

func TestH2GroupPresent_notFound(t *testing.T) {
	h2s := []string{"Introduction"}
	if h2GroupPresent([]string{"Executive Summary", "Summary"}, h2s) {
		t.Error("expected group not to be found")
	}
}

func TestH2GroupIndex_found(t *testing.T) {
	h2s := []string{"Alpha", "Beta", "Gamma"}
	if got := h2GroupIndex([]string{"Beta"}, h2s); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestH2GroupIndex_notFound(t *testing.T) {
	if got := h2GroupIndex([]string{"Missing"}, []string{"Alpha"}); got != -1 {
		t.Errorf("got %d, want -1", got)
	}
}

// --------------------------------------------------------------------------
// artifactValidator — checkSchemaFile
// --------------------------------------------------------------------------

func TestCheckSchemaFile_valid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", "foo.schema.json")
	writeFile(t, p, `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`)
	v := &artifactValidator{}
	v.checkSchemaFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors, got %d", v.errors)
	}
}

func TestCheckSchemaFile_wrongExtension(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", "foo.json")
	writeFile(t, p, `{}`)
	v := &artifactValidator{}
	v.checkSchemaFile(p)
	if v.errors == 0 {
		t.Error("expected error for wrong extension")
	}
}

func TestCheckSchemaFile_invalidJSON(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", "bad.schema.json")
	writeFile(t, p, `not json`)
	v := &artifactValidator{}
	v.checkSchemaFile(p)
	if v.errors == 0 {
		t.Error("expected error for invalid JSON")
	}
}

func TestCheckSchemaFile_missingSchemaURL(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", "no-schema.schema.json")
	writeFile(t, p, `{"type":"object"}`)
	v := &artifactValidator{}
	v.checkSchemaFile(p)
	if v.errors == 0 {
		t.Error("expected error for missing $schema URL")
	}
}

func TestCheckSchemaFile_gitkeep(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", ".gitkeep")
	writeFile(t, p, "")
	v := &artifactValidator{}
	v.checkSchemaFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for .gitkeep, got %d", v.errors)
	}
}

func TestCheckSchemaFile_unreadable(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", "missing.schema.json")
	// File does not exist — ReadFile will fail.
	v := &artifactValidator{}
	v.checkSchemaFile(p)
	if v.errors == 0 {
		t.Error("expected error for unreadable file")
	}
}

// --------------------------------------------------------------------------
// artifactValidator — checkTemplateFile
// --------------------------------------------------------------------------

func TestCheckTemplateFile_validYAML(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "templates", "plan.yaml")
	writeFile(t, p, "key: value\n")
	v := &artifactValidator{}
	v.checkTemplateFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors, got %d", v.errors)
	}
}

func TestCheckTemplateFile_templateSyntaxSkipped(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "templates", "tmpl.yaml")
	writeFile(t, p, "key: {{ .Value }}\n")
	v := &artifactValidator{}
	v.checkTemplateFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for template syntax file, got %d", v.errors)
	}
}

func TestCheckTemplateFile_emptyYAML(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "templates", "empty.yaml")
	writeFile(t, p, "   \n  \n")
	v := &artifactValidator{}
	v.checkTemplateFile(p)
	if v.errors == 0 {
		t.Error("expected error for empty template")
	}
}

func TestCheckTemplateFile_nonYAMLIgnored(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "templates", "readme.md")
	writeFile(t, p, "# Hello\n")
	v := &artifactValidator{}
	v.checkTemplateFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for non-YAML file, got %d", v.errors)
	}
}

func TestCheckTemplateFile_unreadable(t *testing.T) {
	// Pass a path that does not exist — os.ReadFile will fail.
	v := &artifactValidator{}
	v.checkTemplateFile("/nonexistent/assets/templates/missing.yaml")
	if v.errors == 0 {
		t.Error("expected error for unreadable template file")
	}
}

func TestCheckTemplateFile_gitkeep(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "templates", ".gitkeep")
	writeFile(t, p, "")
	v := &artifactValidator{}
	v.checkTemplateFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for .gitkeep, got %d", v.errors)
	}
}

// --------------------------------------------------------------------------
// artifactValidator — checkScriptFile / shebang checks
// --------------------------------------------------------------------------

func TestCheckShellScript_portableShebang(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.sh")
	writeFile(t, p, "#!/usr/bin/env sh\necho hi\n")
	v := &artifactValidator{}
	v.checkShellScript(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors, got %d", v.errors)
	}
}

func TestCheckShellScript_bashWithAnnotation(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.sh")
	writeFile(t, p, "#!/usr/bin/env bash\n# shell: bash\necho hi\n")
	v := &artifactValidator{}
	v.checkShellScript(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for bash with annotation, got %d", v.errors)
	}
}

func TestCheckShellScript_bashWithoutAnnotation(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.sh")
	writeFile(t, p, "#!/usr/bin/env bash\necho hi\n")
	v := &artifactValidator{}
	v.checkShellScript(p)
	if v.errors == 0 {
		t.Error("expected error for bash without annotation")
	}
}

func TestCheckShellScript_badShebang(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.sh")
	writeFile(t, p, "#!/bin/sh\necho hi\n")
	v := &artifactValidator{}
	v.checkShellScript(p)
	if v.errors == 0 {
		t.Error("expected error for non-portable shebang")
	}
}

func TestCheckPythonScript_valid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.py")
	writeFile(t, p, "#!/usr/bin/env python3\nprint('hi')\n")
	v := &artifactValidator{}
	v.checkPythonScript(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors, got %d", v.errors)
	}
}

func TestCheckPythonScript_invalid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.py")
	writeFile(t, p, "#!/usr/bin/env python\nprint('hi')\n")
	v := &artifactValidator{}
	v.checkPythonScript(p)
	if v.errors != 0 {
		t.Errorf("#!/usr/bin/env python should also be valid, got %d errors", v.errors)
	}
}

func TestCheckPythonScript_badShebang(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.py")
	writeFile(t, p, "#!/usr/bin/python3\n")
	v := &artifactValidator{}
	v.checkPythonScript(p)
	if v.errors == 0 {
		t.Error("expected error for non-env shebang")
	}
}

func TestCheckTSScript_valid(t *testing.T) {
	for _, shebang := range []string{
		"#!/usr/bin/env bun",
		"#!/usr/bin/env -S bun",
		"#!/usr/bin/env -S bun run",
	} {
		tmp := t.TempDir()
		p := filepath.Join(tmp, "scripts", "run.ts")
		writeFile(t, p, shebang+"\nconsole.log('hi')\n")
		v := &artifactValidator{}
		v.checkTSScript(p)
		if v.errors != 0 {
			t.Errorf("shebang %q: expected 0 errors, got %d", shebang, v.errors)
		}
	}
}

func TestCheckTSScript_badShebang(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.ts")
	writeFile(t, p, "#!/usr/bin/env node\n")
	v := &artifactValidator{}
	v.checkTSScript(p)
	if v.errors == 0 {
		t.Error("expected error for non-bun shebang in .ts file")
	}
}

func TestCheckJSScript_valid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.js")
	writeFile(t, p, "#!/usr/bin/env node\nconsole.log('hi')\n")
	v := &artifactValidator{}
	v.checkJSScript(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors, got %d", v.errors)
	}
}

func TestCheckJSScript_badShebang(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.js")
	writeFile(t, p, "#!/usr/bin/env bun\n")
	v := &artifactValidator{}
	v.checkJSScript(p)
	if v.errors == 0 {
		t.Error("expected error for non-node shebang in .js file")
	}
}

func TestCheckScriptFile_dispatchesPython(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.py")
	writeFile(t, p, "#!/usr/bin/env python3\nprint('hi')\n")
	v := &artifactValidator{}
	v.checkScriptFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for valid python via checkScriptFile, got %d", v.errors)
	}
}

func TestCheckScriptFile_dispatchesTS(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.ts")
	writeFile(t, p, "#!/usr/bin/env bun\nconsole.log('hi')\n")
	v := &artifactValidator{}
	v.checkScriptFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for valid TS via checkScriptFile, got %d", v.errors)
	}
}

func TestCheckScriptFile_dispatchesJS(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.js")
	writeFile(t, p, "#!/usr/bin/env node\nconsole.log('hi')\n")
	v := &artifactValidator{}
	v.checkScriptFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for valid JS via checkScriptFile, got %d", v.errors)
	}
}

func TestCheckScriptFile_unknownExtension(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.rb")
	writeFile(t, p, "#!/usr/bin/env ruby\n")
	v := &artifactValidator{}
	v.checkScriptFile(p)
	if v.errors == 0 {
		t.Error("expected error for unrecognised script extension")
	}
}

func TestCheckScriptFile_gitkeep(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", ".gitkeep")
	writeFile(t, p, "")
	v := &artifactValidator{}
	v.checkScriptFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for .gitkeep, got %d", v.errors)
	}
}

func TestCheckScriptFile_pycache(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "__pycache__", "mod.pyc")
	writeFile(t, p, "")
	v := &artifactValidator{}
	v.checkScriptFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for __pycache__ file, got %d", v.errors)
	}
}

// --------------------------------------------------------------------------
// artifactValidator — checkSkillDir
// --------------------------------------------------------------------------

func makeSkillDir(t *testing.T, root string, opts map[string]string) string {
	t.Helper()
	skillDir := filepath.Join(root, "skills", "domain", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for relPath, content := range opts {
		writeFile(t, filepath.Join(skillDir, relPath), content)
	}
	return skillDir
}

func TestCheckSkillDir_valid(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md": "---\nname: my-skill\n---\n# Hello\n",
	})
	v := &artifactValidator{repoRoot: tmp}
	v.checkSkillDir(skillDir)
	if v.errors != 0 {
		t.Errorf("expected 0 errors, got %d", v.errors)
	}
}

func TestCheckSkillDir_noSKILLMD(t *testing.T) {
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "skills", "domain", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	// No SKILL.md → returns early, no error.
	if v.errors != 0 {
		t.Errorf("expected 0 errors when SKILL.md absent, got %d", v.errors)
	}
}

func TestCheckSkillDir_unreadableSkillMD(t *testing.T) {
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "skills", "domain", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := filepath.Join(skillDir, "SKILL.md")
	writeFile(t, skillMD, "content")
	// Make unreadable.
	if err := os.Chmod(skillMD, 0o000); err != nil {
		t.Skip("cannot chmod on this platform")
	}
	t.Cleanup(func() { _ = os.Chmod(skillMD, 0o644) })
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors == 0 {
		t.Error("expected error for unreadable SKILL.md")
	}
}

func TestCheckSkillDir_tooManyLines(t *testing.T) {
	tmp := t.TempDir()
	lines := strings.Repeat("x\n", 501)
	skillDir := makeSkillDir(t, tmp, map[string]string{"SKILL.md": lines})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors == 0 {
		t.Error("expected error for SKILL.md exceeding 500 lines")
	}
}

func TestCheckSkillDir_nameMismatch(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md": "---\nname: wrong-name\n---\n",
	})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors == 0 {
		t.Error("expected error for frontmatter name mismatch")
	}
}

func TestCheckSkillDir_dotdotRef(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md": "---\nname: my-skill\n---\nSee ../other for details.\n",
	})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors == 0 {
		t.Error("expected error for ../ outside code block")
	}
}

func TestCheckSkillDir_dotdotInCodeBlock(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md": "---\nname: my-skill\n---\n```\ncd ../other\n```\n",
	})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for ../ inside code block, got %d", v.errors)
	}
}

func TestCheckSkillDir_nonStandardAssetsSubdir(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md":        "---\nname: my-skill\n---\n",
		"assets/misc/foo": "content",
	})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors == 0 {
		t.Error("expected error for non-standard assets/ subdirectory")
	}
}

func TestCheckSkillDir_yamlDirectlyInAssets(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md":           "---\nname: my-skill\n---\n",
		"assets/config.yaml": "key: val\n",
	})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors == 0 {
		t.Error("expected error for YAML directly in assets/")
	}
}

func TestCheckSkillDir_allowedAssetsSubdirs(t *testing.T) {
	tmp := t.TempDir()
	skillDir := makeSkillDir(t, tmp, map[string]string{
		"SKILL.md":                     "---\nname: my-skill\n---\n",
		"assets/templates/.gitkeep":    "",
		"assets/schemas/.gitkeep":      "",
		"assets/requirements/.gitkeep": "",
		"assets/examples/.gitkeep":     "",
	})
	v := &artifactValidator{}
	v.checkSkillDir(skillDir)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for all allowed assets subdirs, got %d", v.errors)
	}
}

// --------------------------------------------------------------------------
// artifactValidator — walkSkillsDir
// --------------------------------------------------------------------------

func TestWalkSkillsDir_noSkillsDir(t *testing.T) {
	tmp := t.TempDir()
	v := &artifactValidator{repoRoot: tmp}
	v.walkSkillsDir() // must not panic
	if v.errors != 0 {
		t.Errorf("expected 0 errors when skills/ absent, got %d", v.errors)
	}
}

func TestWalkSkillsDir_validSkill(t *testing.T) {
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "skills", "domain", "my-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "assets", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(skillDir, "SKILL.md"), "---\nname: my-skill\n---\n")
	writeFile(t, filepath.Join(skillDir, "assets", "schemas", "plan.schema.json"),
		`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`)
	v := &artifactValidator{repoRoot: tmp}
	v.walkSkillsDir()
	if v.errors != 0 {
		t.Errorf("expected 0 errors for valid skill, got %d", v.errors)
	}
}

// --------------------------------------------------------------------------
// checkFile dispatch
// --------------------------------------------------------------------------

func TestCheckFile_dispatchSchema(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "schemas", "bad.json") // wrong extension
	writeFile(t, p, `{}`)
	v := &artifactValidator{}
	v.checkFile(p)
	if v.errors == 0 {
		t.Error("expected error dispatched for schema file")
	}
}

func TestCheckFile_dispatchTemplate(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "assets", "templates", "empty.yaml")
	writeFile(t, p, "   ")
	v := &artifactValidator{}
	v.checkFile(p)
	if v.errors == 0 {
		t.Error("expected error dispatched for template file")
	}
}

func TestCheckFile_dispatchScript(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "scripts", "run.sh")
	writeFile(t, p, "#!/bin/sh\n") // bad shebang
	v := &artifactValidator{}
	v.checkFile(p)
	if v.errors == 0 {
		t.Error("expected error dispatched for script file")
	}
}

func TestCheckFile_noMatch(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "some", "other", "file.txt")
	writeFile(t, p, "content")
	v := &artifactValidator{}
	v.checkFile(p)
	if v.errors != 0 {
		t.Errorf("expected 0 errors for unmatched path, got %d", v.errors)
	}
}

// --------------------------------------------------------------------------
// runValidateReview
// --------------------------------------------------------------------------

// minimalValidReport builds a report that satisfies all required fields from
// the embedded review-report.requirements.json.
func minimalValidReport() string {
	return `---
review_date: 2026-04-28
reviewer: alice
skill_location: skills/domain/my-skill
---
# Skill Evaluation Report: my-skill

## Executive Summary
Some summary.

## Dimension Scores
Scores here.

## Critical Issues
None.

## Top 3 Recommended Improvements
1. Improve D1.
2. Improve D3.
3. Improve D9.

## Detailed Dimension Analysis
Details here.

## Proposed Restructured SKILL.md
Nothing to change.

D1: Knowledge Delta
D2: Mindset + Procedures
D3: Anti-Pattern Quality
D4: Specification Compliance
D5: Progressive Disclosure
D6: Freedom Calibration
D7: Pattern Recognition
D8: Practical Usability
D9: Eval Validation
`
}

func TestRunValidateReview_valid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, minimalValidReport())
	if err := runValidateReview(p, false); err != nil {
		t.Errorf("expected valid report to pass, got: %v", err)
	}
}

func TestRunValidateReview_missingFile(t *testing.T) {
	err := runValidateReview("/nonexistent/report.md", false)
	if err == nil {
		t.Error("expected error for missing report file")
	}
}

func TestRunValidateReview_missingH1(t *testing.T) {
	tmp := t.TempDir()
	report := strings.ReplaceAll(minimalValidReport(), "# Skill Evaluation Report: my-skill", "## Not an H1")
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err == nil {
		t.Error("expected error for missing H1")
	}
}

func TestRunValidateReview_wrongTitlePrefix(t *testing.T) {
	tmp := t.TempDir()
	report := strings.ReplaceAll(minimalValidReport(), "# Skill Evaluation Report: my-skill", "# Wrong Title")
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err == nil {
		t.Error("expected error for wrong title prefix")
	}
}

func TestRunValidateReview_missingFrontmatterKey(t *testing.T) {
	tmp := t.TempDir()
	report := strings.ReplaceAll(minimalValidReport(), "reviewer: alice\n", "")
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err == nil {
		t.Error("expected error for missing required frontmatter key")
	}
}

func TestRunValidateReview_missingRequiredH2(t *testing.T) {
	tmp := t.TempDir()
	report := strings.ReplaceAll(minimalValidReport(), "## Dimension Scores\nScores here.\n\n", "")
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err == nil {
		t.Error("expected error for missing required H2")
	}
}

func TestRunValidateReview_h2OrderViolation(t *testing.T) {
	tmp := t.TempDir()
	// Swap Dimension Scores and Critical Issues to create an order violation.
	report := minimalValidReport()
	report = strings.ReplaceAll(report,
		"## Dimension Scores\nScores here.\n\n## Critical Issues\nNone.\n",
		"## Critical Issues\nNone.\n\n## Dimension Scores\nScores here.\n")
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err == nil {
		t.Error("expected error for H2 order violation")
	}
}

func TestRunValidateReview_missingDimensionLabel(t *testing.T) {
	tmp := t.TempDir()
	report := strings.ReplaceAll(minimalValidReport(), "D1: Knowledge Delta\n", "")
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err == nil {
		t.Error("expected error for missing dimension label")
	}
}

func TestRunValidateReview_withRecommendedSections(t *testing.T) {
	// Include all recommended H2s and commands so those "found" branches are hit.
	tmp := t.TempDir()
	report := minimalValidReport() + `
## Conclusion
Done.

## Files Inventory
Files here.

## Verification
Verified.

## Audit Execution
Ran.

## Score Evolution
Trending up.

## Grade Scale Reference
A = 120+

skill-auditor evaluate
skill-auditor batch
./scripts/detect-duplication.sh
`
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, report)
	if err := runValidateReview(p, false); err != nil {
		t.Errorf("report with all recommended sections should pass, got: %v", err)
	}
}

func TestRunValidateReview_strictRecommended(t *testing.T) {
	// The minimal report omits recommended H2s and commands.
	// Without --strict-recommended those are warnings → should pass.
	// With --strict-recommended those become errors → should fail.
	tmp := t.TempDir()
	p := filepath.Join(tmp, "report.md")
	writeFile(t, p, minimalValidReport())

	if err := runValidateReview(p, false); err != nil {
		t.Errorf("minimal report should pass without strict mode, got: %v", err)
	}
	if err := runValidateReview(p, true); err == nil {
		t.Error("minimal report should fail with --strict-recommended (missing recommended sections)")
	}
}
