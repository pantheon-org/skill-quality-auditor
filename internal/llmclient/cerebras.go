package llmclient

// Cerebras's Chat Completions API is explicitly documented as OpenAI-wire-
// compatible (inference-docs.cerebras.ai/resources/openai), so this adapter
// reuses OpenAIClient rather than duplicating the HTTP/JSON logic — the
// same pattern openai.go already uses for the openai-compatible provider.

func init() {
	register(cerebrasFactory{})
}

type cerebrasFactory struct{}

func (cerebrasFactory) ID() string { return ProviderCerebras }

func (cerebrasFactory) New(cfg Config) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseCerebras
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelCerebras
	}
	return &OpenAIClient{cfg: cfg, http: DefaultHTTPClient}, nil
}
