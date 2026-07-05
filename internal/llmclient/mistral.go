package llmclient

// Mistral's Chat Completions API is OpenAI-wire-compatible (same
// /v1/chat/completions request/response shape), so this adapter reuses
// OpenAIClient rather than duplicating the HTTP/JSON logic — the same
// pattern openai.go already uses for the openai-compatible provider.

func init() {
	register(mistralFactory{})
}

type mistralFactory struct{}

func (mistralFactory) ID() string { return ProviderMistral }

func (mistralFactory) New(cfg Config) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseMistral
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelMistral
	}
	return &OpenAIClient{cfg: cfg, http: DefaultHTTPClient}, nil
}
