package scorer

import (
	"strings"
	"testing"
)

func TestMatchesRegexCI_valid(t *testing.T) {
	if !matchesRegexCI("Hello World", `(?i)hello`) {
		t.Error("expected case-insensitive match")
	}
	if matchesRegexCI("Hello World", `(?i)goodbye`) {
		t.Error("expected no match")
	}
}

func TestMatchesRegexCI_invalidPattern(t *testing.T) {
	// Invalid regex should return false without panicking.
	if matchesRegexCI("some content", `[invalid(regex`) {
		t.Error("invalid regex should return false")
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
