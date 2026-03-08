package workflow

import "fmt"

// StageError is a structured error for stage failures.
// main.go adds the "Error: " prefix when printing to stderr.
// With hint, produces: "<what> — <hint>". Without hint: "<what>".
type StageError struct {
	Stage int
	Name  string
	Err   error
	Hint  string
}

// Error returns the formatted error message.
// With hint: "<what> — <hint>". Without hint: "<what>".
func (e *StageError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s — %s", e.Err.Error(), e.Hint)
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error for errors.Is/errors.As compatibility.
func (e *StageError) Unwrap() error {
	return e.Err
}
