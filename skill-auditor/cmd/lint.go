package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Deprecated: lint is an alias for "validate artifacts". It will be removed in a future release.
var lintCmd = &cobra.Command{
	Use:        "lint [skills-dir]",
	Short:      "Deprecated: use 'validate artifacts' instead",
	Deprecated: "use 'validate artifacts' instead — lint will be removed in a future release",
	Args:       cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: 'lint' is deprecated; switching to 'validate artifacts'")
		// validate artifacts walks skills/ from repo root; the optional skills-dir
		// arg that lint accepted has no equivalent — drop it.
		return validateArtifactsCmd.RunE(cmd, nil)
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
