# 7-Stage Deployment Pipeline

## Stage Sequence

The workflow runs these 7 stages in order, each with a start and completion message printed in `[N/7]` format:

### Stage 1: Build Images
**Message:** `[1/7] Building images...` → `[1/7] Build complete`

Build Docker Compose images defined in provided compose file(s). Discovers which images were built and stores metadata for downstream stages.

### Stage 2: Tag Images
**Message:** `[2/7] Tagging images for transfer...` → `[2/7] Tag complete`

Re-tag images from original names to local registry format (`localhost:5001/*`). Build ImageMap for tracking original ↔ tagged names for restoration on remote.

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

**Current state:** Stub implementation with hardcoded stage messages. Prints all 7 stages with start/completion pairs.

**Data flow:**
- `Run(cfg)` iterates through stages array
- Each stage calls `progress.StageStart()` then `progress.StageComplete()`
- Future: Replace with actual stage implementations that mutate WorkflowState

## Error Handling

Each stage will return error or nil. Workflow stops on first failure and exits with error (fail-fast).
