package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

// ---- canonicalSkillKey ----

func TestCanonicalSkillKey_standard(t *testing.T) {
	repoRoot := "/repo"
	skillPath := "/repo/skills/domain/my-skill/SKILL.md"
	got := canonicalSkillKey(skillPath, repoRoot)
	if got != "domain/my-skill" {
		t.Errorf("got %q, want %q", got, "domain/my-skill")
	}
}

func TestCanonicalSkillKey_notUnderSkillsDir(t *testing.T) {
	// If path is outside skills/, TrimPrefix is a no-op — returns the raw path
	// (minus any SKILL.md suffix). We just verify it doesn't panic.
	got := canonicalSkillKey("/other/path/SKILL.md", "/repo")
	if got == "" {
		t.Error("canonicalSkillKey should return a non-empty string")
	}
}

func TestCanonicalSkillKey_noSKILLMDSuffix(t *testing.T) {
	repoRoot := "/repo"
	// Path already points to directory, not the file.
	got := canonicalSkillKey("/repo/skills/domain/my-skill", repoRoot)
	// Still returns the directory segment without crashing.
	if got == "" {
		t.Error("expected non-empty result")
	}
}

// ---- resolveSkillPath ----

func TestResolveSkillPath_bareKey(t *testing.T) {
	repoRoot := "/repo"
	got := resolveSkillPath("domain/my-skill", repoRoot)
	want := "/repo/skills/domain/my-skill/SKILL.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveSkillPath_bareKeyWithSKILLMD(t *testing.T) {
	repoRoot := "/repo"
	got := resolveSkillPath("domain/my-skill/SKILL.md", repoRoot)
	// Should NOT double-append SKILL.md
	want := "/repo/skills/domain/my-skill/SKILL.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveSkillPath_absolutePath(t *testing.T) {
	got := resolveSkillPath("/abs/path/my-skill", "/repo")
	want := "/abs/path/my-skill/SKILL.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveSkillPath_absolutePathWithSKILLMD(t *testing.T) {
	got := resolveSkillPath("/abs/path/SKILL.md", "/repo")
	if got != "/abs/path/SKILL.md" {
		t.Errorf("got %q, want /abs/path/SKILL.md", got)
	}
}

func TestResolveSkillPath_relativeCurrentDir(t *testing.T) {
	got := resolveSkillPath("./my-skill", "/repo")
	// filepath.Abs will resolve relative to cwd; just check it ends with SKILL.md
	if filepath.Base(got) != "SKILL.md" {
		t.Errorf("expected path to end in SKILL.md, got %q", got)
	}
}

func TestResolveSkillPath_relativeParentDir(t *testing.T) {
	got := resolveSkillPath("../other-skill", "/repo")
	if filepath.Base(got) != "SKILL.md" {
		t.Errorf("expected path to end in SKILL.md, got %q", got)
	}
}

// ---- resolveRepoRoot ----

