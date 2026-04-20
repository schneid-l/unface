// This file defines the error sentinels, the Error path-tracing type, and
// helpers for the soft-error protocol.
package plugin

import (
	"errors"
	"strings"
)

// ErrNotHandled is the soft-error sentinel. Returned by an Un*er method (or
// any adapter) to signal "I cannot handle this particular input; try the
// next candidate in the dispatch pipeline." Any error that wraps
// ErrNotHandled is treated the same by the dispatcher.
var ErrNotHandled = errors.New("unface: not handled")

// Hard error sentinels. Returned as-is or via fmt.Errorf("...: %w", ...) so
// callers can discriminate with errors.Is.
var (
	ErrInvalidDest  = errors.New("unface: dest must be a non-nil pointer")
	ErrNoCoercion   = errors.New("unface: no coercion available")
	ErrSrcNil       = errors.New("unface: src is nil")
	ErrUnknownField = errors.New("unface: unknown field")
	ErrRequired     = errors.New("unface: required field missing")
)

// Skip returns ErrNotHandled. Provided for symmetry with std patterns —
// some authors prefer a call expression over referencing a sentinel.
func Skip() error { return ErrNotHandled }

// Unhandled is an optional interface that errors can implement to opt into
// the soft-error protocol without wrapping ErrNotHandled.
type Unhandled interface {
	Unhandled() bool
}

// IsUnhandled reports whether err is, or wraps, the soft-error signal. It
// accepts both ErrNotHandled (via errors.Is) and the Unhandled interface.
func IsUnhandled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotHandled) {
		return true
	}
	var u Unhandled
	if errors.As(err, &u) {
		return u.Unhandled()
	}
	return false
}

// Error carries a structured coercion failure with a path trace from the
// destination root down to the offending field.
type Error struct {
	Path []string
	Src  any
	Dest any
	Err  error
}

// Error implements the error interface.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString("unface: ")
	if len(e.Path) > 0 {
		b.WriteString(strings.Join(e.Path, "."))
		b.WriteString(": ")
	}
	if e.Err != nil {
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

// Unwrap exposes the inner error for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.Err }

// WithPath prepends segments to the error's path. Returns a new Error so
// the receiver remains safe to share across goroutines.
func (e *Error) WithPath(segments ...string) *Error {
	if e == nil {
		return nil
	}
	out := *e
	out.Path = append(append([]string{}, segments...), e.Path...)
	return &out
}
