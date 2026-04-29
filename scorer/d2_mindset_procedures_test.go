package scorer

import (
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

// nilBridge is a convenience for unit tests that don't exercise the library path.
func nilBridge() *validatorBridge { return &validatorBridge{} }

// ---------------------------------------------------------------------------
// scoreD2Structure (revised: max 3 pts, no ListItemCount)
// ---------------------------------------------------------------------------

func TestD2Structure_AllZeroMarkers(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{
		StrongMarkers:   0,
		WeakMarkers:     0,
		ImperativeRatio: 0,
	}}
	score, diags := scoreD2Structure("some content", b)
	if score != 0 {
		t.Errorf("want 0 for all-zero markers, got %d", score)
	}
	if len(diags) == 0 {
		t.Error("expected a diagnostic warning for all-zero markers")
	}
}

func TestD2Structure_HighRatio(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.45}}
	score, diags := scoreD2Structure("content", b)
	if score != 3 {
		t.Errorf("want 3 for ImperativeRatio=0.45, got %d", score)
	}
	if len(diags) != 0 {
		t.Errorf("want no diagnostics for high ratio, got %v", diags)
	}
}

func TestD2Structure_MidRatio(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.30}}
	score, _ := scoreD2Structure("content", b)
	if score != 2 {
		t.Errorf("want 2 for ImperativeRatio=0.30, got %d", score)
	}
}

func TestD2Structure_LowRatio(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.12}}
	score, _ := scoreD2Structure("content", b)
	if score != 1 {
		t.Errorf("want 1 for ImperativeRatio=0.12, got %d", score)
	}
}

func TestD2Structure_FallbackNumberedList(t *testing.T) {
	score, _ := scoreD2Structure("1. step one\n2. step two", nilBridge())
	if score != 1 {
		t.Errorf("want 1 for fallback numbered list, got %d", score)
	}
}

func TestD2Structure_FallbackNoList(t *testing.T) {
	score, _ := scoreD2Structure("just some prose", nilBridge())
	if score != 0 {
		t.Errorf("want 0 for fallback with no numbered list, got %d", score)
	}
}

// ---------------------------------------------------------------------------
// scorePreconditions (max 2 pts)
// ---------------------------------------------------------------------------

func TestD2Preconditions_ExplicitHeader(t *testing.T) {
	score, _ := scorePreconditions("## Prerequisites\n- X must exist")
	if score != 2 {
		t.Errorf("want 2 for explicit prerequisites header, got %d", score)
	}
}

func TestD2Preconditions_DependsOn(t *testing.T) {
	score, _ := scorePreconditions("This skill depends on the repo being cloned first.")
	if score != 2 {
		t.Errorf("want 2 for depends on pattern, got %d", score)
	}
}

func TestD2Preconditions_MustExist(t *testing.T) {
	score, _ := scorePreconditions("The config file must exist before running this.")
	if score != 2 {
		t.Errorf("want 2 for must exist pattern, got %d", score)
	}
}

func TestD2Preconditions_ImpliedOnly(t *testing.T) {
	score, _ := scorePreconditions("Requires the config to have been initialised before starting.")
	if score != 1 {
		t.Errorf("want 1 for prerequisite statement without header, got %d", score)
	}
}

func TestD2Preconditions_ConditionalActivation(t *testing.T) {
	score, _ := scorePreconditions("Do not use this if the environment is not set up.")
	if score != 1 {
		t.Errorf("want 1 for conditional activation pattern, got %d", score)
	}
}

func TestD2Preconditions_None(t *testing.T) {
	score, diags := scorePreconditions("Just some general guidance text with nothing relevant.")
	if score != 0 {
		t.Errorf("want 0 for no precondition signals, got %d", score)
	}
	if len(diags) == 0 {
		t.Error("expected a diagnostic for missing preconditions")
	}
}

// ---------------------------------------------------------------------------
// scorePostconditions (max 2 pts)
// ---------------------------------------------------------------------------

func TestD2Postconditions_TestGate(t *testing.T) {
	score, _ := scorePostconditions("Run go test ./... before proceeding to the next step.")
	if score != 2 {
		t.Errorf("want 2 for test-gate pattern, got %d", score)
	}
}

func TestD2Postconditions_NpmTest(t *testing.T) {
	score, _ := scorePostconditions("Execute npm test to verify the build passes.")
	if score != 2 {
		t.Errorf("want 2 for npm test pattern, got %d", score)
	}
}

func TestD2Postconditions_HumanApproval(t *testing.T) {
	score, _ := scorePostconditions("Wait for human approval before merging.")
	if score != 2 {
		t.Errorf("want 2 for human-confirmation-gate pattern, got %d", score)
	}
}

func TestD2Postconditions_ArtifactOnly(t *testing.T) {
	score, _ := scorePostconditions("Confirm the output file exists before continuing.")
	if score != 1 {
		t.Errorf("want 1 for artifact-confirmation only, got %d", score)
	}
}

