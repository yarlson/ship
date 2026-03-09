//go:build integration

package stage

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"ship/docker"
	"ship/progress"
	"ship/testctx"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

func requireDocker(t *testing.T) {
	t.Helper()
	cmd := exec.CommandContext(testctx.New(t), "docker", "version")
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping integration test: Docker daemon unavailable: %v", err)
	}
}

func ensureLocalImage(t *testing.T, imageRef string) {
	t.Helper()
	requireDocker(t)
	ctx := testctx.New(t)

	pull := exec.CommandContext(ctx, "docker", "pull", "alpine:latest")
	if out, err := pull.CombinedOutput(); err != nil {
		t.Fatalf("failed to pull alpine: %v\n%s", err, string(out))
	}

	require.NoError(t, docker.TagImage(ctx, "alpine:latest", imageRef))
}
