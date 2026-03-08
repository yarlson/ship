//go:build integration

package main_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupComposeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dockerfile := filepath.Join(dir, "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfile, []byte("FROM alpine:latest\nRUN echo hello\n"), 0o600))

	compose := filepath.Join(dir, "compose.yml")
	content := `services:
  web:
    build: .
    image: ship-test-web:latest
  api:
    build: .
    image: ship-test-api:latest
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

	return compose
}

func TestShip_AllFlags_PrintsSevenStages(t *testing.T) {
	composePath := setupComposeProject(t)

	cmd := exec.CommandContext(context.Background(), binaryPath,
		"--docker-compose", composePath,
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	)
	out, err := cmd.Output()
	require.NoError(t, err, "exit code should be 0")

	stdout := string(out)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	// Filter to only stage progress lines (skip Docker build output).
	var stageLines []string
	stagePattern := regexp.MustCompile(`^\[(\d)/7\]`)
	for _, line := range lines {
		if stagePattern.MatchString(line) {
			stageLines = append(stageLines, line)
		}
	}

	assert.Len(t, stageLines, 14, "expected 14 stage lines (7 starts + 7 completes)")

	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5", "6", "6", "7", "7"}
	for i, line := range stageLines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2, "line %d did not match: %s", i, line)
		assert.Equal(t, expectedOrder[i], matches[1])
	}

	// Contract rule 16: start messages end with "...".
	for i := 0; i < len(stageLines); i += 2 {
		assert.True(t, strings.HasSuffix(stageLines[i], "..."), "start line should end with ...: %s", stageLines[i])
	}

	// Contract rule 12: no ANSI codes in stage lines.
	ansi := regexp.MustCompile(`\x1b\[`)
	for _, line := range stageLines {
		assert.False(t, ansi.MatchString(line), "stage line contains ANSI escape codes: %s", line)
	}

	// Contract rule 13: no emoji in stage lines.
	emoji := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	for _, line := range stageLines {
		assert.False(t, emoji.MatchString(line), "stage line contains emoji: %s", line)
	}

	// Verify build stage reports correct image count.
	assert.Contains(t, stdout, "Build complete (2 images)")
	assert.Contains(t, stdout, "Tag complete")
}

func TestShip_NoBuildServices_FailsWithError(t *testing.T) {
	dir := t.TempDir()
	compose := filepath.Join(dir, "compose.yml")
	content := `services:
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

	cmd := exec.CommandContext(context.Background(), binaryPath,
		"--docker-compose", compose,
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err, "exit code should be non-zero")
	assert.Contains(t, stderr.String(), "No images found after build")
}

func TestShip_TransferTagsExist(t *testing.T) {
	composePath := setupComposeProject(t)

	cmd := exec.CommandContext(context.Background(), binaryPath,
		"--docker-compose", composePath,
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "ship should succeed: %s", string(out))

	// Verify transfer tags were created.
	for _, img := range []string{"localhost:5001/ship-test-web:latest", "localhost:5001/ship-test-api:latest"} {
		inspect := exec.CommandContext(context.Background(), "docker", "image", "inspect", img)
		require.NoError(t, inspect.Run(), "transfer tag should exist: %s", img)
	}

	// Verify image-only service was NOT tagged.
	inspect := exec.CommandContext(context.Background(), "docker", "image", "inspect", "localhost:5001/redis:alpine")
	assert.Error(t, inspect.Run(), "redis should not have a transfer tag")
}
