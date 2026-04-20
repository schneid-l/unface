package plugin_test

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
)

// ExampleNumberOf wraps a concrete integer in the Number abstraction
// and demonstrates the lossless accessors.
func ExampleNumberOf() {
	n, ok := plugin.NumberOf(42)
	if !ok {
		fmt.Println("not a number")
		return
	}
	i, _ := n.Int64()
	f, _ := n.Float64()
	fmt.Println(i, f)
	// Output: 42 42
}

// ExampleMapOf wraps a map[string]any and reads typed values via the
// Map abstraction's accessors.
func ExampleMapOf() {
	m, _ := plugin.MapOf(map[string]any{
		"host": "localhost",
		"port": 8080,
	})
	host, _ := m.GetString("host")
	_, hasPort := m.Get("port")
	fmt.Println("len=", m.Len())
	fmt.Println("host=", host)
	fmt.Println("hasPort=", hasPort)
	// Output:
	// len= 2
	// host= localhost
	// hasPort= true
}

// ExampleListOf wraps a []int, filters odd values, and reports the
// resulting length.
func ExampleListOf() {
	l, _ := plugin.ListOf([]int{1, 2, 3, 4, 5})
	evens := l.Filter(func(v any) bool { return v.(int)%2 == 0 })
	fmt.Println("len=", evens.Len())
	evens.Iter(func(_ int, v any) bool {
		fmt.Println(v)
		return true
	})
	// Output:
	// len= 2
	// 2
	// 4
}

// ExampleCompose bundles two atomic plugins under a new name. The
// composite's ChildNames exposes the bundled plugins (useful for Without).
func ExampleCompose() {
	p1 := plugin.NewPlugin("a")
	p2 := plugin.NewPlugin("b")
	combo := plugin.Compose("ab", p1, p2)
	names := append([]string{combo.Name()}, combo.ChildNames()...)
	sort.Strings(names)
	fmt.Println(strings.Join(names, ","))
	// Output: a,ab,b
}

// color is a tiny enum driven by a plugin that parses "red", "green",
// and "blue".
type color int

const (
	colorRed color = iota + 1
	colorGreen
	colorBlue
)

func (c color) String() string {
	switch c {
	case colorRed:
		return "red"
	case colorGreen:
		return "green"
	case colorBlue:
		return "blue"
	}
	return "?"
}

type colorAdapter struct{ dst *color }

func (a colorAdapter) Unface(src any) error {
	s, ok := src.(string)
	if !ok {
		return plugin.ErrNotHandled
	}
	switch s {
	case "red":
		*a.dst = colorRed
	case "green":
		*a.dst = colorGreen
	case "blue":
		*a.dst = colorBlue
	default:
		return fmt.Errorf("bad color %q", s)
	}
	return nil
}

// ExampleNewPlugin writes a minimal plugin from scratch: a factory that
// matches *color destinations plus an adapter that parses string sources
// into the enum. The plugin is then registered with a Facer.
func ExampleNewPlugin() {
	colorType := reflect.TypeOf(color(0))
	colorPlugin := plugin.NewPlugin("color", plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == colorType },
		func(ptr any) plugin.Adapter { return colorAdapter{dst: ptr.(*color)} },
	))

	f := engine.New(engine.With(colorPlugin))
	var c color
	if err := f.Unface("green", &c); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(c)
	// Output: green
}

// urlDest implements Unstringer. The engine's source-specific dispatch
// path picks it up for a string source before consulting any plugin.
type urlDest struct {
	Scheme, Host string
}

func (u *urlDest) Unstring(s string) error {
	i := strings.Index(s, "://")
	if i < 0 {
		return plugin.ErrNotHandled
	}
	u.Scheme, u.Host = s[:i], s[i+3:]
	return nil
}

// ExampleUnstringer shows how a destination type can opt into string
// coercion by implementing Unstringer. engine.New() with no plugins is
// sufficient — the Un*er dispatch fires before plugins are consulted.
func ExampleUnstringer() {
	f := engine.New()
	var u urlDest
	if err := f.Unface("https://example.com", &u); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(u.Scheme, u.Host)
	// Output: https example.com
}

// resource is a tiny polymorphic type: the "kind" field decides what
// the rest of the map means. It implements Unmapper to run that switch
// itself.
type resource struct {
	Kind string
	Name string
}

func (r *resource) Unmap(m plugin.Map) error {
	kind, ok := m.GetString("kind")
	if !ok {
		return errors.New("kind missing")
	}
	r.Kind = kind
	if name, ok := m.GetString("name"); ok {
		r.Name = name
	}
	return nil
}

// ExampleUnmapper shows polymorphic dispatch: a map source is handed to
// the destination's Unmap method, which reads "kind" and populates
// fields accordingly.
func ExampleUnmapper() {
	f := engine.New()
	var r resource
	if err := f.Unface(map[string]any{"kind": "Service", "name": "web"}, &r); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(r.Kind, r.Name)
	// Output: Service web
}
