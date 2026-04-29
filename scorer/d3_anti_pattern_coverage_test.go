package scorer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

// ── preserved tests from d3_anti_pattern_test.go ────────────────────────────

func TestD3_LibraryStrongMarkers(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{StrongMarkers: 10}}
	content := "---\ndescription: x\n---\nNEVER do this. WHY: because reasons."
	score, _ := scoreD3(content, t.TempDir(), b)
	if score < 5 {
		t.Errorf("want ≥5 with strong markers, got %d", score)
	}
}

func TestD3_FallbackNEVERAndWHY(t *testing.T) {
	content := "---\ndescription: x\n---\nNEVER do this. WHY: because reasons."
	score, _ := scoreD3(content, t.TempDir(), nilBridge())
	if score < 3 {
		t.Errorf("want ≥3 via fallback, got %d", score)
	}
}

func TestD3_BADGOOD(t *testing.T) {
	content := "---\ndescription: x\n---\nBAD: do this GOOD: do that instead. WHY: reasons."
	score, _ := scoreD3(content, t.TempDir(), nilBridge())
	if score < 4 {
		t.Errorf("want ≥4 (WHY: + BAD.*GOOD), got %d", score)
	}
}

func TestD3_AntiPatternInstructionsBonus(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	instrJSON := `{"instructions":[
		{"original_snippets":"NEVER do this","content":"x"},
		{"original_snippets":"ALWAYS validate","content":"x"},
		{"original_snippets":"avoid this pattern","content":"x"},
		{"original_snippets":"do not use","content":"x"},
		{"original_snippets":"anti-pattern here","content":"x"}
	]}`
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte(instrJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill"
	score, diags := scoreD3(content, tmpDir, nilBridge())
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
	if score < 2 {
		t.Errorf("want ≥2 (≥5 anti-pattern instructions bonus), got %d", score)
	}
}

func TestD3_FallbackMultipleNEVER(t *testing.T) {
	// >3 NEVER → +3 via fallback
	content := "---\ndescription: x\n---\nNEVER a. NEVER b. NEVER c. NEVER d."
	score, _ := scoreD3(content, t.TempDir(), nilBridge())
	if score < 3 {
		t.Errorf("want ≥3 for >3 NEVER via fallback, got %d", score)
	}
}

func TestD3_AntiPatternInstructionsThreeBonus(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	instrJSON := `{"instructions":[
		{"original_snippets":"NEVER do this","content":"x"},
		{"original_snippets":"ALWAYS validate","content":"x"},
		{"original_snippets":"avoid this pattern","content":"x"}
	]}`
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte(instrJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill"
	score, _ := scoreD3(content, tmpDir, nilBridge())
	if score < 1 {
		t.Errorf("want ≥1 for 3 anti-pattern instructions, got %d", score)
	}
}

func TestD3_AntiPatternInstructionsArraySnippets(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	instrJSON := `{"instructions":[
		{"type":"anti-pattern","original_snippets":["NEVER do this","also bad"],"content":"x"},
		{"type":"anti-pattern","original_snippets":["ALWAYS validate"],"content":"x"},
		{"type":"anti-pattern","original_snippets":["avoid this"],"content":"x"},
		{"type":"anti-pattern","original_snippets":["do not use"],"content":"x"},
		{"type":"anti-pattern","original_snippets":["bad pattern"],"content":"x"}
	]}`
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte(instrJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill"
	score, diags := scoreD3(content, tmpDir, nilBridge())
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
	if score < 2 {
		t.Errorf("want ≥2 for ≥5 anti-pattern instructions (array snippets), got %d", score)
	}
}

func TestD3_DirectiveLanguage_NilBridgeNoNEVER(t *testing.T) {
	delta := scoreD3DirectiveLanguage("---\ndescription: x\n---\n# Skill\nNo directives here.", nilBridge())
	if delta != 0 {
		t.Errorf("want 0 with no NEVER and nil bridge, got %d", delta)
	}
}

func TestD3_InstructionsParseError(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evalsDir, "instructions.json"), []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: x\n---\n# Skill"
	_, diags := scoreD3(content, tmpDir, nilBridge())
	if len(diags) != 1 || diags[0].severity != "error" || diags[0].Dimension != "D3" {
		t.Errorf("expected D3 error diagnostic, got %v", diags)
	}
}

// ── new TDD tests for per-block SYMPTOM/CONSEQUENCE detection ─────────────────

// full6ComponentBlock has all six components: NEVER, WHY, SYMPTOM, CONSEQUENCE, BAD, GOOD.
const full6ComponentBlock = "---\ndescription: test skill\n---\n\n" +
	"**NEVER** hardcode credentials in source files.\n\n" +
	"**WHY:** Hardcoded secrets are exposed in version control history and logs.\n\n" +
	"**SYMPTOM:**\n" +
	"You see plaintext passwords or API keys in the source tree or git log.\n\n" +
	"**CONSEQUENCE:**\n" +
	"Credential leaks cause account takeovers and compliance violations.\n\n" +
	"**BAD:**\n" +
	"```python\nDB_PASS = \"hunter2\"\n```\n\n" +
	"**GOOD:**\n" +
	"```python\nDB_PASS = os.environ[\"DB_PASS\"]\n```\n"

// four4ComponentBlock has NEVER/WHY/BAD/GOOD only — no SYMPTOM/CONSEQUENCE.
const four4ComponentBlock = "---\ndescription: test skill\n---\n\n" +
	"**NEVER** use eval() with user input.\n\n" +
	"**WHY:** eval() executes arbitrary code, leading to remote code execution.\n\n" +
	"**BAD:**\n" +
	"```js\neval(userInput)\n```\n\n" +
	"**GOOD:**\n" +
	"```js\nJSON.parse(userInput)\n```\n"

// symptomInWHYProse has "symptom" mentioned only in WHY paragraph prose (not a bold header).
const symptomInWHYProse = "---\ndescription: test skill\n---\n\n" +
	"**NEVER** skip input validation.\n\n" +
	"**WHY:** Without validation the symptom of silent corruption is common. Root cause is missing guards.\n\n" +
	"**BAD:**\nAccepting raw form data directly.\n\n" +
	"**GOOD:**\nValidate and sanitise all inputs.\n"

// consequenceInBADBlock has "consequence" mentioned only inside BAD example prose.
const consequenceInBADBlock = "---\ndescription: test skill\n---\n\n" +
	"**NEVER** ignore error returns.\n\n" +
	"**WHY:** Errors carry critical context.\n\n" +
	"**BAD:**\nIgnoring errors has the consequence of silent failure in production.\n\n" +
	"**GOOD:**\nAlways handle or propagate errors explicitly.\n"

// threeValidBlocks has three NEVER-anchored blocks, each with NEVER/WHY/BAD/GOOD.
const threeValidBlocks = "---\ndescription: test skill\n---\n\n" +
	"**NEVER** hardcode secrets.\n\n**WHY:** Version control exposure.\n\n**BAD:** Plain text secrets.\n\n**GOOD:** Use env vars.\n\n" +
	"**NEVER** use eval() with input.\n\n**WHY:** Remote code execution risk.\n\n**BAD:** eval(userInput)\n\n**GOOD:** JSON.parse(userInput)\n\n" +
	"**NEVER** ignore error returns.\n\n**WHY:** Silent failure in production.\n\n**BAD:** Swallow errors silently.\n\n**GOOD:** Handle and propagate errors.\n"

// TestD3_PerBlock_Full6Components verifies a block with all six components scores ≥ 13.
// Uses a library bridge with high StrongMarkers (10 > d3StrongMarkersHigh=8) → directive = 5.
// Expected: directive(5) + BAD/GOOD(2) + WHY(2) + SYMPTOM(2) + CONSEQUENCE(2) = 13.
func TestD3_PerBlock_Full6Components(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{StrongMarkers: 10}}
	score, _ := scoreD3(full6ComponentBlock, t.TempDir(), b)
	if score < 13 {
		t.Errorf("full 6-component block: want score ≥13, got %d", score)
	}
}

// TestD3_PerBlock_4Components verifies NEVER/WHY/BAD/GOOD only scores ≤ 10.
func TestD3_PerBlock_4Components(t *testing.T) {
	score, _ := scoreD3(four4ComponentBlock, t.TempDir(), nilBridge())
	if score > 10 {
		t.Errorf("4-component block (no SYMPTOM/CONSEQUENCE): want score ≤10, got %d", score)
	}
}

// TestD3_PerBlock_SYMPTOMInWHYProse verifies SYMPTOM in WHY prose does not score.
func TestD3_PerBlock_SYMPTOMInWHYProse(t *testing.T) {
	blocks := parseAntiPatternBlocks(symptomInWHYProse)
	if len(blocks) == 0 {
		t.Fatal("expected at least one anti-pattern block, got none")
	}
	s := scoreSymptom(blocks[0])
	if s != 0 {
		t.Errorf("SYMPTOM in WHY prose should not score; want 0, got %d", s)
	}
}

// TestD3_PerBlock_CONSEQUENCEInBADBlock verifies CONSEQUENCE in BAD prose does not score.
func TestD3_PerBlock_CONSEQUENCEInBADBlock(t *testing.T) {
	blocks := parseAntiPatternBlocks(consequenceInBADBlock)
	if len(blocks) == 0 {
		t.Fatal("expected at least one anti-pattern block, got none")
	}
	c := scoreConsequence(blocks[0])
	if c != 0 {
		t.Errorf("CONSEQUENCE in BAD prose should not score; want 0, got %d", c)
	}
}

// TestD3_PerBlock_ParseBlocks verifies that three NEVER markers produce three blocks.
func TestD3_PerBlock_ParseBlocks(t *testing.T) {
	blocks := parseAntiPatternBlocks(threeValidBlocks)
	if len(blocks) != 3 {
		t.Errorf("want 3 anti-pattern blocks, got %d", len(blocks))
	}
}

// TestD3_PerBlock_CountBonus verifies that ≥3 valid blocks earn the count bonus.
func TestD3_PerBlock_CountBonus(t *testing.T) {
	score, _ := scoreD3(threeValidBlocks, t.TempDir(), nilBridge())
	if score < 5 {
		t.Errorf("≥3 valid blocks: want score ≥5 (count bonus applies), got %d", score)
	}
}

// TestParseAntiPatternBlocks_Empty verifies empty content returns no blocks.
func TestParseAntiPatternBlocks_Empty(t *testing.T) {
	blocks := parseAntiPatternBlocks("")
	if len(blocks) != 0 {
		t.Errorf("empty content: want 0 blocks, got %d", len(blocks))
	}
}

// TestParseAntiPatternBlocks_NoNEVER verifies a document with no NEVER markers returns no blocks.
func TestParseAntiPatternBlocks_NoNEVER(t *testing.T) {
	content := "---\ndescription: x\n---\n# Just a regular section\nSome content here."
	blocks := parseAntiPatternBlocks(content)
	if len(blocks) != 0 {
		t.Errorf("no NEVER markers: want 0 blocks, got %d", len(blocks))
	}
}

// TestScoreSymptom_BoldHeader verifies a bold **SYMPTOM:** header with body scores 1.
func TestScoreSymptom_BoldHeader(t *testing.T) {
	block := "**NEVER** do X.\n**WHY:** Because.\n**SYMPTOM:**\nYou see error Y in logs.\n**BAD:**\nfoo\n**GOOD:**\nbar"
	s := scoreSymptom(block)
	if s != 1 {
		t.Errorf("bold SYMPTOM header with body: want 1, got %d", s)
	}
}

// TestScoreSymptom_EmptyHeader verifies an empty **SYMPTOM:** header scores 0.
func TestScoreSymptom_EmptyHeader(t *testing.T) {
	block := "**NEVER** do X.\n**WHY:** Because.\n**SYMPTOM:**\n**BAD:**\nfoo\n**GOOD:**\nbar"
	s := scoreSymptom(block)
	if s != 0 {
		t.Errorf("empty SYMPTOM header (no body): want 0, got %d", s)
	}
}

// TestScoreConsequence_BoldHeader verifies a bold **CONSEQUENCE:** header with body scores 1.
func TestScoreConsequence_BoldHeader(t *testing.T) {
	block := "**NEVER** do X.\n**WHY:** Because.\n**CONSEQUENCE:**\nSystem crashes under load.\n**BAD:**\nfoo\n**GOOD:**\nbar"
	c := scoreConsequence(block)
	if c != 1 {
		t.Errorf("bold CONSEQUENCE header with body: want 1, got %d", c)
	}
}

// TestScoreConsequence_EmptyHeader verifies an empty **CONSEQUENCE:** header scores 0.
func TestScoreConsequence_EmptyHeader(t *testing.T) {
	block := "**NEVER** do X.\n**WHY:** Because.\n**CONSEQUENCE:**\n**BAD:**\nfoo\n**GOOD:**\nbar"
	c := scoreConsequence(block)
	if c != 0 {
		t.Errorf("empty CONSEQUENCE header (no body): want 0, got %d", c)
	}
}
