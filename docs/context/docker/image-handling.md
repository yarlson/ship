# Docker Image Handling

## Current Docker Helpers

### `TransferTag(imageRef string) string`

Converts an original image ref into the temporary transfer tag:

```text
app:latest -> localhost:5001/app:latest
```

### `TransferTags(imageRefs []string) []string`

Converts multiple original image refs into transfer tags in the same order.

### `ImageExists(ctx context.Context, imageRef string) error`

Runs:

```bash
docker image inspect <imageRef>
```

Used during preflight to verify each requested local image exists before any stage runs.

### `TagImage(ctx context.Context, source, target string) error`

Runs:

```bash
docker tag <source> <target>
```

Used by Stage 1 to create the local-registry transfer tag.

## Registry Helpers

`docker/registry.go` owns:

- checking whether a registry is already running on `5001`
- detecting port conflicts
- starting `registry:2`
- pushing an image to the registry

These helpers take a caller-owned `context.Context` because they shell out to Docker or wait on local sockets.

## Deliberate Non-Goals

This module does not decide what images should exist or where they come from.
It only works with image references that were already provided to the transfer workflow.
