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

Ensure local registry is running on `localhost:5001`. Start if not present.

### Stage 4: Push to Local Registry
**Message:** `[4/7] Pushing images to local registry...` → `[4/7] Push complete`

Push tagged images to local registry on `:5001`.

### Stage 5: Establish Tunnel
**Message:** `[5/7] Establishing tunnel...` → `[5/7] Tunnel established`

Start reverse SSH tunnel (background process) from remote host to local registry. Allows remote host to access `localhost:5001`.

### Stage 6: Pull and Restore on Remote
**Message:** `[6/7] Pulling and restoring images on remote host...` → `[6/7] Pull and restore complete`

Remote pulls images from tunnel-accessible registry and re-tags them back to original names.

### Stage 7: Execute Remote Command
**Message:** `[7/7] Running remote command...` → `[7/7] Command complete`

Execute user's deployment command on remote host (e.g., `docker compose up -d`).

## Implementation

**File:** `workflow/workflow.go`

**Current state:** Stages 1-2 are real implementations. Stages 3-7 are stubs with hardcoded progress messages.

**Data flow:**
- `Run(cfg)` calls `stage.Build(cfg.ComposeFiles)` → returns ImageMap
- `Run(cfg)` calls `stage.Tag(imageMap)` → tags all images
- `Run(cfg)` iterates through stub stages 3-7, each prints start/completion
- Each stage calls `progress.StageStart()` then does work, then `progress.StageComplete()`

**Real implementation details:**
- See `docs/context/docker/image-handling.md` for Docker CLI operations
- See `docs/context/stage/implementation.md` for Stage 1-2 flow and error handling

## Error Handling

Each stage will return error or nil. Workflow stops on first failure and exits with error (fail-fast).
