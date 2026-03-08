# Stage Implementations (1-2)

## Overview

Stages 1 and 2 have real implementations. Stages 3-7 are stubs that print hardcoded progress messages.

**File:** `stage/build.go`, `stage/tag.go`

## Stage 1: Build

**Function:** `Build(composeFiles string) (map[string]string, error)`

### Flow

1. Print `[1/7] Building images...` via `progress.StageStart()`
2. Parse comma-separated compose file paths
3. Call `docker.ComposeBuild(files)` — runs `docker compose build`
4. Call `docker.ComposeConfig(files)` — parses config JSON to get images
5. Build ImageMap: map original image ref → transfer tag
6. Print `[1/7] Build complete (N images)` via `progress.StageComplete()`
7. Return (imageMap, nil) on success

### ImageMap Construction

```go
imageMap := make(map[string]string, len(images))
for _, img := range images {
    imageMap[img.Ref()] = docker.TransferTag(img)
}
// Examples:
// "web:latest" → "localhost:5001/web:latest"
// "api:v2" → "localhost:5001/api:v2"
```

### Error Handling

Three failure points:

1. **Build fails:** `Build failed — see docker compose output above`
   - Cause: docker compose build command failed
   - User sees full docker output in stdout/stderr
   - Message directs user to check output

2. **Config parse fails:** `Build failed — <error detail>`
   - Cause: docker compose config parsing failed
   - Error wrapped with context

3. **No images found:** `No images found after build — check that services in the compose file have a 'build' key`
   - Cause: All services either lack build key or image specification
   - Error names the likely cause

### Contract

- **Input:** Comma-separated paths to compose files
- **Output:** ImageMap (original → transfer tag) or error
- **Invariant:** All returned image refs (keys) exist locally
- **Side effect:** Docker images built and cached locally

## Stage 2: Tag

**Function:** `Tag(imageMap map[string]string) error`

### Flow

1. Print `[2/7] Tagging images for transfer...` via `progress.StageStart()`
2. Iterate through imageMap entries
3. For each entry, call `docker.TagImage(original, transfer)` — runs `docker tag`
4. On first error, return immediately (fail fast)
5. Print `[2/7] Tag complete` via `progress.StageComplete()`
6. Return nil on success

### Error Handling

**Single failure point:**

- **Tag fails:** `Failed to tag <original> — <error detail>`
  - Cause: docker tag command failed (image not found, permission denied, etc.)
  - Error names which image failed and why
  - Workflow stops immediately

**Invariant:** All images in imageMap keys must exist locally (guaranteed by Stage 1).

### Contract

- **Input:** ImageMap from Stage 1 (original → transfer tag)
- **Output:** nil on success, error on first failure
- **Invariant:** All keys (original image refs) exist locally
- **Side effect:** New tags created for all images (or partial if error occurs)

## Integration with Workflow

**File:** `workflow/workflow.go`

```go
func Run(cfg cli.Config) error {
    // Stage 1: Build
    imageMap, err := stage.Build(cfg.ComposeFiles)
    if err != nil {
        return err
    }

    // Stage 2: Tag
    if err := stage.Tag(imageMap); err != nil {
        return err
    }

    // Stages 3-7: Stubs
    for i, s := range stubs {
        n := i + 3
        progress.StageStart(n, s.startMsg)
        progress.StageComplete(n, s.completeMsg)
    }

    return nil
}
```

**Flow:**
1. Stage 1 builds and returns ImageMap
2. Stage 2 receives ImageMap, tags all images
3. Stages 3-7 execute with stub implementations
4. On any error, return immediately (fail fast)

## Testing

Both stages have integration tests in `*_integration_test.go` files:

- **TestRun_PrintsAllSevenStages** — Verifies all 7 stages produce `[N/7]` output
- **TestRun_StagesInOrder** — Verifies stage numbers appear in order

Tests use real Docker Compose project with temporary files.
