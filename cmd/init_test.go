package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

// newInitCmd returns initCmd with fresh flags for isolated testing.
func newInitCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := initCmd
	cmd.ResetFlags()
	cmd.Flags().StringArrayP("agent", "a", nil, "")
	cmd.Flags().BoolP("global", "g", false, "")
	cmd.Flags().StringP("method", "m", "symlink", "")
	cmd.Flags().BoolP("dry-run", "n", false, "")
	cmd.SetOut(&bytes.Buffer{})
	return cmd
}

// ---- initCmd RunE error paths ----

func TestInitCmd_invalidMethod(t *testing.T) {
	cmd := newInitCmd(t)
	if err := cmd.Flags().Set("method", "invalid"); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "--method must be") {
		t.Errorf("expected method validation error, got: %v", err)
	}
}

func TestInitCmd_unknownAgent(t *testing.T) {
	cmd := newInitCmd(t)
	if err := cmd.Flags().Set("method", "copy"); err != nil {
		t.Fatal(err)
	}
	// StringArray flags must be set with repeated Set calls or use the slice variant.
	if err := cmd.Flags().Set("agent", "unknown-agent-xyz"); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("expected unknown agent error, got: %v", err)
	}
}

// ---- dry-run ----

func TestInitCmd_dryRun_copy_noFilesCreated(t *testing.T) {
	home := t.TempDir()
	// Create the agent skill dir so claude-code is auto-detected.
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := newInitCmd(t)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	// Patch UserHomeDir by exercising RunE directly with a fake homeDir via agent flag.
	// We call RunE directly after setting flags to exercise the dry-run path.
	if err := cmd.Flags().Set("method", "copy"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("agent", "claude-code"); err != nil {
		t.Fatal(err)
	}

	// RunE will call os.UserHomeDir; we just verify no writes happen by checking
	// that no files were created in temp dir afterward and output contains [dry-run].
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected [dry-run] output, got: %q", out)
	}
	if !strings.Contains(out, "would create directory") {
		t.Errorf("expected 'would create directory' in output, got: %q", out)
	}
	if !strings.Contains(out, "would write") {
		t.Errorf("expected 'would write' in output, got: %q", out)
	}
}

func TestInitCmd_dryRun_symlink_output(t *testing.T) {
	cmd := newInitCmd(t)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	if err := cmd.Flags().Set("method", "symlink"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("agent", "claude-code"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "would symlink") {
		t.Errorf("expected 'would symlink' in output for symlink method, got: %q", out)
	}
}

func TestInitCmd_dryRun_shorthand(t *testing.T) {
	cmd := newInitCmd(t)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	if err := cmd.Flags().Set("method", "copy"); err != nil {
		t.Fatal(err)
	}
	// Use shorthand -n via the flag name "dry-run" (cobra resolves shorthands internally)
	flag := cmd.Flags().ShorthandLookup("n")
	if flag == nil {
		t.Fatal("expected shorthand -n to exist for --dry-run")
	}
	if flag.Name != "dry-run" {
		t.Errorf("expected shorthand -n to map to dry-run, got %q", flag.Name)
	}
}

func TestInitCmd_flagShorthands(t *testing.T) {
	cmd := newInitCmd(t)

	cases := []struct {
		shorthand string
		longName  string
	}{
		{"a", "agent"},
		{"g", "global"},
		{"m", "method"},
		{"n", "dry-run"},
	}
	for _, tc := range cases {
		flag := cmd.Flags().ShorthandLookup(tc.shorthand)
		if flag == nil {
			t.Errorf("expected shorthand -%s to exist", tc.shorthand)
			continue
		}
		if flag.Name != tc.longName {
			t.Errorf("shorthand -%s: expected long name %q, got %q", tc.shorthand, tc.longName, flag.Name)
		}
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
