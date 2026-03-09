package stage

import (
	"context"
	"fmt"

	"ship/docker"
	"ship/progress"
)

// Tag executes Stage 1: re-tag the local image with a localhost:5001/ prefix for transfer.
func Tag(ctx context.Context, originals, transfers []string) error {
	if err := validateImageLists(originals, transfers); err != nil {
		return err
	}

	progress.StageStart(1, progressMessage(len(originals), "Tagging image", "Tagging images")+" for transfer")

	for i, original := range originals {
		if err := docker.TagImage(ctx, original, transfers[i]); err != nil {
			return fmt.Errorf("failed to tag %s — %w", original, err)
		}
	}

	progress.StageComplete(1, "Tag complete")
	return nil
}
