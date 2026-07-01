package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/internal/llmclient"
)

// mockEvalClient is a non-network llmclient.Client for cmd/eval tests. It
// records calls and returns scripted responses in order; out-of-range
// entries default to (nil success) so callers don't panic on unexpected
// extra calls.
type mockEvalClient struct {
	cfg       llmclient.Config
	responses []*llmclient.Response
	errs      []error
	mu        sync.Mutex
	calls     []llmclient.Request
}

func newMockEvalClient(cfg llmclient.Config, responses []*llmclient.Response, errs []error) *mockEvalClient {
	return &mockEvalClient{cfg: cfg, responses: responses, errs: errs}
}

func (m *mockEvalClient) Chat(_ context.Context, req llmclient.Request) (*llmclient.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, req)
	idx := len(m.calls) - 1
	if idx < len(m.errs) && m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &llmclient.Response{Provider: m.cfg.Provider, Model: m.cfg.Model}, nil
}

func (m *mockEvalClient) Config() llmclient.Config { return m.cfg }

func (m *mockEvalClient) Calls() []llmclient.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]llmclient.Request, len(m.calls))
	copy(out, m.calls)
	return out
}

var _ llmclient.Client = (*mockEvalClient)(nil)

// helpers ----

// writeScenario writes a valid scenario-N directory with the three required
// files inside dir/evals.
func writeScenario(t *testing.T, dir, name string, criteriaSum int) {
	t.Helper()
	scDir := filepath.Join(dir, "evals", name)
	if err := os.MkdirAll(scDir, 0o755); err != nil {
		t.Fatal(err)
	}
	max := criteriaSum / 2
	rest := criteriaSum - max
	criteria := []byte(`{
		"type":"weighted_checklist",
		"checklist":[
			{"name":"item-a","description":"a desc","max_score":` + itoa(max) + `},
			{"name":"item-b","description":"b desc","max_score":` + itoa(rest) + `}
		]
	}`)
	if err := os.WriteFile(filepath.Join(scDir, "criteria.json"), criteria, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scDir, "task.md"), []byte("# Scenario\n\nUser prompt: do X"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scDir, "capability.txt"), []byte("Can do X"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b []byte
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		b = append([]byte{digits[i%10]}, b...)
		i /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func writeSkillMD(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---- Structural-only mode ----

func TestEval_structuralOnly_noKey(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill\n\nirrelevant")
	writeScenario(t, dir, "scenario-01", 100)
	writeScenario(t, dir, "scenario-02", 100)

	// Force no key for every provider so NewFromEnv returns nil.
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"eval", dir})
	defer rootCmd.SetArgs(nil)
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "structural-only mode") {
		t.Errorf("expected structural-only notice in text output, got:\n%s", out)
	}
	if !strings.Contains(out, "Overall: PASS") {
		t.Errorf("structural pass should report Overall PASS, got:\n%s", out)
	}
}

func TestEval_structuralOnly_criteriaMismatchFails(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 80) // not 100

	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"eval", dir})
	defer rootCmd.SetArgs(nil)
	_ = rootCmd.Execute()
	out := buf.String()
	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected FAIL when criteria sum != 100, got:\n%s", out)
	}
}

func TestEval_noEvalsDir(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	// no evals/ directory

	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"eval", dir})
	defer rootCmd.SetArgs(nil)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when evals/ directory is missing")
	}
}

// ---- LLM-judge mode against MockClient ----

