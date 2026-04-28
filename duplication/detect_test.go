package duplication

import "testing"

func TestShortKey_withSlash(t *testing.T) {
	got := ShortKey("domain/skill-name")
	if got != "skill-name" {
		t.Errorf("expected 'skill-name', got %q", got)
	}
}

func TestShortKey_noSlash(t *testing.T) {
	got := ShortKey("standalone")
	if got != "standalone" {
		t.Errorf("expected 'standalone', got %q", got)
	}
}

func TestShortKey_multipleSlashes(t *testing.T) {
	got := ShortKey("domain/sub/skill-name")
	if got != "sub/skill-name" {
		t.Errorf("expected 'sub/skill-name', got %q", got)
	}
}

func TestShortKey_empty(t *testing.T) {
	got := ShortKey("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
