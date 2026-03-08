//go:build integration

package stage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_ExecutesSuccessfully(t *testing.T) {
	cfg := testSSHConfig(t)
	cfg.Command = "echo deployed"

	out := captureOutput(func() {
		err := Command(cfg)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[7/7] Running remote command...")
	assert.Contains(t, out, "[7/7] Command complete")
}

func TestCommand_NonZeroExit(t *testing.T) {
	cfg := testSSHConfig(t)
	cfg.Command = "exit 1"

	var err error
	captureOutput(func() {
		err = Command(cfg)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Remote command exited with code 1")
}

func TestCommand_CapturesExitCode(t *testing.T) {
	cfg := testSSHConfig(t)
	cfg.Command = "exit 42"

	var err error
	captureOutput(func() {
		err = Command(cfg)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code 42")
}
