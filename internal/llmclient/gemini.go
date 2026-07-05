package llmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GeminiClient implements Client against the native Google generative
// language API generateContent endpoint (default model gemini-2.0-flash,
// key GEMINI_API_KEY with fallback to GOOGLE_API_KEY).
type GeminiClient struct {
	cfg  Config
	http *http.Client
}

func init() {
	register(geminiFactory{})
}

type geminiFactory struct{}

func (geminiFactory) ID() string { return ProviderGemini }

func (geminiFactory) New(cfg Config) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseGemini
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelGemini
	}
	return &GeminiClient{cfg: cfg, http: DefaultHTTPClient}, nil
}

func (c *GeminiClient) Config() Config { return c.cfg }

// geminiRequest is the generateContent request body.
type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	GenerationConfig  geminiGenConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *geminiError `json:"error,omitempty"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// buildGeminiRequest translates the provider-agnostic Request into Gemini's
// generateContent body, remapping "assistant" to "model" and pulling any
// "system" message out into the top-level systemInstruction field.
func buildGeminiRequest(req Request) geminiRequest {
	var greq geminiRequest
	greq.GenerationConfig = geminiGenConfig{
		MaxOutputTokens: req.MaxTokens,
		Temperature:     req.Temperature,
	}
	for _, m := range req.Messages {
		role := m.Role
		switch role {
		case "assistant":
			role = "model"
		case "system":
			greq.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: m.Content}}}
			continue
		}
		greq.Contents = append(greq.Contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}
	return greq
}

func (c *GeminiClient) Chat(ctx context.Context, req Request) (*Response, error) {
	if c.cfg.Key == "" {
		return nil, errors.New("gemini client has no API key")
	}
	model := req.Model
	if model == "" {
		model = c.cfg.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	body, err := json.Marshal(buildGeminiRequest(req))
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	endpoint := c.cfg.BaseURL + "/v1beta/models/" + url.PathEscape(model) + ":generateContent"
	var resp *http.Response
	for attempt := 0; attempt < MaxRetryAttempts; attempt++ {
		r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		// Native Google API auths via "x-goog-api-key" header; URL alt is acceptable too.
		r.Header.Set("x-goog-api-key", c.cfg.Key)
		r.Header.Set("content-type", "application/json")

		resp, err = c.http.Do(r)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		if IsRetryable(resp.StatusCode) && attempt < MaxRetryAttempts-1 {
			wait := retryDelay(resp.StatusCode, resp.Header, attempt)
			_ = resp.Body.Close()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timeAfter(wait):
			}
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("gemini api status %d: %s", resp.StatusCode, string(raw))
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var gr geminiResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return nil, fmt.Errorf("parse gemini response: %w", err)
	}
	if gr.Error != nil {
		return nil, fmt.Errorf("gemini error: %s (code %d): %s", gr.Error.Status, gr.Error.Code, gr.Error.Message)
	}

	out := &Response{Provider: ProviderGemini, Model: model}
	out.Usage = Usage{
		InputTokens:  gr.UsageMetadata.PromptTokenCount,
		OutputTokens: gr.UsageMetadata.CandidatesTokenCount,
	}
	for _, cand := range gr.Candidates {
		for _, p := range cand.Content.Parts {
			out.Content += p.Text
		}
	}
	return out, nil
}
