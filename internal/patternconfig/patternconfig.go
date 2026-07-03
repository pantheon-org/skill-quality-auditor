// Package patternconfig loads the scoring pattern lists (D1 knowledge-delta
// signals, D6 when-not-to-use phrases, and analysis quality word lists) from
// an embedded YAML config, falling back to built-in defaults if the config
// is missing or malformed. Scoring must never fail because of a bad config.
package patternconfig

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config is the root of the scoring-patterns.yaml document.
type Config struct {
	Version  int      `yaml:"version"`
	Patterns Patterns `yaml:"patterns"`
}

// Patterns groups the pattern lists by the dimension or component that consumes them.
type Patterns struct {
	D1KnowledgeDelta     D1Patterns       `yaml:"d1_knowledge_delta"`
	AnalysisQuality      AnalysisPatterns `yaml:"analysis_quality"`
	D6FreedomCalibration D6Patterns       `yaml:"d6_freedom_calibration"`
}

// D1Patterns holds the beginner/expert register signals used by scoreD1.
type D1Patterns struct {
	BeginnerSignals []string `yaml:"beginner_signals"`
	ExpertSignals   []string `yaml:"expert_signals"`
}

// AnalysisPatterns holds the hedge/vague/passive word lists used by DetectAntiPatternSignals.
type AnalysisPatterns struct {
	HedgeWords      []string `yaml:"hedge_words"`
	VagueWords      []string `yaml:"vague_words"`
	PassivePatterns []string `yaml:"passive_patterns"`
}

// D6Patterns holds the negative-scope phrases used by scoreWhenNotToUse.
type D6Patterns struct {
	WhenNotToUse []string `yaml:"when_not_to_use"`
}

// defaultConfig mirrors the values previously hardcoded in each scorer.
// It is the fallback used whenever the embedded YAML fails to load or parse.
var defaultConfig = Config{
	Version: 1,
	Patterns: Patterns{
		D1KnowledgeDelta: D1Patterns{
			BeginnerSignals: []string{"npm install", "yarn add", "pip install", "getting started", "introduction", "basic syntax", "hello world"},
			ExpertSignals:   []string{"anti-pattern", "NEVER", "ALWAYS", "production", "gotcha", "pitfall"},
		},
		AnalysisQuality: AnalysisPatterns{
			HedgeWords:      []string{"maybe", "perhaps", "might want to", "could be", "feel free", "you might", "possibly"},
			VagueWords:      []string{"do something", "handle appropriately", "as needed", "when necessary", "if applicable"},
			PassivePatterns: []string{"is done", "was created", "can be used", "is used", "are used", "is called", "was called"},
		},
		D6FreedomCalibration: D6Patterns{
			WhenNotToUse: []string{"when not to use", "do not use", "not intended for", "outside the scope", "avoid using"},
		},
	},
}

var (
	mu     sync.RWMutex
	active = defaultConfig
)

// Init loads the pattern config from fs at path and installs it as the active
// config. On any error (missing file, malformed YAML, empty pattern group) it
// logs a warning to stderr and leaves the built-in defaults in place — a bad
// config must degrade scoring quality, never break it.
func Init(fs embed.FS, path string) {
	data, err := fs.ReadFile(path)
	if err != nil {
		warnf("cannot read pattern config %s: %v (using built-in defaults)", path, err)
		return
	}
	cfg, err := parseAndValidate(data)
	if err != nil {
		warnf("pattern config %s: %v (using built-in defaults)", path, err)
		return
	}
	mu.Lock()
	active = cfg
	mu.Unlock()
}

// LoadFromPath reads, parses, and validates a scoring-patterns.yaml from an
// arbitrary OS path, installing it as the active config on success. It
// distinguishes a genuinely absent file (ok=false, err=nil — not configured)
// from a path that exists but is unusable: a directory, a permission error, a
// symlink to a missing target, malformed YAML, or a config missing one or
// more required pattern groups (each ok=false, err=<detail>). The prior
// active config is left untouched unless loading succeeds (ok=true, err=nil).
func LoadFromPath(path string) (Config, bool, error) {
	lstatInfo, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, false, fmt.Errorf("cannot stat pattern config %s: %w", path, err)
	}
	if lstatInfo.Mode()&os.ModeSymlink != 0 {
		if _, err := os.Stat(path); err != nil {
			return Config{}, false, fmt.Errorf("pattern config %s is a symlink to a missing or unreadable target: %w", path, err)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return Config{}, false, fmt.Errorf("cannot stat pattern config %s: %w", path, err)
	}
	if info.IsDir() {
		return Config{}, false, fmt.Errorf("pattern config %s is a directory, not a file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false, fmt.Errorf("cannot read pattern config %s: %w", path, err)
	}
	cfg, err := parseAndValidate(data)
	if err != nil {
		return Config{}, false, fmt.Errorf("pattern config %s: %w", path, err)
	}
	mu.Lock()
	active = cfg
	mu.Unlock()
	return cfg, true, nil
}

// WriteDefault marshals cfg back to YAML, matching the embedded file's
// structure so the written file round-trips through LoadFromPath, and writes
// it to path (creating parent directories as needed). It is the first-run
// auto-generation primitive: callers must treat a returned error as
// non-fatal (log and continue with the in-memory config), never panic.
func WriteDefault(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal pattern config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create pattern config directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write pattern config %s: %w", path, err)
	}
	return nil
}

func warnf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "warning: "+format+"\n", a...)
}

// parseAndValidate unmarshals raw YAML bytes and validates that every
// required pattern group is present and non-empty. It is the single code
// path shared by the embedded loader (Init) and the disk loader
// (LoadFromPath) so both apply identical rules.
func parseAndValidate(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse YAML: %w", err)
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validate(cfg Config) error {
	groups := []struct {
		name string
		vals []string
	}{
		{"d1_knowledge_delta.beginner_signals", cfg.Patterns.D1KnowledgeDelta.BeginnerSignals},
		{"d1_knowledge_delta.expert_signals", cfg.Patterns.D1KnowledgeDelta.ExpertSignals},
		{"analysis_quality.hedge_words", cfg.Patterns.AnalysisQuality.HedgeWords},
		{"analysis_quality.vague_words", cfg.Patterns.AnalysisQuality.VagueWords},
		{"analysis_quality.passive_patterns", cfg.Patterns.AnalysisQuality.PassivePatterns},
		{"d6_freedom_calibration.when_not_to_use", cfg.Patterns.D6FreedomCalibration.WhenNotToUse},
	}
	var missing []string
	for _, g := range groups {
		if len(g.vals) == 0 {
			missing = append(missing, g.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing groups: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Get returns the currently active pattern config — the loaded YAML config
// after a successful Init, or the built-in defaults otherwise.
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// StripCodeBlocks removes fenced code blocks (```...```) from content, so
// pattern matching does not fire on example code. This is the single
// canonical implementation shared by the analysis and scorer packages.
func StripCodeBlocks(content string) string {
	var result strings.Builder
	skip := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			skip = !skip
			continue
		}
		if !skip {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return result.String()
}
