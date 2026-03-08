//go:build e2e

package stage

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"ship/cli"
	"ship/docker"
	"ship/progress"
	"ship/testenv"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

func testSSHConfig(t *testing.T) cli.Config {
	t.Helper()
	cfg := testenv.RequireE2EConfig(t)
	return cli.Config{
		User:    cfg.User,
		Host:    cfg.Host,
		KeyPath: cfg.KeyPath,
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
