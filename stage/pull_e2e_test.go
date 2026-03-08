//go:build e2e

package stage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/ssh"
	"ship/testlock"
)

func TestPull_PullsAndRestoresImages(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	cfg := testSSHConfig(t)

	// Start registry, then tag and push a test image.
	captureOutput(func() {
		require.NoError(t, Registry())
	})
	require.NoError(t, tagAndPushTestImage(t, "alpine:latest", "localhost:5001/ship-pulltest:latest"))

	// Establish a tunnel so the remote can reach our local registry.
	var tp *ssh.TunnelProcess
	captureOutput(func() {
		var err error
		tp, err = Tunnel(cfg)
		require.NoError(t, err)
	})
	t.Cleanup(func() {
		ssh.StopTunnel(tp) //nolint:errcheck // best-effort cleanup
	})

	imageMap := map[string]string{
		"ship-pulltest:latest": "localhost:5001/ship-pulltest:latest",
	}

	out := captureOutput(func() {
		err := Pull(cfg, imageMap)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[6/7] Pulling and restoring images on remote host...")
	assert.Contains(t, out, "[6/7] Pull and restore complete (1 images)")
}

func TestPull_FailsWhenDockerUnreachable(t *testing.T) {
	testlock.Port5001(t)

	cfg := testSSHConfig(t)

	// Use an image that doesn't exist in any registry accessible via tunnel.
	imageMap := map[string]string{
		"nonexistent:latest": "localhost:5001/nonexistent-image-that-does-not-exist:latest",
	}

	// Establish a tunnel.
	var tp *ssh.TunnelProcess
	captureOutput(func() {
		var err error
		tp, err = Tunnel(cfg)
		require.NoError(t, err)
	})
	t.Cleanup(func() {
		ssh.StopTunnel(tp) //nolint:errcheck // best-effort cleanup
	})

	var err error
	captureOutput(func() {
		err = Pull(cfg, imageMap)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to pull images on remote host")
}
