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

	//nolint:gosec // Integration test only; URL host is fixed to the local registry.
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

	original := "ship-inttest-push:latest"
	transfer := "localhost:5001/ship-inttest-push:latest"
	ensureLocalImage(t, original)
	require.NoError(t, Tag(original, transfer))

	captureOutput(func() {
		err := Registry()
		require.NoError(t, err)
	})

	out := captureOutput(func() {
		err := Push(transfer)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "[3/5] Pushing image to local registry...")
	assert.Contains(t, out, "[3/5] Push complete")

	tags := queryRegistryTags(t, "ship-inttest-push")
	assert.Contains(t, tags, "latest")
}

func TestPush_FailsOnBadImageRef(t *testing.T) {
	requireDocker(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })

	captureOutput(func() {
		require.NoError(t, Registry())
	})

	err := Push("localhost:5001/nonexistent:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "localhost:5001/nonexistent:latest")
}
