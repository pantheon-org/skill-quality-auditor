package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// OutputFormat represents the output format for a command.
type OutputFormat int

const (
	// OutputFormatJSON emits JSON output.
	OutputFormatJSON OutputFormat = iota
	// OutputFormatMarkdown emits Markdown output.
	OutputFormatMarkdown
)

// resolveOutputFormat reads the --json and --markdown flags from cmd and returns
// the active OutputFormat. It returns an error if both flags are set simultaneously.
// When neither flag is set, the provided defaultFormat is returned.
func resolveOutputFormat(cmd *cobra.Command, defaultFormat OutputFormat) (OutputFormat, error) {
	asJSON, _ := cmd.Flags().GetBool("json")
	asMarkdown, _ := cmd.Flags().GetBool("markdown")

	if asJSON && asMarkdown {
		return 0, fmt.Errorf("--json and --markdown are mutually exclusive")
	}

	if asJSON {
		return OutputFormatJSON, nil
	}
	if asMarkdown {
		return OutputFormatMarkdown, nil
	}
	return defaultFormat, nil
}
