# Workflow Validation

## Preflight Checks

`workflow.Preflight(cfg)` runs these checks in order:

1. Docker is installed and responsive
2. `ssh` is installed
3. the SSH key path exists if `cfg.KeyPath` is not empty
4. each local Docker image exists
5. SSH connectivity to `cfg.User@cfg.Host` works

Preflight stops on the first failure.

## Error Shape

Representative messages:

- `docker is not installed or not in PATH`
- `SSH key file not found: /path/to/key — verify the -i path`
- `local image not found: app:latest — build or pull it first`
- `SSH connection failed — verify the target and SSH credentials`

## Notes

- Missing key path is allowed if the user relies on normal SSH identity resolution.
- SSH connectivity is tested with a remote `true` command before any transfer stages run.
- Preflight does not create Docker objects or open tunnels.
