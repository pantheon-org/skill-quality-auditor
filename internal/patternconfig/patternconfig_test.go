package patternconfig

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata
var fixtureFS embed.FS

// resetActive restores the package-level singleton to defaultConfig so tests
// don't leak state into one another via Init's global mutation.
func resetActive(t *testing.T) {
	t.Helper()
	mu.Lock()
	active = defaultConfig
	mu.Unlock()
}

func TestGet_DefaultsBeforeInit(t *testing.T) {
	resetActive(t)
	cfg := Get()
	if len(cfg.Patterns.D1KnowledgeDelta.BeginnerSignals) == 0 {
		t.Fatal("expected default beginner signals before Init")
	}
	if cfg.Version != 1 {
		t.Errorf("want version 1, got %d", cfg.Version)
	}
}

func TestInit_ValidConfig_RoundTrip(t *testing.T) {
	resetActive(t)
	Init(fixtureFS, "testdata/valid.yaml")
	cfg := Get()
	if len(cfg.Patterns.D1KnowledgeDelta.BeginnerSignals) != 1 || cfg.Patterns.D1KnowledgeDelta.BeginnerSignals[0] != "custom beginner" {
		t.Errorf("expected loaded config to override defaults, got %v", cfg.Patterns.D1KnowledgeDelta.BeginnerSignals)
	}
}

func TestInit_MissingFile_FallsBackToDefaults(t *testing.T) {
	resetActive(t)
	Init(fixtureFS, "testdata/does-not-exist.yaml")
	cfg := Get()
	if len(cfg.Patterns.D1KnowledgeDelta.BeginnerSignals) != len(defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals) {
		t.Error("expected fallback to defaultConfig when file is missing")
	}
}

func TestInit_MalformedYAML_FallsBackToDefaults(t *testing.T) {
	resetActive(t)
	Init(fixtureFS, "testdata/malformed.yaml")
	cfg := Get()
	if len(cfg.Patterns.D1KnowledgeDelta.BeginnerSignals) != len(defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals) {
		t.Error("expected fallback to defaultConfig when YAML is malformed")
	}
}

func TestInit_EmptyGroup_RejectedFallsBackToDefaults(t *testing.T) {
	resetActive(t)
	Init(fixtureFS, "testdata/empty-group.yaml")
	cfg := Get()
	if len(cfg.Patterns.AnalysisQuality.HedgeWords) != len(defaultConfig.Patterns.AnalysisQuality.HedgeWords) {
		t.Error("expected fallback to defaultConfig when a pattern group is empty")
	}
}

func TestValidate_AllGroupsPresent(t *testing.T) {
	if err := validate(defaultConfig); err != nil {
		t.Errorf("expected defaultConfig to validate cleanly, got %v", err)
	}
}

func TestDefaultConfig_MatchesEmbeddedShippedConfig(t *testing.T) {
	// The shipped cmd/assets/assets/config/scoring-patterns.yaml must match
	// defaultConfig exactly — this guards against the two drifting apart.
	repoRoot := findRepoRoot(t)
	shipped := filepath.Join(repoRoot, "cmd", "assets", "assets", "config", "scoring-patterns.yaml")
	data, err := os.ReadFile(shipped)
	if err != nil {
		t.Skipf("shipped config not found (running outside repo checkout?): %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("shipped config failed to parse: %v", err)
	}
	if err := validate(cfg); err != nil {
		t.Fatalf("shipped config failed validation: %v", err)
	}
	assertSameStrings(t, "beginner_signals", cfg.Patterns.D1KnowledgeDelta.BeginnerSignals, defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals)
	assertSameStrings(t, "expert_signals", cfg.Patterns.D1KnowledgeDelta.ExpertSignals, defaultConfig.Patterns.D1KnowledgeDelta.ExpertSignals)
	assertSameStrings(t, "hedge_words", cfg.Patterns.AnalysisQuality.HedgeWords, defaultConfig.Patterns.AnalysisQuality.HedgeWords)
	assertSameStrings(t, "vague_words", cfg.Patterns.AnalysisQuality.VagueWords, defaultConfig.Patterns.AnalysisQuality.VagueWords)
	assertSameStrings(t, "passive_patterns", cfg.Patterns.AnalysisQuality.PassivePatterns, defaultConfig.Patterns.AnalysisQuality.PassivePatterns)
	assertSameStrings(t, "when_not_to_use", cfg.Patterns.D6FreedomCalibration.WhenNotToUse, defaultConfig.Patterns.D6FreedomCalibration.WhenNotToUse)
}

func assertSameStrings(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: length mismatch, got %d want %d (%v vs %v)", name, len(got), len(want), got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: got %q want %q", name, i, got[i], want[i])
		}
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

func TestLoadFromPath_ValidConfig_OverridesActive(t *testing.T) {
	resetActive(t)
	cfg, ok, err := LoadFromPath(filepath.Join("testdata", "valid.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true for a valid config")
	}
	if cfg.Patterns.D1KnowledgeDelta.BeginnerSignals[0] != "custom beginner" {
		t.Errorf("expected returned config to reflect the file, got %v", cfg.Patterns.D1KnowledgeDelta.BeginnerSignals)
	}
	active := Get()
	if active.Patterns.D1KnowledgeDelta.BeginnerSignals[0] != "custom beginner" {
		t.Errorf("expected active config to be overridden, got %v", active.Patterns.D1KnowledgeDelta.BeginnerSignals)
	}
}

