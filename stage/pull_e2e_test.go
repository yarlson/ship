//go:build e2e

package stage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/ssh"
	"ship/testlock"
)

func TestPull_PullsAndRestoresImage(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	cfg := testSSHConfig(t)
	originals := []string{"ship-pulltest:latest", "ship-pulltest-proxy:v3"}
	transfers := []string{"localhost:5001/ship-pulltest:latest", "localhost:5001/ship-pulltest-proxy:v3"}

	captureOutput(func() {
		require.NoError(t, Registry())
	})
	for _, transfer := range transfers {
		require.NoError(t, tagAndPushTestImage(t, "alpine:latest", transfer))
	}

	var tp *ssh.TunnelProcess
	captureOutput(func() {
		var err error
		tp, err = Tunnel(cfg)
		require.NoError(t, err)
	})
	t.Cleanup(func() {
		ssh.StopTunnel(tp) //nolint:errcheck // best-effort cleanup
	})

	out := captureOutput(func() {
		err := Pull(cfg, originals, transfers)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[5/5] Pulling and restoring images on remote host...")
	assert.Contains(t, out, "[5/5] Pull and restore complete")
}

func TestPull_FailsWhenImageUnavailable(t *testing.T) {
	testlock.Port5001(t)

	cfg := testSSHConfig(t)

	var tp *ssh.TunnelProcess
	captureOutput(func() {
		var err error
		tp, err = Tunnel(cfg)
		require.NoError(t, err)
	})
	t.Cleanup(func() {
		ssh.StopTunnel(tp) //nolint:errcheck // best-effort cleanup
	})

	err := Pull(cfg, []string{"nonexistent:latest"}, []string{"localhost:5001/nonexistent-image-that-does-not-exist:latest"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull image on remote host")
}
