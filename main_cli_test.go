package main_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "ship-cli")
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