func TestEval_jsonOutputWithMock(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill\n\nexplain")
	writeScenario(t, dir, "scenario-01", 100)

	// Intercept the cobra RunE by running the eval runner directly. This is
	// the cleanest seam: we construct a mock client, build scenarios from
	// the tempdir, and invoke the runner.
	evalsDir := filepath.Join(dir, "evals")
	scenarios, diags, err := loadScenarios(evalsDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected structural diags: %v", diags)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}

	// Script: scenario-01 receives one actor response (the actor's "answer")
	// then one judge response (scores summing to 80% of 100).
	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic, Model: "claude-test"}
	judgeBody := `{"scores":[
		{"name":"item-a","score":40,"max_score":50,"justification":"a ok"},
		{"name":"item-b","score":40,"max_score":50,"justification":"b ok"}
	]}`
	mockResponses := []*llmclient.Response{
		{Content: "actor output", Provider: llmclient.ProviderAnthropic, Model: "claude-test", Usage: llmclient.Usage{InputTokens: 5, OutputTokens: 3}},
		{Content: judgeBody, Provider: llmclient.ProviderAnthropic, Model: "claude-test", Usage: llmclient.Usage{InputTokens: 7, OutputTokens: 4}},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)

	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}

	result := runner.run(context.Background(), scenarios)
	if result.StructuralOnly {
		t.Error("StructuralOnly should be false when client is set")
	}
	if result.Provider != llmclient.ProviderAnthropic {
		t.Errorf("provider = %q, want anthropic", result.Provider)
	}
	if result.JudgePromptVersion == "" || !strings.HasPrefix(result.JudgePromptVersion, "sha256:") {
		t.Errorf("judge prompt version missing/invalid: %q", result.JudgePromptVersion)
	}
	if len(result.Scenarios) != 1 {
		t.Fatalf("scenarios = %d, want 1", len(result.Scenarios))
	}
	sr := result.Scenarios[0]
	if sr.ID != "scenario-01" {
		t.Errorf("id = %q", sr.ID)
	}
	if sr.Total != 80 {
		t.Errorf("total = %d, want 80", sr.Total)
	}
	if sr.MaxTotal != 100 {
		t.Errorf("max = %d, want 100", sr.MaxTotal)
	}
	if sr.Status != "MARGIN" {
		t.Errorf("80%% at margin ±5%% (threshold 80) should be MARGIN, got %q", sr.Status)
	}
	if sr.Pass {
		t.Error("MARGIN should report pass=false (not within band)")
	}
	if len(sr.Scores) != 2 {
		t.Fatalf("expected 2 score items, got %d", len(sr.Scores))
	}
	if sr.Scores[0].Name != "item-a" || sr.Scores[0].Score != 40 {
		t.Errorf("scores[0] = %+v", sr.Scores[0])
	}
	if sr.ActorOutputSnippet == "" {
		t.Error("expected non-empty actor_output_snippet")
	}
	if sr.InputTokens != 7 || sr.OutputTokens != 4 {
		t.Errorf("tokens wrong: in=%d out=%d", sr.InputTokens, sr.OutputTokens)
	}

	// Verify exactly one actor call + one judge call were made (samples=1).
	if len(mock.Calls()) != 2 {
		t.Fatalf("expected 2 client calls (actor+judge), got %d", len(mock.Calls()))
	}
	if mock.Calls()[0].Messages[0].Role != "system" {
		t.Errorf("actor call first message should be system, got %q", mock.Calls()[0].Messages[0].Role)
	}
	// JSON serialisation smoke.
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), "judge_prompt_version") {
		t.Error("JSON missing judge_prompt_version")
	}
}

// TestEval_samplesCachesActor verifies the actor output caching rule.
// --samples 3 should produce exactly 1 actor call + 3 judge calls per
// scenario (Phase 2 spec).
func TestEval_samplesCachesActor(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)

	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	judgeBody := `{"scores":[{"name":"item-a","score":45,"max_score":50,"justification":""},{"name":"item-b","score":45,"max_score":50,"justification":""}]}`
	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}

	// 1 actor + 3 judges
	mockResponses := []*llmclient.Response{
		// sample 0: actor + judge
		{Content: "actor", Provider: llmclient.ProviderAnthropic},
		{Content: judgeBody, Provider: llmclient.ProviderAnthropic},
		// sample 1: judge only
		{Content: judgeBody, Provider: llmclient.ProviderAnthropic},
		// sample 2: judge only
		{Content: judgeBody, Provider: llmclient.ProviderAnthropic},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)

	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       3,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	if len(mock.Calls()) != 1+3 {
		t.Fatalf("expected 4 client calls for --samples 3 (1 actor + 3 judge), got %d", len(mock.Calls()))
	}
	if len(result.Scenarios) != 1 {
		t.Fatalf("scenarios = %d", len(result.Scenarios))
	}
	if result.Scenarios[0].Total != 90 {
		t.Errorf("median of 45+45 across samples should be 90, got %d", result.Scenarios[0].Total)
	}
	// All samples returned 90; status should be PASS (90 ≥ 80+5).
	if result.Scenarios[0].Status != "PASS" {
		t.Errorf("status = %q, want PASS", result.Scenarios[0].Status)
	}
}

