package engine

import (
	"fmt"
	"reflect"
	"time"

	"github.com/schneid-l/unface/internal/reflectutil"
	"github.com/schneid-l/unface/plugin"
)

// Facer coerces src values into dest values using a configured set of
// plugins. It is safe for concurrent use once constructed.
type Facer struct {
	cfg *plugin.Config
}

// New builds a Facer with the given options. No plugins are loaded unless
// explicitly requested via With(...).
func New(opts ...Option) *Facer {
	c := plugin.NewDefaultConfig()
	for _, o := range opts {
		o(c)
	}
	return &Facer{cfg: c}
}

// Strict is a Facer with no plugins. Only Un*er methods on dest are honored.
var Strict = New()

// Config returns the Facer's underlying configuration pointer. Intended
// for introspection by built-in adapters that need to dispatch recursively.
// Callers should treat the returned value as read-only.
func (f *Facer) Config() *plugin.Config { return f.cfg }

// Unface coerces src into dest.
func (f *Facer) Unface(src, dest any, opts ...Option) error {
	cfg := f.cfg
	if len(opts) > 0 {
		cfg = cfg.Clone()
		for _, o := range opts {
			o(cfg)
		}
	}
	return f.dispatch(cfg, src, dest)
}

// Dispatch is the low-level entry point used by built-in adapters to
// re-enter the dispatcher with an explicit Config. It performs the same
// pointer-resolution and fast-path/plugin walk as Unface.
func (f *Facer) Dispatch(cfg *plugin.Config, src, dest any) error {
	if cfg == nil {
		cfg = f.cfg
	}
	return f.dispatch(cfg, src, dest)
}

func (f *Facer) dispatch(cfg *plugin.Config, src, dest any) error {
	if dest == nil {
		return plugin.ErrInvalidDest
	}
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Pointer || dv.IsNil() {
		return plugin.ErrInvalidDest
	}

	switch cfg.PointerMode {
	case plugin.PointerResolveNone:
		return translateTerminal(f.dispatchOne(cfg, src, dest))

	case plugin.PointerResolveDeep:
		destLevels := pointerLadder(dest)
		srcLevels := srcLadder(src)
		for _, d := range destLevels {
			for _, s := range srcLevels {
				err := f.dispatchOne(cfg, s, d)
				if err == nil {
					return nil
				}
				if !plugin.IsUnhandled(err) {
					return err
				}
			}
		}
		return plugin.ErrNoCoercion

	case plugin.PointerResolveFlat:
		fallthrough
	default:
		if _, err := reflectutil.DerefToAddressable(dest); err != nil {
			return fmt.Errorf("%w", plugin.ErrInvalidDest)
		}
		flatDest := innermostDestPtr(dest)
		flatSrc := flattenSrc(src)
		return translateTerminal(f.dispatchOne(cfg, flatSrc, flatDest))
	}
}

// translateTerminal converts the dispatcher's internal soft "unhandled"
// terminal into the public ErrNoCoercion. Used by Flat and None modes where
// there is exactly one dispatch attempt.
func translateTerminal(err error) error {
	if err != nil && plugin.IsUnhandled(err) {
		return plugin.ErrNoCoercion
	}
	return err
}

// dispatchOne runs steps 2-5 of the dispatch order for a single (src, dest)
// pair. dest must be a non-nil pointer.
func (f *Facer) dispatchOne(cfg *plugin.Config, src, dest any) error {
	if u, ok := dest.(plugin.Unfacer); ok {
		if err := callUnfacer(u, src, cfg); err == nil {
			return nil
		} else if !plugin.IsUnhandled(err) {
			return err
		}
	}

	if assignDirect(src, dest) {
		return nil
	}

	if err := tryInterfaceDispatch(src, dest); err == nil {
		return nil
	} else if !plugin.IsUnhandled(err) {
		return err
	} else if cfg.SoftHandler != nil {
		cfg.SoftHandler(src, dest, err)
	}

	destType := reflect.TypeOf(dest).Elem()
	for _, p := range cfg.Plugins {
		for _, fac := range p.Factories() {
			if !fac.Matches(destType) {
				continue
			}
			adapter := fac.For(dest)
			if ca, ok := adapter.(plugin.CfgAware); ok {
				adapter = ca.WithConfig(cfg)
			}
			if err := callUnfacer(adapter, src, cfg); err == nil {
				return nil
			} else if !plugin.IsUnhandled(err) {
				return err
			}
		}
	}
	return plugin.ErrNotHandled
}

// callUnfacer invokes u.Unface(src) and forwards soft-error observations.
func callUnfacer(u plugin.Unfacer, src any, cfg *plugin.Config) error {
	err := u.Unface(src)
	if err != nil && plugin.IsUnhandled(err) && cfg.SoftHandler != nil {
		cfg.SoftHandler(src, u, err)
	}
	return err
}

