//go:build e2e

package ssh

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/testctx"
	"ship/testenv"
)

func TestRunRemoteCommand_Success(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)

	result, err := RunRemoteCommand(testctx.New(t), cfg.KeyPath, 22, cfg.User, cfg.Host, "echo hello")
	require.NoError(t, err)
	assert.Equal(t, "hello\n", result.Stdout)
}

func TestRunRemoteCommand_NonZeroExit(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)

	_, err := RunRemoteCommand(testctx.New(t), cfg.KeyPath, 22, cfg.User, cfg.Host, "exit 42")
	require.Error(t, err)

	var remoteErr *RemoteCommandError
	require.True(t, errors.As(err, &remoteErr))
	assert.Equal(t, 42, remoteErr.ExitCode)
}

func TestRunRemoteCommand_StderrCapture(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)

	result, err := RunRemoteCommand(testctx.New(t), cfg.KeyPath, 22, cfg.User, cfg.Host, "echo err >&2")
	require.NoError(t, err)
	assert.Equal(t, "err\n", result.Stderr)
}

func TestStartTunnel_ProcessAlive(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)

	tp, err := StartTunnel(testctx.New(t), cfg.KeyPath, 22, cfg.User, cfg.Host)
	require.NoError(t, err)
	require.NotNil(t, tp)
	t.Cleanup(func() {
		ctx, cancel := testctx.Background()
		defer cancel()

		StopTunnel(ctx, tp) //nolint:errcheck // best-effort cleanup in tests
	})

	// Process should be alive — Exited channel should not be closed yet.
	select {
	case <-tp.Exited():
		t.Fatal("tunnel process exited unexpectedly")
	default:
		// Expected — process is alive.
	}
}

func TestStopTunnel_Cleanup(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)

	tp, err := StartTunnel(testctx.New(t), cfg.KeyPath, 22, cfg.User, cfg.Host)
	require.NoError(t, err)

	err = StopTunnel(testctx.New(t), tp)
	require.NoError(t, err)

	// Process should have exited — Exited channel should be closed.
	select {
	case <-tp.Exited():
		// Expected — process is dead.
	default:
		t.Fatal("tunnel process still running after StopTunnel")
	}
}
