package diagnosis

import (
	"fmt"
	"runtime/debug"
)

type stackError struct {
	cause error
	stack string
}

func (e *stackError) Error() string {
	return fmt.Sprintf("%v\ncall stack:\n%s", e.cause, e.stack)
}

func (e *stackError) Unwrap() error {
	return e.cause
}

// newStackErrorf must be called where the error is produced. Callers that
// merely receive an error should wrap it with %w and return it unchanged so
// the stack points to the original failure site.
func newStackErrorf(format string, args ...any) error {
	return &stackError{cause: fmt.Errorf(format, args...), stack: string(debug.Stack())}
}
