package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// TransferTag returns the localhost:5001/ prefixed tag for an image reference.
func TransferTag(imageRef string) string {
	return "localhost:5001/" + imageRef
}

// TransferTags returns the localhost:5001/ prefixed tags for image references.
func TransferTags(imageRefs []string) []string {
	transferRefs := make([]string, 0, len(imageRefs))
	for _, imageRef := range imageRefs {
		transferRefs = append(transferRefs, TransferTag(imageRef))
	}

	return transferRefs
}

// ImageExists verifies that a local image reference exists.
func ImageExists(ctx context.Context, imageRef string) error {
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageRef)
	if out, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(msg, "No such object") || strings.Contains(msg, "No such image") {
			return fmt.Errorf("local image not found: %s — build or pull it first", imageRef)
		}
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("failed to inspect local image: %s — %s", imageRef, msg)
	}

	return nil
}

// TagImage runs docker tag to create a new tag for an image.
func TagImage(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, "docker", "tag", source, target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker tag %s → %s: %s", source, target, strings.TrimSpace(string(out)))
	}
	return nil
}
