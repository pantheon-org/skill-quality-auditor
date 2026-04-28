package duplication

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInventory_findsSkillFiles(t *testing.T) {
	dir := t.TempDir()
	writeSkill(t, dir, "domain-a/skill-one/SKILL.md", "## Anti-patterns\nNEVER do this")
	writeSkill(t, dir, "domain-a/skill-two/SKILL.md", "## Overview\nSome content here")
	writeSkill(t, dir, "domain-b/skill-three/SKILL.md", "## Best practices\nAlways do this")

	entries, err := Inventory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestInventory_setsKeyAndContent(t *testing.T) {
	dir := t.TempDir()
	writeSkill(t, dir, "tools/my-tool/SKILL.md", "some content")

	entries, err := Inventory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Key != "tools/my-tool" {
		t.Errorf("expected key 'tools/my-tool', got %q", e.Key)
	}
	if e.Content != "some content" {
		t.Errorf("expected content 'some content', got %q", e.Content)
	}
	if e.Path == "" {
		t.Error("path should not be empty")
	}
}

func TestInventory_emptyDir(t *testing.T) {
	dir := t.TempDir()
	entries, err := Inventory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestInventory_ignoresNonSkillFiles(t *testing.T) {
	dir := t.TempDir()
	writeSkill(t, dir, "tool/foo/SKILL.md", "real skill")
	writeSkill(t, dir, "tool/foo/README.md", "not a skill")

	entries, err := Inventory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestInventory_nonexistentDir(t *testing.T) {
	_, err := Inventory("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func writeSkill(t *testing.T, base, rel, content string) {
	t.Helper()
	full := filepath.Join(base, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
