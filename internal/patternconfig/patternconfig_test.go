package patternconfig

import (
	"embed"
	"os"
	"path/filepath"
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

func TestStripCodeBlocks(t *testing.T) {
	content := "before\n```go\nfoo := 1\n```\nafter"
	got := StripCodeBlocks(content)
	if got != "before\nafter\n" {
		t.Errorf("got %q", got)
	}
}