func TestLoadFromPath_MissingFile_SilentlySkipped(t *testing.T) {
	resetActive(t)
	_, ok, err := LoadFromPath(filepath.Join("testdata", "does-not-exist.yaml"))
	if err != nil {
		t.Errorf("expected nil error for an absent file, got %v", err)
	}
	if ok {
		t.Error("expected ok=false for an absent file")
	}
	if Get().Patterns.D1KnowledgeDelta.BeginnerSignals[0] != defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals[0] {
		t.Error("expected active config to remain untouched when the file is absent")
	}
}

func TestLoadFromPath_Directory_ReturnsDistinctError(t *testing.T) {
	resetActive(t)
	dir := t.TempDir()
	_, ok, err := LoadFromPath(dir)
	if ok {
		t.Error("expected ok=false for a directory")
	}
	if err == nil {
		t.Fatal("expected an error for a directory")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("expected error to mention 'directory', got %v", err)
	}
	if Get().Patterns.D1KnowledgeDelta.BeginnerSignals[0] != defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals[0] {
		t.Error("expected active config to remain untouched when the path is a directory")
	}
}

func TestLoadFromPath_PermissionDenied_ReturnsError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits are not enforced")
	}
	resetActive(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "unreadable.yaml")
	if err := os.WriteFile(path, []byte("version: 1\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	_, ok, err := LoadFromPath(path)
	if ok {
		t.Error("expected ok=false for a permission-denied file")
	}
	if err == nil {
		t.Fatal("expected an error for a permission-denied file")
	}
	if Get().Patterns.D1KnowledgeDelta.BeginnerSignals[0] != defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals[0] {
		t.Error("expected active config to remain untouched on a permission error")
	}
}

func TestLoadFromPath_SymlinkToMissingTarget_ReturnsError(t *testing.T) {
	resetActive(t)
	dir := t.TempDir()
	link := filepath.Join(dir, "dangling.yaml")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist.yaml"), link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	_, ok, err := LoadFromPath(link)
	if ok {
		t.Error("expected ok=false for a dangling symlink")
	}
	if err == nil {
		t.Fatal("expected an error for a dangling symlink")
	}
	if Get().Patterns.D1KnowledgeDelta.BeginnerSignals[0] != defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals[0] {
		t.Error("expected active config to remain untouched for a dangling symlink")
	}
}

func TestLoadFromPath_MalformedYAML_ReturnsDistinctError(t *testing.T) {
	resetActive(t)
	_, ok, err := LoadFromPath(filepath.Join("testdata", "malformed.yaml"))
	if ok {
		t.Error("expected ok=false for malformed YAML")
	}
	if err == nil || !strings.Contains(err.Error(), "parse YAML") {
		t.Errorf("expected a parse error, got %v", err)
	}
	if Get().Patterns.D1KnowledgeDelta.BeginnerSignals[0] != defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals[0] {
		t.Error("expected active config to remain untouched for malformed YAML")
	}
}

func TestLoadFromPath_MissingGroups_ReturnsDistinctError(t *testing.T) {
	resetActive(t)
	_, ok, err := LoadFromPath(filepath.Join("testdata", "partial-groups.yaml"))
	if ok {
		t.Error("expected ok=false for a config missing a pattern group")
	}
	if err == nil || !strings.Contains(err.Error(), "missing groups") {
		t.Errorf("expected a missing-groups error naming the group, got %v", err)
	}
	if !strings.Contains(err.Error(), "d6_freedom_calibration.when_not_to_use") {
		t.Errorf("expected error to name the missing group, got %v", err)
	}
	if Get().Patterns.D1KnowledgeDelta.BeginnerSignals[0] != defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals[0] {
		t.Error("expected active config to remain untouched when groups are missing")
	}
}

func TestLoadFromPath_EmptyGroup_ReturnsDistinctError(t *testing.T) {
	resetActive(t)
	_, ok, err := LoadFromPath(filepath.Join("testdata", "empty-group.yaml"))
	if ok {
		t.Error("expected ok=false for an empty pattern group")
	}
	if err == nil || !strings.Contains(err.Error(), "missing groups") {
		t.Errorf("expected a missing-groups error, got %v", err)
	}
}

func TestWriteDefault_RoundTrips(t *testing.T) {
	resetActive(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "generated", "scoring-patterns.yaml")
	if err := WriteDefault(path, defaultConfig); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}
	cfg, ok, err := LoadFromPath(path)
	if err != nil || !ok {
		t.Fatalf("expected the written file to load back cleanly, ok=%v err=%v", ok, err)
	}
	assertSameStrings(t, "beginner_signals", cfg.Patterns.D1KnowledgeDelta.BeginnerSignals, defaultConfig.Patterns.D1KnowledgeDelta.BeginnerSignals)
	assertSameStrings(t, "when_not_to_use", cfg.Patterns.D6FreedomCalibration.WhenNotToUse, defaultConfig.Patterns.D6FreedomCalibration.WhenNotToUse)
}

func TestWriteDefault_SurfacesWriteFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits are not enforced")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	path := filepath.Join(dir, "subdir", "scoring-patterns.yaml")
	if err := WriteDefault(path, defaultConfig); err == nil {
		t.Error("expected WriteDefault to surface a write failure, got nil")
	}
}

func TestStripCodeBlocks(t *testing.T) {
	content := "before\n```go\nfoo := 1\n```\nafter"
	got := StripCodeBlocks(content)
	if got != "before\nafter\n" {
		t.Errorf("got %q", got)
	}
}
