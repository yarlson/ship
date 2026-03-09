//go:build integration

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/testctx"
)

func TestCheckDocker_Available(t *testing.T) {
	cmdErr := checkDocker(testctx.New(t))
	if cmdErr != nil {
		t.Skipf("skipping integration test: Docker daemon unavailable: %v", cmdErr)
	}
	assert.NoError(t, cmdErr)
}

func TestCheckSSH_Available(t *testing.T) {
	err := checkSSH()
	assert.NoError(t, err)
}

func TestCheckSSHConnectivity_Unreachable(t *testing.T) {
	err := checkSSHConnectivity(testctx.New(t), cli.Config{
		User:    "root",
		Host:    "192.0.2.1",
		KeyPath: "/tmp/nonexistent-key-for-test",
		Port:    22,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH connection failed")
}
