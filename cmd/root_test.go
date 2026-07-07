package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"testing"
)

func TestVersionCmd_run(t *testing.T) {
	// Exercises the versionCmd.Run closure (fmt.Printf path).
	versionCmd.Run(versionCmd, []string{})
}

func TestReleaseDate(t *testing.T) {
	origDate, origReader := buildDate, buildInfoReader
	t.Cleanup(func() { buildDate, buildInfoReader = origDate, origReader })

	// Disable the VCS fallback so each case exercises buildDate alone.
	buildInfoReader = func() (*debug.BuildInfo, bool) { return nil, false }

	cases := []struct {
		name      string
		buildDate string
		want      string
	}{
		{"rfc3339 renders DD-MM-YYYY", "2026-07-06T09:00:00Z", "06-07-2026"},
		{"empty date yields empty", "", ""},
		{"malformed date yields empty", "not-a-date", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buildDate = tc.buildDate
			if got := releaseDate(); got != tc.want {
				t.Errorf("releaseDate() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReleaseDate_vcsFallback(t *testing.T) {
	origDate, origReader := buildDate, buildInfoReader
	t.Cleanup(func() { buildDate, buildInfoReader = origDate, origReader })

	// No injected buildDate: releaseDate must fall back to the vcs.time stamp.
	buildDate = ""
	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.time", Value: "2026-01-02T03:04:05Z"},
		}}, true
	}
	if got, want := releaseDate(), "02-01-2026"; got != want {
		t.Errorf("releaseDate() vcs fallback = %q, want %q", got, want)
	}
}

func TestVersionString(t *testing.T) {
	origDate, origReader := buildDate, buildInfoReader
	t.Cleanup(func() { buildDate, buildInfoReader = origDate, origReader })
	buildInfoReader = func() (*debug.BuildInfo, bool) { return nil, false }

	buildDate = "2026-07-06T09:00:00Z"
	if got, want := versionString(), "skill-auditor v"+version+" (released 06-07-2026)"; got != want {
		t.Errorf("versionString() with date = %q, want %q", got, want)
	}

	buildDate = ""
	if got, want := versionString(), "skill-auditor v"+version; got != want {
		t.Errorf("versionString() without date = %q, want %q", got, want)
	}
}

func TestExitCodeFor_default(t *testing.T) {
	if code := exitCodeFor(fmt.Errorf("generic failure")); code != 1 {
		t.Errorf("want 1, got %d", code)
	}
}

func TestExitCodeFor_exitCoder(t *testing.T) {
	if code := exitCodeFor(criticalDuplicationError{}); code != 2 {
		t.Errorf("want 2, got %d", code)
	}
}

func TestExitCodeFor_wrappedExitCoder(t *testing.T) {
	wrapped := fmt.Errorf("running duplication: %w", criticalDuplicationError{})
	if code := exitCodeFor(wrapped); code != 2 {
		t.Errorf("want 2 for wrapped exitCoder, got %d", code)
	}
}

// TestMain isolates every test in this package from the real per-OS user
// config directory. Without this, any test that runs a command through
// rootCmd.Execute() triggers resolveConfig's default-path tier, which can
// auto-generate a real scoring-patterns.yaml under the developer's actual
// home directory (e.g. ~/Library/Application Support on macOS). Individual
// tests that need to exercise the default-path tier explicitly still
// override userConfigDir themselves, pointing it at a t.TempDir().
func TestMain(m *testing.M) {
	scratch, err := os.MkdirTemp("", "skill-auditor-test-userconfig-*")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	userConfigDir = func() (string, error) { return scratch, nil }

	os.Exit(m.Run())
}

// resetConfigFlags restores the persistent -c/--config and --no-user-config
// flag state to its zero value. pflag does not reset a flag to its default
// when the flag is simply absent from a later argv, so the ~15 cmd/*_test.go
// files that drive the shared rootCmd via SetArgs/Execute would otherwise
// leak flag values from one test into the next.
func resetConfigFlags(t *testing.T) {
	t.Helper()
	configFlag = ""
	noUserConfigFlag = false
}
