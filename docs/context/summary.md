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
- `stage/` — stage implementations (Build, Tag, Registry, Push)
- `docker/` — Docker CLI utilities (compose build, config parsing, image tagging, registry operations)
- Future: `ssh/` for tunnel and remote operations

**Data flow:** CLI flags → Config → WorkflowState (shared across stages) → result

## Core Flow

1. **Entry** — `main.go` calls `cli.Parse(args)` to extract Config
2. **Validation** — Config has required flags or returns error
3. **Orchestration** — `workflow.Run(cfg)` executes 7 stages sequentially
4. **Progress** — Each stage prints `[N/7] message` lines
5. **Exit** — Workflow returns nil on success or error on failure

## System State

**Implemented stages:**

- Stage 1 (Build) — Real: runs `docker compose build`, discovers built images via `docker compose config`
- Stage 2 (Tag) — Real: re-tags images with `localhost:5001/` prefix using ImageMap pattern
- Stage 3 (Registry) — Real: checks registry status, detects port conflicts, starts registry:2 container on :5001
- Stage 4 (Push) — Real: pushes transfer-tagged images to local registry on :5001
- Stages 5-7 — Stub implementations with hardcoded progress messages

**Data flow across stages:**

- Stage 1 returns ImageMap (original name → transfer tag mapping)
- Stage 2 receives ImageMap and tags all images
- Stage 3 checks/starts registry (independent operation)
- Stage 4 receives ImageMap and pushes images
- Stages 5-7 stubs execute independently (stubs don't use image data)

**Module boundaries enforced:**

- `cli` — only flag parsing, no I/O
- `docker` — Docker CLI operations (compose build, config parsing, tag, registry check/start, push)
- `stage` — stage implementations (Build, Tag, Registry, Push) with progress integration
- `workflow` — pipeline sequencing, stage invocation
- `progress` — output formatting, testable via Writer var

## Capabilities

✓ Parse and validate all required CLI flags
✓ Display help text with usage examples
✓ Fail fast with clear error messages naming what failed
✓ Print stage progress in `[N/7]` format
✓ Run 7-stage workflow in sequence
✓ Build Docker Compose images and discover built images (Stage 1)
✓ Tag images with local registry prefix (Stage 2)
✓ Check registry status, detect port conflicts, start registry on :5001 (Stage 3)
✓ Push transfer-tagged images to local registry (Stage 4)
✓ Testable stage functions with mocked output

## Tech Stack

- **Language:** Go 1.22+
- **CLI parsing:** `flag` stdlib (no cobra/urfave)
- **Testing:** `testify` (require/assert)
- **Code quality:** `golangci-lint` (gci, gofmt, gofumpt, goimports)
- **Deployment target:** Remote host via SSH with local registry tunnel
- **Build system:** `go build`, `docker`, `ssh` (via CLI, not SDKs)
