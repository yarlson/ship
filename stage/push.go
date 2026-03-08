package stage

import (
	"fmt"

	"ship/docker"
	"ship/progress"
)

// Push executes Stage 3: push the transfer-tagged image to the local registry.
func Push(transfers []string) error {
	progress.StageStart(3, "Pushing image to local registry")

	for _, transfer := range transfers {
		if err := docker.PushImage(transfer); err != nil {
			return fmt.Errorf("failed to push %s — %w", transfer, err)
		}
	}

	progress.StageComplete(3, "Push complete")
	return nil
}
