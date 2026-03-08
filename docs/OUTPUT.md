# Output Rules

All user-facing output must follow these rules. Full spec: `.snap/sessions/default/tasks/DESIGN.md`

## Stage Progress

```
[N/7] <Present participle>...       ← stage start (stdout)
[N/7] <Noun> complete (<detail>)    ← stage end (stdout)
```

Examples:

```
[1/7] Building images...
[1/7] Build complete (3 images)
[5/7] Establishing tunnel to deploy.example.com...
[5/7] Tunnel established
```

## Errors

```
Error: <what failed> — <remediation hint>    ← always to stderr
```

Examples:

```
Error: Build failed — see docker compose output above
Error: SSH tunnel failed — connection refused (verify --host and --key)
Error: Port 5001 already in use — stop the existing process or free the port
```

## Success Summary

```
Ship complete
  Host:     <host>
  Images:   <original names, comma-separated>
  Command:  <command>
  Status:   Success
```

Labels are 10 chars wide including colon. Image names use original compose refs, never `localhost:5001/` transfer tags.

## Hard Rules

- No ANSI escape codes or color (MVP)
- No emoji
- No first person ("I", "we")
- No second person in progress/summary ("you") — allowed only in error hints
- No hedging ("might", "possibly", "try to")
- Errors reference the relevant `--flag` when caused by user input
- Errors include the actual value (path, host) when not sensitive
- Docker/SSH passthrough output is never reformatted
- Stage start line must be printed before its error
- No blank lines between stage progress lines
- Missing-flag errors list ALL missing flags in one message, followed by usage line

## Terminology

| Use          | Avoid                   |
| ------------ | ----------------------- |
| stage        | step, phase, task       |
| image        | container, artifact     |
| transfer     | upload, sync            |
| remote host  | server, target, machine |
| tunnel       | connection, link        |
| compose file | config, manifest        |