// assignDirect handles dest = *T, src = T (exactly) for value-kind types
// where direct assignment produces an independent value, plus the
// special case of an empty-interface dest (which accepts any value).
//
// Reference-like kinds (slice, map, chan) are deliberately skipped when
// the source type matches, so that plugins can deep-copy them and avoid
// accidental aliasing with the caller's value.
func assignDirect(src, dest any) bool {
	if src == nil || dest == nil {
		return false
	}
	dp := reflect.ValueOf(dest)
	if dp.Kind() != reflect.Pointer || dp.IsNil() {
		return false
	}
	ev := dp.Elem()
	if !ev.CanSet() {
		return false
	}
	sv := reflect.ValueOf(src)
	if !sv.IsValid() {
		return false
	}

	// Interface dest (e.g. *any, *MyIface): assign if src implements it.
	if ev.Kind() == reflect.Interface {
		if sv.Type().Implements(ev.Type()) {
			ev.Set(sv)
			return true
		}
		return false
	}

	if sv.Type() != ev.Type() {
		return false
	}

	// Skip aliasing-prone kinds so the plugin layer can copy.
	switch ev.Kind() {
	case reflect.Slice, reflect.Map, reflect.Chan:
		return false
	default:
		ev.Set(sv)
		return true
	}
}

// tryInterfaceDispatch dispatches based on src's kind. Returns ErrNotHandled
// if none of the Un*er interfaces match or the matching one declines.
func tryInterfaceDispatch(src, dest any) error {
	if src == nil {
		if u, ok := dest.(plugin.Unniler); ok {
			return u.Unnil()
		}
		return plugin.ErrNotHandled
	}
	switch v := src.(type) {
	case bool:
		if u, ok := dest.(plugin.Unbooler); ok {
			return u.Unbool(v)
		}
	case string:
		if u, ok := dest.(plugin.Unstringer); ok {
			return u.Unstring(v)
		}
	case []byte:
		if u, ok := dest.(plugin.Unbyteser); ok {
			return u.Unbytes(v)
		}
		if u, ok := dest.(plugin.Unstringer); ok {
			return u.Unstring(string(v))
		}
	case time.Time:
		if u, ok := dest.(plugin.Untimer); ok {
			return u.Untime(v)
		}
	case time.Duration:
		if u, ok := dest.(plugin.Undurationer); ok {
			return u.Unduration(v)
		}
	}
	if n, ok := plugin.NumberOf(src); ok {
		if u, ok := dest.(plugin.Unnumberer); ok {
			return u.Unnumber(n)
		}
	}
	if l, ok := plugin.ListOf(src); ok {
		if u, ok := dest.(plugin.Unlister); ok {
			return u.Unlist(l)
		}
	}
	if m, ok := plugin.MapOf(src); ok {
		if u, ok := dest.(plugin.Unmapper); ok {
			return u.Unmap(m)
		}
	}
	return plugin.ErrNotHandled
}

// pointerLadder returns a chain of pointer layers from the caller-supplied
// dest pointer down to a pointer to the innermost non-pointer element.
// Nil intermediates are allocated along the way. For dest = ****T, the
// returned slice is [****T, ***T, **T, *T] (outer to inner).
func pointerLadder(dest any) []any {
	out := []any{dest}
	rv := reflect.ValueOf(dest)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			break
		}
		inner := rv.Elem()
		if inner.Kind() != reflect.Pointer {
			break
		}
		if inner.IsNil() {
			if inner.CanSet() {
				inner.Set(reflect.New(inner.Type().Elem()))
			} else {
				break
			}
		}
		out = append(out, inner.Interface())
		rv = inner
	}
	return out
}

// srcLadder returns a chain of the source value from the caller-supplied
// level down through pointer indirections to the innermost non-pointer
// value (or nil).
func srcLadder(src any) []any {
	if src == nil {
		return []any{nil}
	}
	out := []any{src}
	rv := reflect.ValueOf(src)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			out = append(out, nil)
			return out
		}
		rv = rv.Elem()
		out = append(out, rv.Interface())
	}
	return out
}

// innermostDestPtr returns a pointer to the innermost non-pointer element
// of dest. Assumes dest is a non-nil pointer; intermediates have been
// allocated via reflectutil.DerefToAddressable.
func innermostDestPtr(dest any) any {
	rv := reflect.ValueOf(dest)
	for rv.Kind() == reflect.Pointer && rv.Elem().Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv.Interface()
}

// flattenSrc walks through src pointer indirections until a non-pointer
// value (or nil) is reached.
func flattenSrc(src any) any {
	if src == nil {
		return nil
	}
	rv := reflect.ValueOf(src)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}
