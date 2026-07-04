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

// AnthropicClient implements Client against the Anthropic Messages API
// (default model claude-sonnet-4-6, key ANTHROPIC_API_KEY).
type AnthropicClient struct {
	cfg  Config
	http *http.Client
}

func init() {
	register(anthropicFactory{})
}

type anthropicFactory struct{}

func (anthropicFactory) ID() string { return ProviderAnthropic }

func (anthropicFactory) New(cfg Config) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseAnthropic
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelAnthropic
	}
	return &AnthropicClient{cfg: cfg, http: DefaultHTTPClient}, nil
}

func (c *AnthropicClient) Config() Config { return c.cfg }

// anthropicRequest is the Messages API request body.
type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *anthropicError `json:"error,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *AnthropicClient) Chat(ctx context.Context, req Request) (*Response, error) {
	if c.cfg.Key == "" {
		return nil, errors.New("anthropic client has no API key")
	}
	model := req.Model
	if model == "" {
		model = c.cfg.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	sys, msgs := splitSystem(req.Messages)
	body, err := json.Marshal(anthropicRequest{
		Model:       model,
		Messages:    msgs,
		System:      sys,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	url := c.cfg.BaseURL + "/v1/messages"
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		r.Header.Set("x-api-key", c.cfg.Key)
		r.Header.Set("anthropic-version", "2023-06-01")
		r.Header.Set("content-type", "application/json")

		resp, err = c.http.Do(r)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		// Retryable: 429 or 5xx.
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
		return nil, fmt.Errorf("anthropic api status %d: %s", resp.StatusCode, string(raw))
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ar anthropicResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return nil, fmt.Errorf("parse anthropic response: %w", err)
	}
	if ar.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s: %s", ar.Error.Type, ar.Error.Message)
	}

	out := &Response{Provider: ProviderAnthropic, Model: model, Usage: Usage{
		InputTokens:  ar.Usage.InputTokens,
		OutputTokens: ar.Usage.OutputTokens,
	}}
	// Anthropic returns content as a list of blocks; concatenate text blocks.
	var b []byte
	for _, block := range ar.Content {
		b = append(b, block.Text...)
	}
	out.Content = string(b)
	return out, nil
}

// splitSystem pulls the leading system message out of the messages list so
// it can be passed via Anthropic's top-level "system" field. Non-system
// messages are passed through unchanged.
func splitSystem(msgs []Message) (string, []anthropicMessage) {
	var system string
	var out []anthropicMessage
	for _, m := range msgs {
		if m.Role == "system" {
			if system != "" {
				system += "\n"
			}
			system += m.Content
			continue
		}
		out = append(out, anthropicMessage(m))
	}
	return system, out
}

// timeAfter is a package-level var so tests can stub timing. Default uses
// the standard library time.After.
var timeAfter = defaultAfter
