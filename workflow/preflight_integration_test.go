//go:build integration

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDocker_Available(t *testing.T) {
	cmdErr := checkDocker()
	if cmdErr != nil {
		t.Skipf("skipping integration test: Docker daemon unavailable: %v", cmdErr)
	}
	assert.NoError(t, cmdErr)
}

func TestCheckDockerCompose_Available(t *testing.T) {
	err := checkDockerCompose()
	assert.NoError(t, err)
}

func TestCheckSSH_Available(t *testing.T) {
	err := checkSSH()
	assert.NoError(t, err)
}

func TestCheckSSHConnectivity_Unreachable(t *testing.T) {
	// Use bad key file for instant failure instead of unreachable IP (which hangs).
	err := checkSSHConnectivity("/tmp/nonexistent-key-for-test", "root", "192.0.2.1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH connection failed")
}
