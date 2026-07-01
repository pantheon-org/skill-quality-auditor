package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/internal/llmclient"
	"github.com/spf13/cobra"
)

// evalResult is the JSON shape produced by the eval runner. See Phase 2 of
// the native-eval-runner plan for the contract.
type evalResult struct {
	SkillPath          string           `json:"skill_path"`
	Provider           string           `json:"provider,omitempty"`
	Model              string           `json:"model,omitempty"`
	JudgeModel         string           `json:"judge_model,omitempty"`
	JudgePromptVersion string           `json:"judge_prompt_version,omitempty"`
	Samples            int              `json:"samples"`
	Timestamp          string           `json:"timestamp"`
	Scenarios          []scenarioResult `json:"scenarios"`
	OverallPass        bool             `json:"overall_pass"`
	StructuralOnly     bool             `json:"structural_only,omitempty"`
}

type scenarioResult struct {
	ID                 string                `json:"id"`
	Capability         string                `json:"capability,omitempty"`
	ActorOutputSnippet string                `json:"actor_output_snippet,omitempty"`
	Scores             []llmclient.JudgeItem `json:"scores,omitempty"`
	Total              int                   `json:"total"`
	MaxTotal           int                   `json:"max_total,omitempty"`
	Pass               bool                  `json:"pass"`
	Status             string                `json:"status,omitempty"` // PASS/MARGIN/PASS/FAIL or "structural"
	Diagnostic         string                `json:"diagnostic,omitempty"`
	InputTokens        int                   `json:"input_tokens,omitempty"`
	OutputTokens       int                   `json:"output_tokens,omitempty"`
}

func init() {
	var flags struct {
		provider     string
		model        string
		judgeModel   string
		failBelow    int
		writeSummary bool
		asJSON       bool
		samples      int
		margin       int
		costLog      bool
		repoRoot     string
	}

	cmd := &cobra.Command{
		Use:   "eval <skill>",
		Short: "Run eval scenarios against an LLM (native eval runner)",
		Long: "Runs the eval scenarios under <skill>/evals/ against an LLM provider, " +
			"grading actor outputs with an LLM judge. With no provider key in the " +
			"environment the runner degrades to structural-only mode (schema " +
			"consistency gate, no semantic grading).",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepoRoot(flags.repoRoot)
			if err != nil {
				return fmt.Errorf("cannot determine repo root: %w", err)
			}
			skillPath := resolveSkillPath(args[0], repoRoot)
			skillDir := filepath.Dir(skillPath)
			evalsDir := filepath.Join(skillDir, "evals")

			skillContent, err := os.ReadFile(skillPath)
			if err != nil {
				return fmt.Errorf("read SKILL.md: %w", err)
			}

			scenarios, structDiags, err := loadScenarios(evalsDir)
			if err != nil {
				return err
			}
			if len(scenarios) == 0 {
				if len(structDiags) > 0 {
					return fmt.Errorf("evals/ directory is missing or has no valid scenarios: %s", strings.Join(structDiags, "; "))
				}
				return fmt.Errorf("no scenario-N directories found under %s", evalsDir)
			}

			// Select LLM provider (returns nil client when no key — graceful
			// degradation to structural-only mode per ADR-007 #5).
			client, err := llmclient.NewFromEnv(flags.provider)
			if err != nil {
				return fmt.Errorf("llm provider selection: %w", err)
			}

			var costLog io.Writer
			if flags.costLog {
				costLog = cmd.ErrOrStderr()
			}

			runner := &evalRunner{
				skillPath:     skillPath,
				skillContent:  string(skillContent),
				client:        client,
				model:         flags.model,
				judgeModel:    flags.judgeModel,
				samples:       flags.samples,
				margin:        flags.margin,
				costLog:       costLog,
				structDiags:   structDiags,
				maxConcurrent: 3,
			}

			result := runner.run(cmd.Context(), scenarios)

			// Honour --fail-below (percentage). Default 0 means structural
			// gate only.
			if flags.failBelow > 0 {
				for _, s := range result.Scenarios {
					if !s.Pass {
						return fmt.Errorf("scenario %s did not pass: %s", s.ID, s.Diagnostic)
					}
					if s.MaxTotal > 0 {
						pct := (s.Total * 100) / s.MaxTotal
						if pct < flags.failBelow {
							return fmt.Errorf("scenario %s scored %d%% (below threshold %d%%)", s.ID, pct, flags.failBelow)
						}
					}
				}
			}

			if flags.writeSummary {
				if err := writeEvalSummary(evalsDir, result); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warn: --write-summary failed: %v\n", err)
				}
			}

			out := cmd.OutOrStdout()
			if flags.asJSON {
				data, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal result: %w", err)
				}
				fmt.Fprintln(out, string(data))
			} else {
				printEvalText(out, result, flags.failBelow)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.provider, "provider", "", "override LLM provider: anthropic|openai|gemini|openai-compatible (default: LLM_PROVIDER env, else anthropic)")
	cmd.Flags().StringVar(&flags.model, "model", "", "override actor model (default: per-provider)")
	cmd.Flags().StringVar(&flags.judgeModel, "judge-model", "", "override judge model (defaults to --model)")
	cmd.Flags().IntVar(&flags.failBelow, "fail-below", 0, "exit non-zero if any scenario's percentage score is below this (default 0: structural-only gate)")
	cmd.Flags().BoolVar(&flags.writeSummary, "write-summary", false, "update evals/summary.json in place (local use only — never in CI)")
	cmd.Flags().BoolVarP(&flags.asJSON, "json", "j", false, "emit machine-readable JSON to stdout")
	cmd.Flags().IntVar(&flags.samples, "samples", 1, "run each scenario N times and report the median score (ADR-018 #2)")
	cmd.Flags().IntVar(&flags.margin, "margin", 5, "advisory band width percentage for pass/fail reporting")
	cmd.Flags().BoolVar(&flags.costLog, "cost-log", false, "log raw token usage (input/output) to stderr for cost derivation (ADR-018 #4)")
	cmd.Flags().StringVarP(&flags.repoRoot, "repo-root", "r", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(cmd)
}

