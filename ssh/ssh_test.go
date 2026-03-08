package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSSHArgs_RemoteCommand(t *testing.T) {
	args := BuildRemoteCommandArgs("/home/user/.ssh/id_rsa", "deploy", "10.0.0.5", "echo hello")

	require.Contains(t, args, "-i")
	require.Contains(t, args, "/home/user/.ssh/id_rsa")
	require.Contains(t, args, "deploy@10.0.0.5")
	require.Contains(t, args, "echo hello")

	// Verify argument order: -i comes before keyPath.
	for i, a := range args {
		if a == "-i" {
			assert.Equal(t, "/home/user/.ssh/id_rsa", args[i+1])
			break
		}
	}
}

func TestBuildSSHArgs_Tunnel(t *testing.T) {
	args := BuildTunnelArgs("/home/user/.ssh/id_rsa", "root", "example.com")

	require.Contains(t, args, "-i")
	require.Contains(t, args, "/home/user/.ssh/id_rsa")
	require.Contains(t, args, "-R")
	require.Contains(t, args, "5001:localhost:5001")
	require.Contains(t, args, "-N")
	require.Contains(t, args, "root@example.com")
}

func TestNoKeyContentsInArgs(t *testing.T) {
	keyPath := "/home/user/.ssh/id_rsa"

	remoteArgs := BuildRemoteCommandArgs(keyPath, "user", "host", "cmd")
	tunnelArgs := BuildTunnelArgs(keyPath, "user", "host")

	allArgs := make([]string, 0, len(remoteArgs)+len(tunnelArgs))
	allArgs = append(allArgs, remoteArgs...)
	allArgs = append(allArgs, tunnelArgs...)
	for _, arg := range allArgs {
		assert.NotContains(t, arg, "BEGIN")
		assert.NotContains(t, arg, "PRIVATE")
		assert.NotContains(t, arg, "KEY")
	}
}
