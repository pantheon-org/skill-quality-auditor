package llmclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---- Mock client for use by callers (cmd/eval_test.go) ----

// MockClient is a non-network Client used by tests of code that depends
// on llmclient.Client. It records calls and returns scripted responses in
// order; a nil Response triggers the configured error.
type MockClient struct {
	cfg       Config
	responses []*Response
	errs      []error
	calls     []Request
}

// NewMock constructs a mock with the given scripted responses and errors.
// responses and errs are pulled by index; out-of-range entries default to
// (nil, nil) so the call returns nothing without panicking.
func NewMock(cfg Config, responses []*Response, errs []error) *MockClient {
	return &MockClient{cfg: cfg, responses: responses, errs: errs}
}

func (m *MockClient) Chat(_ context.Context, req Request) (*Response, error) {
	m.calls = append(m.calls, req)
	idx := len(m.calls) - 1
	if idx < len(m.errs) && m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &Response{Provider: m.cfg.Provider, Model: m.cfg.Model}, nil
}

func (m *MockClient) Config() Config { return m.cfg }

// Calls returns a copy of recorded requests (actor + judge, in order).
func (m *MockClient) Calls() []Request {
	out := make([]Request, len(m.calls))
	copy(out, m.calls)
	return out
}

// ---- Provider selection ----

func TestNewFromEnv_unknownProvider(t *testing.T) {
	_, err := NewFromEnv("not-a-provider")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), "not-a-provider") {
		t.Errorf("error should mention bad id, got %q", err)
	}
	if !strings.Contains(err.Error(), "anthropic") {
		t.Errorf("error should list valid providers, got %q", err)
	}
}

func TestNewFromEnv_defaultIsAnthropic(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	c, err := NewFromEnv("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client when ANTHROPIC_API_KEY is set")
	}
	if c.Config().Provider != ProviderAnthropic {
		t.Errorf("provider = %q, want anthropic", c.Config().Provider)
	}
	if c.Config().Model != DefaultModelAnthropic {
		t.Errorf("model = %q, want %q", c.Config().Model, DefaultModelAnthropic)
	}
}

func TestNewFromEnv_noKeyReturnsNil(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	c, err := NewFromEnv("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != nil {
		t.Errorf("expected nil client when no key is set, got %T", c)
	}
}

func TestNewFromEnv_openAIProvider(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-oai")
	c, err := NewFromEnv(ProviderOpenAI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client for openai provider with key")
	}
	if c.Config().Provider != ProviderOpenAI {
		t.Errorf("provider = %q, want openai", c.Config().Provider)
	}
	if c.Config().Model != DefaultModelOpenAI {
		t.Errorf("model = %q, want %q", c.Config().Model, DefaultModelOpenAI)
	}
}

func TestNewFromEnv_geminiFallsBackToGoogleKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "g-key")
	c, err := NewFromEnv(ProviderGemini)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client when GOOGLE_API_KEY is set")
	}
	if c.Config().Key != "g-key" {
		t.Errorf("key = %q, want g-key", c.Config().Key)
	}
}

func TestNewFromEnv_openAICompatibleRequiresBaseURL(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	t.Setenv("LLM_BASE_URL", "")
	_, err := NewFromEnv(ProviderOpenAICompatible)
	if err == nil {
		t.Fatal("expected error when openai-compatible has no LLM_BASE_URL")
	}
}

func TestNewFromEnv_openAICompatibleWithBaseURL(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	t.Setenv("LLM_BASE_URL", "http://localhost:11434/v1")
	c, err := NewFromEnv(ProviderOpenAICompatible)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.Config().Provider != ProviderOpenAICompatible {
		t.Errorf("provider = %q, want openai-compatible", c.Config().Provider)
	}
	if c.Config().BaseURL != "http://localhost:11434/v1" {
		t.Errorf("base = %q, want http://localhost:11434/v1", c.Config().BaseURL)
	}
}

func TestNewFromEnv_noPartialAuthFallback(t *testing.T) {
	// LLM_PROVIDER=openai with only ANTHROPIC_API_KEY set must degrade
	// (return nil), not silently fall back to anthropic.
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "sk-anth")
	c, err := NewFromEnv(ProviderOpenAI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != nil {
		t.Errorf("expected nil (no OpenAI key); got %T provider=%q", c, c.Config().Provider)
	}
}

