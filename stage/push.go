package stage

import (
	"fmt"

	"ship/docker"
	"ship/progress"
)

// Push executes Stage 4: push all transfer-tagged images to the local registry.
// imageMap maps original image ref → localhost:5001/ transfer tag.
func Push(imageMap map[string]string) error {
	progress.StageStart(4, "Pushing images to local registry")

	count := 0
	for _, transfer := range imageMap {
		if err := docker.PushImage(transfer); err != nil {
			return fmt.Errorf("Failed to push %s — %w", transfer, err) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}
		count++
	}

	progress.StageComplete(4, fmt.Sprintf("Push complete (%d images)", count))
	return nil
}