// scenarioInput is everything needed to run one scenario.
type scenarioInput struct {
	ID           string // "scenario-01"
	TaskPrompt   string // contents of task.md
	CriteriaJSON []byte // contents of criteria.json
	Capability   string // contents of capability.txt
}

// evalRunner executes scenarios against an LLM client (or structurally
// when the client is nil).
type evalRunner struct {
	skillPath     string
	skillContent  string
	client        llmclient.Client
	model         string
	judgeModel    string
	samples       int
	margin        int
	costLog       io.Writer
	structDiags   []string // structural problems surfaced by loadScenarios
	maxConcurrent int
}

func (r *evalRunner) run(ctx context.Context, scenarios []scenarioInput) evalResult {
	result := evalResult{
		SkillPath:          r.skillPath,
		Samples:            r.samples,
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		JudgePromptVersion: llmclient.PromptVersion(),
		StructuralOnly:     r.client == nil,
	}
	if r.client != nil {
		result.Provider = r.client.Config().Provider
		result.Model = r.model
		if result.Model == "" {
			result.Model = r.client.Config().Model
		}
		result.JudgeModel = r.judgeModel
		if result.JudgeModel == "" {
			result.JudgeModel = result.Model
		}
	}

	if r.samples < 1 {
		r.samples = 1
	}

	judgeTemp := 0.0
	result.Scenarios = make([]scenarioResult, len(scenarios))

	// Bounded worker pool: max 3 concurrent scenarios per Phase 1 reliability spec.
	sem := make(chan struct{}, maxInt(r.maxConcurrent, 1))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, sc := range scenarios {
		wg.Add(1)
		go func(idx int, s scenarioInput) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var sr scenarioResult
			if r.client == nil {
				sr = r.runStructural(s)
			} else {
				sr = r.runLLM(ctx, s, judgeTemp)
			}
			sr.ID = s.ID
			if s.Capability != "" {
				sr.Capability = strings.TrimSpace(s.Capability)
			}

			mu.Lock()
			result.Scenarios[idx] = sr
			mu.Unlock()
		}(i, sc)
	}
	wg.Wait()

	// Sort scenarios by ID for stable output.
	sort.Slice(result.Scenarios, func(i, j int) bool {
		return result.Scenarios[i].ID < result.Scenarios[j].ID
	})

	// Append structural diagnostics (e.g. malformed scenario dirs) as a
	// synthetic summary diagnostic.
	if len(r.structDiags) > 0 && len(result.Scenarios) == 0 {
		result.Scenarios = append(result.Scenarios, scenarioResult{
			ID:         "structural",
			Status:     "FAIL",
			Diagnostic: strings.Join(r.structDiags, "; "),
		})
	}

	overall := true
	for _, s := range result.Scenarios {
		if !s.Pass {
			overall = false
			break
		}
	}
	result.OverallPass = overall
	return result
}

