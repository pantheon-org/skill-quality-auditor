package reporter

import (
	"os"
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

func makeResult() *scorer.Result {
	return &scorer.Result{
		Skill:    "agentic-harness/skill-quality-auditor",
		Total:    122,
		MaxTotal: 140,
		Grade:    "B+",
		Dimensions: map[string]int{
			"knowledgeDelta":          14,
			"mindsetProcedures":       12,
			"antiPatternQuality":      9,
			"specificationCompliance": 13,
			"progressiveDisclosure":   15,
			"freedomCalibration":      12,
			"patternRecognition":      9,
			"practicalUsability":      13,
			"evalValidation":          15,
		},
		Errors:   0,
		Warnings: 0,
	}
}

func TestFormat_withDiagnostics(t *testing.T) {
	r := makeResult()
	r.Errors = 1
	r.Warnings = 1
	r.ErrorDetails = []scorer.Diagnostic{
		{Dimension: "D1", Message: "missing expert keywords"},
	}
	r.WarningDetails = []scorer.Diagnostic{
		{Dimension: "D4", Message: "harness path found"},
	}
	out := Format(r)
	if !strings.Contains(out, "Diagnostics:") {
		t.Error("Format should include Diagnostics section when errors/warnings present")
	}
	if !strings.Contains(out, "[E]") {
		t.Error("errors should appear with [E] prefix")
	}
	if !strings.Contains(out, "[W]") {
		t.Error("warnings should appear with [W] prefix")
	}
	if !strings.Contains(out, "missing expert keywords") {
		t.Error("error message should appear in output")
	}
}

func TestFormat_missingDimension(t *testing.T) {
	r := makeResult()
	delete(r.Dimensions, "evalValidation")
	out := Format(r)
	// should still format without panicking
	if !strings.Contains(out, "Knowledge Delta") {
		t.Error("other dimensions should still appear after missing key")
	}
}

func TestStore_writesAnalysisAndRemediation(t *testing.T) {
	r := makeResult()
	r.Date = "2026-04-28"
	root := t.TempDir()
	if err := Store(root, "domain/skill", r); err != nil {
		t.Fatalf("Store returned error: %v", err)
	}
	for _, name := range []string{"audit.json", "Analysis.md", "Remediation.md"} {
		path := root + "/.context/audits/domain/skill/2026-04-28/" + name
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestFormatBasic(t *testing.T) {
	r := makeResult()
	out := Format(r)

	checks := []string{
		"Skill: agentic-harness/skill-quality-auditor",
		"Grade: B+ (122/140)",
		"Knowledge Delta",
		"Mindset + Procedures",
		"Anti-Pattern Quality",
		"Specification Compliance",
		"Progressive Disclosure",
		"Freedom Calibration",
		"Pattern Recognition",
		"Practical Usability",
		"Eval Validation",
		"Errors: 0  Warnings: 0",
	}

	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("Format output missing %q\ngot:\n%s", want, out)
		}
	}
}

// ---- Analysis ----

func TestAnalysis_noErrors(t *testing.T) {
	r := makeResult()
	out := Analysis(r)
	if !strings.Contains(out, "# Skill Audit") {
		t.Error("Analysis should contain heading")
	}
	if !strings.Contains(out, "No errors or warnings.") {
		t.Errorf("clean result should show 'No errors or warnings.', got:\n%s", out)
	}
	if strings.Contains(out, "### Errors") {
		t.Error("clean result should not have Errors section")
	}
}

func TestAnalysis_withErrors(t *testing.T) {
	r := makeResult()
	r.ErrorDetails = []scorer.Diagnostic{
		{Dimension: "D1", Message: "missing expert keywords"},
	}
	out := Analysis(r)
	if !strings.Contains(out, "### Errors") {
		t.Error("result with errors should have Errors section")
	}
	if !strings.Contains(out, "missing expert keywords") {
		t.Error("error message should appear in output")
	}
	if strings.Contains(out, "No errors or warnings.") {
		t.Error("should not show 'No errors or warnings' when errors exist")
	}
}

func TestAnalysis_withWarnings(t *testing.T) {
	r := makeResult()
	r.WarningDetails = []scorer.Diagnostic{
		{Dimension: "D4", Message: "harness path found"},
	}
	out := Analysis(r)
	if !strings.Contains(out, "### Warnings") {
		t.Error("result with warnings should have Warnings section")
	}
	if !strings.Contains(out, "harness path found") {
		t.Error("warning message should appear in output")
	}
}

func TestAnalysis_missingDimensionKey(t *testing.T) {
	r := makeResult()
	// Remove one dimension — the loop should skip it without panicking.
	delete(r.Dimensions, "evalValidation")
	out := Analysis(r)
	if !strings.Contains(out, "Knowledge Delta") {
		t.Error("other dimensions should still appear")
	}
}

func TestAnalysis_containsDimensionTable(t *testing.T) {
	r := makeResult()
	out := Analysis(r)
	if !strings.Contains(out, "| Dimension | Score | Max |") {
		t.Error("Analysis should include dimension score table")
	}
}

// ---- Remediation ----

func TestRemediation_allAtMax(t *testing.T) {
	r := makeResult()
	for _, d := range scorer.AllDimensions {
		r.Dimensions[d.Key] = d.Max
	}
	out := Remediation(r)
	if !strings.Contains(out, "Nothing to remediate") {
		t.Errorf("perfect score should produce 'Nothing to remediate', got:\n%s", out)
	}
}

func TestRemediation_plural_one(t *testing.T) {
	if plural(1) != "" {
		t.Error("plural(1) should return empty string")
	}
}

func TestRemediation_plural_many(t *testing.T) {
	for _, n := range []int{0, 2, 10, -1} {
		if plural(n) != "s" {
			t.Errorf("plural(%d) should return 's'", n)
		}
	}
}

func TestRemediation_withDiagnostics(t *testing.T) {
	r := makeResult()
	r.Dimensions["knowledgeDelta"] = 5 // large gap
	r.ErrorDetails = []scorer.Diagnostic{
		{Dimension: "D1", Message: "missing expert keywords"},
	}
	out := Remediation(r)
	if !strings.Contains(out, "missing expert keywords") {
		t.Error("remediation should include diagnostic messages")
	}
}

func TestRemediation_warningDiagnostics(t *testing.T) {
	r := makeResult()
	r.Dimensions["evalValidation"] = 5 // gap
	r.WarningDetails = []scorer.Diagnostic{
		scorer.NewWarnDiag("D9", "low coverage warning"),
	}
	out := Remediation(r)
	if !strings.Contains(out, "low coverage warning") {
		t.Errorf("remediation should include warning diagnostic messages, got:\n%s", out)
	}
	if !strings.Contains(out, "⚠️") {
		t.Errorf("expected ⚠️ icon for warning-severity diagnostic, got:\n%s", out)
	}
}

func TestRemediation_errorSeverityIcon(t *testing.T) {
	r := makeResult()
	r.Dimensions["knowledgeDelta"] = 5
	r.ErrorDetails = []scorer.Diagnostic{
		scorer.NewErrorDiag("D1", "critical error here"),
	}
	out := Remediation(r)
	if !strings.Contains(out, "🔴") {
		t.Errorf("expected 🔴 icon for error-severity diagnostic, got:\n%s", out)
	}
}

func TestFormatDimensionOrder(t *testing.T) {
	r := makeResult()
	out := Format(r)

	labels := []string{
		"Knowledge Delta",
		"Mindset + Procedures",
		"Anti-Pattern Quality",
		"Specification Compliance",
		"Progressive Disclosure",
		"Freedom Calibration",
		"Pattern Recognition",
		"Practical Usability",
		"Eval Validation",
	}

	pos := -1
	for _, label := range labels {
		idx := strings.Index(out, label)
		if idx < 0 {
			t.Errorf("dimension %q not found in output", label)
			continue
		}
		if idx <= pos {
			t.Errorf("dimension %q appears before previous (pos %d <= %d)", label, idx, pos)
		}
		pos = idx
	}
}
