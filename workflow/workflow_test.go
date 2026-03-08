//go:build integration

package workflow

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/progress"
	"ship/testlock"
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
    image: ship-wftest-web:latest
  api:
    build: .
    image: ship-wftest-api:latest
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

	return compose
}

func TestRun_PrintsAllSevenStages(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	cfg := cli.Config{
		ComposeFiles: composePath,
		User:         "deploy",
		Host:         "10.0.0.5",
		KeyPath:      "~/.ssh/id_ed25519",
		Command:      "docker compose up -d",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// Progress output captured via progress.Writer contains only stage lines.
	for i := 1; i <= 7; i++ {
		pattern := regexp.MustCompile(`\[` + string(rune('0'+i)) + `/7\]`)
		assert.True(t, pattern.MatchString(out), "missing stage %d", i)
	}
}

func TestRun_StagesInOrder(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	cfg := cli.Config{
		ComposeFiles: composePath,
		User:         "deploy",
		Host:         "10.0.0.5",
		KeyPath:      "~/.ssh/id_ed25519",
		Command:      "docker compose up -d",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// Extract only stage progress lines from output.
	stagePattern := regexp.MustCompile(`^\[(\d)/7\]`)
	var stageLines []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if stagePattern.MatchString(line) {
			stageLines = append(stageLines, line)
		}
	}

	require.Len(t, stageLines, 14, "expected 14 stage lines (7 starts + 7 completes)")

	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5", "6", "6", "7", "7"}
	for i, line := range stageLines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2, "line %d did not match stage pattern: %s", i, line)
		assert.Equal(t, expectedOrder[i], matches[1], "line %d: expected stage %s, got %s", i, expectedOrder[i], matches[1])
	}
}

func TestRun_ReturnsNilOnSuccess(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	cfg := cli.Config{
		ComposeFiles: composePath,
		User:         "deploy",
		Host:         "10.0.0.5",
		KeyPath:      "~/.ssh/id_ed25519",
		Command:      "docker compose up -d",
	}

	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()

	err := Run(cfg)
	assert.NoError(t, err)
}
