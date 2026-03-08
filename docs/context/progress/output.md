# Progress Output

Progress output follows a fixed `[N/5]` format.

## Start Line

```text
[N/5] message...
```

Example:

```text
[4/5] Establishing tunnel to 46.101.213.82...
```

## Complete Line

```text
[N/5] message
```

Example:

```text
[4/5] Tunnel established
```

## Current Stage Messages

- `[1/5] Tagging image for transfer...`
- `[1/5] Tag complete`
- `[2/5] Starting local registry...`
- `[2/5] Registry ready`
- `[3/5] Pushing image to local registry...`
- `[3/5] Push complete`
- `[4/5] Establishing tunnel to <host>...`
- `[4/5] Tunnel established`
- `[5/5] Pulling and restoring image on remote host...`
- `[5/5] Pull and restore complete`

## Testability

`progress.Writer` is replaceable in tests so progress output can be captured without asserting on real stdout.