func TestResolveRepoRoot_flagValue(t *testing.T) {
	got, err := resolveRepoRoot("/explicit/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/explicit/root" {
		t.Errorf("got %q, want /explicit/root", got)
	}
}

func TestResolveRepoRoot_autoDetect(t *testing.T) {
	// Create a temp dir with a .git marker so findRepoRoot succeeds.
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "subdir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Change cwd into subdir so auto-detect walks up to tmp.
	orig, _ := os.Getwd()
	if err := os.Chdir(subdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck

	got, err := resolveRepoRoot("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

// ---- findRepoRoot ----

func TestFindRepoRoot_gitMarker(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findRepoRoot(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

func TestFindRepoRoot_gomodMarker(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := findRepoRoot(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

func TestFindRepoRoot_walksUp(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findRepoRoot(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

func TestFindRepoRoot_noMarker(t *testing.T) {
	// $TMPDIR on this machine may sit under ~/.config/claude which has .git,
	// so we create an isolated subtree directly under /private/tmp (macOS's
	// real /tmp) which is guaranteed to be outside any user repository.
	tmp, err := os.MkdirTemp("/private/tmp", "no-repo-*")
	if err != nil {
		t.Skip("cannot create temp dir under /private/tmp:", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmp) }) //nolint:errcheck
	subdir := filepath.Join(tmp, "deep", "subdir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err = findRepoRoot(subdir)
	if err == nil {
		t.Error("expected error when no .git or go.mod found")
	}
}

// ---- skillBaseName ----

func TestSkillBaseName_withSlash(t *testing.T) {
	got := skillBaseName("domain/skill-name")
	if got != "skill-name" {
		t.Errorf("got %q, want 'skill-name'", got)
	}
}

func TestSkillBaseName_noSlash(t *testing.T) {
	got := skillBaseName("skill-name")
	if got != "skill-name" {
		t.Errorf("got %q, want 'skill-name'", got)
	}
}

// ---- dateFromAuditPath / latestAuditJSON / loadAuditJSON ----

func TestDateFromAuditPath(t *testing.T) {
	path := "/repo/.context/audits/domain/skill/2026-04-28/audit.json"
	got := dateFromAuditPath(path)
	if got != "2026-04-28" {
		t.Errorf("got %q, want '2026-04-28'", got)
	}
}

func TestLatestAuditJSON_findsLatest(t *testing.T) {
	base := t.TempDir()
	for _, date := range []string{"2026-01-01", "2026-03-15", "2026-04-28"} {
		dir := filepath.Join(base, date)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "audit.json"), []byte(`{}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	path, date, err := latestAuditJSON(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if date != "2026-04-28" {
		t.Errorf("expected latest date '2026-04-28', got %q", date)
	}
	if filepath.Base(filepath.Dir(path)) != "2026-04-28" {
		t.Errorf("path should point to 2026-04-28 dir, got %q", path)
	}
}

func TestLatestAuditJSON_emptyDir(t *testing.T) {
	base := t.TempDir()
	_, _, err := latestAuditJSON(base)
	if err == nil {
		t.Error("expected error for empty audits dir")
	}
}

func TestLatestAuditJSON_missingAuditJSON(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "2026-04-28")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Directory exists but no audit.json inside it
	_, _, err := latestAuditJSON(base)
	if err == nil {
		t.Error("expected error when audit.json is missing")
	}
}

func TestLoadAuditJSON_valid(t *testing.T) {
	f := filepath.Join(t.TempDir(), "audit.json")
	data := `{"skill":"domain/skill","total":100,"maxTotal":140,"grade":"B"}`
	if err := os.WriteFile(f, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := loadAuditJSON(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Total != 100 {
		t.Errorf("got total %d, want 100", r.Total)
	}
}

func TestLoadAuditJSON_invalidJSON(t *testing.T) {
	f := filepath.Join(t.TempDir(), "audit.json")
	if err := os.WriteFile(f, []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadAuditJSON(f)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadAuditJSON_missingFile(t *testing.T) {
	_, err := loadAuditJSON("/nonexistent/audit.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---- resolveSkillPath error cases ----

// TestResolveSkillPath_dirWithNoSKILLMD verifies that scorer.Score returns an
// error when the resolved path points to a directory that has no SKILL.md.
func TestResolveSkillPath_dirWithNoSKILLMD(t *testing.T) {
	tmp := t.TempDir() // a real directory, but contains no SKILL.md
	path := resolveSkillPath(tmp, tmp)
	// path will be <tmp>/SKILL.md which does not exist — scorer.Score must error
	_, err := scorer.Score(path)
	if err == nil {
		t.Error("expected error when SKILL.md is absent from the directory")
	}
}

// TestResolveSkillPath_nonExistentPath verifies that scorer.Score returns an
// error when the resolved path does not exist at all.
func TestResolveSkillPath_nonExistentPath(t *testing.T) {
	path := resolveSkillPath("/nonexistent/domain/skill", "/nonexistent")
	_, err := scorer.Score(path)
	if err == nil {
		t.Error("expected error for a path that does not exist")
	}
}

// ---- fileExists ----

func TestFileExists_exists(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !fileExists(f) {
		t.Errorf("fileExists should return true for existing file")
	}
}

func TestFileExists_notExists(t *testing.T) {
	if fileExists("/nonexistent/path/file.txt") {
		t.Error("fileExists should return false for nonexistent file")
	}
}

func TestFileExists_directory(t *testing.T) {
	tmp := t.TempDir()
	if !fileExists(tmp) {
		t.Error("fileExists should return true for directories")
	}
}
