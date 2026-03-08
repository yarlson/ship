# CLI Parsing

## Command Shape

```bash
ship [-i key] [-p port] user@host image[:tag] [image[:tag]...]
```

## Parsed Fields

`cli.Config` contains:

- `Images`
- `User`
- `Host`
- `KeyPath`
- `Port`

## Parsing Rules

1. Parse optional flags first:
   - `-i`, `--identity-file`
   - `-p`, `--port`
2. Expect at least two positional arguments after flags:
   - `<user@host>`
   - one or more `<image[:tag]>`
3. Split the target positional on `@` into `User` and `Host`.
4. Reject missing or malformed positional arguments.
5. Reject empty `-i` values and non-positive ports.

## Error Style

Examples:

- `missing required arguments: <user@host>, <image[:tag]>`
- `invalid target: example.com — expected <user@host>`
- `empty -i flag — provide the path to an SSH private key`

`main.go` prints these as `Error: <message>`.
