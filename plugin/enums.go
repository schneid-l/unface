package plugin

// MatchMode controls how struct field names are matched against source map
// keys.
type MatchMode int

const (
	// MatchFold is the default mode: case-insensitive, with snake_case /
	// kebab-case folded into CamelCase. "http_port", "http-port",
	// "httpPort", and "HTTPPort" all match the field HTTPPort.
	MatchFold MatchMode = iota
	// MatchInsensitive matches case-insensitively; separators are
	// significant.
	MatchInsensitive
	// MatchExact requires byte-for-byte equality.
	MatchExact
)

// UnknownPolicy controls what the struct walker does with source keys that
// don't map to any destination field (and are not absorbed by a remainder).
type UnknownPolicy int

const (
	// UnknownIgnore is the default: unknown keys are silently discarded.
	UnknownIgnore UnknownPolicy = iota
	// UnknownError returns an error wrapping ErrUnknownField on the first
	// unknown key.
	UnknownError
	// UnknownWarn invokes a user-supplied handler and continues.
	UnknownWarn
)

// UnknownHandler is called when OnUnknown(UnknownWarn, handler) is set.
type UnknownHandler = func(field string, value any)

// SoftErrorHandler observes soft errors as the dispatcher unwinds. It cannot
// change behavior; useful for debug logging.
type SoftErrorHandler = func(src, dest any, err error)

// PointerResolution controls how multi-level pointers are handled on src and
// dest. See engine.WithPointerResolve for details.
type PointerResolution int

const (
	// PointerResolveFlat is the default: flatten multi-level pointers to
	// the innermost non-pointer element once before dispatch. Nil
	// intermediates on dest are allocated along the way; src is
	// dereferenced in-place.
	PointerResolveFlat PointerResolution = iota
	// PointerResolveNone disables dereferencing. Dispatch runs at the
	// caller-supplied depth. A method set on *T will not fire for a **T
	// dest.
	PointerResolveNone
	// PointerResolveDeep tries every pointer level on both src and dest.
	// If a level returns ErrNotHandled, the dispatcher unwraps one layer
	// and retries. Dest outer->inner; per dest level, src outer->inner.
	PointerResolveDeep
)
