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

func TestDetect_truncatesLargeCorpus(t *testing.T) {
	entries := make([]SkillEntry, MaxDetectEntries+10)
	for i := range entries {
		entries[i] = SkillEntry{Key: "a/skill", Content: "unique content alpha beta gamma delta epsilon"}
	}
	// Should not panic and must not process more than MaxDetectEntries entries.
	// We verify indirectly: if the cap were absent, the identical content would
	// produce O(MaxDetectEntries+10)² pairs — with the cap it produces exactly
	// MaxDetectEntries*(MaxDetectEntries-1)/2 at most (all identical, all Critical).
	pairs := Detect(entries)
	maxPairs := MaxDetectEntries * (MaxDetectEntries - 1) / 2
	if len(pairs) > maxPairs {
		t.Errorf("Detect returned %d pairs but corpus was capped at %d entries (max %d pairs)", len(pairs), MaxDetectEntries, maxPairs)
	}
}
