# CLI Parsing

## Command Shape

```bash
ship [-i key] [-p port] user@host image[:tag]
```

## Parsed Fields

`cli.Config` contains:

- `Image`
- `User`
- `Host`
- `KeyPath`
- `Port`

## Parsing Rules

1. Parse optional flags first:
   - `-i`, `--identity-file`
   - `-p`, `--port`
2. Expect exactly two positional arguments after flags:
   - `<user@host>`
   - `<image[:tag]>`
3. Split the target positional on `@` into `User` and `Host`.
4. Reject missing, malformed, or extra positional arguments.
5. Reject empty `-i` values and non-positive ports.

## Error Style

Examples:

- `missing required arguments: <user@host>, <image[:tag]>`
- `invalid target: example.com — expected <user@host>`
- `unexpected arguments: extra`
- `empty -i flag — provide the path to an SSH private key`

`main.go` prints these as `Error: <message>`.
