package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransferTag(t *testing.T) {
	assert.Equal(t, "localhost:5001/app:latest", TransferTag("app:latest"))
}

func TestTransferTag_WithExistingRegistry(t *testing.T) {
	assert.Equal(t, "localhost:5001/ghcr.io/org/app:v1", TransferTag("ghcr.io/org/app:v1"))
}

func TestTransferTag_MultipleSlashes(t *testing.T) {
	assert.Equal(t, "localhost:5001/registry.example.com/org/app:v1", TransferTag("registry.example.com/org/app:v1"))
}

func TestParseRegistryContainerFilter(t *testing.T) {
	assert.True(t, ParseRegistryContainerFilter("abc123def456\n"))
}

func TestParseRegistryContainerFilter_Empty(t *testing.T) {
	assert.False(t, ParseRegistryContainerFilter(""))
	assert.False(t, ParseRegistryContainerFilter("  \n"))
}
