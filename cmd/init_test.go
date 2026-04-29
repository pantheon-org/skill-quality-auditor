package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- resolveTargets ----

func TestResolveTargets_specificKnownID(t *testing.T) {
	agents, err := resolveTargets([]string{"claude-code"}, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 1 || agents[0].ID != "claude-code" {
		t.Errorf("expected [claude-code], got %v", agents)
	}
}

func TestResolveTargets_unknownID(t *testing.T) {
	_, err := resolveTargets([]string{"does-not-exist-xyz"}, t.TempDir(), false)
	if err == nil {
		t.Error("expected error for unknown agent ID")
	}
}

func TestResolveTargets_multipleIDs(t *testing.T) {
	agents, err := resolveTargets([]string{"claude-code", "cursor"}, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestResolveTargets_autoDetect_nonePresent(t *testing.T) {
	agents, err := resolveTargets(nil, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("expected 0 agents for empty home dir, got %d", len(agents))
	}
}

func TestResolveTargets_autoDetect_detected(t *testing.T) {
	home := t.TempDir()
	// claude-code GlobalPath is ".claude/skills"
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	agents, err := resolveTargets(nil, home, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, a := range agents {
		if a.ID == "claude-code" {
			found = true
		}
	}
	if !found {
		t.Error("expected claude-code to be auto-detected")
	}
}

// ---- writeCanonical ----

func TestWriteCanonical(t *testing.T) {
	home := t.TempDir()
	dest, err := writeCanonical(home)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("canonical SKILL.md not created at %s: %v", dest, err)
	}
	expected := filepath.Join(home, ".local", "share", skillName, "SKILL.md")
	if dest != expected {
		t.Errorf("got path %q, want %q", dest, expected)
	}
}

func TestWriteCanonical_alsoWritesRefs(t *testing.T) {
	home := t.TempDir()
	if _, err := writeCanonical(home); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	refsDir := filepath.Join(home, ".local", "share", skillName, "references")
	if _, err := os.Stat(refsDir); err != nil {
		t.Errorf("references dir not created: %v", err)
	}
}

// ---- initCmd RunE error paths ----

func TestInitCmd_invalidMethod(t *testing.T) {
	origMethod := initMethod
	initMethod = "invalid"
	defer func() { initMethod = origMethod }()

	err := initCmd.RunE(initCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "--method must be") {
		t.Errorf("expected method validation error, got: %v", err)
	}
}

func TestInitCmd_unknownAgent(t *testing.T) {
	origMethod, origAgents := initMethod, initAgents
	initMethod = "copy"
	initAgents = []string{"unknown-agent-xyz"}
	defer func() { initMethod = origMethod; initAgents = origAgents }()

	err := initCmd.RunE(initCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("expected unknown agent error, got: %v", err)
	}
}

// ---- writeRefs ----

func TestWriteRefs(t *testing.T) {
	dest := t.TempDir()
	if err := writeRefs(dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	refsDir := filepath.Join(dest, "references")
	entries, err := os.ReadDir(refsDir)
	if err != nil {
		t.Fatalf("references dir not created: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one reference file to be written")
	}
}
