package llmclient

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// JudgePrompt is the pinned rubric prompt used by the eval runner's judge
// step (Phase 2 of the native eval runner plan). The SHA256 of JudgePrompt
// is reported as judge_prompt_version in JSON output so score trends are
// comparable only within a single prompt version.
//
// Modifying this prompt invalidates prior score trends. Bump the version by
// editing the prompt; the hash is computed at runtime in PromptVersion().
const JudgePrompt = `You are grading whether an AI agent correctly completed a task using a reference skill.

You will receive:
  1. The reference skill content that the agent had access to.
  2. The task prompt the agent was asked to complete.
  3. The agent's output.
  4. A weighted checklist of criteria (each with a name, description, and max_score). The max_score values sum to 100.

Grade each checklist item independently. For each item award an integer score from 0 up to its max_score, with a one-sentence justification. The justifications must reference the agent's output specifically (what was present or missing).

Output a single JSON object with this exact shape and nothing else:
{"scores": [{"name": "<criterion name>", "score": N, "max_score": M, "justification": "<one sentence>"}]}

Do not include any text before or after the JSON object. The "scores" array must contain exactly one entry per checklist item, in the order given.`

// _actorSystemPrompt is prepended to the actor's messages when the runner
// invokes the model against a skill. It frames the model as an agent that
// has the skill available and is responding to the task prompt.
const actorSystemPrompt = `You are an AI agent. Use the following skill content as your working instructions for the task that follows. Respond as the agent would, producing the artefacts the task asks for. Do not preface your answer with commentary — produce the requested output directly.`

// ActorMessages builds the message list for an actor call. skillContent is
// the SKILL.md text the agent is presumed to be operating under; taskPrompt
// is the scenario's user-facing prompt (typically the contents of task.md).
func ActorMessages(skillContent, taskPrompt string) []Message {
	return []Message{
		{Role: "system", Content: actorSystemPrompt + "\n\n--- SKILL START ---\n" + skillContent + "\n--- SKILL END ---"},
		{Role: "user", Content: taskPrompt},
	}
}

// JudgeMessages builds the message list for a judge call. The judge is
// pinned at temperature 0 by the caller for determinism.
func JudgeMessages(skillContent, taskPrompt, actorOutput string, criteriaJSON []byte) ([]Message, error) {
	var cd struct {
		Checklist []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			MaxScore    int    `json:"max_score"`
		} `json:"checklist"`
	}
	if err := json.Unmarshal(criteriaJSON, &cd); err != nil {
		return nil, fmt.Errorf("parse criteria.json: %w", err)
	}

	var b strings.Builder
	b.WriteString(JudgePrompt)
	b.WriteString("\n\n--- SKILL CONTENT START ---\n")
	b.WriteString(skillContent)
	b.WriteString("\n--- SKILL CONTENT END ---\n\n")
	b.WriteString("--- TASK PROMPT START ---\n")
	b.WriteString(taskPrompt)
	b.WriteString("\n--- TASK PROMPT END ---\n\n")
	b.WriteString("--- AGENT OUTPUT START ---\n")
	b.WriteString(actorOutput)
	b.WriteString("\n--- AGENT OUTPUT END ---\n\n")
	b.WriteString("--- CHECKLIST START ---\n")
	b.WriteString("The checklist items (in order) are:\n")
	for i, item := range cd.Checklist {
		fmt.Fprintf(&b, "%d. name=%q description=%q max_score=%d\n", i+1, item.Name, item.Description, item.MaxScore)
	}
	b.WriteString("--- CHECKLIST END ---\n")

	return []Message{{Role: "user", Content: b.String()}}, nil
}

// JudgeResult is the parsed judge response: a list of per-item scores.
type JudgeResult struct {
	Scores []JudgeItem `json:"scores"`
}

// JudgeItem is one checklist item scored by the judge.
type JudgeItem struct {
	Name          string `json:"name"`
	Score         int    `json:"score"`
	MaxScore      int    `json:"max_score"`
	Justification string `json:"justification"`
}

// ParseJudgeResponse extracts the scores array from a judge model's output.
// The judge is instructed to emit a single JSON object; if the output has
// extra prose around it, this falls back to scanning for the first "{"
// through the last "}". Returns an error if no valid JSON object could be
// parsed or if "scores" is missing.
func ParseJudgeResponse(raw string) (JudgeResult, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return JudgeResult{}, fmt.Errorf("empty judge response")
	}

	// Try a direct unmarshal first.
	var result JudgeResult
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		if result.Scores == nil {
			return JudgeResult{}, fmt.Errorf("judge response missing \"scores\" array")
		}
		return result, nil
	}

	// Fall back to extracting the outermost {...} block (handles cases
	// where the model prefaces or trails the JSON with prose).
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return JudgeResult{}, fmt.Errorf("no JSON object found in judge response")
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return JudgeResult{}, fmt.Errorf("parse judge JSON: %w", err)
	}
	if result.Scores == nil {
		return JudgeResult{}, fmt.Errorf("judge response missing \"scores\" array")
	}
	return result, nil
}

// PromptVersion returns the SHA256 hash of JudgePrompt, prefixed with
// "sha256:". This is the value reported as judge_prompt_version in the
// eval runner's JSON output.
func PromptVersion() string {
	sum := sha256.Sum256([]byte(JudgePrompt))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// MaxOutputTokensFromCriteria reads the optional max_output_tokens field
// from criteria.json. Returns defaultTokens when the field is absent or
// unparsable.
func MaxOutputTokensFromCriteria(criteriaJSON []byte, defaultTokens int) int {
	var cd struct {
		MaxOutputTokens int `json:"max_output_tokens"`
	}
	if err := json.Unmarshal(criteriaJSON, &cd); err != nil {
		return defaultTokens
	}
	if cd.MaxOutputTokens > 0 {
		return cd.MaxOutputTokens
	}
	return defaultTokens
}
