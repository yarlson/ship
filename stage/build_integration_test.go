//go:build integration

package stage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/progress"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

func setupComposeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dockerfile := filepath.Join(dir, "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfile, []byte("FROM alpine:latest\nRUN echo hello\n"), 0o600))

	compose := filepath.Join(dir, "compose.yml")
	content := `services:
  web:
    build: .
    image: ship-inttest-web:latest
  api:
    build: .
    image: ship-inttest-api:latest
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

	return compose
}

func TestBuild_RealComposeFile(t *testing.T) {
	composePath := setupComposeProject(t)

	var imageMap map[string]string
	out := captureOutput(func() {
		var err error
		imageMap, err = Build(composePath)
		require.NoError(t, err)
	})

	// Verify 2 images discovered (redis excluded).
	require.Len(t, imageMap, 2)

	assert.Contains(t, imageMap, "ship-inttest-web:latest")
	assert.Contains(t, imageMap, "ship-inttest-api:latest")
	assert.Equal(t, "localhost:5001/ship-inttest-web:latest", imageMap["ship-inttest-web:latest"])
	assert.Equal(t, "localhost:5001/ship-inttest-api:latest", imageMap["ship-inttest-api:latest"])

	// Verify progress output.
	assert.Contains(t, out, "[1/7] Building images...")
	assert.Contains(t, out, "[1/7] Build complete (2 images)")
}

func TestBuild_NoImages(t *testing.T) {
	dir := t.TempDir()
	compose := filepath.Join(dir, "compose.yml")
	content := `services:
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

	_ = captureOutput(func() {
		_, err := Build(compose)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "No images found after build")
	})
}
