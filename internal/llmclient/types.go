// Package llmclient provides a provider-agnostic abstraction over LLM
// chat completion APIs (Anthropic, OpenAI, Gemini, Mistral, Cerebras, and
// OpenAI-compatible gateways such as Ollama and vLLM).
//
// The Client interface is the model-call boundary shared by every provider
// implementation. NewFromEnv selects a provider based on the LLM_PROVIDER
// env var (or the --provider flag) and reads the provider-specific key,
// returning nil when the selected provider has no key configured so that
// callers can degrade gracefully to structural-only mode (per ADR-007 #3
// and #5; recorded as ADR-025).
package llmclient

import "context"

// Message is a single chat message.
type Message struct {
	Role    string `json:"role"` // "user", "assistant", or "system"
	Content string `json:"content"`
}

// Request is the input to a Client.Chat call.
type Request struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Response is the output of a Client.Chat call.
type Response struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
	// Provider is the selected provider id (anthropic, openai, gemini,
	// mistral, cerebras, openai-compatible) recorded in the response so
	// callers can surface it in their own output (e.g. the eval runner's
	// JSON report).
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// Usage reports raw per-call token counts. Dollar conversion is left to
// the operator (ADR-018 #4): the runner logs raw token usage to stderr
// via --cost-log.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Config carries the resolved provider settings for a single Client.
type Config struct {
	Provider    string  // "anthropic" | "openai" | "gemini" | "mistral" | "cerebras" | "openai-compatible"
	Model       string  // resolved model id (env override or provider default)
	BaseURL     string  // endpoint (overridden via LLM_BASE_URL when set)
	Key         string  // provider-specific API key
	Temperature float64 // default temperature for actor calls; judges pin to 0
}

// Provider is the factory interface implemented by each provider adapter.
// Implementations register themselves via the providers registry map and
// are invoked by NewFromEnv after env-driven selection.
type Provider interface {
	New(cfg Config) (Client, error)
	ID() string
}

// Client is the model-call boundary (ADR-007 #3). Every provider
// implementation satisfies this interface; callers must use NewFromEnv
// rather than constructing a specific implementation directly so that
// provider selection stays env-driven.
type Client interface {
	Chat(ctx context.Context, req Request) (*Response, error)
	Config() Config
}
