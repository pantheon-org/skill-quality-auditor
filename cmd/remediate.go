package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pantheon-org/skill-quality-auditor/reporter"
	"github.com/spf13/cobra"
)

var remediateCmd = &cobra.Command{
	Use:   "remediate <skill>",
	Short: "Generate or validate a remediation plan",
	Long: `Generate a schema-compliant remediation plan from a stored audit result.

With --validate, parses an existing plan file and checks it against the
remediation-plan.schema.json constraints instead of generating a new plan.

Use --json to emit the plan as JSON instead of Markdown.
Use --dry-run to print the plan to stdout without writing to .context/plans/.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRootFlag, _ := cmd.Flags().GetString("repo-root")
		repoRoot, err := resolveRepoRoot(repoRootFlag)
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}

		skillArg := args[0]
		validate, _ := cmd.Flags().GetBool("validate")
		targetScore, _ := cmd.Flags().GetInt("target-score")

		if validate {
			return runValidate(cmd, skillArg, repoRoot)
		}
		return runGenerate(cmd, skillArg, repoRoot, targetScore)
	},
}

func runGenerate(cmd *cobra.Command, skillArg, repoRoot string, targetScore int) error {
	out := cmd.OutOrStdout()
	asJSON, _ := cmd.Flags().GetBool("json")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

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

	var content string
	var ext string
	if asJSON {
		raw, err := reporter.RemediationPlanJSON(result, targetScore, auditPath, date)
		if err != nil {
			return fmt.Errorf("generate plan: %w", err)
		}
		content = string(raw)
		ext = ".json"
	} else {
		plan, err := reporter.RemediationPlan(result, targetScore, auditPath, date)
		if err != nil {
			return fmt.Errorf("generate plan: %w", err)
		}
		content = plan
		ext = ".md"
	}

	if dryRun {
		fmt.Fprint(out, content)
		return nil
	}

	outDir := filepath.Join(repoRoot, ".context", "plans")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create plans dir: %w", err)
	}

	skillBase := filepath.Base(skillArg)
	outFile := filepath.Join(outDir, skillBase+"-remediation-plan"+ext)
	if err := os.WriteFile(outFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write plan: %w", err)
	}

	fmt.Fprint(out, content)
	fmt.Fprintf(cmd.ErrOrStderr(), "plan written to %s\n", outFile)
	return nil
}

func runValidate(cmd *cobra.Command, planArg, repoRoot string) error {
	out := cmd.OutOrStdout()
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
		fmt.Fprintf(out, "✓ %s — valid\n", planPath)
		return nil
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "✗ %s — %d validation error(s):\n", planPath, len(errs))
	for _, e := range errs {
		fmt.Fprintf(cmd.ErrOrStderr(), "  • %s\n", e)
	}
	return fmt.Errorf("validation failed (%d errors)", len(errs))
}

func init() {
	remediateCmd.Flags().IntP("target-score", "t", 0, "target score (default: current+20, max 140)")
	remediateCmd.Flags().BoolP("validate", "v", false, "validate an existing plan file instead of generating")
	remediateCmd.Flags().StringP("repo-root", "r", "", "repo root (auto-detected if empty)")
	remediateCmd.Flags().BoolP("json", "j", false, "emit the plan as JSON instead of Markdown")
	remediateCmd.Flags().BoolP("dry-run", "n", false, "print plan to stdout without writing to .context/plans/")
	rootCmd.AddCommand(remediateCmd)
}