func TestNewFromEnv_modelAndBaseURLOverrides(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-x")
	t.Setenv("LLM_MODEL", "claude-opus-4-20250514")
	t.Setenv("LLM_BASE_URL", "https://gateway.example.com")
	c, err := NewFromEnv(ProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Config().Model != "claude-opus-4-20250514" {
		t.Errorf("model override not honoured, got %q", c.Config().Model)
	}
	if c.Config().BaseURL != "https://gateway.example.com" {
		t.Errorf("base url override not honoured, got %q", c.Config().BaseURL)
	}
}

// ---- Anthropic adapter against httptest ----

func TestAnthropicClient_wrongStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":"bad"}}`))
	}))
	defer srv.Close()

	c, err := MustNew(Config{Provider: ProviderAnthropic, Key: "sk", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}, MaxTokens: 8})
	if err == nil {
		t.Fatal("expected error from 400 status")
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Errorf("error should mention status, got %q", err)
	}
}

func TestAnthropicClient_happyPath(t *testing.T) {
	var capturedBody []byte
	var capturedAuth string
	var capturedVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("x-api-key")
		capturedVersion = r.Header.Get("anthropic-version")
		capturedBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{
			"content":[{"type":"text","text":"hello judge"}],
			"usage":{"input_tokens":7,"output_tokens":2}
		}`))
	}))
	defer srv.Close()

	c, err := MustNew(Config{Provider: ProviderAnthropic, Key: "sk-abc", BaseURL: srv.URL, Model: "claude-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := c.Chat(context.Background(), Request{
		Messages:    []Message{{Role: "system", Content: "be brief"}, {Role: "user", Content: "ping"}},
		MaxTokens:   16,
		Temperature: 0,
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Content != "hello judge" {
		t.Errorf("content = %q, want 'hello judge'", resp.Content)
	}
	if resp.Provider != ProviderAnthropic {
		t.Errorf("provider = %q, want anthropic", resp.Provider)
	}
	if resp.Usage.InputTokens != 7 || resp.Usage.OutputTokens != 2 {
		t.Errorf("usage = %+v", resp.Usage)
	}
	if capturedAuth != "sk-abc" {
		t.Errorf("auth header = %q, want sk-abc", capturedAuth)
	}
	if capturedVersion == "" {
		t.Error("anthropic-version header not set")
	}
	var body anthropicRequest
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("parse body: %v", err)
	}
	if body.System != "be brief" {
		t.Errorf("system field = %q, want 'be brief'", body.System)
	}
	if len(body.Messages) != 1 || body.Messages[0].Role != "user" || body.Messages[0].Content != "ping" {
		t.Errorf("messages = %+v", body.Messages)
	}
	if body.Model != "claude-test" {
		t.Errorf("model = %q, want claude-test", body.Model)
	}
}

// ---- OpenAI adapter against httptest ----

func TestOpenAIClient_happyPath(t *testing.T) {
	var capturedAuth string
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		capturedBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"hi back"}}],
			"usage":{"prompt_tokens":10,"completion_tokens":2}
		}`))
	}))
	defer srv.Close()

	c, err := MustNew(Config{Provider: ProviderOpenAI, Key: "oai-key", BaseURL: srv.URL, Model: "gpt-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := c.Chat(context.Background(), Request{
		Messages:  []Message{{Role: "user", Content: "hi"}},
		MaxTokens: 16,
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Content != "hi back" {
		t.Errorf("content = %q, want 'hi back'", resp.Content)
	}
	if resp.Provider != ProviderOpenAI {
		t.Errorf("provider = %q, want openai", resp.Provider)
	}
	if !strings.HasPrefix(capturedAuth, "Bearer oai-key") {
		t.Errorf("auth = %q, want Bearer oai-key", capturedAuth)
	}
	var body openAIRequest
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("parse body: %v", err)
	}
	if body.Messages[0].Content != "hi" {
		t.Errorf("content = %q, want 'hi'", body.Messages[0].Content)
	}
}

func TestOpenAIClient_compatibleProviderIDReported(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"x"}}],"usage":{}}`))
	}))
	defer srv.Close()

	c, err := MustNew(Config{Provider: ProviderOpenAICompatible, Key: "k", BaseURL: srv.URL, Model: "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := c.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Provider != ProviderOpenAICompatible {
		t.Errorf("provider = %q, want openai-compatible", resp.Provider)
	}
}

// ---- Gemini adapter against httptest ----

func TestGeminiClient_happyPath(t *testing.T) {
	var capturedKey string
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedKey = r.Header.Get("x-goog-api-key")
		capturedBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{
			"candidates":[{"content":{"parts":[{"text":"gemini reply"}]}}],
			"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":3}
		}`))
	}))
	defer srv.Close()

	c, err := MustNew(Config{Provider: ProviderGemini, Key: "gem-key", BaseURL: srv.URL, Model: "gemini-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := c.Chat(context.Background(), Request{
		Messages: []Message{
			{Role: "system", Content: "sys"},
			{Role: "user", Content: "u"},
			{Role: "assistant", Content: "a"},
			{Role: "user", Content: "u2"},
		},
		MaxTokens: 12,
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Content != "gemini reply" {
		t.Errorf("content = %q, want 'gemini reply'", resp.Content)
	}
	if resp.Provider != ProviderGemini {
		t.Errorf("provider = %q, want gemini", resp.Provider)
	}
	if capturedKey != "gem-key" {
		t.Errorf("key header = %q, want gem-key", capturedKey)
	}
	// The full URL (path containing "/v1beta/models/<model>:generateContent")
	// is captured on the server side in r.URL.Path; we cannot observe it from
	// this side of httptest, so no inline assertion is made here.
	var body geminiRequest
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("parse body: %v", err)
	}
	if body.SystemInstruction == nil || body.SystemInstruction.Parts[0].Text != "sys" {
		t.Errorf("systemInstruction = %+v", body.SystemInstruction)
	}
	if len(body.Contents) != 3 {
		t.Fatalf("expected 3 user/model contents, got %d", len(body.Contents))
	}
	if body.Contents[1].Role != "model" {
		t.Errorf("assistant should be remapped to 'model' role, got %q", body.Contents[1].Role)
	}
}

