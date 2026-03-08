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
- `workflow/` — stage orchestration
- `progress/` — progress output formatting
- Future: `stage/`, `docker/`, `ssh/` (stubs in workflow)

**Data flow:** CLI flags → Config → WorkflowState (shared across stages) → result

## Core Flow

1. **Entry** — `main.go` calls `cli.Parse(args)` to extract Config
2. **Validation** — Config has required flags or returns error
3. **Orchestration** — `workflow.Run(cfg)` executes 7 stages sequentially
4. **Progress** — Each stage prints `[N/7] message` lines
5. **Exit** — Workflow returns nil on success or error on failure

## System State

**Current implementation:** Stub stages with hardcoded stage pipeline.
- Stages run in sequence with progress output
- No actual Docker, SSH, or registry operations yet
- All required flags parsed and validated
- Error handling scaffolding in place

**Module boundaries enforced:**
- `cli` — only flag parsing, no I/O
- `workflow` — stage sequencing, no external processes
- `progress` — output formatting, testable via Writer var

## Capabilities

✓ Parse and validate all required CLI flags
✓ Display help text with usage examples
✓ Fail fast with clear error messages naming what failed
✓ Print stage progress in `[N/7]` format
✓ Run 7-stage workflow in sequence
✓ Testable stage functions with mocked output

## Tech Stack

- **Language:** Go 1.22+
- **CLI parsing:** `flag` stdlib (no cobra/urfave)
- **Testing:** `testify` (require/assert)
- **Code quality:** `golangci-lint` (gci, gofmt, gofumpt, goimports)
- **Deployment target:** Remote host via SSH with local registry tunnel
- **Build system:** `go build`, `docker`, `ssh` (via CLI, not SDKs)
