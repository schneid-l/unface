package plugin

import "time"

// Unfacer is the master consumer. A destination implementing Unfacer handles
// every kind of src itself. The Facer dispatcher prefers Unfacer over the
// more specific Un*er interfaces.
type Unfacer interface {
	Unface(src any) error
}

// Unstringer consumes a string source.
type Unstringer interface {
	Unstring(s string) error
}

// Unbooler consumes a bool source.
type Unbooler interface {
	Unbool(b bool) error
}

// Unbyteser consumes a []byte source. If a destination implements both
// Unbyteser and Unstringer, the dispatcher prefers Unbyteser for []byte src.
type Unbyteser interface {
	Unbytes(b []byte) error
}

// Unruner consumes a single rune source.
type Unruner interface {
	Unrune(r rune) error
}

// Unniler consumes an explicit nil source.
type Unniler interface {
	Unnil() error
}

// Untimer consumes a time.Time source directly (bypassing string parsing).
type Untimer interface {
	Untime(t time.Time) error
}

// Undurationer consumes a time.Duration source.
type Undurationer interface {
	Unduration(d time.Duration) error
}
