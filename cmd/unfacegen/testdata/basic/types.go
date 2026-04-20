// Package basic is the unfacegen golden input. It declares types with
// various Un*er methods; the golden .expected file next to it shows the
// Unfacer methods the generator should produce.
package basic

import (
	"fmt"
	"time"

	"github.com/schneid-l/unface"
)

// URL implements Unstringer and Unmapper. Generated Unface must dispatch
// string src to Unstring and map src to Unmap.
type URL struct {
	Scheme string
	Host   string
}

func (u *URL) Unstring(s string) error {
	u.Host = s
	return nil
}

func (u *URL) Unmap(m unface.Map) error {
	if v, ok := m.GetString("scheme"); ok {
		u.Scheme = v
	}
	if v, ok := m.GetString("host"); ok {
		u.Host = v
	}
	return nil
}

// Counter implements only Unnumberer. Generated Unface must use NumberOf.
type Counter struct{ N int64 }

func (c *Counter) Unnumber(n unface.Number) error {
	v, ok := n.Int64()
	if !ok {
		return fmt.Errorf("counter: not int64-representable")
	}
	c.N = v
	return nil
}

// Event implements Untimer. Generated Unface must dispatch time.Time src.
type Event struct{ When time.Time }

func (e *Event) Untime(t time.Time) error {
	e.When = t
	return nil
}

// Flags implements Unbooler + Unstringer.
type Flags struct{ Enabled bool }

func (f *Flags) Unbool(b bool) error     { f.Enabled = b; return nil }
func (f *Flags) Unstring(s string) error { f.Enabled = s == "yes"; return nil }
