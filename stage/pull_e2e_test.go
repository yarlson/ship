//go:build e2e

package stage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/ssh"
	"ship/testctx"
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
		require.NoError(t, Registry(testctx.New(t)))
	})
	for _, transfer := range transfers {
		require.NoError(t, tagAndPushTestImage(t, "alpine:latest", transfer))
	}

	var tp *ssh.TunnelProcess
	captureOutput(func() {
		var err error
		tp, err = Tunnel(testctx.New(t), cfg)
		require.NoError(t, err)
	})
	t.Cleanup(func() {
		ctx, cancel := testctx.Background()
		defer cancel()

		ssh.StopTunnel(ctx, tp) //nolint:errcheck // best-effort cleanup
	})

	out := captureOutput(func() {
		err := Pull(testctx.New(t), cfg, originals, transfers)
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
		tp, err = Tunnel(testctx.New(t), cfg)
		require.NoError(t, err)
	})
	t.Cleanup(func() {
		ctx, cancel := testctx.Background()
		defer cancel()

		ssh.StopTunnel(ctx, tp) //nolint:errcheck // best-effort cleanup
	})

	err := Pull(testctx.New(t), cfg, []string{"nonexistent:latest"}, []string{"localhost:5001/nonexistent-image-that-does-not-exist:latest"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull image on remote host")
}
