package stage

import (
	"fmt"

	"ship/docker"
	"ship/progress"
)

// Tag executes Stage 2: re-tag images with localhost:5001/ prefix for transfer.
// imageMap maps original image ref → transfer tag.
func Tag(imageMap map[string]string) error {
	progress.StageStart(2, "Tagging images for transfer")

	for original, transfer := range imageMap {
		if err := docker.TagImage(original, transfer); err != nil {
			return fmt.Errorf("Failed to tag %s — %w", original, err) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}
	}

	progress.StageComplete(2, "Tag complete")
	return nil
}
