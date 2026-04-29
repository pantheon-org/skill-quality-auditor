package scorer

import (
	"strings"
	"testing"
)

func TestReD2MindsetHeader(t *testing.T) {
	if !reD2MindsetHeader.MatchString("## Mindset\n") {
		t.Error("expected match on ## Mindset")
	}
	if !reD2MindsetHeader.MatchString("## Philosophy\n") {
		t.Error("expected match on ## Philosophy")
	}
	if reD2MindsetHeader.MatchString("## Usage\n") {
		t.Error("expected no match on ## Usage")
	}
}

func TestReD3BadGood(t *testing.T) {
	if !reD3BadGood.MatchString("BAD: do this GOOD: do that") {
		t.Error("expected BAD/GOOD match")
	}
	if reD3BadGood.MatchString("no contrast here") {
		t.Error("expected no match without BAD...GOOD")
	}
}

func TestExtractFrontmatterField_present(t *testing.T) {
	content := "---\ndescription: my description\nauthor: alice\n---\n# Body"
	got := extractFrontmatterField(content, "description")
	if got != "my description" {
		t.Errorf("got %q, want 'my description'", got)
	}
}

func TestExtractFrontmatterField_quoted(t *testing.T) {
	content := "---\ndescription: \"quoted value\"\n---\n# Body"
	got := extractFrontmatterField(content, "description")
	if got != "quoted value" {
		t.Errorf("expected quotes stripped, got %q", got)
	}
}

func TestExtractFrontmatterField_missing(t *testing.T) {
	content := "---\nauthor: alice\n---\n# Body"
	got := extractFrontmatterField(content, "description")
	if got != "" {
		t.Errorf("expected empty string for missing field, got %q", got)
	}
}

func TestExtractFrontmatterField_noFrontmatter(t *testing.T) {
	content := "# No frontmatter here\nJust content."
	got := extractFrontmatterField(content, "description")
	if got != "" {
		t.Errorf("expected empty string with no frontmatter, got %q", got)
	}
}

func TestParseFrontmatter_multiLineValue(t *testing.T) {
	content := "---\ndescription: |\n  multi\n  line\nname: my-skill\n---\n# Body"
	fm := parseFrontmatter(content)
	if fm.Name != "my-skill" {
		t.Errorf("got name %q, want 'my-skill'", fm.Name)
	}
}

func TestParseFrontmatter_nameField(t *testing.T) {
	content := "---\nname: cool-skill\ndescription: does things\n---\n# Body"
	fm := parseFrontmatter(content)
	if fm.Name != "cool-skill" {
		t.Errorf("got name %q, want 'cool-skill'", fm.Name)
	}
	if fm.Description != "does things" {
		t.Errorf("got description %q, want 'does things'", fm.Description)
	}
}

func TestRemoveCodeBlocks_languageTagged(t *testing.T) {
	content := "before\n```bash\ncode here\n```\nafter"
	got := removeCodeBlocks(content)
	if strings.Contains(got, "code here") {
		t.Error("removeCodeBlocks should strip content from language-tagged blocks")
	}
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Error("removeCodeBlocks should preserve text outside code blocks")
	}
}
