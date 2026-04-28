package cmd

import (
	"bytes"
	"path/filepath"
	"testing"
)

// ---- agentByID ----

func TestAgentByID_found(t *testing.T) {
	a, ok := agentByID("claude-code")
	if !ok {
		t.Fatal("expected claude-code to be found")
	}
	if a.ID != "claude-code" {
		t.Errorf("got ID %q, want claude-code", a.ID)
	}
}

func TestAgentByID_notFound(t *testing.T) {
	_, ok := agentByID("unknown-agent-xyz")
	if ok {
		t.Error("expected not-found for unknown agent")
	}
}

// ---- batch command ----

const fixturesBase = "../testdata/fixtures"

func runBatch(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"batch"}, args...))
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	return buf.String(), err
}

func TestBatchCmd_singleSkill(t *testing.T) {
	skillPath := filepath.Join(fixturesBase, "skill-full")
	_, err := runBatch(t, skillPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBatchCmd_multipleSkills(t *testing.T) {
	full := filepath.Join(fixturesBase, "skill-full")
	minimal := filepath.Join(fixturesBase, "skill-minimal")
	_, err := runBatch(t, full, minimal)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBatchCmd_nonexistentSkill(t *testing.T) {
	_, err := runBatch(t, "/nonexistent/SKILL.md")
	// Error should not propagate as a command error — it prints ERROR line instead
	_ = err
}

func TestBatchCmd_jsonFlag(t *testing.T) {
	skillPath := filepath.Join(fixturesBase, "skill-full")
	out, err := runBatch(t, "--json", skillPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) > 0 && out[0] != '[' {
		t.Errorf("expected JSON array output, got: %s", out[:min(50, len(out))])
	}
}

func TestBatchCmd_failBelowPasses(t *testing.T) {
	minimal := filepath.Join(fixturesBase, "skill-minimal")
	_, err := runBatch(t, "--fail-below", "F", minimal)
	if err != nil {
		t.Errorf("unexpected error when all skills meet F threshold: %v", err)
	}
}

func TestBatchCmd_failBelowTriggered(t *testing.T) {
	minimal := filepath.Join(fixturesBase, "skill-minimal")
	_, err := runBatch(t, "--fail-below", "A+", minimal)
	if err == nil {
		t.Error("expected error when skill is below A+ threshold")
	}
}

func TestBatchCmd_unknownGrade(t *testing.T) {
	skillPath := filepath.Join(fixturesBase, "skill-full")
	_, err := runBatch(t, "--fail-below", "BOGUS", skillPath)
	if err == nil {
		t.Error("expected error for unknown grade")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
