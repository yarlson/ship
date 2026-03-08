//go:build e2e

package ssh

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireE2EPrereqs(t *testing.T, keyPath, user, host string) {
	t.Helper()

	dockerCmd := exec.CommandContext(context.Background(), "docker", "version")
	if err := dockerCmd.Run(); err != nil {
		t.Skipf("skipping e2e test: Docker daemon unavailable: %v", err)
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Skipf("skipping e2e test: SSH key unavailable: %v", err)
	}

	sshCmd := exec.CommandContext(context.Background(), "ssh",
		"-i", keyPath,
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		user+"@"+host,
		"true",
	)
	if err := sshCmd.Run(); err != nil {
		t.Skipf("skipping e2e test: SSH test host unavailable: %v", err)
	}
}

func testSSHConfig(t *testing.T) (keyPath, user, host string) {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	keyPath = home + "/.ssh/id_rsa"
	user = "root"
	host = "46.101.213.82"
	requireE2EPrereqs(t, keyPath, user, host)
	return keyPath, user, host
}

func TestRunRemoteCommand_Success(t *testing.T) {
	keyPath, user, host := testSSHConfig(t)

	stdout, _, exitCode, err := RunRemoteCommand(keyPath, user, host, "echo hello")
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "hello\n", stdout)
}

func TestRunRemoteCommand_NonZeroExit(t *testing.T) {
	keyPath, user, host := testSSHConfig(t)

	_, _, exitCode, err := RunRemoteCommand(keyPath, user, host, "exit 42")
	require.Error(t, err)
	assert.Equal(t, 42, exitCode)
}

func TestRunRemoteCommand_StderrCapture(t *testing.T) {
	keyPath, user, host := testSSHConfig(t)

	_, stderr, _, err := RunRemoteCommand(keyPath, user, host, "echo err >&2")
	require.NoError(t, err)
	assert.Equal(t, "err\n", stderr)
}

func TestStartTunnel_ProcessAlive(t *testing.T) {
	keyPath, user, host := testSSHConfig(t)

	tp, err := StartTunnel(keyPath, user, host)
	require.NoError(t, err)
	require.NotNil(t, tp)
	t.Cleanup(func() {
		StopTunnel(tp) //nolint:errcheck // best-effort cleanup in tests
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
	keyPath, user, host := testSSHConfig(t)

	tp, err := StartTunnel(keyPath, user, host)
	require.NoError(t, err)

	err = StopTunnel(tp)
	require.NoError(t, err)

	// Process should have exited — Exited channel should be closed.
	select {
	case <-tp.Exited():
		// Expected — process is dead.
	default:
		t.Fatal("tunnel process still running after StopTunnel")
	}
}
