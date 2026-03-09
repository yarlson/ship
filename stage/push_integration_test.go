//go:build integration

package stage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/testctx"
	"ship/testlock"
)

func queryRegistryTags(t *testing.T, name string) []string {
	t.Helper()
	req, err := http.NewRequestWithContext(testctx.New(t), http.MethodGet, fmt.Sprintf("http://localhost:5001/v2/%s/tags/list", name), http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Tags []string `json:"tags"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result.Tags
}

func TestRegistryAndPush_HappyPath(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	originals := []string{"ship-inttest-push:latest", "ship-inttest-proxy:v3"}
	transfers := []string{"localhost:5001/ship-inttest-push:latest", "localhost:5001/ship-inttest-proxy:v3"}
	for _, original := range originals {
		ensureLocalImage(t, original)
	}
	require.NoError(t, Tag(testctx.New(t), originals, transfers))

	captureOutput(func() {
		err := Registry(testctx.New(t))
		require.NoError(t, err)
	})

	out := captureOutput(func() {
		err := Push(testctx.New(t), transfers)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[3/5] Pushing images to local registry...")
	assert.Contains(t, out, "[3/5] Push complete")

	tags := queryRegistryTags(t, "ship-inttest-push")
	assert.Contains(t, tags, "latest")
	tags = queryRegistryTags(t, "ship-inttest-proxy")
	assert.Contains(t, tags, "v3")
}

func TestPush_FailsOnBadImageRef(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	captureOutput(func() {
		require.NoError(t, Registry(testctx.New(t)))
	})

	err := Push(testctx.New(t), []string{"localhost:5001/nonexistent:latest"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "localhost:5001/nonexistent:latest")
}
