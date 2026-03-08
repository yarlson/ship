# Practices & Conventions

## Code Quality

- **Formatting** — Enforced by `golangci-lint run --fix` (gci, gofmt, gofumpt, goimports). No manual style debates.
- **Import order** — Standard lib → external → local (enforced by gci).
- **Error wrapping** — Always wrap errors with context using `fmt.Errorf()`. Never swallow errors.
- **Linter findings** — All enabled linter rules are treated as errors, not warnings.

## Testing

- **Framework** — `testify` (`require` for fatal assertions, `assert` for non-fatal).
- **Table-driven tests** — Used for functions with multiple cases.
- **Output testing** — Mock `progress.Writer` to capture output in tests instead of checking stdout.
- **TDD approach** — Write failing test first, then minimal code to pass.

## CLI & Error Handling

- **Flag library** — Use stdlib `flag`, not cobra/urfave.
- **Help text** — Defined as constant in `cli.go`, includes examples.
- **Validation** — All required flags checked. Missing flags listed by name.
- **Error messages** — User-facing errors name what failed and what to check (no secrets in output).

## Module Boundaries

- **`cli`** — Only flag parsing. No I/O beyond parsing.
- **`workflow`** — Stage orchestration. Calls stage functions. Owns sequence.
- **`progress`** — Output formatting. Testable via `Writer` var.
- **`docker`** — Docker CLI operations (build, config, tag). Wraps `docker` CLI. Returns errors, not exit codes.
- **`stage`** — Stage implementations. Each stage function: calls docker/SSH utilities, calls progress functions, returns data for next stage or error.
- **External processes** — Shell out to `docker` and `ssh` CLIs, not Go SDKs.

## Stage Implementation Pattern

- **Stage functions:** Named `stage.Build(cfg)`, `stage.Tag(imageMap)`, etc. Each takes required inputs and returns (result, error).
- **Progress calls:** Each stage opens with `progress.StageStart(n, msg)` and closes with `progress.StageComplete(n, msg)`. Between them are the real operations.
- **Error flow:** Stages return errors which bubble up to `workflow.Run()`. Errors contain context (what operation, why it failed).
- **Data passing:** ImageMap pattern used to pass image metadata between stages (Stage 1 → Stage 2 → Stages 3-7 stubs).
- **Docker operations:** `docker.ComposeBuild()`, `docker.ComposeConfig()`, `docker.TagImage()` handle CLI invocation and error wrapping.

## Secrets & Output

- **Key paths OK** — Safe to log `~/.ssh/id_rsa` paths.
- **Key contents forbidden** — Never log private key material or remote passwords.
