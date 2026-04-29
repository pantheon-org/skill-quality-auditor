package scorer

import "regexp"

var (
	reD2MindsetHeader = regexp.MustCompile(`(?im)##\s*(mindset|philosophy|principles)`)
	reD2NumberedList  = regexp.MustCompile(`(?m)^\s*[0-9]+\.`)

	// Precondition signals.
	reD2PreHeader      = regexp.MustCompile(`(?im)##\s*(prerequisites?|preconditions?|requirements?|before you (start|begin))`)
	reD2PreGuard       = regexp.MustCompile(`(?i)(only (invoke|use|run|apply) (this|when)|depends on|must (exist|be present|be configured|have run))`)
	reD2PreImplied     = regexp.MustCompile(`(?i)(prerequisite[s]?|requires? .* to (have|be)|before (starting|running|invoking|applying))`)
	reD2PreConditional = regexp.MustCompile(`(?i)(do not|don't) (use|invoke|run) (this|if|unless|when)`)

	// Postcondition signals.
	reD2PostTestVerify = regexp.MustCompile(`(?i)((run|execute) .*(test[s]?|spec[s]?|lint|check|verify)|go test|npm test|pytest)`)
	reD2PostHumanGate  = regexp.MustCompile(`(?i)(wait for|requires?|ask for) .*(approval|confirmation|sign.?off|review)`)
	reD2PostCIGate     = regexp.MustCompile(`(?i)(must pass|pipeline|ci|cd).*(pass|succeed|green)`)
	reD2PostArtifact   = regexp.MustCompile(`(?i)(confirm|verify|check) .*(file|artifact|output|result) (exists?|is present|was created)`)
	reD2PostExtState   = regexp.MustCompile(`(?i)(assert|ensure|validate) .*(external|remote|database|api|service)`)

	// Decision-point signals.
	reD2DPExplicitBranch = regexp.MustCompile(`(?i)if .*(fail[s]?|error[s]?|unexpected|not found|missing)`)
	reD2DPOtherwise      = regexp.MustCompile(`(?i)\botherwise\b`)
	reD2DPCondHeader     = regexp.MustCompile(`(?im)##\s*(troubleshooting|error handling|if .* fails?|when .* fails?)`)
	reD2DPFallback       = regexp.MustCompile(`(?i)(fall(back| back) to|revert to|retry)`)
	reD2DPStop           = regexp.MustCompile(`(?i)(stop|abort|halt|do not continue) (if|when|unless)`)
	reD2DPEscalate       = regexp.MustCompile(`(?i)(escalate|raise|report) .*(to|with) .*(human|user|team|engineer)`)
)

// scoreD2 — Mindset + Procedures (max: 15)
//
// PKO-aligned breakdown:
//   - Mindset/philosophy heading    : 3 pts
//   - Structural/procedural density : 3 pts
//   - When/when-not guidance        : 3 pts
//   - Preconditions                 : 2 pts
//   - Postconditions                : 2 pts
//   - Decision points               : 2 pts
func scoreD2(content string, b *validatorBridge) (int, []Diagnostic) {
	score := 0
	var diags []Diagnostic

	if reD2MindsetHeader.MatchString(content) {
		score += 3
	}

	delta, structDiags := scoreD2Structure(content, b)
	score += delta
	diags = append(diags, structDiags...)

	if countPattern(content, "when to use") > 0 || countPattern(content, "when to apply") > 0 ||
		countPattern(content, "when not to") > 0 {
		score += 3
	}

	prePts, preDiags := scorePreconditions(content)
	score += prePts
	diags = append(diags, preDiags...)

	postPts, postDiags := scorePostconditions(content)
	score += postPts
	diags = append(diags, postDiags...)

	dpPts, dpDiags := scoreDecisionPoints(content)
	score += dpPts
	diags = append(diags, dpDiags...)

	if score > 15 {
		score = 15
	}
	return score, diags
}

// scoreD2Structure — structural / procedural density (max: 3 pts).
// ListItemCount is no longer used; ImperativeRatio drives the full budget.
func scoreD2Structure(content string, b *validatorBridge) (int, []Diagnostic) {
	if b.Content != nil {
		if b.Content.StrongMarkers == 0 && b.Content.WeakMarkers == 0 && b.Content.ImperativeRatio == 0 {
			return 0, []Diagnostic{warnDiag("D2", "no imperative or directive markers detected — procedural guidance may be missing")}
		}
		switch {
		case b.Content.ImperativeRatio >= 0.4:
			return 3, nil
		case b.Content.ImperativeRatio >= 0.25:
			return 2, nil
		case b.Content.ImperativeRatio >= 0.1:
			return 1, nil
		default:
			return 0, nil
		}
	}
	if reD2NumberedList.MatchString(content) {
		return 1, nil
	}
	return 0, nil
}

// scorePreconditions — explicit entry conditions (max: 2 pts).
//
//	2 pts: explicit guard clause, entry-condition header, or dependency declaration.
//	1 pt:  prerequisite statement or conditional-activation pattern only.
func scorePreconditions(content string) (int, []Diagnostic) {
	if reD2PreHeader.MatchString(content) || reD2PreGuard.MatchString(content) {
		return 2, nil
	}
	if reD2PreImplied.MatchString(content) || reD2PreConditional.MatchString(content) {
		return 1, nil
	}
	return 0, []Diagnostic{warnDiag("D2", "no precondition signals detected — add explicit entry conditions (e.g. ## Prerequisites)")}
}

// scorePostconditions — external checkpoints (max: 2 pts).
//
//	2 pts: test/lint verification, CI/CD gate, or human confirmation gate.
//	1 pt:  artifact confirmation or external-state assertion only.
func scorePostconditions(content string) (int, []Diagnostic) {
	if reD2PostTestVerify.MatchString(content) || reD2PostHumanGate.MatchString(content) || reD2PostCIGate.MatchString(content) {
		return 2, nil
	}
	if reD2PostArtifact.MatchString(content) || reD2PostExtState.MatchString(content) {
		return 1, nil
	}
	return 0, []Diagnostic{warnDiag("D2", "no postcondition signals detected — add external verification steps (e.g. run tests, confirm output)")}
}

// scoreDecisionPoints — branching / error-handling logic (max: 2 pts).
//
//	2 pts: explicit branch (if...fail/error) or conditional header.
//	1 pt:  fallback instruction, stop condition, or escalation signal only.
func scoreDecisionPoints(content string) (int, []Diagnostic) {
	if reD2DPExplicitBranch.MatchString(content) || reD2DPOtherwise.MatchString(content) || reD2DPCondHeader.MatchString(content) {
		return 2, nil
	}
	if reD2DPFallback.MatchString(content) || reD2DPStop.MatchString(content) || reD2DPEscalate.MatchString(content) {
		return 1, nil
	}
	return 0, []Diagnostic{warnDiag("D2", "no decision-point signals detected — add branching logic for error cases (e.g. ## Troubleshooting)")}
}
