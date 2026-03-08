//go:build integration

package stage

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

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

func requireDocker(t *testing.T) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "docker", "version")
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping integration test: Docker daemon unavailable: %v", err)
	}
}

func ensureLocalImage(t *testing.T, imageRef string) {
	t.Helper()
	requireDocker(t)

	pull := exec.CommandContext(context.Background(), "docker", "pull", "alpine:latest")
	if out, err := pull.CombinedOutput(); err != nil {
		t.Fatalf("failed to pull alpine: %v\n%s", err, string(out))
	}

	require.NoError(t, docker.TagImage("alpine:latest", imageRef))
}
