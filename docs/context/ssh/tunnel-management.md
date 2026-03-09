# SSH Tunnel Management

## Responsibilities

The `ssh` module owns:

- argument construction for SSH commands
- remote command execution
- reverse tunnel startup
- reverse tunnel shutdown
- respecting caller-provided cancellation and deadlines for blocking SSH operations

## Remote Commands

Ship uses remote SSH commands internally only for the transfer path:

- `docker pull <transfer>`
- `docker tag <transfer> <original>`
- preflight `true` connectivity check

## Tunnel Shape

The reverse tunnel uses:

```text
ssh -R 5001:localhost:5001 -N user@host
```

Optional SSH key and port are included when configured.

## Lifecycle

- `RunRemoteCommand(ctx, ...)` executes one remote command with caller-owned cancellation
- `StartTunnel(ctx, ...)` starts the background SSH process
- `TunnelProcess.Exited()` exposes whether the process has exited
- `StopTunnel(ctx, ...)` sends SIGTERM, waits, then SIGKILLs if required

The workflow defers cleanup immediately after tunnel establishment and may use a bounded cleanup context so teardown still runs after workflow cancellation.
