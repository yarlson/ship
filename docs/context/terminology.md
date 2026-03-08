# Terminology

**CLI config** — Parsed command-line input stored in `cli.Config`. Contains `Image`, `User`, `Host`, optional `KeyPath`, and `Port`.

**Original image** — The image reference the user passed to `ship`, for example `app:latest`.

**Transfer tag** — The temporary local-registry form of the image ref, for example `localhost:5001/app:latest`.

**Local registry** — A registry container running locally and exposed on port `5001`. Used only as the transfer bridge.

**Reverse tunnel** — An SSH `-R 5001:localhost:5001` tunnel that lets the remote host reach the local registry.

**Remote restore** — The remote `docker tag <transfer> <original>` step that restores the original image ref after pull.

**Stage** — One of the 5 ordered workflow steps: tag, registry, push, tunnel, pull/restore.

**Preflight** — The checks that run before the first stage: Docker, SSH, optional key path, local image existence, and SSH connectivity.

**Progress output** — The `[N/5]` lines printed by the workflow to show stage start and completion.
