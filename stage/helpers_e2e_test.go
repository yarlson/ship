//go:build e2e

package stage

import (
	"bytes"
	"os/exec"
	"testing"

	"ship/cli"
	"ship/docker"
	"ship/progress"
	"ship/testctx"
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
		Port:    22,
	}
}

// tagAndPushTestImage tags a source image with the given transfer tag and pushes it to the local registry.
func tagAndPushTestImage(t *testing.T, source, transferTag string) error {
	t.Helper()
	ctx := testctx.New(t)

	// Pull the source image if not present.
	pull := exec.CommandContext(ctx, "docker", "pull", source)
	if out, err := pull.CombinedOutput(); err != nil {
		t.Logf("pull output: %s", string(out))
		return err
	}

	if err := docker.TagImage(ctx, source, transferTag); err != nil {
		return err
	}

	return docker.PushImage(ctx, transferTag)
}
