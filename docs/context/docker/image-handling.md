# Docker Image Handling

## Current Docker Helpers

### `TransferTag(imageRef string) string`

Converts an original image ref into the temporary transfer tag:

```text
app:latest -> localhost:5001/app:latest
```

### `ImageExists(imageRef string) error`

Runs:

```bash
docker image inspect <imageRef>
```

Used during preflight to verify the local image exists before any stage runs.

### `TagImage(source, target string) error`

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

## Deliberate Non-Goals

This module does not:

- parse Docker Compose config
- build Docker images
- discover multiple service images