func (r *evalRunner) runStructural(s scenarioInput) scenarioResult {
	// Structural-only mode: a schema-consistency gate, not a semantic
	// quality gate (per Phase 2 note in the native-eval-runner plan). Pass
	// when the scenario dir exists with the required files and the
	// criteria.json checklist sums to 100.
	maxTotal := 0
	if len(s.CriteriaJSON) > 0 {
		maxTotal = sumCriteria(s.CriteriaJSON)
	}
	sr := scenarioResult{
		MaxTotal: maxTotal,
		Status:   "structural",
	}
	if maxTotal != 100 {
		sr.Pass = false
		sr.Diagnostic = fmt.Sprintf("criteria.json checklist does not sum to 100 (got %d)", maxTotal)
		return sr
	}
	if s.TaskPrompt == "" {
		sr.Pass = false
		sr.Diagnostic = "task.md missing or empty"
		return sr
	}
	if s.Capability == "" {
		sr.Pass = false
		sr.Diagnostic = "capability.txt missing or empty"
		return sr
	}
	sr.Pass = true
	return sr
}

func (r *evalRunner) runLLM(ctx context.Context, s scenarioInput, judgeTemp float64) scenarioResult {
	actorMsgs := llmclient.ActorMessages(r.skillContent, s.TaskPrompt)
	maxOutput := llmclient.MaxOutputTokensFromCriteria(s.CriteriaJSON, 4096)

	actorReq := llmclient.Request{Messages: actorMsgs, MaxTokens: maxOutput}
	if r.model != "" {
		actorReq.Model = r.model
	}
	// Plan Phase 2 Decision #1: judge output is a single JSON object (no streaming).
	resp, err := r.client.Chat(ctx, actorReq)
	if err != nil {
		return scenarioResult{Pass: false, Status: "FAIL", Diagnostic: fmt.Sprintf("actor call failed: %v", err)}
	}
	actorOutput := resp.Content
	if r.costLog != nil {
		fmt.Fprintf(r.costLog, "scenario %s actor tokens: in=%d out=%d\n", s.ID, resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}

	// Re-sample only the judge when --samples > 1 (Phase 1 caching rule).
	judgeMsgs, err := llmclient.JudgeMessages(r.skillContent, s.TaskPrompt, actorOutput, s.CriteriaJSON)
	if err != nil {
		return scenarioResult{
			Pass:               false,
			Status:             "FAIL",
			ActorOutputSnippet: snippet(actorOutput),
			Diagnostic:         fmt.Sprintf("build judge messages: %v", err),
		}
	}

	judgeReq := llmclient.Request{Messages: judgeMsgs, Temperature: judgeTemp, MaxTokens: maxOutput}
	if r.judgeModel != "" {
		judgeReq.Model = r.judgeModel
	} else if r.model != "" {
		judgeReq.Model = r.model
	}

	type sample struct {
		scores []llmclient.JudgeItem
		in     int
		out    int
	}
	samples := make([]sample, 0, r.samples)
	for i := 0; i < r.samples; i++ {
		jResp, err := r.client.Chat(ctx, judgeReq)
		if err != nil {
			return scenarioResult{
				Pass:               false,
				Status:             "FAIL",
				ActorOutputSnippet: snippet(actorOutput),
				Diagnostic:         fmt.Sprintf("judge call %d failed: %v", i+1, err),
			}
		}
		if r.costLog != nil {
			fmt.Fprintf(r.costLog, "scenario %s judge[%d] tokens: in=%d out=%d\n", s.ID, i+1, jResp.Usage.InputTokens, jResp.Usage.OutputTokens)
		}

		res, parseErr := llmclient.ParseJudgeResponse(jResp.Content)
		if parseErr != nil {
			// Judge output validation: retry once with a reminder.
			if i == 0 && r.samples == 1 {
				reminderReq := judgeReq
				reminderReq.Messages = append([]llmclient.Message{{Role: "user", Content: "Your previous response was not valid JSON or was missing the \"scores\" array. Output ONLY the JSON object."}}, judgeReq.Messages...)
				jResp2, err2 := r.client.Chat(ctx, reminderReq)
				if err2 == nil {
					res, parseErr = llmclient.ParseJudgeResponse(jResp2.Content)
				}
			}
			if parseErr != nil {
				return scenarioResult{
					Pass:               false,
					Status:             "FAIL",
					ActorOutputSnippet: snippet(actorOutput),
					Diagnostic:         fmt.Sprintf("judge returned malformed JSON: %v", parseErr),
				}
			}
		}
		samples = append(samples, sample{scores: res.Scores, in: jResp.Usage.InputTokens, out: jResp.Usage.OutputTokens})
	}

	// Aggregate per-item medians across samples.
	itemCount := len(samples[0].scores)
	medians := make([]llmclient.JudgeItem, itemCount)
	for i := 0; i < itemCount; i++ {
		vals := make([]int, 0, len(samples))
		var item llmclient.JudgeItem
		for _, s := range samples {
			if i < len(s.scores) {
				vals = append(vals, s.scores[i].Score)
				item = s.scores[i] // preserve name/justification from first sample
			}
		}
		item.Score = median(vals)
		medians[i] = item
	}

	// Total = sum of median per-item scores.
	maxTotal := sumCriteria(s.CriteriaJSON)
	total := 0
	for _, it := range medians {
		total += it.Score
	}
	total = clampInt(total, 0, maxTotal)

	// Pass/margin/fail against the advisory threshold (max 100).
	pct := 0
	if maxTotal > 0 {
		pct = (total * 100) / maxTotal
	}
	status := "PASS"
	passThreshold := 80 // advisory threshold per Phase 2 spec
	if pct < passThreshold-r.margin {
		status = "FAIL"
	} else if pct < passThreshold+r.margin {
		status = "MARGIN"
	}

	// Aggregate cost tokens across samples.
	inTok, outTok := 0, 0
	for _, s := range samples {
		inTok += s.in
		outTok += s.out
	}

	return scenarioResult{
		ActorOutputSnippet: snippet(actorOutput),
		Scores:             medians,
		Total:              total,
		MaxTotal:           maxTotal,
		Pass:               status == "PASS",
		Status:             status,
		InputTokens:        inTok,
		OutputTokens:       outTok,
	}
}

// loadScenarios discovers scenario-N directories under evalsDir and loads
// the three required files for each. Returns the parsed scenarios, a list
// of structural diagnostics (strings), and any I/O error from reading
// the evals dir itself.
func loadScenarios(evalsDir string) ([]scenarioInput, []string, error) {
	entries, err := os.ReadDir(evalsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, []string{fmt.Sprintf("evals/ directory missing at %q", evalsDir)}, nil
		}
		return nil, nil, err
	}
	var scenarios []scenarioInput
	var diags []string
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "scenario-") {
			continue
		}
		scDir := filepath.Join(evalsDir, e.Name())
		task, err := os.ReadFile(filepath.Join(scDir, "task.md"))
		if err != nil {
			diags = append(diags, fmt.Sprintf("%s: missing task.md", e.Name()))
			continue
		}
		criteria, err := os.ReadFile(filepath.Join(scDir, "criteria.json"))
		if err != nil {
			diags = append(diags, fmt.Sprintf("%s: missing criteria.json", e.Name()))
			continue
		}
		capability, err := os.ReadFile(filepath.Join(scDir, "capability.txt"))
		if err != nil {
			diags = append(diags, fmt.Sprintf("%s: missing capability.txt", e.Name()))
			continue
		}
		scenarios = append(scenarios, scenarioInput{
			ID:           e.Name(),
			TaskPrompt:   string(task),
			CriteriaJSON: criteria,
			Capability:   string(capability),
		})
	}
	if len(scenarios) == 0 && len(diags) == 0 {
		diags = append(diags, "no scenario-N directories found under evals/")
	}
	return scenarios, diags, nil
}

