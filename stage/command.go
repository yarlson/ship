package stage

import (
	"fmt"
	"os"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
)

// Command executes Stage 7: run the user-provided command on the remote host.
func Command(cfg cli.Config) error {
	progress.StageStart(7, "Running remote command")

	stdout, stderr, exitCode, err := ssh.RunRemoteCommand(cfg.KeyPath, cfg.User, cfg.Host, cfg.Command)

	// Passthrough output (contract rule 28: not reformatted).
	if stdout != "" {
		fmt.Fprint(progress.Writer, stdout)
	}
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	if err != nil || exitCode != 0 {
		return fmt.Errorf("Remote command exited with code %d — see output above", exitCode) //nolint:staticcheck // user-facing message per DESIGN.md spec
	}

	progress.StageComplete(7, "Command complete")
	return nil
}
