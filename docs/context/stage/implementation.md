# Stage Implementations (1-4)

## Overview

Stages 1-4 have real implementations. Stages 5-7 are stubs that print hardcoded progress messages.

**Files:** `stage/build.go`, `stage/tag.go`, `stage/registry.go`, `stage/push.go`

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

## Stage 3: Registry

**Function:** `Registry() error`

### Flow

1. Print `[3/7] Starting local registry...` via `progress.StageStart()`
2. Call `docker.CheckRegistryRunning()` — check if `registry:2` on :5001 exists
3. If already running, skip to completion
4. If not running, call `docker.CheckPortConflict()` — detect if :5001 is occupied
5. If port conflict detected, return error with instructions
6. Call `docker.StartRegistry()` — start `registry:2` container with port mapping
7. Wait for registry to accept TCP connections (up to 3 seconds with polling)
8. Print `[3/7] Registry ready` via `progress.StageComplete()`

### Error Handling

Three failure points:

1. **Registry check fails:** `Failed to check registry status — <error>`
   - Cause: docker ps command failed
   - Error wrapped with context

2. **Port conflict detected:** `Port 5001 already in use — stop the existing process or free the port`
   - Cause: Another container or process occupies :5001
   - Message names the problem and action

3. **Registry fails to start:** `Failed to start registry — <error>` or `registry started but not accepting connections on port 5001`
   - Cause: docker run failed or registry not ready after 3 seconds
   - Error identifies the specific issue

### Contract

- **Input:** None (reads system state)
- **Output:** nil on success, error on failure
- **Side effect:** Docker registry running on :5001 (or already was)
- **Data passing:** None to next stage (registry state is implicit)

## Stage 4: Push

**Function:** `Push(imageMap map[string]string) error`

### Flow

1. Print `[4/7] Pushing images to local registry...` via `progress.StageStart()`
2. Iterate through ImageMap values (transfer tags only, keys ignored)
3. For each transfer tag, call `docker.PushImage(transfer)` — runs `docker push`
4. Count successful pushes
5. On first error, return immediately (fail fast)
6. Print `[4/7] Push complete (N images)` via `progress.StageComplete()`
7. Return nil on success

### Error Handling

**Single failure point:**

- **Push fails:** `Failed to push <image> — <error>`
  - Cause: docker push command failed (registry unreachable, invalid ref, etc.)
  - Error names which image failed and why
  - Workflow stops immediately

### Invariant

- ImageMap must be non-empty (guaranteed by Stage 1)
- All transfer tags must be valid Docker image references
- Registry must be running (guaranteed by Stage 3)

### Contract

- **Input:** ImageMap from Stage 1 (original → transfer tag mapping)
- **Output:** nil on success, error on first failure
- **Side effect:** Images pushed to local registry on :5001
- **Data passing:** None to next stage

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

    // Stage 3: Registry
    if err := stage.Registry(); err != nil {
        return err
    }

    // Stage 4: Push
    if err := stage.Push(imageMap); err != nil {
        return err
    }

    // Stages 5-7: Stubs
    for i, s := range stubs {
        n := i + 5
        progress.StageStart(n, s.startMsg)
        progress.StageComplete(n, s.completeMsg)
    }

    return nil
}
```

**Flow:**

1. Stage 1 builds and returns ImageMap
2. Stage 2 receives ImageMap, tags all images
3. Stage 3 ensures registry is running on :5001
4. Stage 4 pushes all transfer-tagged images to registry
5. Stages 5-7 execute with stub implementations
6. On any error, return immediately (fail fast)

## Testing

Stages 1-4 have integration tests in `*_integration_test.go` files:

- **TestRun_PrintsAllSevenStages** — Verifies all 7 stages produce `[N/7]` output
- **TestRun_StagesInOrder** — Verifies stage numbers appear in order
- **TestRegistry_ChecksAndStartsRegistry** — Verifies registry startup and port conflict detection
- **TestPush_PushesAllImages** — Verifies all transfer-tagged images are pushed

Tests use real Docker Compose projects and Docker CLI operations.
