# Terminology

- **transfer tag** — `localhost:5001/<name>:<tag>` format used to push/pull images through the local registry and SSH tunnel
- **tunnel** — reverse SSH port forward (`-R 5001:localhost:5001`) allowing the remote host to pull from the local registry
- **TunnelProcess** — Go struct wrapping `*exec.Cmd` with a `done` channel for safe lifecycle management of the background SSH process
- **stage** — one of 7 sequential steps in the deployment pipeline
- **progress.Writer** — swappable `io.Writer` used for all user-facing output (testable by replacing with `bytes.Buffer`)