// sumCriteria returns the sum of check-list max_score values in a
// criteria.json blob. Returns 0 on parse failure.
func sumCriteria(criteriaJSON []byte) int {
	var cd struct {
		Checklist []struct {
			MaxScore int `json:"max_score"`
		} `json:"checklist"`
	}
	if err := json.Unmarshal(criteriaJSON, &cd); err != nil {
		return 0
	}
	total := 0
	for _, item := range cd.Checklist {
		total += item.MaxScore
	}
	return total
}

// writeEvalSummary writes a minimal summary.json compatible with the D9
// scorer. Local-authoring only — never called from CI (--write-summary in
// CI is a misuse; the runner is read-only in CI per ADR-018).
func writeEvalSummary(evalsDir string, result evalResult) error {
	// Preserve any existing instructions_coverage if present.
	var existing map[string]any
	summaryPath := filepath.Join(evalsDir, "summary.json")
	if data, err := os.ReadFile(summaryPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	if existing == nil {
		existing = map[string]any{}
	}
	passing := 0
	for _, s := range result.Scenarios {
		if s.Pass {
			passing++
		}
	}
	coverage := 0
	if len(result.Scenarios) > 0 {
		coverage = (passing * 100) / len(result.Scenarios)
	}
	existing["instructions_coverage"] = map[string]any{
		"coverage_percentage":       coverage,
		"eval_runner_timestamp":     result.Timestamp,
		"eval_runner_judge_version": result.JudgePromptVersion,
		"provider":                  result.Provider,
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(summaryPath, data, 0o644)
}

// printEvalText emits the human-readable scenario table.
func printEvalText(out io.Writer, result evalResult, failBelow int) {
	fmt.Fprintf(out, "Eval results for %s\n", result.SkillPath)
	if result.StructuralOnly {
		fmt.Fprintln(out, "  (structural-only mode — no LLM key configured)")
	}
	if result.Provider != "" {
		fmt.Fprintf(out, "  provider=%s model=%s judge=%s samples=%d\n", result.Provider, result.Model, result.JudgeModel, result.Samples)
	}
	for i, s := range result.Scenarios {
		fmt.Fprintf(out, "Scenario %d/%d: %s\n", i+1, len(result.Scenarios), s.ID)
		for _, item := range s.Scores {
			fmt.Fprintf(out, "  ✓ %s: %d/%d — %s\n", item.Name, item.Score, item.MaxScore, item.Justification)
		}
		if s.MaxTotal > 0 {
			fmt.Fprintf(out, "  Total: %d/%d → %s", s.Total, s.MaxTotal, s.Status)
			if failBelow > 0 {
				fmt.Fprintf(out, " (gate: %d%%)", failBelow)
			}
			fmt.Fprintln(out)
		} else if s.Diagnostic != "" {
			fmt.Fprintf(out, "  %s\n", s.Diagnostic)
		}
	}
	status := "FAIL"
	if result.OverallPass {
		status = "PASS"
	}
	fmt.Fprintf(out, "\nOverall: %s\n", status)
}

// snippet returns the first 200 characters of s, trimmed of leading
// whitespace. Used to expose a small portion of the actor's output in the
// JSON report without dumping the full text.
func snippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

func median(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]int, len(vals))
	copy(sorted, vals)
	sort.Ints(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if hi > 0 && v > hi {
		return hi
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Compile-time guard: keep errors referenced to avoid an unused import
// warning when callers trim their handler trees.
var _ = errors.New