func TestD2Postconditions_None(t *testing.T) {
	score, diags := scorePostconditions("Nothing here that verifies any outcome at all.")
	if score != 0 {
		t.Errorf("want 0 for no postcondition signals, got %d", score)
	}
	if len(diags) == 0 {
		t.Error("expected a diagnostic for missing postconditions")
	}
}

// ---------------------------------------------------------------------------
// scoreDecisionPoints (max 2 pts)
// ---------------------------------------------------------------------------

func TestD2DecisionPoints_ExplicitBranch(t *testing.T) {
	score, _ := scoreDecisionPoints("If the command fails, roll back the changes.")
	if score != 2 {
		t.Errorf("want 2 for explicit branch on failure, got %d", score)
	}
}

func TestD2DecisionPoints_ConditionalHeader(t *testing.T) {
	score, _ := scoreDecisionPoints("## Troubleshooting\nIf the build fails, check the logs.")
	if score != 2 {
		t.Errorf("want 2 for troubleshooting conditional header, got %d", score)
	}
}

func TestD2DecisionPoints_Otherwise(t *testing.T) {
	score, _ := scoreDecisionPoints("Run the script. Otherwise, use the fallback approach.")
	if score != 2 {
		t.Errorf("want 2 for otherwise pattern, got %d", score)
	}
}

func TestD2DecisionPoints_FallbackOnly(t *testing.T) {
	score, _ := scoreDecisionPoints("If that doesn't work, retry the operation.")
	if score != 1 {
		t.Errorf("want 1 for retry/fallback without explicit branch, got %d", score)
	}
}

func TestD2DecisionPoints_StopCondition(t *testing.T) {
	score, _ := scoreDecisionPoints("Abort when the connection is lost.")
	if score != 1 {
		t.Errorf("want 1 for stop condition, got %d", score)
	}
}

func TestD2DecisionPoints_None(t *testing.T) {
	score, diags := scoreDecisionPoints("Just do the task, no conditionals mentioned.")
	if score != 0 {
		t.Errorf("want 0 for no decision-point signals, got %d", score)
	}
	if len(diags) == 0 {
		t.Error("expected a diagnostic for missing decision points")
	}
}

// ---------------------------------------------------------------------------
// scoreD2 integration — wiring checks (updated for new PKO budget)
// ---------------------------------------------------------------------------

func TestD2_WhenToUse(t *testing.T) {
	content := "---\ndescription: x\n---\n## When to Use\nuse it when needed\n\n## When NOT to Use\nnot this time"
	score, _ := scoreD2(content, nilBridge())
	// 3 (when guidance combined bucket)
	if score != 3 {
		t.Errorf("want 3, got %d", score)
	}
}

func TestD2_MindsetHeading(t *testing.T) {
	content := "---\ndescription: x\n---\n## Mindset\nthink carefully"
	score, _ := scoreD2(content, nilBridge())
	// 3 (mindset heading now worth 3 pts)
	if score != 3 {
		t.Errorf("want 3, got %d", score)
	}
}

func TestD2_PhilosophyHeading(t *testing.T) {
	content := "---\ndescription: x\n---\n## Philosophy\nthink this way"
	score, _ := scoreD2(content, nilBridge())
	if score != 3 {
		t.Errorf("want 3, got %d", score)
	}
}

func TestD2_FallbackNumberedList(t *testing.T) {
	content := "---\ndescription: x\n---\n1. step one\n2. step two\n3. step three"
	score, _ := scoreD2(content, nilBridge())
	if score < 1 {
		t.Errorf("numbered list via fallback should score >=1, got %d", score)
	}
}

func TestD2_LibraryImperativeRatioMid(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.3}}
	content := "---\ndescription: x\n---\nsome content"
	score, _ := scoreD2(content, b)
	// 2 pts (structure mid-ratio; ListItemCount dropped)
	if score < 2 {
		t.Errorf("want >=2 for mid-range imperative ratio, got %d", score)
	}
}

func TestD2_LibraryImperativeRatioLow(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.15}}
	content := "---\ndescription: x\n---\nsome content"
	score, _ := scoreD2(content, b)
	// 1 pt (structure low ratio)
	if score < 1 {
		t.Errorf("want >=1 for low imperative ratio, got %d", score)
	}
}

func TestD2_LibraryImperativeRatio(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.45}}
	content := "---\ndescription: x\n---\n## When to Use\ndo this"
	score, _ := scoreD2(content, b)
	// 3 (structure high) + 3 (when-to-use) = 6
	if score != 6 {
		t.Errorf("want 6, got %d", score)
	}
}

func TestD2_Cap15(t *testing.T) {
	b := &validatorBridge{Content: &types.ContentReport{ImperativeRatio: 0.5}}
	content := "## Mindset\nThink first.\n\n## Prerequisites\nDepends on X.\n\n## When to Use\nwhen needed\n\n## Troubleshooting\nIf step fails, retry.\n\nRun go test ./... before proceeding.\n"
	score, _ := scoreD2(content, b)
	if score > 15 {
		t.Errorf("score must not exceed 15, got %d", score)
	}
}
