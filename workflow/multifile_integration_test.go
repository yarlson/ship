//go:build integration

package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/testlock"
)

func setupMultiFileComposeProject(t *testing.T) (baseCompose, overrideCompose string) {
	t.Helper()
	dir := t.TempDir()

	dockerfile := filepath.Join(dir, "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfile, []byte("FROM alpine:latest\nRUN echo hello\n"), 0o600))

	base := filepath.Join(dir, "compose.yml")
	baseContent := `services:
  web:
    build:
      context: .
      platforms:
        - linux/amd64
    image: ship-mftest-web:latest
    platform: linux/amd64
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(base, []byte(baseContent), 0o600))

	override := filepath.Join(dir, "compose.prod.yml")
	overrideContent := `services:
  web:
    environment:
      - NODE_ENV=production
`
	require.NoError(t, os.WriteFile(override, []byte(overrideContent), 0o600))

	return base, override
}

func TestMultiFileCompose_FullWorkflow_E2E(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	basePath, overridePath := setupMultiFileComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{basePath, overridePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployed",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// All 7 stages should complete.
	assert.Contains(t, out, "[1/7]")
	assert.Contains(t, out, "[7/7]")

	// Success summary should show original image names.
	assert.Contains(t, out, "Ship complete")
	assert.Contains(t, out, "ship-mftest-web:latest")
	assert.NotContains(t, out, "localhost:5001/")

	// Only built image transferred — redis should not appear in summary.
	assert.NotContains(t, out, "redis")

	// Build should report 1 image (only web has build key).
	assert.Contains(t, out, "Build complete (1 images)")
}

func TestMultiFileCompose_PulledImageExcluded(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	basePath, overridePath := setupMultiFileComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{basePath, overridePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployed",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// redis:alpine should not appear in the Images line of the summary.
	assert.Contains(t, out, "ship-mftest-web:latest")
	assert.NotContains(t, out, "redis:alpine")
}

func TestMultiFileCompose_MissingSecondFile(t *testing.T) {
	basePath, _ := setupMultiFileComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{basePath, "/tmp/nonexistent-compose-file.yml"},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployed",
	}

	out := captureOutput(func() {
		err := Preflight(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Compose file not found: /tmp/nonexistent-compose-file.yml")
	})

	// No stage progress should appear.
	assert.NotContains(t, out, "[1/7]")
}
