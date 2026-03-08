package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseComposeConfig_ServicesWithBuild(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"web": {"build": {"context": "."}, "image": "web:latest"},
			"api": {"build": {"context": "./api"}, "image": "api:v2"}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	assert.Len(t, images, 2)

	byName := map[string]Image{}
	for _, img := range images {
		byName[img.Name] = img
	}

	assert.Equal(t, "latest", byName["web"].Tag)
	assert.Equal(t, "v2", byName["api"].Tag)
}

func TestParseComposeConfig_MixedServices(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"web": {"build": {"context": "."}, "image": "web:latest"},
			"api": {"build": {"context": "./api"}, "image": "api:latest"},
			"redis": {"image": "redis:alpine"}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	assert.Len(t, images, 2)

	names := map[string]bool{}
	for _, img := range images {
		names[img.Name] = true
	}
	assert.True(t, names["web"])
	assert.True(t, names["api"])
	assert.False(t, names["redis"])
}

func TestParseComposeConfig_NoBuildServices(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"redis": {"image": "redis:alpine"},
			"postgres": {"image": "postgres:16"}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestParseComposeConfig_NoTagImpliesLatest(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"web": {"build": {"context": "."}, "image": "myapp"}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, "myapp", images[0].Name)
	assert.Equal(t, "latest", images[0].Tag)
}

func TestParseComposeConfig_ExplicitTag(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"web": {"build": {"context": "."}, "image": "myapp:v2"}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, "myapp", images[0].Name)
	assert.Equal(t, "v2", images[0].Tag)
}

func TestParseComposeConfig_NestedBuildPath(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"worker": {"build": {"context": "./services/worker", "dockerfile": "Dockerfile.prod"}, "image": "worker:latest"}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, "worker", images[0].Name)
}

func TestParseComposeConfig_ServiceWithBuildButNoImage(t *testing.T) {
	input := `{
		"name": "myproject",
		"services": {
			"web": {"build": {"context": "."}}
		}
	}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	require.Len(t, images, 1)
	// Docker Compose default: <project>-<service>
	assert.Equal(t, "myproject-web", images[0].Name)
	assert.Equal(t, "latest", images[0].Tag)
}

func TestParseComposeConfig_EmptyJSON(t *testing.T) {
	input := `{"name": "myproject", "services": {}}`

	images, err := ParseComposeConfig([]byte(input))
	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestParseComposeConfig_MalformedJSON(t *testing.T) {
	input := `{not valid json`

	_, err := ParseComposeConfig([]byte(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse compose config")
}

func TestTransferTag(t *testing.T) {
	img := Image{Name: "app", Tag: "latest"}
	assert.Equal(t, "localhost:5001/app:latest", TransferTag(img))
}

func TestTransferTag_WithExistingRegistry(t *testing.T) {
	img := Image{Name: "ghcr.io/org/app", Tag: "v1"}
	assert.Equal(t, "localhost:5001/ghcr.io/org/app:v1", TransferTag(img))
}

func TestTransferTag_MultipleSlashes(t *testing.T) {
	img := Image{Name: "registry.example.com/org/app", Tag: "v1"}
	assert.Equal(t, "localhost:5001/registry.example.com/org/app:v1", TransferTag(img))
}

func TestParseRegistryContainerFilter(t *testing.T) {
	assert.True(t, ParseRegistryContainerFilter("abc123def456\n"))
}

func TestParseRegistryContainerFilter_Empty(t *testing.T) {
	assert.False(t, ParseRegistryContainerFilter(""))
	assert.False(t, ParseRegistryContainerFilter("  \n"))
}
