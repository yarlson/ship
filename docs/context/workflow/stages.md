# Workflow Stages

The workflow runs 5 stages after preflight succeeds.

## Stage 1: Tag

**Message:** `[1/5] Tagging image for transfer...` → `[1/5] Tag complete`

- input: original image ref
- action: create `localhost:5001/<image>` transfer tag
- implementation: `stage.Tag(original, transfer)`

## Stage 2: Registry

**Message:** `[2/5] Starting local registry...` → `[2/5] Registry ready`

- input: none
- action: reuse or start a local `registry:2` container bound to port `5001`
- implementation: `stage.Registry()`

## Stage 3: Push

**Message:** `[3/5] Pushing image to local registry...` → `[3/5] Push complete`

- input: transfer tag
- action: push the transfer tag into the local registry
- implementation: `stage.Push(transfer)`

## Stage 4: Tunnel

**Message:** `[4/5] Establishing tunnel to <host>...` → `[4/5] Tunnel established`

- input: SSH config
- action: open reverse SSH tunnel `5001:localhost:5001`
- implementation: `stage.Tunnel(cfg)`

## Stage 5: Pull And Restore

**Message:** `[5/5] Pulling and restoring image on remote host...` → `[5/5] Pull and restore complete`

- input: SSH config, original image ref, transfer tag
- action:
  - run remote `docker pull <transfer>`
  - run remote `docker tag <transfer> <original>`
- implementation: `stage.Pull(cfg, original, transfer)`

## Cleanup

The workflow owns tunnel cleanup. After Stage 4 succeeds, `workflow.Run()` defers `ssh.StopTunnel()` so the tunnel is closed on success and on failure.
