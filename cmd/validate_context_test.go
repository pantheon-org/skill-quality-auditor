package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

// newValidatorForTest builds a validator rooted at the real repo so it loads the
// live schemas from .context/plugins/.../assets/schemas/.
func newValidatorForTest(t *testing.T) *contextSchemaValidator {
	t.Helper()
	root, err := resolveRepoRoot("")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	v, err := newContextSchemaValidator(root)
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}
	return v
}

func writeMD(t *testing.T, body string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "fixture.md")
	if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return f
}

func TestValidateContext_typoKeyRejected(t *testing.T) {
	v := newValidatorForTest(t)
	// `valeu` is a typo the shell validator cannot catch; additionalProperties:false must.
	f := writeMD(t, "---\ntitle: \"x\"\ntype: PLAN\nstatus: DRAFT\ndate: 2026-07-06\nvaleu: HIGH\n---\n# x\n")
	v.checkFile(f)
	if v.errors == 0 {
		t.Error("expected a typo'd top-level key to be rejected")
	}
}

func TestValidateContext_validFilePasses(t *testing.T) {
	v := newValidatorForTest(t)
	f := writeMD(t, "---\ntitle: \"x\"\ntype: FINDING\nstatus: ACTIVE\ndate: 2026-07-06\nvalue: HIGH\nthemes:\n  - GOVERNANCE\n---\n# x\n")
	v.checkFile(f)
	if v.errors != 0 {
		t.Errorf("expected valid file to pass, got %d error(s)", v.errors)
	}
}

// Regression: an unquoted ISO date must NOT be coerced to a timestamp and then
// fail the YYYY-MM-DD pattern (the bug the yamlNodeToJSON converter fixes).
func TestValidateContext_unquotedDateNotCoerced(t *testing.T) {
	v := newValidatorForTest(t)
	f := writeMD(t, "---\ntitle: \"x\"\ntype: ANALYSIS\nstatus: DONE\ndate: 2026-04-29\n---\n# x\n")
	v.checkFile(f)
	if v.errors != 0 {
		t.Errorf("unquoted date should validate as a string, got %d error(s)", v.errors)
	}
}

// A generated remediation plan carries skill_name (routing signature) and the
// generator's structured keys; it must validate against remediation-plan.schema.json.
func TestValidateContext_generatedRemediationPlanPasses(t *testing.T) {
	v := newValidatorForTest(t)
	plan, err := reporter.RemediationPlan(makeResultForValidate(84), 0, ".context/audits/my-skill/2026-04-27/Analysis.md", "2026-04-27")
	if err != nil {
		t.Fatalf("generate plan: %v", err)
	}
	f := filepath.Join(t.TempDir(), "my-skill-remediation-plan-2026-04-27.md")
	if err := os.WriteFile(f, []byte(plan), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	v.checkFile(f)
	if v.errors != 0 {
		t.Errorf("expected generated remediation plan to validate, got %d error(s)", v.errors)
	}
}

// walkMarkdown finds *.md under a directory and skips a plugins/ subdir, so the
// validator can be pointed at any context-file location, not just .context.
func TestWalkMarkdown_findsMarkdownAndSkipsPlugins(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.md"), "x")
	mustWrite(t, filepath.Join(root, "sub", "b.md"), "x")
	mustWrite(t, filepath.Join(root, "notes.txt"), "x")       // non-md ignored
	mustWrite(t, filepath.Join(root, "plugins", "p.md"), "x") // plugins skipped

	got := walkMarkdown(root)
	if len(got) != 2 {
		t.Fatalf("expected 2 markdown files (a.md, sub/b.md), got %d: %v", len(got), got)
	}
	for _, g := range got {
		if strings.Contains(g, string(os.PathSeparator)+"plugins"+string(os.PathSeparator)) {
			t.Errorf("plugins/ file should be skipped, got %s", g)
		}
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func makeResultForValidate(total int) *scorer.Result {
	dims := map[string]int{}
	for _, d := range scorer.AllDimensions {
		dims[d.Key] = d.Max * total / 140
	}
	return &scorer.Result{
		Skill:      "my-skill",
		Date:       "2026-04-27",
		Total:      total,
		MaxTotal:   140,
		Grade:      scorer.Grade(total),
		Dimensions: dims,
	}
}