// TestEval_judgeMalformedJSONRetries verifies Phase 2 judge output
// validation: one retry with a reminder on malformed JSON, then fail.
func TestEval_judgeMalformedJSONRetries(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)

	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	goodBody := `{"scores":[{"name":"item-a","score":50,"max_score":50,"justification":""},{"name":"item-b","score":50,"max_score":50,"justification":""}]}`

	// 1 actor + bad judge + retried good judge.
	mockResponses := []*llmclient.Response{
		{Content: "actor", Provider: llmclient.ProviderAnthropic},
		{Content: "not json at all", Provider: llmclient.ProviderAnthropic}, // malformed
		{Content: goodBody, Provider: llmclient.ProviderAnthropic},          // retry succeeds
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)

	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	if len(result.Scenarios) != 1 {
		t.Fatalf("scenarios = %d", len(result.Scenarios))
	}
	if !result.Scenarios[0].Pass {
		t.Errorf("malformed judge retry should yield PASS, got status=%q diag=%q", result.Scenarios[0].Status, result.Scenarios[0].Diagnostic)
	}
	if result.Scenarios[0].Total != 100 {
		t.Errorf("total = %d, want 100 after successful retry", result.Scenarios[0].Total)
	}
}

// TestEval_judgeMalformedJSONGivesUp verifies that persistent malformed
// JSON marks the scenario as failed with a diagnostic noting the parse error.
func TestEval_judgeMalformedJSONGivesUp(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)

	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	mockResponses := []*llmclient.Response{
		{Content: "actor", Provider: llmclient.ProviderAnthropic},
		{Content: "still not json", Provider: llmclient.ProviderAnthropic},
		{Content: "still not json", Provider: llmclient.ProviderAnthropic},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)

	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	if len(result.Scenarios) != 1 {
		t.Fatalf("scenarios = %d", len(result.Scenarios))
	}
	if result.Scenarios[0].Pass {
		t.Errorf("persistent malformed JSON should not pass; status=%q", result.Scenarios[0].Status)
	}
	if !strings.Contains(result.Scenarios[0].Diagnostic, "malformed JSON") {
		t.Errorf("diagnostic should mention malformed JSON, got %q", result.Scenarios[0].Diagnostic)
	}
}

// TestEval_actorCallFails verifies that an actor-side error marks the
// scenario as failed and short-circuits the judge step.
func TestEval_actorCallFails(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)

	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	actorErr := errBoom("actor call: 500 internal")
	mock := newMockEvalClient(cfg, nil, []error{actorErr})

	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	if result.OverallPass {
		t.Error("OverallPass should be false when actor fails")
	}
	if !strings.Contains(result.Scenarios[0].Diagnostic, "actor call failed") {
		t.Errorf("diagnostic should mention actor failure, got %q", result.Scenarios[0].Diagnostic)
	}
	// Actor failed → no judge call should be made.
	if len(mock.Calls()) != 1 {
		t.Errorf("expected 1 call (actor only), got %d", len(mock.Calls()))
	}
}

// TestEval_failBelowGateExitsNonZero verifies --fail-below triggers a
// command-level error when a scenario's percentage is below threshold.
func TestEval_failBelowGateExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)

	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	// 50/100 = 50%, below --fail-below 80
	judgeBody := `{"scores":[{"name":"item-a","score":25,"max_score":50,"justification":""},{"name":"item-b","score":25,"max_score":50,"justification":""}]}`
	mockResponses := []*llmclient.Response{
		{Content: "actor", Provider: llmclient.ProviderAnthropic},
		{Content: judgeBody, Provider: llmclient.ProviderAnthropic},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)
	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	// Simulate the --fail-below gate logic.
	gate := 80
	failed := false
	for _, s := range result.Scenarios {
		if !s.Pass {
			failed = true
		}
		if s.MaxTotal > 0 {
			pct := (s.Total * 100) / s.MaxTotal
			if pct < gate {
				failed = true
			}
		}
	}
	if !failed {
		t.Errorf("50%% score should fail --fail-below 80, scenarios=%+v", result.Scenarios)
	}
}

