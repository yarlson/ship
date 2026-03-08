//go:build integration

package stage

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/testlock"
)

// registryRunning checks if a registry:2 container is running on port 5001.
func registryRunning(t *testing.T) bool {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "docker", "ps", "-q", "--filter", "ancestor=registry:2", "--filter", "publish=5001")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func TestRegistry_StartsContainer(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	captureOutput(func() {
		err := Registry()
		require.NoError(t, err)
	})

	assert.True(t, registryRunning(t), "registry container should be running on port 5001")
}

func TestRegistry_ReusesExisting(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	// Start registry manually first.
	cmd := exec.CommandContext(context.Background(), "docker", "run", "-d", "-p", "5001:5000", "registry:2")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to start registry: %s", string(out))

	captureOutput(func() {
		err := Registry()
		require.NoError(t, err)
	})

	// Verify only one registry container on 5001.
	cmd = exec.CommandContext(context.Background(), "docker", "ps", "-q", "--filter", "ancestor=registry:2", "--filter", "publish=5001")
	out, err = cmd.Output()
	require.NoError(t, err)
	ids := strings.Split(strings.TrimSpace(string(out)), "\n")
	assert.Len(t, ids, 1, "should have exactly one registry container, not two")
}

func TestRegistry_PortConflict(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)

	// Start a non-registry container on port 5001.
	cmd := exec.CommandContext(context.Background(), "docker", "run", "-d", "-p", "5001:80", "nginx:alpine")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to start nginx: %s", string(out))
	nginxID := strings.TrimSpace(string(out))
	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup in tests
		exec.CommandContext(context.Background(), "docker", "rm", "-f", nginxID).Run()
		testlock.WaitPort5001Free(t)
	})

	captureOutput(func() {
		err := Registry()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Port 5001 already in use")
	})
}

func TestRegistry_ProgressOutput(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	out := captureOutput(func() {
		err := Registry()
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[2/5] Starting local registry...")
	assert.Contains(t, out, "[2/5] Registry ready")
}
