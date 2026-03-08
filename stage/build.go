package stage

import (
	"fmt"
	"strings"

	"ship/docker"
	"ship/progress"
)

// Build executes Stage 1: docker compose build and image discovery.
// Returns a map of original image ref → localhost:5001/ transfer tag.
func Build(composeFiles string) (map[string]string, error) {
	progress.StageStart(1, "Building images")

	files := strings.Split(composeFiles, ",")

	if err := docker.ComposeBuild(files); err != nil {
		return nil, fmt.Errorf("Build failed — see docker compose output above")
	}

	images, err := docker.ComposeConfig(files)
	if err != nil {
		return nil, fmt.Errorf("Build failed — %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("No images found after build — check that services in the compose file have a 'build' key") //nolint:staticcheck // user-facing message per DESIGN.md spec
	}

	imageMap := make(map[string]string, len(images))
	for _, img := range images {
		imageMap[img.Ref()] = docker.TransferTag(img)
	}

	progress.StageComplete(1, fmt.Sprintf("Build complete (%d images)", len(images)))
	return imageMap, nil
}
