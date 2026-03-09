# Workflow Stages

The workflow runs 5 stages after preflight succeeds.

## Stage 1: Tag

**Message:** `[1/5] Tagging image for transfer...` or `[1/5] Tagging images for transfer...` → `[1/5] Tag complete`

- input: original image refs
- action: create `localhost:5001/<image>` transfer tags
- implementation: `stage.Tag(ctx, originals, transfers)`

## Stage 2: Registry

**Message:** `[2/5] Starting local registry...` → `[2/5] Registry ready`

- input: none
- action: reuse or start a local `registry:2` container bound to port `5001`
- implementation: `stage.Registry(ctx)`

## Stage 3: Push

**Message:** `[3/5] Pushing image to local registry...` or `[3/5] Pushing images to local registry...` → `[3/5] Push complete`

- input: transfer tags
- action: push the transfer tags into the local registry
- implementation: `stage.Push(ctx, transfers)`

## Stage 4: Tunnel

**Message:** `[4/5] Establishing tunnel to <host>...` → `[4/5] Tunnel established`

- input: SSH config
- action: open reverse SSH tunnel `5001:localhost:5001`
- implementation: `stage.Tunnel(ctx, cfg)`

## Stage 5: Pull And Restore

**Message:** `[5/5] Pulling and restoring image on remote host...` or `[5/5] Pulling and restoring images on remote host...` → `[5/5] Pull and restore complete`

- input: SSH config, original image refs, transfer tags
- action:
  - run remote `docker pull <transfer>` for each image
  - run remote `docker tag <transfer> <original>` for each image
- implementation: `stage.Pull(ctx, cfg, originals, transfers)`

## Cleanup

The workflow owns tunnel cleanup. After Stage 4 succeeds, `workflow.Run(ctx, cfg)` defers `ssh.StopTunnel(cleanupCtx, tp)` so the tunnel is closed on success and on failure, even if the main workflow context was canceled.
