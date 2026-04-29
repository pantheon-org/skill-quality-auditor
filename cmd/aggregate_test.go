package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeAggSkillsDir creates a skills dir with skills matching a given family prefix.
func makeAggSkillsDir(t *testing.T, family string, n int) string {
	t.Helper()
	root := t.TempDir()
	for i := 0; i < n; i++ {
		dir := filepath.Join(root, "domain", fmt.Sprintf("%s-skill-%d", family, i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := strings.Repeat(fmt.Sprintf("This is %s skill %d. It handles thing %d.\n", family, i, i), 20)
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// runAggregate executes aggregateCmd with the given flags and returns (stdout, error).
func runAggregate(t *testing.T, family, repoRoot, skillsDir string, dryRun bool, posArgs []string) (string, error) {
	t.Helper()
	cmd := aggregateCmd
	cmd.ResetFlags()
	cmd.Flags().String("family", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("skills-dir", "", "")
	cmd.Flags().String("repo-root", "", "")

	if err := cmd.Flags().Set("family", family); err != nil {
		t.Fatalf("set family: %v", err)
	}
	if err := cmd.Flags().Set("repo-root", repoRoot); err != nil {
		t.Fatalf("set repo-root: %v", err)
	}
	if skillsDir != "" {
		if err := cmd.Flags().Set("skills-dir", skillsDir); err != nil {
			t.Fatalf("set skills-dir: %v", err)
		}
	}
	if dryRun {
		if err := cmd.Flags().Set("dry-run", "true"); err != nil {
			t.Fatalf("set dry-run: %v", err)
		}
	}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, posArgs)
	return buf.String(), err
}

func TestAggregateCmd_missingFamily(t *testing.T) {
	cmd := aggregateCmd
	cmd.ResetFlags()
	cmd.Flags().String("family", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("skills-dir", "", "")
	cmd.Flags().String("repo-root", "", "")

	err := cmd.RunE(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "--family is required") {
		t.Errorf("expected --family required error, got: %v", err)
	}
}

func TestAggregateCmd_args0SkillsDir(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "test", 2)
	repoRoot := t.TempDir()
	// Pass skills dir as positional arg — exercises the args[0] branch.
	if _, err := runAggregate(t, "test", repoRoot, "", true, []string{skillsDir}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAggregateCmd_dryRun(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "myskill", 2)
	repoRoot := t.TempDir()
	if _, err := runAggregate(t, "myskill", repoRoot, skillsDir, true, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAggregateCmd_writeOutput(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "writeskill", 2)
	repoRoot := t.TempDir()
	if _, err := runAggregate(t, "writeskill", repoRoot, skillsDir, false, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAggregateCmd_noFamilyMatch(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "other", 2)
	repoRoot := t.TempDir()
	_, err := runAggregate(t, "nomatch", repoRoot, skillsDir, false, []string{})
	if err == nil || !strings.Contains(err.Error(), "no skills found") {
		t.Errorf("expected 'no skills found' error, got: %v", err)
	}
}
