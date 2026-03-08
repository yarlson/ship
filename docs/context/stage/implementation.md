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

## Stage 5: Tunnel

**Function:** `Tunnel(cfg cli.Config) (*ssh.TunnelProcess, error)`

### Flow

1. Print `[5/7] Establishing tunnel to <host>...` via `progress.StageStart()`
2. Call `ssh.StartTunnel(keyPath, user, host)` — starts background SSH process
3. Reverse tunnel command: `ssh -R 5001:localhost:5001 user@host -N`
4. Wait up to 2 seconds for tunnel to establish (select on `tp.Exited()`)
5. If process exits during wait, tunnel failed
6. Print `[5/7] Tunnel established` via `progress.StageComplete()`
7. Return (TunnelProcess, nil) on success

### Error Handling

**Single failure point:**

- **Tunnel fails:** `SSH tunnel failed — connection refused (verify --host and --key)`
  - Cause: SSH connection refused, bad key, unreachable host
  - Error guides user to check host and key path
  - Process exits during 2-second wait window

### Contract

- **Input:** CLI config with SSH credentials (key, user, host)
- **Output:** TunnelProcess handle for cleanup, or error
- **Side effect:** Reverse SSH tunnel established and running in background
- **Data passing:** TunnelProcess stored in workflow.State for deferred cleanup

## Stage 6: Pull

**Function:** `Pull(cfg cli.Config, imageMap map[string]string) error`

### Flow

1. Print `[6/7] Pulling and restoring images on remote host` via `progress.StageStart()`
2. Receive ImageMap from Stage 1 (original → transfer tag)
3. Iterate through ImageMap:
   - For each image, execute `docker pull <transfer-tag>` on remote via SSH
   - Execute `docker tag <transfer-tag> <original>` to restore original name
4. Count successful image restores
5. On first error, return immediately (fail fast)
6. Print `[6/7] Pull and restore complete (N images)` via `progress.StageComplete()`
7. Return nil on success

### Error Handling

**Two failure points:**

1. **Pull fails:** `Failed to pull images on remote host — verify Docker is running on <host>`
   - Cause: docker pull command failed (registry unreachable, image not found, etc.)
   - Error names likely cause
   - Workflow stops immediately

2. **Tag fails:** `Failed to restore image tag on remote host — <original>`
   - Cause: docker tag command failed on remote
   - Error names which image failed
   - Workflow stops immediately

### Contract

- **Input:** CLI config (SSH credentials), ImageMap from Stage 1
- **Output:** nil on success, error on first failure
- **Side effect:** Images pulled on remote and re-tagged to original names
- **Data passing:** None to next stage

## Stage 7: Command

**Function:** `Command(cfg cli.Config) error`

### Flow

1. Print `[7/7] Running remote command` via `progress.StageStart()`
2. Execute user-provided command from cfg.Command on remote via SSH
3. Capture stdout and stderr from remote command
4. Pass through stdout to progress.Writer (unformatted, direct passthrough)
5. Pass through stderr to os.Stderr (direct passthrough)
6. Check exit code
7. Print `[7/7] Command complete` via `progress.StageComplete()`
8. Return error if exit code is non-zero

### Error Handling

**Single failure point:**

- **Command fails:** `Remote command exited with code <N> — see output above`
  - Cause: remote command returned non-zero exit code
  - Error names the exit code
  - Command output is already printed above
  - Workflow stops immediately

### Contract

- **Input:** CLI config (SSH credentials, command)
- **Output:** nil on success (exit 0), error on failure (non-zero exit)
- **Side effect:** Deployment command executes on remote, output visible to user
- **Data passing:** None (final stage)

### Output Rules

- **Passthrough, not reformatted** — Command output printed as-is, maintaining formatting
- **No suppression** — User sees all output from deployment command
- **Interleaved streams** — stdout to progress.Writer, stderr to os.Stderr

## Integration with Workflow

**File:** `workflow/workflow.go`

```go
func Run(cfg cli.Config) error {
    state := &State{}

    // Stage 1-4: Build, Tag, Registry, Push
    imageMap, err := stage.Build(cfg.ComposeFiles)
    if err != nil {
        return err
    }
    state.ImageMap = imageMap
    // ... tag, registry, push ...

    // Stage 5: Tunnel
    tp, err := stage.Tunnel(cfg)
    if err != nil {
        return err
    }
    state.TunnelCmd = tp
    defer cleanupTunnel(state)  // Ensure tunnel is stopped

    // Stage 6: Pull
    if err := stage.Pull(cfg, imageMap); err != nil {
        return err
    }

    // Stage 7: Command
    if err := stage.Command(cfg); err != nil {
        return err
    }

    printSummary(cfg, state)
    return nil
}
```

**Flow:**

1. Stage 1 builds and returns ImageMap
2. Stage 2 tags all images locally
3. Stage 3 ensures registry is running on :5001
4. Stage 4 pushes all images to registry
5. Stage 5 establishes tunnel, stores TunnelProcess in state
6. Stage 6 pulls and restores images on remote
7. Stage 7 executes deployment command on remote
8. Deferred cleanup stops tunnel (SIGTERM → wait 5s → SIGKILL)
9. On any error, cleanup still runs and returns error
10. On success, print summary with images and host

## Testing

All stages have integration tests in `*_integration_test.go` files:

- **stage/tunnel_integration_test.go** — Verifies tunnel establishment and bad-host error handling
- **stage/pull_integration_test.go** — Verifies image pull and tag restoration on remote
- **stage/command_integration_test.go** — Verifies command execution and output passthrough
- **ship_integration_test.go** — Full end-to-end workflow tests

Tests use real Docker, SSH, and remote host for comprehensive validation.
