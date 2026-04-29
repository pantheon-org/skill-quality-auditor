package cmd

import "testing"

func TestVersionCmd_run(t *testing.T) {
	// Exercises the versionCmd.Run closure (fmt.Printf path).
	versionCmd.Run(versionCmd, []string{})
}
