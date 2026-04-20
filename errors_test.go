package unface_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/schneid-l/unface"
)

func TestErrNotHandledIsSentinel(t *testing.T) {
	if unface.ErrNotHandled == nil {
		t.Fatal("ErrNotHandled must be non-nil")
	}
	wrapped := fmt.Errorf("wrap: %w", unface.ErrNotHandled)
	if !errors.Is(wrapped, unface.ErrNotHandled) {
		t.Fatal("errors.Is must recognize wrapped ErrNotHandled")
	}
}

func TestSkipReturnsErrNotHandled(t *testing.T) {
	if !errors.Is(unface.Skip(), unface.ErrNotHandled) {
		t.Fatal("Skip() must return an error matching ErrNotHandled")
	}
}

func TestUnhandledInterfaceRecognised(t *testing.T) {
	err := customUnhandled{}
	if !unface.IsUnhandled(err) {
		t.Fatal("IsUnhandled must recognise errors implementing Unhandled")
	}
	if !unface.IsUnhandled(unface.ErrNotHandled) {
		t.Fatal("IsUnhandled must recognise the sentinel")
	}
	if unface.IsUnhandled(errors.New("other")) {
		t.Fatal("IsUnhandled must reject unrelated errors")
	}
	if unface.IsUnhandled(nil) {
		t.Fatal("IsUnhandled(nil) must be false")
	}
}

type customUnhandled struct{}

func (customUnhandled) Error() string   { return "custom unhandled" }
func (customUnhandled) Unhandled() bool { return true }

func TestErrorTypeFormatsPath(t *testing.T) {
	e := &unface.Error{Path: []string{"config", "server", "port"}, Err: errors.New("bad")}
	got := e.Error()
	want := "unface: config.server.port: bad"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if !errors.Is(e, e.Err) {
		t.Fatal("Error.Unwrap must expose inner error")
	}
}

func TestErrorTypeEmptyPath(t *testing.T) {
	e := &unface.Error{Err: errors.New("bad")}
	if got := e.Error(); got != "unface: bad" {
		t.Fatalf("got %q", got)
	}
}

func TestErrorWithPath(t *testing.T) {
	inner := &unface.Error{Path: []string{"port"}, Err: errors.New("bad")}
	outer := inner.WithPath("server")
	if outer.Error() != "unface: server.port: bad" {
		t.Fatalf("got %q", outer.Error())
	}
	// Original must be untouched (shared safely).
	if inner.Error() != "unface: port: bad" {
		t.Fatalf("inner mutated: %q", inner.Error())
	}
}

func TestSentinelDistinct(t *testing.T) {
	sentinels := []error{
		unface.ErrNotHandled,
		unface.ErrInvalidDest,
		unface.ErrNoCoercion,
		unface.ErrSrcNil,
		unface.ErrUnknownField,
		unface.ErrRequired,
	}
	seen := map[error]bool{}
	for _, s := range sentinels {
		if s == nil {
			t.Fatal("sentinel nil")
		}
		if seen[s] {
			t.Fatalf("duplicate sentinel: %v", s)
		}
		seen[s] = true
	}
}
