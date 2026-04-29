package scorer

import (
	"strings"
	"testing"
)

func TestCountLines(t *testing.T) {
	content := "line one\n\nline three\nline four\n"
	if n := countLines(content); n != 3 {
		t.Errorf("want 3 non-empty lines, got %d", n)
	}
}

func TestErrDiag(t *testing.T) {
	d := errDiag("D1", "something failed")
	if d.Dimension != "D1" || d.Message != "something failed" || d.severity != "error" {
		t.Errorf("unexpected errDiag: %+v", d)
	}
}

func TestWarnDiag(t *testing.T) {
	d := warnDiag("D4", "missing ref")
	if d.Dimension != "D4" || d.Message != "missing ref" || d.severity != "warning" {
		t.Errorf("unexpected warnDiag: %+v", d)
	}
}

func TestCountPattern(t *testing.T) {
	if n := countPattern("NEVER do this. Never again.", "never"); n != 2 {
		t.Errorf("want 2, got %d", n)
	}
}

func TestNewErrorDiag(t *testing.T) {
	d := NewErrorDiag("D2", "missing section")
	if d.Dimension != "D2" || d.Message != "missing section" || d.Severity() != "error" {
		t.Errorf("unexpected NewErrorDiag: %+v", d)
	}
}

func TestNewWarnDiag(t *testing.T) {
	d := NewWarnDiag("D5", "low score")
	if d.Dimension != "D5" || d.Message != "low score" || d.Severity() != "warning" {
		t.Errorf("unexpected NewWarnDiag: %+v", d)
	}
}

func TestAllDimensions_count(t *testing.T) {
	if len(AllDimensions) != 9 {
		t.Errorf("expected 9 dimensions, got %d", len(AllDimensions))
	}
}

func TestAllDimensions_uniqueFields(t *testing.T) {
	codes := map[string]bool{}
	keys := map[string]bool{}
	for _, d := range AllDimensions {
		if codes[d.Code] {
			t.Errorf("duplicate Code: %s", d.Code)
		}
		if keys[d.Key] {
			t.Errorf("duplicate Key: %s", d.Key)
		}
		codes[d.Code] = true
		keys[d.Key] = true
		if d.Max <= 0 {
			t.Errorf("dimension %s has non-positive Max: %d", d.Code, d.Max)
		}
	}
}

func TestAllDimensions_totalMax(t *testing.T) {
	total := 0
	for _, d := range AllDimensions {
		total += d.Max
	}
	if total != 140 {
		t.Errorf("expected AllDimensions total Max=140, got %d", total)
	}
}

func TestBuildDimensionMap_keysMatchAllDimensions(t *testing.T) {
	scores := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m := buildDimensionMap(scores)
	if len(m) != len(AllDimensions) {
		t.Errorf("expected %d keys, got %d", len(AllDimensions), len(m))
	}
	for i, d := range AllDimensions {
		if m[d.Key] != i+1 {
			t.Errorf("dimension %s: want %d, got %d", d.Key, i+1, m[d.Key])
		}
	}
}

func TestRemoveCodeBlocks(t *testing.T) {
	content := "before\n```\ncode line\n```\nafter\n"
	result := removeCodeBlocks(content)
	if strings.Contains(result, "code line") {
		t.Error("removeCodeBlocks should strip code block contents")
	}
	if !strings.Contains(result, "before") || !strings.Contains(result, "after") {
		t.Error("removeCodeBlocks should preserve non-code content")
	}
}