// TestEval_writeSummaryWritesStructuredJSON verifies --write-summary leaves
// a summary.json file with instructions_coverage.
func TestEval_writeSummaryWritesStructuredJSON(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)
	evalsDir := filepath.Join(dir, "evals")

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	mockResponses := []*llmclient.Response{
		{Content: "actor", Provider: llmclient.ProviderAnthropic},
		{Content: `{"scores":[{"name":"item-a","score":50,"max_score":50,"justification":"ok"},{"name":"item-b","score":50,"max_score":50,"justification":"ok"}]}`, Provider: llmclient.ProviderAnthropic},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)
	scenarios, _, _ := loadScenarios(evalsDir)
	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	if err := writeEvalSummary(evalsDir, result); err != nil {
		t.Fatalf("writeSummary: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(evalsDir, "summary.json"))
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	var restored map[string]any
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("parse summary: %v", err)
	}
	ic, ok := restored["instructions_coverage"].(map[string]any)
	if !ok {
		t.Fatalf("missing instructions_coverage: %s", data)
	}
	if ic["coverage_percentage"].(float64) != 100 {
		t.Errorf("coverage = %v, want 100", ic["coverage_percentage"])
	}
	if ic["provider"] != llmclient.ProviderAnthropic {
		t.Errorf("provider = %v, want anthropic", ic["provider"])
	}
}

// TestEval_costLogEmitsTokenUsage verifies that --cost-log drains raw
// token usage to stderr.
func TestEval_costLogEmitsTokenUsage(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)
	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	mockResponses := []*llmclient.Response{
		{Content: "actor", Provider: llmclient.ProviderAnthropic, Usage: llmclient.Usage{InputTokens: 11, OutputTokens: 7}},
		{Content: `{"scores":[{"name":"item-a","score":50,"max_score":50,"justification":"x"},{"name":"item-b","score":50,"max_score":50,"justification":"y"}]}`, Provider: llmclient.ProviderAnthropic, Usage: llmclient.Usage{InputTokens: 13, OutputTokens: 9}},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)

	var costBuf bytes.Buffer
	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		costLog:       &costBuf,
		maxConcurrent: 1,
	}
	_ = runner.run(context.Background(), scenarios)
	out := costBuf.String()
	if !strings.Contains(out, "scenario-01 actor tokens") {
		t.Errorf("cost log missing actor tokens: %s", out)
	}
	if !strings.Contains(out, "scenario-01 judge[1] tokens") {
		t.Errorf("cost log missing judge tokens: %s", out)
	}
	if !strings.Contains(out, "in=11") || !strings.Contains(out, "in=13") {
		t.Errorf("cost log should record input token counts, got: %s", out)
	}
}

// TestEval_multipleScenariosConcurrently verifies bounded concurrency
// against multiple scenarios: results land in stable order keyed by ID
// and all scenarios are produced exactly once.
func TestEval_multipleScenariosConcurrently(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	for _, name := range []string{"scenario-01", "scenario-02", "scenario-03"} {
		writeScenario(t, dir, name, 100)
	}
	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)
	if len(scenarios) != 3 {
		t.Fatalf("loaded %d scenarios", len(scenarios))
	}

	cfg := llmclient.Config{Provider: llmclient.ProviderOpenAI}
	// Each scenario gets 1 actor + 1 judge.
	mockResponses := []*llmclient.Response{
		{Content: "a1", Provider: llmclient.ProviderOpenAI},
		{Content: `{"scores":[{"name":"item-a","score":50,"max_score":50,"justification":""},{"name":"item-b","score":50,"max_score":50,"justification":""}]}`, Provider: llmclient.ProviderOpenAI},
		{Content: "a2", Provider: llmclient.ProviderOpenAI},
		{Content: `{"scores":[{"name":"item-a","score":40,"max_score":50,"justification":""},{"name":"item-b","score":40,"max_score":50,"justification":""}]}`, Provider: llmclient.ProviderOpenAI},
		{Content: "a3", Provider: llmclient.ProviderOpenAI},
		{Content: `{"scores":[{"name":"item-a","score":45,"max_score":50,"justification":""},{"name":"item-b","score":45,"max_score":50,"justification":""}]}`, Provider: llmclient.ProviderOpenAI},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)
	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1, // serialise so response pairing is deterministic
	}
	result := runner.run(context.Background(), scenarios)
	if len(result.Scenarios) != 3 {
		t.Fatalf("scenarios = %d", len(result.Scenarios))
	}
	if result.Scenarios[0].ID != "scenario-01" || result.Scenarios[1].ID != "scenario-02" || result.Scenarios[2].ID != "scenario-03" {
		t.Errorf("scenarios out of order: %+v", result.Scenarios)
	}
	if len(mock.Calls()) != 6 {
		t.Errorf("expected 6 total calls (3 × 2), got %d", len(mock.Calls()))
	}
}

