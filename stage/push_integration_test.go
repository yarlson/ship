//go:build integration

package stage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/testlock"
)

func queryRegistryTags(t *testing.T, name string) []string {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("http://localhost:5001/v2/%s/tags/list", name), http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // test-only, URL is hardcoded localhost
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Tags []string `json:"tags"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result.Tags
}

func TestRegistryAndPush_HappyPath_E2E(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	composePath := setupComposeProject(t)

	// Stage 1: Build.
	var imageMap map[string]string
	captureOutput(func() {
		var err error
		imageMap, err = Build([]string{composePath})
		require.NoError(t, err)
	})

	// Stage 2: Tag.
	captureOutput(func() {
		err := Tag(imageMap)
		require.NoError(t, err)
	})

	// Stage 3: Registry.
	captureOutput(func() {
		err := Registry()
		require.NoError(t, err)
	})

	assert.True(t, registryRunning(t), "registry should be running")

	// Stage 4: Push.
	out := captureOutput(func() {
		err := Push(imageMap)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[4/7] Pushing images to local registry...")
	assert.Contains(t, out, "[4/7] Push complete (2 images)")

	// Verify images are queryable via registry HTTP API.
	for _, name := range []string{"ship-inttest-web", "ship-inttest-api"} {
		tags := queryRegistryTags(t, name)
		assert.Contains(t, tags, "latest", "image %s should have latest tag in registry", name)
	}
}

func TestPush_PushesImagesToRegistry(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	composePath := setupComposeProject(t)

	// Build and tag.
	var imageMap map[string]string
	captureOutput(func() {
		var err error
		imageMap, err = Build([]string{composePath})
		require.NoError(t, err)
	})
	captureOutput(func() {
		require.NoError(t, Tag(imageMap))
	})

	// Ensure registry running.
	captureOutput(func() {
		require.NoError(t, Registry())
	})

	// Push.
	captureOutput(func() {
		err := Push(imageMap)
		require.NoError(t, err)
	})

	// Verify via registry HTTP API.
	for _, name := range []string{"ship-inttest-web", "ship-inttest-api"} {
		tags := queryRegistryTags(t, name)
		assert.NotEmpty(t, tags, "image %s should have tags in registry", name)
	}
}

func TestPush_ReportsImageCount(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	composePath := setupComposeProject(t)

	var imageMap map[string]string
	captureOutput(func() {
		var err error
		imageMap, err = Build([]string{composePath})
		require.NoError(t, err)
	})
	captureOutput(func() {
		require.NoError(t, Tag(imageMap))
	})

	captureOutput(func() {
		require.NoError(t, Registry())
	})

	out := captureOutput(func() {
		err := Push(imageMap)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Push complete (2 images)")
}

func TestPush_EmptyImageMap(t *testing.T) {
	out := captureOutput(func() {
		err := Push(map[string]string{})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[4/7] Pushing images to local registry...")
	assert.Contains(t, out, "[4/7] Push complete (0 images)")
}

func TestPush_FailsOnBadImageRef(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	// Ensure registry is running.
	captureOutput(func() {
		require.NoError(t, Registry())
	})

	imageMap := map[string]string{
		"nonexistent:latest": "localhost:5001/nonexistent:latest",
	}

	captureOutput(func() {
		err := Push(imageMap)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "localhost:5001/nonexistent:latest")
	})
}
