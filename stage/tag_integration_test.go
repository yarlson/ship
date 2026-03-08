//go:build integration

package stage

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTag_CreatesTransferTags(t *testing.T) {
	composePath := setupComposeProject(t)

	// First build images.
	var imageMap map[string]string
	captureOutput(func() {
		var err error
		imageMap, err = Build(composePath)
		require.NoError(t, err)
	})

	// Now tag them.
	captureOutput(func() {
		err := Tag(imageMap)
		require.NoError(t, err)
	})

	// Verify transfer tags exist via docker image inspect.
	for _, transfer := range imageMap {
		cmd := exec.CommandContext(context.Background(), "docker", "image", "inspect", transfer)
		require.NoError(t, cmd.Run(), "transfer tag should exist: %s", transfer)
	}
}

func TestBuildAndTag_HappyPath_E2E(t *testing.T) {
	composePath := setupComposeProject(t)

	// Stage 1: Build.
	var imageMap map[string]string
	out := captureOutput(func() {
		var err error
		imageMap, err = Build(composePath)
		require.NoError(t, err)
	})

	require.Len(t, imageMap, 2)
	assert.Contains(t, out, "[1/7] Building images...")
	assert.Contains(t, out, "[1/7] Build complete (2 images)")

	// Stage 2: Tag.
	out = captureOutput(func() {
		err := Tag(imageMap)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[2/7] Tagging images for transfer...")
	assert.Contains(t, out, "[2/7] Tag complete")

	// Verify transfer tags exist; image-only service excluded.
	for original, transfer := range imageMap {
		inspect := exec.CommandContext(context.Background(), "docker", "image", "inspect", transfer)
		require.NoError(t, inspect.Run(), "transfer tag should exist: %s (from %s)", transfer, original)
	}

	// Verify redis was NOT tagged.
	inspect := exec.CommandContext(context.Background(), "docker", "image", "inspect", "localhost:5001/redis:alpine")
	assert.Error(t, inspect.Run(), "redis should not have a transfer tag")
}
