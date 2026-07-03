package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/pantheon-org/skill-quality-auditor/internal/patternconfig"
)

// exitCoder is implemented by errors that require a specific process exit
// code instead of the default 1 (e.g. duplication's Critical-pair gate).
type exitCoder interface {
	ExitCode() int
}

// buildVersion is injected by GoReleaser via ldflags at release time.
// It takes precedence over the version embedded in tile.json.
var buildVersion string

var version = func() string {
	if buildVersion != "" {
		return buildVersion
	}
	var tile struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(embeddedTile, &tile); err != nil || tile.Version == "" {
		return "unknown"
	}
	return tile.Version
}()

var rootCmd = &cobra.Command{
	Use:   "skill-auditor",
	Short: "Audit skill quality using the 9-dimension framework",
	Long:  "skill-auditor evaluates skills against the 9-dimension quality framework, combining skill-validator structural checks with custom D1-D9 scoring.",
}

// NewRootCmd returns a root command with all subcommands registered.
// Passing a non-nil out wires it as the default output writer, enabling
// test code to capture output without mutating os.Stdout.
func NewRootCmd(out io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:   rootCmd.Use,
		Short: rootCmd.Short,
		Long:  rootCmd.Long,
	}
	root.Version = version
	if out != nil {
		root.SetOut(out)
	}
	for _, sub := range rootCmd.Commands() {
		root.AddCommand(sub)
	}
	return root
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "skill-auditor v%s\n", version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(exitCodeFor(err))
	}
}

// exitCodeFor returns the process exit code for a command error: 1 by
// default, or whatever an exitCoder-implementing error specifies.
func exitCodeFor(err error) int {
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

func init() {
	patternconfig.Init(embeddedConfig, "assets/assets/config/scoring-patterns.yaml")
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = version
}
