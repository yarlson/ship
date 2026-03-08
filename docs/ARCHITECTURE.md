# Architecture

## Module Map

```
ship (binary)
├── main.go              # Entry point — parse flags, invoke workflow
├── cli/
│   └── cli.go           # Flag parsing, validation, help text
├── workflow/
│   ├── workflow.go      # Orchestrator — runs stages 1-7 in sequence
│   ├── preflight.go     # Pre-workflow environment checks
│   └── errors.go        # StageError type
├── stage/
│   ├── build.go         # Stage 1: docker compose build
│   ├── tag.go           # Stage 2: re-tag images with localhost:5001/ prefix
│   ├── registry.go      # Stage 3: ensure local registry container on :5001
│   ├── push.go          # Stage 4: push to local registry
│   ├── tunnel.go        # Stage 5: reverse SSH tunnel
│   ├── pull.go          # Stage 6: remote pull + restore original tags
│   └── command.go       # Stage 7: execute remote command via SSH
├── docker/
│   └── docker.go        # Wrappers for docker/compose CLI invocations
├── ssh/
│   └── ssh.go           # SSH command execution and tunnel management
└── progress/
    └── progress.go      # Stage progress printer ([1/7] Building...)
```

## Module Boundaries

| Module     | Responsibility                                                 | Boundary                                        |
| ---------- | -------------------------------------------------------------- | ----------------------------------------------- |
| `cli`      | Parse and validate CLI flags. Return typed `Config` or error.  | No I/O beyond flag parsing.                     |
| `workflow` | Execute stages in order, handle fail-fast, print summary.      | Calls stage functions. Owns the stage sequence. |
| `stage`    | Each stage: `func(cfg, state) error`. Read/write shared state. | Calls `docker` and `ssh` modules.               |
| `docker`   | Thin wrappers around `docker`/`docker compose` CLI.            | Executes external processes via `os/exec`.      |
| `ssh`      | Remote commands and tunnel lifecycle.                          | Executes `ssh` binary via `os/exec`.            |
| `progress` | Format and print `[N/7]` stage lines.                          | Writes to stdout.                               |

## Data Flow

```
CLI Flags → cli.Parse() → Config
  → workflow.Preflight(Config) → error or proceed
  → workflow.Run(Config)
    → Stage 1: Build → discover images → []Image
    → Stage 2: Tag → ImageMap (original ↔ localhost:5001/*)
    → Stage 3: Registry → ensure localhost:5001
    → Stage 4: Push → images in local registry
    → Stage 5: Tunnel → SSH reverse tunnel (background process)
    → Stage 6: Pull+Restore → remote has original-named images
    → Stage 7: Command → user's deploy command on remote
  → Print summary → exit 0
```

## Key Design Decisions

- **Shell out to CLIs** — not Go SDK libraries. Smaller binary, fewer deps, identical behavior to manual commands.
- **No goroutines in workflow** — stages are sequential. Only the SSH tunnel runs as a background process, cleaned up via defer.
- **Shared mutable state** — `WorkflowState` struct with `ImageMap` and `TunnelCmd`, passed by pointer. No channels, no mutexes.
- **Fixed registry port** — `localhost:5001` is hardcoded (M-11).

## Detailed Specs

Full requirements, design rules, and task breakdown live in `.snap/sessions/default/tasks/`:

- `PRD.md` — requirements (M-1 through M-11, S-1 through S-5)
- `DESIGN.md` — voice, terminology, 30 contract rules, UI state matrix
- `TECHNOLOGY.md` — engineering north star, testing strategy, CI/release
- `TASKS.md` — task plan and dependency graph
- `TASK0.md`–`TASK5.md` — individual task specs with acceptance criteria