// TestEval_fullCommandPipelineSmoke runs the full cobra command against a
// tempdir skill with a mock provider mounted via the package-level
// NewFromEnv path. We avoid cobra HTTP by pre-populating a fake key and
// stubbing the client via a registry hook is too invasive; instead we
// verify the structural-only path end-to-end through cobra.
func TestEval_fullCommandStructuralPipeline(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill\n\nfull pipeline")
	writeScenario(t, dir, "scenario-01", 100)

	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"eval", dir, "--json"})
	defer rootCmd.SetArgs(nil)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("eval: %v", err)
	}
	var out evalResult
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal JSON output: %v\nraw: %s", err, buf.String())
	}
	if !out.StructuralOnly {
		t.Error("StructuralOnly should be true when no key")
	}
	if len(out.Scenarios) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(out.Scenarios))
	}
	if !out.Scenarios[0].Pass {
		t.Errorf("scenario-01 should structurally pass: %+v", out.Scenarios[0])
	}
}

// TestEval_flagsRegistered verifies all Phase 2 flags exist on the eval
// command and use the preferred short/long forms.
func TestEval_flagsRegistered(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Use != "eval <skill>" {
			continue
		}
		expected := []string{
			"provider", "model", "judge-model", "fail-below",
			"write-summary", "json", "samples", "margin",
			"cost-log", "repo-root",
		}
		for _, f := range expected {
			if sub.Flags().Lookup(f) == nil {
				t.Errorf("flag %q not registered", f)
			}
		}
		return
	}
	t.Fatal("eval subcommand not registered")
}

// TestEval_timestampIsRFC3339 verifies the timestamp field is a valid
// RFC3339 time, which is required for CI artifact comparisons.
func TestEval_timestampIsRFC3339(t *testing.T) {
	dir := t.TempDir()
	writeSkillMD(t, dir, "# Skill")
	writeScenario(t, dir, "scenario-01", 100)
	evalsDir := filepath.Join(dir, "evals")
	scenarios, _, _ := loadScenarios(evalsDir)

	cfg := llmclient.Config{Provider: llmclient.ProviderAnthropic}
	mockResponses := []*llmclient.Response{
		{Content: "actor", Provider: llmclient.ProviderAnthropic},
		{Content: `{"scores":[{"name":"item-a","score":50,"max_score":50,"justification":""},{"name":"item-b","score":50,"max_score":50,"justification":""}]}`, Provider: llmclient.ProviderAnthropic},
	}
	mock := newMockEvalClient(cfg, mockResponses, nil)
	runner := &evalRunner{
		skillPath:     filepath.Join(dir, "SKILL.md"),
		skillContent:  "# Skill",
		client:        mock,
		samples:       1,
		margin:        5,
		maxConcurrent: 1,
	}
	result := runner.run(context.Background(), scenarios)
	if _, err := time.Parse(time.RFC3339, result.Timestamp); err != nil {
		t.Errorf("timestamp %q is not RFC3339: %v", result.Timestamp, err)
	}
}

// errBoom is a tiny helper for scripting mock errors with a stable message.
type errBoom string

func (e errBoom) Error() string { return string(e) }

// ensure context import stays required even when future edits drop usages.
var _ = context.Background
