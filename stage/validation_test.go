package stage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/testctx"
)

func TestTag_RejectsMismatchedImageLists(t *testing.T) {
	err := Tag(testctx.New(t), []string{"app:latest"}, nil)

	require.Error(t, err)
	assert.EqualError(t, err, "image list mismatch: 1 originals, 0 transfers")
}

func TestPull_RejectsMismatchedImageLists(t *testing.T) {
	err := Pull(testctx.New(t), cli.Config{}, []string{"app:latest"}, nil)

	require.Error(t, err)
	assert.EqualError(t, err, "image list mismatch: 1 originals, 0 transfers")
}
