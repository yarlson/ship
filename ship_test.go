package main_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "ship-e2e")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %s\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tmpDir, "ship")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func TestShip_Help_PrintsUsage(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), binaryPath, "--help")
	out, err := cmd.Output()
	require.NoError(t, err, "exit code should be 0 for --help")

	stdout := string(out)
	assert.Contains(t, stdout, "ship — build, transfer, and deploy Docker Compose images to a remote host")
	assert.Contains(t, stdout, "--docker-compose")
	assert.Contains(t, stdout, "--user")
	assert.Contains(t, stdout, "--host")
	assert.Contains(t, stdout, "--key")
	assert.Contains(t, stdout, "--command")
	assert.Contains(t, stdout, "Examples:")
}

func TestShip_AllFlags_PrintsSevenStages(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), binaryPath,
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	)
	out, err := cmd.Output()
	require.NoError(t, err, "exit code should be 0")

	stdout := string(out)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	assert.Len(t, lines, 14, "expected 14 lines (7 starts + 7 completes)")

	// Verify sequential stage numbers.
	stagePattern := regexp.MustCompile(`^\[(\d)/7\]`)
	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5", "6", "6", "7", "7"}
	for i, line := range lines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2, "line %d did not match: %s", i, line)
		assert.Equal(t, expectedOrder[i], matches[1])
	}

	// Contract rule 16: start messages end with "...".
	for i := 0; i < 14; i += 2 {
		assert.True(t, strings.HasSuffix(lines[i], "..."), "start line should end with ...: %s", lines[i])
	}

	// Contract rule 12: no ANSI codes.
	ansi := regexp.MustCompile(`\x1b\[`)
	assert.False(t, ansi.MatchString(stdout), "output contains ANSI escape codes")

	// Contract rule 13: no emoji.
	emoji := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	assert.False(t, emoji.MatchString(stdout), "output contains emoji")

	// Contract rule 30: no blank lines.
	assert.NotContains(t, stdout, "\n\n", "output contains blank lines")
}

func TestShip_NoFlags_PrintsMissingFlagsError(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), binaryPath)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err, "exit code should be non-zero")
	errOut := stderr.String()
	assert.Contains(t, errOut, "Error: Missing required flags:")
	assert.Contains(t, errOut, "--docker-compose")
	assert.Contains(t, errOut, "--user")
	assert.Contains(t, errOut, "--host")
	assert.Contains(t, errOut, "--key")
	assert.Contains(t, errOut, "--command")
	assert.Contains(t, errOut, "Usage: ship")
}

func TestShip_PartialFlags_ListsMissingOnes(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), binaryPath, "--host", "example.com", "--user", "deploy")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err, "exit code should be non-zero")
	errOut := stderr.String()
	assert.Contains(t, errOut, "--docker-compose")
	assert.Contains(t, errOut, "--key")
	assert.Contains(t, errOut, "--command")
	// Should not list the flags that were provided.
	assert.NotContains(t, errOut, "--host,")
	assert.NotContains(t, errOut, "--user,")
}
