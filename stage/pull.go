package stage

import (
	"fmt"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
)

// Pull executes Stage 6: pull images from the tunnel endpoint on the remote host
// and restore their original tags.
// imageMap maps original image ref → localhost:5001/ transfer tag.
func Pull(cfg cli.Config, imageMap map[string]string) error {
	progress.StageStart(6, "Pulling and restoring images on remote host")

	count := 0
	for original, transfer := range imageMap {
		// Pull the image via the tunnel.
		pullCmd := fmt.Sprintf("docker pull %s", transfer)
		_, _, exitCode, err := ssh.RunRemoteCommand(cfg.KeyPath, cfg.User, cfg.Host, pullCmd)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("Failed to pull images on remote host — verify Docker is running on %s", cfg.Host) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}

		// Restore the original tag.
		tagCmd := fmt.Sprintf("docker tag %s %s", transfer, original)
		_, _, exitCode, err = ssh.RunRemoteCommand(cfg.KeyPath, cfg.User, cfg.Host, tagCmd)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("Failed to restore image tag on remote host — %s", original) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}

		count++
	}

	progress.StageComplete(6, fmt.Sprintf("Pull and restore complete (%d images)", count))
	return nil
}
