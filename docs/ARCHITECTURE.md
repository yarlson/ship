# Architecture

## Module Map

```
ship (binary)
├── main.go              # Entry point — parse args, invoke workflow
├── cli/
│   └── cli.go           # SSH-style CLI parsing and validation
├── workflow/
│   ├── workflow.go      # Orchestrator — runs stages 1-5 in sequence
│   ├── preflight.go     # Pre-workflow environment checks
│   └── errors.go        # StageError type
├── stage/
│   ├── tag.go           # Stage 1: create localhost:5001/ transfer tag
│   ├── registry.go      # Stage 2: ensure local registry container on :5001
│   ├── push.go          # Stage 3: push transfer tag to local registry
│   ├── tunnel.go        # Stage 4: reverse SSH tunnel
│   └── pull.go          # Stage 5: remote pull + restore original tag
├── docker/
│   ├── docker.go        # Local image checks and tagging helpers
│   └── registry.go      # Local registry lifecycle helpers
├── ssh/
│   └── ssh.go           # SSH command execution and tunnel management
└── progress/
    └── progress.go      # Stage progress printer ([1/5] Tagging...)
```

## Module Boundaries

| Module     | Responsibility                                                          | Boundary                                   |
| ---------- | ----------------------------------------------------------------------- | ------------------------------------------ |
| `cli`      | Parse `ship [-i key] [-p port] user@host image[:tag] [image[:tag]...]`. | No I/O beyond flag parsing.                |
| `workflow` | Execute preflight and the 5 transfer stages.                            | Calls stage functions.                     |
| `stage`    | Perform transfer stages in order.                                       | Calls `docker` and `ssh` modules.          |
| `docker`   | Thin wrappers around Docker CLI operations.                             | Executes external processes via `os/exec`. |
| `ssh`      | Remote commands and tunnel lifecycle.                                   | Executes `ssh` via `os/exec`.              |
| `progress` | Format and print `[N/5]` stage lines.                                   | Writes to stdout.                          |

## Data Flow

```
CLI Args → cli.Parse() → Config
  → workflow.Preflight(Config) → error or proceed
  → workflow.Run(Config)
    → Stage 1: Tag → localhost:5001/<image> for each image
    → Stage 2: Registry → ensure localhost:5001
    → Stage 3: Push → transfer tags available in local registry
    → Stage 4: Tunnel → SSH reverse tunnel (background process)
    → Stage 5: Pull+Restore → remote host has original image tags
  → Print summary → exit 0
```

## Key Design Decisions

- Shell out to `docker` and `ssh` instead of using SDKs.
- Keep the workflow narrow: transfer one or more images, do not deploy them.
- Use a fixed local registry port, `5001`.
- Keep the original image tag on the remote host after transfer.
- Clean up the SSH tunnel with `defer` once the workflow returns.
