package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestVersionCmd_run(t *testing.T) {
	// Exercises the versionCmd.Run closure (fmt.Printf path).
	versionCmd.Run(versionCmd, []string{})
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
