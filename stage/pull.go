package stage

import (
	"fmt"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
)

// Pull executes Stage 5: pull the image from the tunnel endpoint on the remote host
// and restore its original tag.
func Pull(cfg cli.Config, originals, transfers []string) error {
	progress.StageStart(5, progressMessage(len(originals), "Pulling and restoring image", "Pulling and restoring images")+" on remote host")

	for i, original := range originals {
		transfer := transfers[i]

		pullCmd := fmt.Sprintf("docker pull %s", transfer)
		_, _, exitCode, err := ssh.RunRemoteCommand(cfg.KeyPath, cfg.Port, cfg.User, cfg.Host, pullCmd)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to pull image on remote host — verify Docker is running on %s", cfg.Host)
		}

		tagCmd := fmt.Sprintf("docker tag %s %s", transfer, original)
		_, _, exitCode, err = ssh.RunRemoteCommand(cfg.KeyPath, cfg.Port, cfg.User, cfg.Host, tagCmd)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to restore image tag on remote host — %s", original)
		}
	}

	progress.StageComplete(5, "Pull and restore complete")
	return nil
}
