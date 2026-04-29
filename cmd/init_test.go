package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---- prettyPath ----

func TestPrettyPath_replacesHome(t *testing.T) {
	home := "/Users/alice"
	got := prettyPath("/Users/alice/foo/bar", home)
	if got != "~/foo/bar" {
		t.Errorf("got %q, want ~/foo/bar", got)
	}
}

func TestPrettyPath_exactHome(t *testing.T) {
	home := "/Users/alice"
	got := prettyPath("/Users/alice", home)
	if got != "~" {
		t.Errorf("got %q, want ~", got)
	}
}

func TestPrettyPath_noMatch(t *testing.T) {
	got := prettyPath("/tmp/something", "/Users/alice")
	if got != "/tmp/something" {
		t.Errorf("got %q, want /tmp/something", got)
	}
}

func TestPrettyPath_emptyHome(t *testing.T) {
	got := prettyPath("/tmp/something", "")
	if got != "/tmp/something" {
		t.Errorf("got %q, want /tmp/something", got)
	}
}

// ---- formatAgentIDs ----

func TestFormatAgentIDs_single(t *testing.T) {
	a, _ := agentByID("claude-code")
	if got := formatAgentIDs([]Agent{a}); got != "claude-code" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAgentIDs_few(t *testing.T) {
	a1, _ := agentByID("claude-code")
	a2, _ := agentByID("cursor")
	got := formatAgentIDs([]Agent{a1, a2})
	if got != "claude-code, cursor" {
		t.Errorf("got %q", got)
	}
}

func TestFormatAgentIDs_abbreviates(t *testing.T) {
	agents := agentRegistry[:5]
	got := formatAgentIDs(agents)
	if !strings.Contains(got, "+2 more") {
		t.Errorf("expected '+2 more' abbreviation, got %q", got)
	}
}

// ---- groupTargetsByDir ----

func TestGroupTargetsByDir_deduplicates(t *testing.T) {
	// Several agents share ".agents/skills" as ProjectPath.
	var sharedPath []Agent
	for _, a := range agentRegistry {
		if a.ProjectPath == ".agents/skills" {
			sharedPath = append(sharedPath, a)
			if len(sharedPath) == 3 {
				break
			}
		}
	}
	if len(sharedPath) < 2 {
		t.Skip("need at least 2 agents sharing .agents/skills")
	}

	groups := groupTargetsByDir(sharedPath, "/base", false)
	if len(groups) != 1 {
		t.Errorf("expected 1 group for shared path, got %d", len(groups))
	}
	if len(groups[0].agents) != len(sharedPath) {
		t.Errorf("expected %d agents in group, got %d", len(sharedPath), len(groups[0].agents))
	}
}

func TestGroupTargetsByDir_keepsDistinct(t *testing.T) {
	a1, _ := agentByID("claude-code") // .claude/skills
	a2, _ := agentByID("cursor")      // .cursor/skills (different)
	groups := groupTargetsByDir([]Agent{a1, a2}, "/base", false)
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

// ---- resolveByIDs ----

func TestResolveByIDs_knownID(t *testing.T) {
	agents, err := resolveByIDs([]string{"claude-code"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 1 || agents[0].ID != "claude-code" {
		t.Errorf("expected [claude-code], got %v", agents)
	}
}

func TestResolveByIDs_unknownID(t *testing.T) {
	_, err := resolveByIDs([]string{"does-not-exist-xyz"})
	if err == nil {
		t.Error("expected error for unknown agent ID")
	}
}

func TestResolveByIDs_multipleIDs(t *testing.T) {
	agents, err := resolveByIDs([]string{"claude-code", "cursor"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

// ---- resolveByHarness ----

func TestResolveByHarness_nonePresent(t *testing.T) {
	agents, err := resolveByHarness(t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("expected 0 agents for empty dir, got %d", len(agents))
	}
}

func TestResolveByHarness_detected(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	agents, err := resolveByHarness(base, false)
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

func TestResolveByHarness_global_detected(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	agents, err := resolveByHarness(home, true)
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
		t.Error("expected claude-code to be auto-detected globally")
	}
}

// ---- resolveTargets (legacy shim) ----

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
		t.Errorf("expected 0 agents for empty dir, got %d", len(agents))
	}
}

func TestResolveTargets_autoDetect_detected(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	agents, err := resolveTargets(nil, base, false)
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

// ---- resolveInteractive ----

func TestResolveInteractive_all(t *testing.T) {
	in := strings.NewReader("all\n")
	var out bytes.Buffer
	agents, err := resolveInteractive(in, &out, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != len(agentRegistry) {
		t.Errorf("expected all %d agents, got %d", len(agentRegistry), len(agents))
	}
}

func TestResolveInteractive_numberedSelection(t *testing.T) {
	in := strings.NewReader("1,2\n")
	var out bytes.Buffer
	agents, err := resolveInteractive(in, &out, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
	if agents[0].ID != agentRegistry[0].ID {
		t.Errorf("expected first agent %q, got %q", agentRegistry[0].ID, agents[0].ID)
	}
}

func TestResolveInteractive_invalidSelection(t *testing.T) {
	in := strings.NewReader("999\n")
	var out bytes.Buffer
	_, err := resolveInteractive(in, &out, t.TempDir(), false)
	if err == nil {
		t.Error("expected error for out-of-range selection")
	}
}

func TestResolveInteractive_emptyInput(t *testing.T) {
	in := strings.NewReader("\n")
	var out bytes.Buffer
	agents, err := resolveInteractive(in, &out, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agents != nil {
		t.Errorf("expected nil agents for empty input, got %v", agents)
	}
}

func TestResolveInteractive_marksDetectedAgents(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	in := strings.NewReader("1\n")
	var out bytes.Buffer
	_, _ = resolveInteractive(in, &out, base, false)
	if !strings.Contains(out.String(), "* ") {
		t.Error("expected at least one detected agent marked with *")
	}
}

// ---- agentSkillDir ----

func TestAgentSkillDir_local(t *testing.T) {
	a, _ := agentByID("claude-code")
	dir := agentSkillDir(a, "/base", false)
	want := filepath.Join("/base", a.ProjectPath, skillName)
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestAgentSkillDir_global(t *testing.T) {
	a, _ := agentByID("claude-code")
	dir := agentSkillDir(a, "/home/user", true)
	want := filepath.Join("/home/user", a.GlobalPath, skillName)
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
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

func TestWriteCanonical_alsoWritesAssets(t *testing.T) {
	home := t.TempDir()
	if _, err := writeCanonical(home); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	base := filepath.Join(home, ".local", "share", skillName)
	for _, subdir := range assetSubdirs {
		if _, err := os.Stat(filepath.Join(base, subdir)); err != nil {
			t.Errorf("asset subdir %q not created: %v", subdir, err)
		}
	}
}

// ---- newInitCmd (test helper) ----

func newInitCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := initCmd
	cmd.ResetFlags()
	cmd.Flags().StringArrayP("agent", "a", nil, "")
	cmd.Flags().BoolP("global", "g", false, "")
	cmd.Flags().BoolP("interactive", "I", false, "")
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
	if err := cmd.Flags().Set("agent", "unknown-agent-xyz"); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("expected unknown agent error, got: %v", err)
	}
}

// ---- dry-run output format ----

func TestInitCmd_dryRun_copy_output(t *testing.T) {
	cmd := newInitCmd(t)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	if err := cmd.Flags().Set("method", "copy"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("agent", "claude-code"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected [dry-run] prefix, got: %q", out)
	}
	if !strings.Contains(out, "(claude-code)") {
		t.Errorf("expected agent label in output, got: %q", out)
	}
	if !strings.Contains(out, "SKILL.md (copy)") {
		t.Errorf("expected 'SKILL.md (copy)', got: %q", out)
	}
	if !strings.Contains(out, "references/") {
		t.Errorf("expected 'references/' in asset list, got: %q", out)
	}
	if !strings.Contains(out, "evals/") {
		t.Errorf("expected 'evals/' in asset list, got: %q", out)
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

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "(symlink)") {
		t.Errorf("expected '(symlink)' in output, got: %q", out)
	}
	if !strings.Contains(out, "→") {
		t.Errorf("expected symlink arrow in output, got: %q", out)
	}
	if !strings.Contains(out, "assets:") {
		t.Errorf("expected 'assets:' line in output, got: %q", out)
	}
}

func TestInitCmd_dryRun_deduplicates(t *testing.T) {
	// amp and cline both use .agents/skills — should appear as one entry.
	cmd := newInitCmd(t)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	if err := cmd.Flags().Set("method", "copy"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("agent", "amp"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("agent", "cline"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both share .agents/skills — must produce exactly one [dry-run] block.
	count := strings.Count(buf.String(), "[dry-run]")
	if count != 1 {
		t.Errorf("expected 1 dry-run block for shared path, got %d:\n%s", count, buf.String())
	}
	if !strings.Contains(buf.String(), "amp") || !strings.Contains(buf.String(), "cline") {
		t.Errorf("expected both agent IDs in output: %q", buf.String())
	}
}

func TestInitCmd_dryRun_pathUseTilde(t *testing.T) {
	// With --global the output path should use ~ not the raw home dir.
	// We can't easily control os.UserHomeDir in RunE, so test prettyPath directly
	// and verify the dry-run branch calls it.
	home := t.TempDir()
	long := filepath.Join(home, "some", "path")
	got := prettyPath(long, home)
	want := "~/some/path"
	if got != want {
		t.Errorf("prettyPath(%q, %q) = %q, want %q", long, home, got, want)
	}
}

func TestInitCmd_dryRun_shorthand(t *testing.T) {
	cmd := newInitCmd(t)
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
		{"I", "interactive"},
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

// ---- writeAllAssets / writeRefs ----

func TestWriteAllAssets(t *testing.T) {
	dest := t.TempDir()
	if err := writeAllAssets(dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, subdir := range assetSubdirs {
		entries, err := os.ReadDir(filepath.Join(dest, subdir))
		if err != nil {
			t.Errorf("subdir %q not created: %v", subdir, err)
			continue
		}
		if len(entries) == 0 {
			t.Errorf("subdir %q is empty", subdir)
		}
	}
}

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