func TestGeminiClient_wrongStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":500,"message":"boom","status":"INTERNAL"}}`))
	}))
	defer srv.Close()

	c, err := MustNew(Config{Provider: ProviderGemini, Key: "k", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "x"}}})
	if err == nil {
		t.Fatal("expected error from 500 status")
	}
}

// ---- Retry behaviour (drives the shared retry path) ----

func TestAnthropicClient_retriesOn5xx(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("slow down"))
			return
		}
		_, _ = w.Write([]byte(`{"content":[{"text":"ok"}],"usage":{}}`))
	}))
	defer srv.Close()

	// Make Backoff effectively zero so the test runs fast.
	origAfter := timeAfter
	timeAfter = func(_ time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		close(ch)
		return ch
	}
	defer func() { timeAfter = origAfter }()

	c, err := MustNew(Config{Provider: ProviderAnthropic, Key: "sk", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := c.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatalf("expected eventual success, got error: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("content = %q, want ok", resp.Content)
	}
	if calls != 3 {
		t.Errorf("expected 3 attempts (2 retries), got %d", calls)
	}
}

func TestAnthropicClient_givesUpAfterThreeAttempts(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	origAfter := timeAfter
	timeAfter = func(_ time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		close(ch)
		return ch
	}
	defer func() { timeAfter = origAfter }()

	c, err := MustNew(Config{Provider: ProviderAnthropic, Key: "sk", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "x"}}})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 3 {
		t.Errorf("expected 3 attempts, got %d", calls)
	}
}

// ---- Mock client behaviour ----

func TestMockClient_recordsCallsAndScriptsResponses(t *testing.T) {
	mock := NewMock(Config{Provider: ProviderAnthropic, Model: "m"},
		[]*Response{{Content: "actor-output", Provider: ProviderAnthropic, Model: "m"}},
		nil)
	got, err := mock.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "u"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != "actor-output" {
		t.Errorf("content = %q", got.Content)
	}
	if len(mock.Calls()) != 1 {
		t.Errorf("calls = %d, want 1", len(mock.Calls()))
	}
}

func TestMockClient_returnsScriptedError(t *testing.T) {
	mock := NewMock(Config{Provider: ProviderAnthropic}, nil, []error{errors.New("boom")})
	_, err := mock.Chat(context.Background(), Request{})
	if err == nil || err.Error() != "boom" {
		t.Errorf("want boom error, got %v", err)
	}
}

