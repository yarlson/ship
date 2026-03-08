//go:build integration

package stage

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTag_CreatesTransferTag(t *testing.T) {
	requireDocker(t)
	originals := []string{"ship-inttest-app:latest", "ship-inttest-traefik:v3"}
	transfers := []string{"localhost:5001/ship-inttest-app:latest", "localhost:5001/ship-inttest-traefik:v3"}
	for _, original := range originals {
		ensureLocalImage(t, original)
	}

	out := captureOutput(func() {
		err := Tag(originals, transfers)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[1/5] Tagging image for transfer...")
	assert.Contains(t, out, "[1/5] Tag complete")

	for _, transfer := range transfers {
		cmd := exec.CommandContext(context.Background(), "docker", "image", "inspect", transfer)
		require.NoError(t, cmd.Run(), "transfer tag should exist: %s", transfer)
	}
}

func TestTag_FailsOnMissingSourceImage(t *testing.T) {
	requireDocker(t)

	err := Tag([]string{"ship-missing-image:latest"}, []string{"localhost:5001/ship-missing-image:latest"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to tag ship-missing-image:latest")
}
