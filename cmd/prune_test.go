package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeAuditTree builds a .context/audits/<skill>/<date>/ tree under root and
// returns the audit root path.
func makeAuditTree(t *testing.T, root string, skill string, dates []string) string {
	t.Helper()
	auditRoot := filepath.Join(root, ".context", "audits")
	for _, d := range dates {
		dir := filepath.Join(auditRoot, skill, d)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "audit.json"), `{"grade":"A"}`)
	}
	return auditRoot
}

// runPrune exercises the prune command's RunE directly with a synthetic repo root.
func runPrune(t *testing.T, _ string, keep int) error {
	t.Helper()
	cmd := pruneCmd
	cmd.ResetFlags()
	cmd.Flags().Int("keep", 5, "")
	if err := cmd.Flags().Set("keep", fmt.Sprintf("%d", keep)); err != nil {
		t.Fatalf("set keep: %v", err)
	}
	cmd.SetOut(&bytes.Buffer{})
	return cmd.RunE(cmd, []string{})
}

// --------------------------------------------------------------------------
// prune tests
// --------------------------------------------------------------------------

func TestPrune_noAuditDir(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	// Point resolveRepoRoot at tmp (no .context/audits exists).
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runPrune(t, tmp, 5); err != nil {
		t.Errorf("expected nil when no audit dir, got: %v", err)
	}
}

func TestPrune_keepsRecentDeletesOld(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")

	dates := []string{"2026-01-01", "2026-02-01", "2026-03-01", "2026-04-01", "2026-04-15", "2026-04-28"}
	makeAuditTree(t, tmp, "my-skill", dates)

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runPrune(t, tmp, 3); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	skillAuditDir := filepath.Join(tmp, ".context", "audits", "my-skill")
	remaining, err := os.ReadDir(skillAuditDir)
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range remaining {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) != 3 {
		t.Errorf("expected 3 dirs kept, got %d: %v", len(dirs), dirs)
	}
	// The three newest should be kept.
	for _, want := range []string{"2026-04-28", "2026-04-15", "2026-04-01"} {
		found := false
		for _, d := range dirs {
			if d == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to be kept, remaining: %v", want, dirs)
		}
	}
}

func TestPrune_keepMoreThanExist(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	makeAuditTree(t, tmp, "my-skill", []string{"2026-04-01", "2026-04-28"})

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runPrune(t, tmp, 10); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	skillAuditDir := filepath.Join(tmp, ".context", "audits", "my-skill")
	remaining, _ := os.ReadDir(skillAuditDir)
	var dirs []string
	for _, e := range remaining {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) != 2 {
		t.Errorf("expected both dirs kept, got %d: %v", len(dirs), dirs)
	}
}

func TestPrune_preservesLatestSymlink(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	makeAuditTree(t, tmp, "my-skill", []string{"2026-04-01", "2026-04-28"})

	skillAuditDir := filepath.Join(tmp, ".context", "audits", "my-skill")
	latestPath := filepath.Join(skillAuditDir, "latest")
	if err := os.Symlink("2026-04-28", latestPath); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runPrune(t, tmp, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// latest symlink must still exist
	if _, err := os.Lstat(latestPath); err != nil {
		t.Errorf("expected 'latest' symlink to be preserved, got: %v", err)
	}
}

func TestPrune_multipleSkills(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	makeAuditTree(t, tmp, "skill-a", []string{"2026-01-01", "2026-02-01", "2026-03-01", "2026-04-28"})
	makeAuditTree(t, tmp, "skill-b", []string{"2026-03-01", "2026-04-28"})

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runPrune(t, tmp, 2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for skill, wantCount := range map[string]int{"skill-a": 2, "skill-b": 2} {
		d := filepath.Join(tmp, ".context", "audits", skill)
		entries, _ := os.ReadDir(d)
		var dirs []string
		for _, e := range entries {
			if e.IsDir() {
				dirs = append(dirs, e.Name())
			}
		}
		if len(dirs) != wantCount {
			t.Errorf("%s: expected %d dirs, got %d: %v", skill, wantCount, len(dirs), dirs)
		}
	}
}

func TestPrune_nonDirEntriesSkipped(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "go.mod"), "module x\n")
	auditRoot := filepath.Join(tmp, ".context", "audits")
	if err := os.MkdirAll(auditRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a regular file directly in the audit root — should be skipped.
	writeFile(t, filepath.Join(auditRoot, "README.md"), "# readme\n")

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runPrune(t, tmp, 5); err != nil {
		t.Errorf("non-dir entry should be skipped without error, got: %v", err)
	}
}

func TestPruneCmd_helpFlagMentionsKeep(t *testing.T) {
	usage := pruneCmd.UsageString()
	if !strings.Contains(usage, "--keep") {
		t.Error("expected --keep flag to appear in usage string")
	}
}
