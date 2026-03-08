package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Image represents a Docker image with name and tag.
type Image struct {
	Name string
	Tag  string
}

// Ref returns the full image reference as name:tag.
func (i Image) Ref() string {
	return i.Name + ":" + i.Tag
}

// TransferTag returns the localhost:5001/ prefixed tag for an image.
func TransferTag(img Image) string {
	return "localhost:5001/" + img.Ref()
}

// ComposeBuild runs docker compose build with the given compose files.
// Stdout and stderr are connected directly to the parent process for passthrough output.
func ComposeBuild(composeFiles []string) error {
	args := make([]string, 0, 1+2*len(composeFiles)+1)
	args = append(args, "compose")
	for _, f := range composeFiles {
		args = append(args, "-f", f)
	}
	args = append(args, "build")

	cmd := exec.CommandContext(context.Background(), "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose build: %w", err)
	}
	return nil
}

// ComposeConfig runs docker compose config and returns the images from services with build keys.
func ComposeConfig(composeFiles []string) ([]Image, error) {
	args := make([]string, 0, 1+2*len(composeFiles)+3)
	args = append(args, "compose")
	for _, f := range composeFiles {
		args = append(args, "-f", f)
	}
	args = append(args, "config", "--format", "json")

	cmd := exec.CommandContext(context.Background(), "docker", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker compose config: %w", err)
	}

	return ParseComposeConfig(out)
}

// composeConfig is the subset of docker compose config JSON we need.
type composeConfig struct {
	Name     string                    `json:"name"`
	Services map[string]composeService `json:"services"`
}

type composeService struct {
	Build json.RawMessage `json:"build"`
	Image string          `json:"image"`
}

// ParseComposeConfig parses docker compose config JSON output and returns images
// from services that have a build key.
func ParseComposeConfig(data []byte) ([]Image, error) {
	var cfg composeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse compose config: %w", err)
	}

	var images []Image
	for svcName, svc := range cfg.Services {
		if len(svc.Build) == 0 || string(svc.Build) == "null" {
			continue
		}

		imageRef := svc.Image
		if imageRef == "" {
			imageRef = cfg.Name + "-" + svcName
		}

		images = append(images, parseImageRef(imageRef))
	}

	return images, nil
}

// parseImageRef splits an image reference into name and tag.
// If no tag is present, defaults to "latest".
func parseImageRef(ref string) Image {
	// Handle images with registry prefix (e.g., ghcr.io/org/app:v1)
	// The tag is after the last colon, but only if the last colon is after the last slash.
	lastColon := strings.LastIndex(ref, ":")
	lastSlash := strings.LastIndex(ref, "/")

	if lastColon > lastSlash && lastColon != -1 {
		return Image{
			Name: ref[:lastColon],
			Tag:  ref[lastColon+1:],
		}
	}

	return Image{
		Name: ref,
		Tag:  "latest",
	}
}

// TagImage runs docker tag to create a new tag for an image.
func TagImage(source, target string) error {
	cmd := exec.CommandContext(context.Background(), "docker", "tag", source, target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker tag %s → %s: %s", source, target, strings.TrimSpace(string(out)))
	}
	return nil
}
