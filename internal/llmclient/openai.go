package llmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// OpenAIClient implements Client against the OpenAI Chat Completions API
// (default model gpt-4o, key OPENAI_API_KEY). Honours LLM_BASE_URL so it
// also serves as the openai-compatible adapter for Ollama, vLLM, and
// internal gateways.
type OpenAIClient struct {
	cfg  Config
	http *http.Client
}

func init() {
	register(openAIFactory{})
	register(openAICompatibleFactory{})
}

type openAIFactory struct{}

func (openAIFactory) ID() string { return ProviderOpenAI }

func (openAIFactory) New(cfg Config) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseOpenAI
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelOpenAI
	}
	return &OpenAIClient{cfg: cfg, http: DefaultHTTPClient}, nil
}

type openAICompatibleFactory struct{}

func (openAICompatibleFactory) ID() string { return ProviderOpenAICompatible }

func (openAICompatibleFactory) New(cfg Config) (Client, error) {
	if cfg.BaseURL == "" {
		// NewFromEnv should have errored earlier; defend in depth.
		return nil, errors.New("openai-compatible provider requires LLM_BASE_URL")
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelOpenAI
	}
	return &OpenAIClient{cfg: cfg, http: DefaultHTTPClient}, nil
}

func (c *OpenAIClient) Config() Config { return c.cfg }

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *openAIError `json:"error,omitempty"`
}

type openAIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *OpenAIClient) Chat(ctx context.Context, req Request) (*Response, error) {
	if c.cfg.Key == "" {
		return nil, errors.New("openai client has no API key")
	}
	model := req.Model
	if model == "" {
		model = c.cfg.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	body, err := json.Marshal(openAIRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	url := c.cfg.BaseURL + "/v1/chat/completions"
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		r.Header.Set("Authorization", "Bearer "+c.cfg.Key)
		r.Header.Set("content-type", "application/json")

		resp, err = c.http.Do(r)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		if IsRetryable(resp.StatusCode) && attempt < 2 {
			_ = resp.Body.Close()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timeAfter(Backoff(attempt)):
			}
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("openai api status %d: %s", resp.StatusCode, string(raw))
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var or openAIResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return nil, fmt.Errorf("parse openai response: %w", err)
	}
	if or.Error != nil {
		return nil, fmt.Errorf("openai error: %s: %s", or.Error.Type, or.Error.Message)
	}

	out := &Response{Provider: c.cfg.Provider, Model: model}
	out.Usage = Usage{
		InputTokens:  or.Usage.PromptTokens,
		OutputTokens: or.Usage.CompletionTokens,
	}
	for _, ch := range or.Choices {
		out.Content += ch.Message.Content
	}
	return out, nil
}
