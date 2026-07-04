package llmclient

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Provider IDs are stable string identifiers for each shipped provider.
const (
	ProviderAnthropic        = "anthropic"
	ProviderOpenAI           = "openai"
	ProviderGemini           = "gemini"
	ProviderOpenAICompatible = "openai-compatible"
)

// Default models per provider (recorded in ADR-025). Override via LLM_MODEL.
const (
	DefaultModelAnthropic = "claude-sonnet-4-6"
	DefaultModelOpenAI    = "gpt-4o"
	DefaultModelGemini    = "gemini-2.0-flash"
)

// Default endpoints per provider. Override via LLM_BASE_URL.
const (
	DefaultBaseAnthropic = "https://api.anthropic.com"
	DefaultBaseOpenAI    = "https://api.openai.com"
	DefaultBaseGemini    = "https://generativelanguage.googleapis.com"
)

// providers is the registry of shipped provider factories. Each adapter
// registers itself via an init() in its own file.
var providers = map[string]Provider{}

// register adds a provider to the registry. Called from each adapter's init().
func register(p Provider) {
	providers[p.ID()] = p
}

// DefaultHTTPClient is the shared HTTP client used by every provider adapter.
// It has a 60s timeout per the Phase 1 reliability spec. Tests override by
// constructing the adapter directly rather than via NewFromEnv.
var DefaultHTTPClient = &http.Client{Timeout: 60 * time.Second}

// Selection captures the resolved provider configuration returned by
// NewFromEnv. When Key is empty, callers MUST degrade to structural-only
// mode rather than proceeding with an unauthenticated request (ADR-007 #5).
type Selection struct {
	Provider string
	Config   Config
}

// NewFromEnv selects a provider based on the LLM_PROVIDER env var (or
// --provider flag value passed in via providerOverride), reads the
// provider-specific key, model, and base URL from the environment, and
// returns a constructed Client.
//
// Returns (nil, nil) when the selected provider has no key configured so
// callers can degrade gracefully to structural-only mode (ADR-007 #5).
// Returns an error only when the selected provider id is unknown or when
// openai-compatible is selected without an LLM_BASE_URL (no safe default
// exists for local gateways).
func NewFromEnv(providerOverride string) (Client, error) {
	provider := providerOverride
	if provider == "" {
		provider = os.Getenv("LLM_PROVIDER")
	}
	if provider == "" {
		provider = ProviderAnthropic
	}

	factory, ok := providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown llm provider %q (valid: %s)", provider, validProviders())
	}

	cfg := Config{Provider: provider}

	switch provider {
	case ProviderAnthropic:
		cfg.Key = os.Getenv("ANTHROPIC_API_KEY")
		cfg.BaseURL = envOrDefault("LLM_BASE_URL", DefaultBaseAnthropic)
		cfg.Model = envOrDefault("LLM_MODEL", DefaultModelAnthropic)
	case ProviderOpenAI:
		cfg.Key = os.Getenv("OPENAI_API_KEY")
		cfg.BaseURL = envOrDefault("LLM_BASE_URL", DefaultBaseOpenAI)
		cfg.Model = envOrDefault("LLM_MODEL", DefaultModelOpenAI)
	case ProviderGemini:
		cfg.Key = firstNonEmpty(os.Getenv("GEMINI_API_KEY"), os.Getenv("GOOGLE_API_KEY"))
		cfg.BaseURL = envOrDefault("LLM_BASE_URL", DefaultBaseGemini)
		cfg.Model = envOrDefault("LLM_MODEL", DefaultModelGemini)
	case ProviderOpenAICompatible:
		cfg.Key = os.Getenv("OPENAI_API_KEY") // gateways usually proxy OpenAI auth
		cfg.BaseURL = os.Getenv("LLM_BASE_URL")
		if cfg.BaseURL == "" {
			return nil, errors.New("openai-compatible provider requires LLM_BASE_URL (no canonical default exists)")
		}
		cfg.Model = envOrDefault("LLM_MODEL", DefaultModelOpenAI)
	}

	if cfg.Key == "" {
		// Graceful degradation: no key for the selected provider.
		return nil, nil
	}

	return factory.New(cfg)
}

// MustNew is a convenience for tests that have already validated the env.
// Production callers should use NewFromEnv to receive the nil degradation
// signal.
func MustNew(cfg Config) (Client, error) {
	factory, ok := providers[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", cfg.Provider)
	}
	return factory.New(cfg)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func validProviders() string {
	ids := make([]string, 0, len(providers))
	for id := range providers {
		ids = append(ids, id)
	}
	return strings.Join(ids, ", ")
}

// IsRetryable reports whether an HTTP status warrants a retry (429 or 5xx).
// Used by adapters that share retry logic.
func IsRetryable(status int) bool {
	return status == 429 || (status >= 500 && status < 600)
}

// Backoff returns the delay for a given retry attempt (0-indexed) using
// exponential backoff with a small full-jitter cap. Adapters may use this
// in their retry loops.
func Backoff(attempt int) time.Duration {
	base := 500 * time.Millisecond
	max := 8 * time.Second
	d := base
	for i := 0; i < attempt && d < max; i++ {
		d *= 2
	}
	if d > max {
		d = max
	}
	return d
}
