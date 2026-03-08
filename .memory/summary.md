# Ship CLI

## What

Go CLI tool that builds Docker Compose images locally, transfers them to a remote host via SSH tunnel and local registry, and runs a deployment command — all in one invocation.

## Architecture

7-stage sequential pipeline orchestrated by `workflow.Run`:

1. **Build** — `docker compose build` via CLI
2. **Tag** — create `localhost:5001/<name>:<tag>` transfer tags
3. **Registry** — start local Docker registry on port 5001
4. **Push** — push transfer-tagged images to local registry
5. **Tunnel** — reverse SSH tunnel (`-R 5001:localhost:5001`) so remote can reach local registry
6. **Pull** — remote `docker pull` from tunnel, then `docker tag` to restore original names
7. **Command** — run user-specified deploy command on remote host

## Core Flow

```
CLI flags → cli.Parse → workflow.Run(cfg)
  → stage.Build → stage.Tag → stage.Registry → stage.Push
  → stage.Tunnel → stage.Pull → stage.Command → printSummary
```

## Package Structure

- `cli/` — flag parsing, Config struct
- `docker/` — compose config parsing, image tag/push/pull (shells out to `docker`)
- `ssh/` — remote command execution, tunnel lifecycle management
- `stage/` — one function per pipeline stage
- `workflow/` — orchestration, state management, cleanup, summary output
- `progress/` — `Writer` (swappable `io.Writer`) for testable progress output
- `testlock/` — port 5001 mutex and registry cleanup for parallel integration tests

## Tech Stack

- Go (standard library + `os/exec` for Docker/SSH)
- Docker Compose (build), Docker Registry (transfer)
- SSH CLI (tunnel, remote commands)
- testify (assert/require) for testing