func TestMockClient_defaultResponseWhenOutOfScripts(t *testing.T) {
	mock := NewMock(Config{Provider: ProviderOpenAI, Model: "gpt-test"}, nil, nil)
	got, err := mock.Chat(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Model != "gpt-test" || got.Provider != ProviderOpenAI {
		t.Errorf("default response shape wrong: %+v", got)
	}
}

// ---- Prompt building & judge parsing ----

func TestParseJudgeResponse_directJSON(t *testing.T) {
	raw := `{"scores":[{"name":"a","score":10,"max_score":10,"justification":"x"}]}`
	r, err := ParseJudgeResponse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(r.Scores) != 1 || r.Scores[0].Score != 10 {
		t.Errorf("parsed wrong: %+v", r)
	}
}

func TestParseJudgeResponse_withPrefaceAndTrailer(t *testing.T) {
	raw := `Here is the result: {"scores":[{"name":"a","score":5,"max_score":10,"justification":"half"}]} done.`
	r, err := ParseJudgeResponse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if r.Scores[0].Score != 5 {
		t.Errorf("score wrong: %+v", r)
	}
}

func TestParseJudgeResponse_missingScores(t *testing.T) {
	_, err := ParseJudgeResponse(`{"foo":1}`)
	if err == nil {
		t.Fatal("expected error when scores missing")
	}
}

func TestParseJudgeResponse_empty(t *testing.T) {
	_, err := ParseJudgeResponse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseJudgeResponse_noJSON(t *testing.T) {
	_, err := ParseJudgeResponse("hello world no braces")
	if err == nil {
		t.Fatal("expected error when no JSON object present")
	}
}

func TestPromptVersion_stableAndPrefixed(t *testing.T) {
	v := PromptVersion()
	if !strings.HasPrefix(v, "sha256:") {
		t.Fatalf("version should be sha256-prefixed, got %q", v)
	}
	if v != PromptVersion() {
		t.Error("PromptVersion should be deterministic within a run")
	}
}

func TestActorMessages_includesSystemAndUser(t *testing.T) {
	msgs := ActorMessages("SKILL", "TASK")
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" || !strings.Contains(msgs[0].Content, "SKILL") {
		t.Errorf("system message wrong: %+v", msgs[0])
	}
	if msgs[1].Role != "user" || msgs[1].Content != "TASK" {
		t.Errorf("user message wrong: %+v", msgs[1])
	}
}

func TestJudgeMessages_includesAllSections(t *testing.T) {
	criteria := []byte(`{"checklist":[{"name":"A","description":"a desc","max_score":10},{"name":"B","description":"b desc","max_score":20}]}`)
	msgs, err := JudgeMessages("SKILLCONTENT", "TASKPROMPT", "ACTOROUTPUT", criteria)
	if err != nil {
		t.Fatalf("judge messages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	s := msgs[0].Content
	for _, want := range []string{"SKILL CONTENT START", "SKILLCONTENT", "TASK PROMPT START", "TASKPROMPT", "AGENT OUTPUT START", "ACTOROUTPUT", "CHECKLIST START", `A`, `B`, "max_score=10", "max_score=20"} {
		if !strings.Contains(s, want) {
			t.Errorf("judge message missing %q", want)
		}
	}
}

func TestJudgeMessages_invalidCriteria(t *testing.T) {
	_, err := JudgeMessages("", "", "", []byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid criteria JSON")
	}
}

func TestMaxOutputTokensFromCriteria_default(t *testing.T) {
	if got := MaxOutputTokensFromCriteria([]byte(`{}`), 4096); got != 4096 {
		t.Errorf("default not returned, got %d", got)
	}
	if got := MaxOutputTokensFromCriteria([]byte(`{"max_output_tokens":1024}`), 4096); got != 1024 {
		t.Errorf("override not returned, got %d", got)
	}
	if got := MaxOutputTokensFromCriteria([]byte("bad"), 4096); got != 4096 {
		t.Errorf("bad json should return default, got %d", got)
	}
}

// ---- Fixture-driven test using a real evals/ directory ----

func TestJudgeMessages_againstFixtureScenario(t *testing.T) {
	dir := t.TempDir()
	scDir := filepath.Join(dir, "scenario-01")
	if err := os.MkdirAll(scDir, 0o755); err != nil {
		t.Fatal(err)
	}
	criteria := []byte(`{
		"type":"weighted_checklist",
		"checklist":[
			{"name":"x","description":"x desc","max_score":50},
			{"name":"y","description":"y desc","max_score":50}
		]
	}`)
	if err := os.WriteFile(filepath.Join(scDir, "criteria.json"), criteria, 0o644); err != nil {
		t.Fatal(err)
	}
	msgs, err := JudgeMessages("SKILL", "TASK", "OUTPUT", criteria)
	if err != nil {
		t.Fatalf("judge messages: %v", err)
	}
	if !strings.Contains(msgs[0].Content, `max_score=50`) {
		t.Error("max_score not embedded in judge prompt")
	}
}

// ---- Helper tests ----

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		status int
		want   bool
	}{
		{200, false}, {400, false}, {401, false},
		{429, true}, {500, true}, {503, true}, {599, true}, {600, false},
	}
	for _, c := range cases {
		if got := IsRetryable(c.status); got != c.want {
			t.Errorf("IsRetryable(%d) = %v, want %v", c.status, got, c.want)
		}
	}
}

func TestBackoff_monotonicAndCapped(t *testing.T) {
	prev := Backoff(0)
	for i := 1; i < 20; i++ {
		cur := Backoff(i)
		if cur <= 0 {
			t.Errorf("backoff must be positive at attempt %d", i)
		}
		// Should never exceed cap.
		if cur > 8*time.Second {
			t.Errorf("backoff at attempt %d exceeds cap: %v", i, cur)
		}
		if cur < prev {
			// Only allowed if we've hit the cap.
			if prev != 8*time.Second {
				t.Errorf("backoff decreased before cap: attempt %d got %v (prev %v)", i, cur, prev)
			}
		}
		prev = cur
	}
}
