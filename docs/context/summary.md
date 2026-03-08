# Ship — Current State Summary

## What

Ship is a Go CLI tool that orchestrates a 7-stage deployment pipeline:

1. Build Docker Compose images locally
2. Tag images with local registry prefix
3. Start local registry container
4. Push images to local registry
5. Establish reverse SSH tunnel
6. Pull and restore images on remote host
7. Execute deployment command on remote host

All stages run in a single CLI invocation with progress reporting.

## Architecture

Ship follows a modular stage-based design:

```
main → cli.Parse() → workflow.Run()
                ↓
          [7 stages in sequence]
                ↓
          progress reporting
```

**Core modules:**

- `cli/` — flag parsing and validation
- `workflow/` — stage orchestration and pipeline sequencing
- `progress/` — progress output formatting
- `stage/` — stage implementations (Build, Tag, Registry, Push, Tunnel, Pull, Command)
- `docker/` — Docker CLI utilities (compose build, config parsing, image tagging, registry operations)
- `ssh/` — SSH tunnel and remote command execution (reverse tunnel, command output, process lifecycle)

**Data flow:** CLI flags → Config → WorkflowState (shared across stages) → result

## Core Flow

1. **Entry** — `main.go` calls `cli.Parse(args)` to extract Config
2. **Validation** — Config has required flags or returns error
3. **Preflight checks** — `workflow.Preflight(cfg)` validates Docker, Docker Compose V2, SSH, key file, and connectivity before any stages run
4. **Orchestration** — `workflow.Run(cfg)` executes preflight checks then 7 stages sequentially
5. **Progress** — Each stage prints `[N/7] message` lines
6. **Exit** — Workflow returns nil on success or error on failure

## System State

**Preflight validation:**

Before any stages run, `Preflight(cfg)` checks:

- Docker is installed and accessible
- Docker Compose V2 plugin is available
- SSH client is installed
- SSH key file exists and is readable
- SSH connectivity to remote host works

If any check fails, the workflow exits immediately with a descriptive error message and hint.

**Implemented stages:**

- Stage 1 (Build) — runs `docker compose build`, discovers built images via `docker compose config`
- Stage 2 (Tag) — re-tags images with `localhost:5001/` prefix using ImageMap pattern
- Stage 3 (Registry) — checks registry status, detects port conflicts, starts registry:2 container on :5001
- Stage 4 (Push) — pushes transfer-tagged images to local registry on :5001
- Stage 5 (Tunnel) — establishes reverse SSH tunnel via `ssh -R 5001:localhost:5001`, allows remote to access local registry
- Stage 6 (Pull) — executes `docker pull` on remote via SSH tunnel, restores original image tags
- Stage 7 (Command) — executes user-provided deployment command on remote, passes through output

**Data flow across stages:**

- Stage 1 returns ImageMap (original name → transfer tag mapping)
- Stage 2 receives ImageMap and tags all images
- Stage 3 checks/starts registry (independent operation)
- Stage 4 receives ImageMap and pushes images
- Stage 5 returns TunnelProcess handle (for cleanup)
- Stage 6 receives ImageMap and pulls/restores images via tunnel
- Stage 7 executes command via SSH, uses tunnel established in Stage 5

**Module boundaries enforced:**

- `cli` — only flag parsing, no I/O
- `docker` — Docker CLI operations (compose build, config parsing, tag, registry check/start, push)
- `ssh` — SSH CLI operations (remote command execution, reverse tunnel lifecycle)
- `stage` — stage implementations (Build, Tag, Registry, Push, Tunnel, Pull, Command) with progress integration
- `workflow` — pipeline sequencing, stage invocation, tunnel lifecycle cleanup
- `progress` — output formatting, testable via Writer var

## Capabilities

✓ Parse and validate all required CLI flags
✓ Support single or multiple Docker Compose files (comma-separated `--docker-compose`)
✓ Display help text with usage examples
✓ Run preflight checks before pipeline: Docker, Docker Compose V2, SSH, key file, compose files, connectivity
✓ Fail fast with clear error messages naming what failed and what to check
✓ Print stage progress in `[N/7]` format
✓ Run 7-stage workflow in sequence with cleanup
✓ Build Docker Compose images from multiple files and discover built images (Stage 1)
✓ Tag images with local registry prefix (Stage 2)
✓ Check registry status, detect port conflicts, start registry on :5001 (Stage 3)
✓ Push transfer-tagged images to local registry (Stage 4)
✓ Establish reverse SSH tunnel to remote host (Stage 5)
✓ Pull images on remote and restore original tags via tunnel (Stage 6)
✓ Execute deployment command on remote with output passthrough (Stage 7)
✓ Testable stage functions with mocked output
✓ Print success summary with deployed images and target host

## Tech Stack

- **Language:** Go 1.22+
- **CLI parsing:** `flag` stdlib (no cobra/urfave)
- **Testing:** `testify` (require/assert)
- **Code quality:** `golangci-lint` (gci, gofmt, gofumpt, goimports)
- **Deployment target:** Remote host via SSH with local registry tunnel
- **Build system:** `go build`, `docker`, `ssh` (via CLI, not SDKs)
