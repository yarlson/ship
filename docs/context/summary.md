# Ship Summary

## What It Is

Ship is a Go CLI that transfers one or more existing local Docker images to a remote host over SSH.

It does not build images.
It does not read Docker Compose files.
It does not run a deploy hook on the remote host.

Its only job is to make sure images that exist locally end up on the remote host under the same original tags.

## CLI Shape

```bash
ship [-i key] [-p port] user@host image[:tag] [image[:tag]...]
```

- `user@host` identifies the SSH target
- `image[:tag]` arguments identify the local Docker images to transfer
- `-i` optionally selects an SSH identity file
- `-p` optionally selects a non-default SSH port

## Transfer Flow

Ship runs these stages in order:

1. tag each local image as `localhost:5001/<image>`
2. ensure a local registry is running on `:5001`
3. push the transfer tags to that local registry
4. open an SSH reverse tunnel so the remote host can reach the local registry
5. run remote `docker pull` and `docker tag` commands to restore the original image tags

## Preflight

Before stage 1 starts, Ship checks:

- Docker is available locally
- `ssh` is available locally
- the SSH key exists if `-i` was provided
- each local image exists
- SSH connectivity to the target works

## Main Modules

- `cli` parses the SSH-style command line
- `workflow` runs preflight and the 5 stages
- `stage` contains the stage implementations
- `docker` wraps local Docker CLI operations
- `ssh` wraps SSH remote commands and tunnel lifecycle
- `progress` prints stage output in `[N/5]` format

## Capabilities

- transfer one or more local images to one remote host
- preserve the original image tag on the remote side
- reuse SSH instead of requiring a hosted or remote registry
- fail fast when local prerequisites or SSH access are missing
- print deterministic stage progress and a short success summary
