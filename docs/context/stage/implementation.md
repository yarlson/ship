# Stage Implementation

## Stage Signatures

- `Tag(ctx context.Context, originals, transfers []string) error`
- `Registry(ctx context.Context) error`
- `Push(ctx context.Context, transfers []string) error`
- `Tunnel(ctx context.Context, cfg cli.Config) (*ssh.TunnelProcess, error)`
- `Pull(ctx context.Context, cfg cli.Config, originals, transfers []string) error`

## Common Pattern

Every stage follows the same structure:

1. call `progress.StageStart()`
2. perform the actual Docker or SSH work using the caller-provided `ctx`
3. return an error immediately on failure
4. call `progress.StageComplete()` on success

## Stage Responsibilities

### Tag

- use `docker.TagImage(ctx, original, transfer)` for each image
- fail if any original image ref does not exist locally

### Registry

- detect whether `registry:2` is already serving port `5001`
- detect port conflicts before trying to start a registry
- start `registry:2` only when needed

### Push

- push each transfer tag to the local registry

### Tunnel

- create one reverse SSH tunnel to the target host
- return a `TunnelProcess` handle to the workflow for cleanup

### Pull

- run remote `docker pull` for each transfer tag
- run remote `docker tag` to restore each original image ref
- keep remote execution limited to the transfer steps the workflow owns
