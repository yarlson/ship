//go:build e2e

package stage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/ssh"
)

func TestTunnel_EstablishesAndStoresInState(t *testing.T) {
	cfg := testSSHConfig(t)

	var tunnelCmd interface{}
	out := captureOutput(func() {
		cmd, err := Tunnel(cfg)
		require.NoError(t, err)
		require.NotNil(t, cmd)
		tunnelCmd = cmd
		t.Cleanup(func() {
			ssh.StopTunnel(cmd) //nolint:errcheck // best-effort cleanup in tests
		})
	})

	assert.NotNil(t, tunnelCmd)
	assert.Contains(t, out, "[5/7] Establishing tunnel to 46.101.213.82...")
	assert.Contains(t, out, "[5/7] Tunnel established")
}

func TestTunnel_FailsOnBadHost(t *testing.T) {
	cfg := testSSHConfig(t)
	// Use a bad key so SSH exits immediately with auth failure.
	badKey := filepath.Join(t.TempDir(), "bad_key")
	require.NoError(t, os.WriteFile(badKey, []byte("not-a-key"), 0o600))
	cfg.KeyPath = badKey

	var err error
	captureOutput(func() {
		_, err = Tunnel(cfg)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH tunnel failed")
}
