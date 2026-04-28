package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = func() string {
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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("skill-auditor v%s\n", version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = version
}
