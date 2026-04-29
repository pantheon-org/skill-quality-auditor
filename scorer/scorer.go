package scorer

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// Score evaluates a skill at skillPath and returns a Result.
func Score(ctx context.Context, skillPath string) (*Result, error) {
	contentBytes, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, err
	}
	evalsDir := filepath.Join(filepath.Dir(skillPath), "evals")
	return ScoreFromContent(ctx, skillPath, string(contentBytes), evalsDir)
}

// ScoreFromContent scores a skill from pre-loaded content and an evals directory path.
func ScoreFromContent(_ context.Context, skillPath, content, evalsDir string) (*Result, error) {
	skillDir := filepath.Dir(skillPath)
	bridge := newValidatorBridge(skillDir)

	// D5 returns metadata beyond the score; capture via closure side-effect.
	var lines, refCount int
	var hasRefs bool

	registry := []dimensionEntry{
		{AllDimensions[0], func(c, dir string, _ *validatorBridge) (int, []Diagnostic) {
			return scoreD1(c, dir)
		}},
		{AllDimensions[1], func(c, _ string, b *validatorBridge) (int, []Diagnostic) {
			return scoreD2(c, b)
		}},
		{AllDimensions[2], scoreD3},
		{AllDimensions[3], scoreD4},
		{AllDimensions[4], func(c, dir string, b *validatorBridge) (int, []Diagnostic) {
			s, l, rc, hr := scoreD5WithMeta(c, dir, b)
			lines, refCount, hasRefs = l, rc, hr
			return s, nil
		}},
		{AllDimensions[5], func(_, _ string, b *validatorBridge) (int, []Diagnostic) {
			return scoreD6(b)
		}},
		{AllDimensions[6], func(_, _ string, b *validatorBridge) (int, []Diagnostic) {
			return scoreD7(b)
		}},
		{AllDimensions[7], func(c, _ string, b *validatorBridge) (int, []Diagnostic) {
			return scoreD8(c, b)
		}},
		{AllDimensions[8], func(_, _ string, _ *validatorBridge) (int, []Diagnostic) {
			return scoreD9(evalsDir, skillPath)
		}},
	}

	scores := make([]int, len(registry))
	var allDiags []Diagnostic
	total := 0
	for i, entry := range registry {
		s, diags := entry.fn(content, skillDir, bridge)
		scores[i] = s
		total += s
		allDiags = append(allDiags, diags...)
	}

	var errorDetails, warningDetails []Diagnostic
	for _, d := range allDiags {
		if d.severity == "error" {
			errorDetails = append(errorDetails, d)
		} else {
			warningDetails = append(warningDetails, d)
		}
	}
	if !hasRefs {
		warningDetails = append(warningDetails, warnDiag("D5", "no references/ directory (progressive disclosure missing)"))
	}

	return &Result{
		Skill:                     skillPath,
		Date:                      time.Now().Format("2006-01-02"),
		Total:                     total,
		MaxTotal:                  140,
		Grade:                     Grade(total),
		Lines:                     lines,
		HasReferences:             hasRefs,
		ReferenceCount:            refCount,
		ReferenceSectionCompliant: isReferenceSectionCompliant(content),
		Errors:                    len(errorDetails),
		Warnings:                  len(warningDetails),
		ErrorDetails:              errorDetails,
		WarningDetails:            warningDetails,
		Dimensions:                buildDimensionMap(scores),
	}, nil
}
