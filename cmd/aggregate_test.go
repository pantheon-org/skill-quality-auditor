package cmd

import (
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

func TestAggregateCmd_missingFamily(t *testing.T) {
	origFamily := aggFamily
	aggFamily = ""
	defer func() { aggFamily = origFamily }()

	err := aggregateCmd.RunE(aggregateCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "--family is required") {
		t.Errorf("expected --family required error, got: %v", err)
	}
}

func TestAggregateCmd_args0SkillsDir(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "test", 2)
	repoRoot := t.TempDir()

	origFamily, origRoot, origDry := aggFamily, aggRepoRoot, aggDryRun
	aggFamily = "test"
	aggRepoRoot = repoRoot
	aggDryRun = true
	defer func() { aggFamily = origFamily; aggRepoRoot = origRoot; aggDryRun = origDry }()

	// Pass skills dir as positional arg — exercises the args[0] branch.
	err := aggregateCmd.RunE(aggregateCmd, []string{skillsDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAggregateCmd_dryRun(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "myskill", 2)
	repoRoot := t.TempDir()

	origFamily, origRoot, origSkills, origDry := aggFamily, aggRepoRoot, aggSkillsDir, aggDryRun
	aggFamily = "myskill"
	aggRepoRoot = repoRoot
	aggSkillsDir = skillsDir
	aggDryRun = true
	defer func() {
		aggFamily = origFamily
		aggRepoRoot = origRoot
		aggSkillsDir = origSkills
		aggDryRun = origDry
	}()

	if err := aggregateCmd.RunE(aggregateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAggregateCmd_writeOutput(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "writeskill", 2)
	repoRoot := t.TempDir()

	origFamily, origRoot, origSkills, origDry := aggFamily, aggRepoRoot, aggSkillsDir, aggDryRun
	aggFamily = "writeskill"
	aggRepoRoot = repoRoot
	aggSkillsDir = skillsDir
	aggDryRun = false
	defer func() {
		aggFamily = origFamily
		aggRepoRoot = origRoot
		aggSkillsDir = origSkills
		aggDryRun = origDry
	}()

	if err := aggregateCmd.RunE(aggregateCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAggregateCmd_noFamilyMatch(t *testing.T) {
	skillsDir := makeAggSkillsDir(t, "other", 2)
	repoRoot := t.TempDir()

	origFamily, origRoot, origSkills := aggFamily, aggRepoRoot, aggSkillsDir
	aggFamily = "nomatch"
	aggRepoRoot = repoRoot
	aggSkillsDir = skillsDir
	defer func() { aggFamily = origFamily; aggRepoRoot = origRoot; aggSkillsDir = origSkills }()

	err := aggregateCmd.RunE(aggregateCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "no skills found") {
		t.Errorf("expected 'no skills found' error, got: %v", err)
	}
}
