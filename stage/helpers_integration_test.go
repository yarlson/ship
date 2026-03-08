//go:build integration

package stage

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/docker"
)

func testSSHConfig(t *testing.T) cli.Config {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	return cli.Config{
		User:    "root",
		Host:    "46.101.213.82",
		KeyPath: home + "/.ssh/id_rsa",
		Command: "echo test",
	}
}

// tagAndPushTestImage tags a source image with the given transfer tag and pushes it to the local registry.
func tagAndPushTestImage(t *testing.T, source, transferTag string) error {
	t.Helper()

	// Pull the source image if not present.
	pull := exec.CommandContext(context.Background(), "docker", "pull", source)
	if out, err := pull.CombinedOutput(); err != nil {
		t.Logf("pull output: %s", string(out))
		return err
	}

	if err := docker.TagImage(source, transferTag); err != nil {
		return err
	}

	return docker.PushImage(transferTag)
}
