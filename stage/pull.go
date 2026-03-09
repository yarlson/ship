package stage

import (
	"context"
	"errors"
	"fmt"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
)

// Pull executes Stage 5: pull the image from the tunnel endpoint on the remote host
// and restore its original tag.
func Pull(ctx context.Context, cfg cli.Config, originals, transfers []string) error {
	if err := validateImageLists(originals, transfers); err != nil {
		return err
	}

	progress.StageStart(5, progressMessage(len(originals), "Pulling and restoring image", "Pulling and restoring images")+" on remote host")

	for i, original := range originals {
		transfer := transfers[i]

		pullCmd := fmt.Sprintf("docker pull %s", transfer)
		if _, err := ssh.RunRemoteCommand(ctx, cfg.KeyPath, cfg.Port, cfg.User, cfg.Host, pullCmd); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			return fmt.Errorf("failed to pull image on remote host — verify Docker is running on %s", cfg.Host)
		}

		tagCmd := fmt.Sprintf("docker tag %s %s", transfer, original)
		if _, err := ssh.RunRemoteCommand(ctx, cfg.KeyPath, cfg.Port, cfg.User, cfg.Host, tagCmd); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			return fmt.Errorf("failed to restore image tag on remote host — %s", original)
		}
	}

	progress.StageComplete(5, "Pull and restore complete")
	return nil
}
