# SSH Tunnel Management

## Responsibilities

The `ssh` module owns:

- argument construction for SSH commands
- remote command execution
- reverse tunnel startup
- reverse tunnel shutdown

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

- `StartTunnel()` starts the background SSH process
- `TunnelProcess.Exited()` exposes whether the process has exited
- `StopTunnel()` sends SIGTERM, waits, then SIGKILLs if required

The workflow defers cleanup immediately after tunnel establishment.
