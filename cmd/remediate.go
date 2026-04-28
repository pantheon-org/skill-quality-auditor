package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/spf13/cobra"
)

var (
	remTargetScore int
	remValidate    bool
	remRepoRoot    string
)

var remediateCmd = &cobra.Command{
	Use:   "remediate <skill>",
	Short: "Generate or validate a remediation plan",
	Long: `Generate a schema-compliant remediation plan from a stored audit result.

With --validate, parses an existing plan file and checks it against the
remediation-plan.schema.json constraints instead of generating a new plan.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot(remRepoRoot)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillArg := args[0]

		if remValidate {
			return runValidate(skillArg, repoRoot)
		}
		return runGenerate(skillArg, repoRoot)
	},
}

func runGenerate(skillArg, repoRoot string) error {
	// Locate the most recent stored audit for this skill.
	auditsBase := filepath.Join(repoRoot, ".context", "audits", skillArg)
	auditJSON, auditDate, err := latestAuditJSON(auditsBase)
	if err != nil {
		return fmt.Errorf("no stored audit found for %q — run 'evaluate %s --store' first: %w", skillArg, skillArg, err)
	}

	result, err := loadAuditJSON(auditJSON)
	if err != nil {
		return fmt.Errorf("load audit: %w", err)
	}

	auditPath := fmt.Sprintf(".context/audits/%s/%s/Analysis.md", skillArg, auditDate)
	date := time.Now().Format("2006-01-02")

	plan, err := reporter.RemediationPlan(result, remTargetScore, auditPath, date)
	if err != nil {
		return fmt.Errorf("generate plan: %w", err)
	}

	outDir := filepath.Join(repoRoot, ".context", "plans")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create plans dir: %w", err)
	}

	skillBase := filepath.Base(skillArg)
	outFile := filepath.Join(outDir, skillBase+"-remediation-plan.md")
	if err := os.WriteFile(outFile, []byte(plan), 0o644); err != nil {
		return fmt.Errorf("write plan: %w", err)
	}

	fmt.Print(plan)
	fmt.Fprintf(os.Stderr, "plan written to %s\n", outFile)
	return nil
}

func runValidate(planArg, repoRoot string) error {
	// planArg may be a bare skill name or a direct file path.
	planPath := planArg
	if !filepath.IsAbs(planArg) && !pathExists(planArg) {
		skillBase := filepath.Base(planArg)
		planPath = filepath.Join(repoRoot, ".context", "plans", skillBase+"-remediation-plan.md")
	}

	if !pathExists(planPath) {
		return fmt.Errorf("plan file not found: %s", planPath)
	}

	errs := reporter.ValidateRemediationPlan(planPath)
	if len(errs) == 0 {
		fmt.Printf("✓ %s — valid\n", planPath)
		return nil
	}

	fmt.Fprintf(os.Stderr, "✗ %s — %d validation error(s):\n", planPath, len(errs))
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  • %s\n", e)
	}
	return fmt.Errorf("validation failed (%d errors)", len(errs))
}

func init() {
	remediateCmd.Flags().IntVar(&remTargetScore, "target-score", 0, "target score (default: current+20, max 140)")
	remediateCmd.Flags().BoolVar(&remValidate, "validate", false, "validate an existing plan file instead of generating")
	remediateCmd.Flags().StringVar(&remRepoRoot, "repo-root", "", "repo root (auto-detected if empty)")
	rootCmd.AddCommand(remediateCmd)
}
