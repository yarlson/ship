# Output Rules

## Stage Progress

```
[N/5] <Present participle>...  ← stage start (stdout)
[N/5] <Noun> complete          ← stage end (stdout)
```

Examples:

```
[1/5] Tagging image for transfer...
[1/5] Tag complete
[4/5] Establishing tunnel to deploy.example.com...
[4/5] Tunnel established
```

## Errors

```
Error: <what failed> — <remediation hint>    ← always to stderr
```

Examples:

```
Error: Local image not found: app:latest — build or pull it first
Error: SSH tunnel failed — verify the target and SSH credentials
Error: Port 5001 already in use — stop the existing process or free the port
```

## Success Summary

```
Ship complete
  Host:     <host>
  Image:    <original image tag>
  Status:   Success
```

Labels are 10 chars wide including colon. The summary shows the original image ref, never the `localhost:5001/` transfer tag.

## Hard Rules

- No ANSI escape codes or color
- No emoji
- No hedging
- Stage start line must be printed before its error
- No blank lines between stage progress lines
- Docker and SSH passthrough output is never reformatted
