# Go Conventions

## Style

- Go 1.22+
- `flag` stdlib for CLI parsing — no cobra/urfave
- Formatting enforced by `golangci-lint run --fix` (gci, gofmt, gofumpt, goimports)
- Linter config in `.golangci.yaml` — treat all enabled linter findings as errors

## Error Handling

- Every exported function that can fail returns `error`
- Wrap errors with context: `fmt.Errorf("stage %d: %w", n, err)`
- Never swallow errors — propagate or log and exit
- Use `StageError` type (in `workflow/errors.go`) for stage failures: carries stage number, name, underlying error, and actionable hint

## Assertions

- Use `github.com/stretchr/testify` for all test assertions (`require` for fatal, `assert` for non-fatal)
- Prefer `require` when a failure makes subsequent assertions meaningless
- Table-driven tests for functions with multiple input/output cases

## Tag Conventions

- JSON fields: `snake_case`
- YAML fields: `snake_case`
- See `.golangci.yaml` tagliatelle config for full mapping

## Import Order

Enforced by gci (configured in `.golangci.yaml`):

1. Standard library
2. External packages
3. Local module
