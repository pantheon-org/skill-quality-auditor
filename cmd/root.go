package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"

	"github.com/pantheon-org/skill-quality-auditor/internal/patternconfig"
)

// exitCoder is implemented by errors that require a specific process exit
// code instead of the default 1 (e.g. duplication's Critical-pair gate).
type exitCoder interface {
	ExitCode() int
}

// embeddedConfigPath is the location of the embedded scoring-patterns.yaml
// inside embeddedConfig (see cmd/embed.go). It is the tier-4/5 fallback for
// every scoring command, and the only source ever used by "eval".
const embeddedConfigPath = "assets/assets/config/scoring-patterns.yaml"

// userConfigDirName/userConfigFileName name the default per-OS config
// location and the opportunistic CWD override:
//
//	Linux:   $XDG_CONFIG_HOME/skill-quality-auditor/scoring-patterns.yaml (or ~/.config/... )
//	macOS:   ~/Library/Application Support/skill-quality-auditor/scoring-patterns.yaml
//	Windows: %AppData%\skill-quality-auditor\scoring-patterns.yaml
const (
	userConfigDirName  = "skill-quality-auditor"
	userConfigFileName = "scoring-patterns.yaml"
)

// userConfigDir is a seam over os.UserConfigDir so tests can point the
// default config-directory tier at a scratch directory instead of the real
// per-OS user config location.
var userConfigDir = os.UserConfigDir

// configFlag and noUserConfigFlag back the persistent -c/--config and
// --no-user-config flags declared in init() below. They are package vars
// (rather than fields on a struct) because they must be visible to
// PersistentPreRunE at the root level, ahead of any subcommand's own flags.
var (
	configFlag       string
	noUserConfigFlag bool
)

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

// buildDate is injected by GoReleaser via ldflags at release time as an
// RFC3339 timestamp (the release commit's date). It is empty for local
// `go build`/`go install` builds, where releaseDate falls back to the Go
// toolchain's VCS stamp instead.
var buildDate string

// buildInfoReader is a seam over debug.ReadBuildInfo so tests can control the
// VCS-stamp fallback without depending on how the test binary was built.
var buildInfoReader = debug.ReadBuildInfo

// versionString renders the version line, appending the release date in the
// organisation locale (DD-MM-YYYY) when a build date is known.
func versionString() string {
	line := "skill-auditor v" + version
	if d := releaseDate(); d != "" {
		line += " (released " + d + ")"
	}
	return line
}

// releaseDate returns the build date formatted DD-MM-YYYY, or "" when no date
// is available or it cannot be parsed. It prefers the GoReleaser-injected
// buildDate and falls back to the toolchain's vcs.time stamp for source builds.
func releaseDate() string {
	raw := buildDate
	if raw == "" {
		raw = vcsTime()
	}
	if raw == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return ""
	}
	return t.Format("02-01-2006")
}

// vcsTime returns the commit timestamp the Go toolchain embeds for VCS builds,
// or "" if it is unavailable (e.g. module-cache installs).
func vcsTime() string {
	info, ok := buildInfoReader()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.time" {
			return s.Value
		}
	}
	return ""
}

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
		fmt.Fprintln(cmd.OutOrStdout(), versionString())
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
	rootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "path to a scoring-patterns.yaml override (hard error if missing or invalid)")
	rootCmd.PersistentFlags().BoolVar(&noUserConfigFlag, "no-user-config", false, "ignore CWD/user config files and score with the embedded/built-in patterns only")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return resolveConfig(cmd.Name() == "eval")
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = version
}

// resolveConfig implements the 5-tier scoring-pattern precedence:
//
//  1. Explicit -c/--config flag — hard error if missing or invalid.
//  2. ./scoring-patterns.yaml in the current working directory — malformed
//     warns and falls through; absent is silently skipped. Never auto-created.
//  3. The default per-OS config-directory path — malformed warns and falls
//     through; absent falls through to tier 4 and is then auto-generated
//     from whatever tier 4/5 resolves to, best-effort.
//  4. The embedded scoring-patterns.yaml.
//  5. The hardcoded defaults built into internal/patternconfig.
//
// isEval and --no-user-config both skip straight to tier 4/5, ignoring the
// flag, the CWD file, and the default path entirely — eval scenarios must be
// reproducible across machines and CI runners.
func resolveConfig(isEval bool) error {
	if isEval || noUserConfigFlag {
		patternconfig.Init(embeddedConfig, embeddedConfigPath)
		return nil
	}

	if configFlag != "" {
		_, ok, err := patternconfig.LoadFromPath(configFlag)
		if err != nil {
			return fmt.Errorf("--config %s: %w", configFlag, err)
		}
		if !ok {
			return fmt.Errorf("--config %s: no such file", configFlag)
		}
		return nil
	}

	if cwd, err := os.Getwd(); err == nil {
		cwdPath := filepath.Join(cwd, userConfigFileName)
		switch _, ok, err := patternconfig.LoadFromPath(cwdPath); {
		case err != nil:
			warnf("ignoring %s: %v", cwdPath, err)
		case ok:
			return nil
		}
	}

	if defaultPath, err := defaultConfigPath(); err == nil {
		switch _, ok, err := patternconfig.LoadFromPath(defaultPath); {
		case err != nil:
			warnf("ignoring %s: %v", defaultPath, err)
		case ok:
			return nil
		default:
			patternconfig.Init(embeddedConfig, embeddedConfigPath)
			if werr := patternconfig.WriteDefault(defaultPath, patternconfig.Get()); werr != nil {
				warnf("could not write default pattern config to %s: %v", defaultPath, werr)
			}
			return nil
		}
	}

	patternconfig.Init(embeddedConfig, embeddedConfigPath)
	return nil
}

// defaultConfigPath returns the per-OS default config-directory path for
// scoring-patterns.yaml (see userConfigDirName/userConfigFileName above).
func defaultConfigPath() (string, error) {
	dir, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, userConfigDirName, userConfigFileName), nil
}

func warnf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "warning: "+format+"\n", a...)
}
