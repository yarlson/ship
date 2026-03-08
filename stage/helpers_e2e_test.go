//go:build e2e

package stage

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/docker"
	"ship/progress"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

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

func testSSHConfig(t *testing.T) cli.Config {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	keyPath := home + "/.ssh/id_rsa"
	requireE2EPrereqs(t, keyPath, "root", "46.101.213.82")
	return cli.Config{
		User:    "root",
		Host:    "46.101.213.82",
		KeyPath: keyPath,
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
