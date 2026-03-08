# Stage Implementation

## Stage Signatures

- `Tag(original, transfer string) error`
- `Registry() error`
- `Push(transfer string) error`
- `Tunnel(cfg cli.Config) (*ssh.TunnelProcess, error)`
- `Pull(cfg cli.Config, original, transfer string) error`

## Common Pattern

Every stage follows the same structure:

1. call `progress.StageStart()`
2. perform the actual Docker or SSH work
3. return an error immediately on failure
4. call `progress.StageComplete()` on success

## Stage Responsibilities

### Tag

- use `docker.TagImage(original, transfer)`
- fail if the original image ref does not exist locally

### Registry

- detect whether `registry:2` is already serving port `5001`
- detect port conflicts before trying to start a registry
- start `registry:2` only when needed

### Push

- push exactly one transfer tag to the local registry

### Tunnel

- create one reverse SSH tunnel to the target host
- return a `TunnelProcess` handle to the workflow for cleanup

### Pull

- run remote `docker pull` for the transfer tag
- run remote `docker tag` to restore the original image ref
- do not execute arbitrary user-supplied shell commands
