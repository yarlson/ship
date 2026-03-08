package stage

import (
	"fmt"

	"ship/docker"
	"ship/progress"
)

// Tag executes Stage 1: re-tag the local image with a localhost:5001/ prefix for transfer.
func Tag(original, transfer string) error {
	progress.StageStart(1, "Tagging image for transfer")

	if err := docker.TagImage(original, transfer); err != nil {
		return fmt.Errorf("failed to tag %s — %w", original, err)
	}

	progress.StageComplete(1, "Tag complete")
	return nil
}
