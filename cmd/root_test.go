package cmd

import (
	"fmt"
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
