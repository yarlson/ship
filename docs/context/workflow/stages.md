# 7-Stage Deployment Pipeline

## Stage Sequence

The workflow runs these 7 stages in order, each with a start and completion message printed in `[N/7]` format:

### Stage 1: Build Images

**Message:** `[1/7] Building images...` → `[1/7] Build complete (N images)`

**Implementation:** `stage.Build(composeFiles)`

1. Run `docker compose build` with provided compose file(s)
2. Run `docker compose config --format json` to discover services with build keys
3. Parse compose config JSON to extract image names and tags
4. Build ImageMap: maps original image ref → `localhost:5001/` transfer tag
5. Return ImageMap to be used in Stage 2

**Error cases:**

- Docker compose build fails (command error)
- Compose config parsing fails (JSON error)
- No images found (services without build key)

### Stage 2: Tag Images

**Message:** `[2/7] Tagging images for transfer...` → `[2/7] Tag complete`

**Implementation:** `stage.Tag(imageMap)`

1. Receive ImageMap from Stage 1 (original ref → transfer tag mapping)
2. Iterate through ImageMap
3. Run `docker tag original transfer-tag` for each image
4. Return nil on success, error on first tag failure

**Error cases:**

- Docker tag command fails (image not found, permission denied, etc.)

**Invariant:** ImageMap keys (original image refs) must exist locally after Stage 1.

### Stage 3: Start Local Registry

**Message:** `[3/7] Starting local registry...` → `[3/7] Registry ready`

**Implementation:** `stage.Registry()`

1. Check if a `registry:2` container is already running on port 5001 via `docker ps`
2. If running, complete immediately
3. If not running, check for port conflicts (existing containers or non-Docker processes on :5001)
4. If port is free, start `registry:2` container with port mapping `5001:5000`
5. Wait for registry to accept TCP connections on `:5001` (up to 3 seconds)

**Error cases:**

- Docker command fails (`docker ps`, `docker run`)
- Port 5001 already in use (named what to check)
- Registry started but fails to accept connections (timeout)

### Stage 4: Push to Local Registry

**Message:** `[4/7] Pushing images to local registry...` → `[4/7] Push complete (N images)`

**Implementation:** `stage.Push(imageMap)`

1. Receive ImageMap from earlier stages
2. Iterate through transfer tags (values in ImageMap)
3. For each transfer tag, run `docker push` to registry on `:5001`
4. On first error, return immediately (fail fast)
5. Return completion message with image count

**Error cases:**

- Docker push fails for any image (invalid reference, registry unreachable, etc.)
- ImageMap is empty (would be caught earlier in workflow)

### Stage 5: Establish Tunnel

**Message:** `[5/7] Establishing tunnel to <host>...` → `[5/7] Tunnel established`

**Implementation:** `stage.Tunnel(cfg)`

1. Call `ssh.StartTunnel(keyPath, user, host)` to start reverse tunnel background process
2. Reverse tunnel forwards remote port 5001 to local 5001 (`ssh -R 5001:localhost:5001`)
3. Wait up to 2 seconds for tunnel to establish, checking if process exited early
4. Return TunnelProcess handle for lifecycle management and cleanup
5. Return error if tunnel fails to establish

**Error cases:**

- SSH tunnel connection fails (bad host, bad key, SSH server down)
- Process exits during setup (connection refused)

**Invariant:** Tunnel must stay alive for Stages 6-7. Workflow defers `cleanupTunnel()` to stop it on exit.

### Stage 6: Pull and Restore on Remote

**Message:** `[6/7] Pulling and restoring images on remote host` → `[6/7] Pull and restore complete (N images)`

**Implementation:** `stage.Pull(cfg, imageMap)`

1. Receive ImageMap from Stage 1 (original image ref → `localhost:5001/` transfer tag)
2. Iterate through ImageMap entries
3. For each image:
   - Execute `docker pull <transfer-tag>` on remote via SSH
   - Execute `docker tag <transfer-tag> <original>` on remote to restore original name
4. Count successful image restores
5. Return error on first failure (fail fast)

**Error cases:**

- Docker pull fails on remote (registry unreachable, image not found)
- Docker tag fails on remote (permission denied, etc.)
- SSH command execution fails

### Stage 7: Execute Remote Command

**Message:** `[7/7] Running remote command` → `[7/7] Command complete`

**Implementation:** `stage.Command(cfg)`

1. Execute user-provided deployment command on remote via SSH (from cfg.Command)
2. Capture command stdout and stderr
3. Pass through stdout to progress.Writer (unformatted)
4. Pass through stderr to os.Stderr
5. Return error if exit code is non-zero
6. User sees command output in real-time

**Error cases:**

- SSH command execution fails
- Remote command exits with non-zero code

## Implementation

**File:** `workflow/workflow.go`

**Current state:** All stages 1-7 have real implementations. Workflow manages ImageMap and TunnelProcess state across stages.

**Data flow:**

- `Run(cfg)` calls `stage.Build(cfg.ComposeFiles)` → returns ImageMap, stored in State
- `Run(cfg)` calls `stage.Tag(imageMap)` → tags all images locally
- `Run(cfg)` calls `stage.Registry()` → starts local registry if needed
- `Run(cfg)` calls `stage.Push(imageMap)` → pushes images to local registry
- `Run(cfg)` calls `stage.Tunnel(cfg)` → starts reverse tunnel, returns TunnelProcess
- `Run(cfg)` calls `stage.Pull(cfg, imageMap)` → pulls and restores images on remote via tunnel
- `Run(cfg)` calls `stage.Command(cfg)` → executes deployment command on remote via tunnel
- Deferred `cleanupTunnel()` stops TunnelProcess after all stages complete or on error
- Each stage calls `progress.StageStart()` then does work, then `progress.StageComplete()`

**Real implementation details:**

- See `docs/context/docker/image-handling.md` for Docker CLI operations
- See `docs/context/stage/implementation.md` for Stage 1-4 flow and error handling

## Error Handling

Each stage will return error or nil. Workflow stops on first failure and exits with error (fail-fast).
