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
	original := "ship-pulltest:latest"
	transfer := "localhost:5001/ship-pulltest:latest"

	captureOutput(func() {
		require.NoError(t, Registry())
	})
	require.NoError(t, tagAndPushTestImage(t, "alpine:latest", transfer))

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
		err := Pull(cfg, original, transfer)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[5/5] Pulling and restoring image on remote host...")
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

	err := Pull(cfg, "nonexistent:latest", "localhost:5001/nonexistent-image-that-does-not-exist:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull image on remote host")
}
