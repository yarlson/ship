package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/testctx"
)

var binaryPath string

func TestMain(m *testing.M) {
	ctx, cancel := testctx.Background()

	tmpDir, err := os.MkdirTemp("", "ship-cli")
	if err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %s\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tmpDir, "ship")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()
	cancel()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func TestShip_Help_PrintsUsage(t *testing.T) {
	cmd := exec.CommandContext(testctx.New(t), binaryPath, "--help")
	out, err := cmd.Output()
	require.NoError(t, err)

	stdout := string(out)
	assert.Contains(t, stdout, "ship — transfer local Docker images to a remote host over SSH")
	assert.Contains(t, stdout, "<user@host> <image[:tag]> [<image[:tag]>...]")
	assert.Contains(t, stdout, "-i, --identity-file")
	assert.Contains(t, stdout, "-p, --port")
}

func TestShip_NoArgs_PrintsMissingArgumentsError(t *testing.T) {
	cmd := exec.CommandContext(testctx.New(t), binaryPath)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err)
	errOut := stderr.String()
	assert.Contains(t, errOut, "Error: missing required arguments:")
	assert.Contains(t, errOut, "<user@host>")
	assert.Contains(t, errOut, "<image[:tag]>")
	assert.Contains(t, errOut, "Usage: ship")
}

func TestShip_OnlyTarget_PrintsMissingImageError(t *testing.T) {
	cmd := exec.CommandContext(testctx.New(t), binaryPath, "deploy@example.com")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err)
	assert.Contains(t, stderr.String(), "missing required arguments: <image[:tag]>")
}
