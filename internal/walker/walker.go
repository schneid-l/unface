// Package walker walks a map-like src into a struct destination,
// honoring the tag grammar and the instance/struct-level match and
// unknown policies. It is invoked by StructPlugin but also exposed so
// codegen can call into it directly.
package walker

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/schneid-l/unface/internal/strcase"
	"github.com/schneid-l/unface/internal/tag"
	"github.com/schneid-l/unface/plugin"
)

// DispatchFunc is the callback the walker uses to coerce a single src into
// a dest pointer. The closure must respect the caller's effective Config
// (match mode, unknown policy, tag fallback, plugin set). In the typical
// case this is a method value over *engine.Facer: `f.Unface`.
type DispatchFunc func(src, dest any) error

// fieldInfo describes one destination field after tag parsing.
type fieldInfo struct {
	index    []int // reflect.Value.FieldByIndex path
	tag      tag.FieldTag
	name     string // effective source name
	fieldTyp reflect.Type
}

// isIndexable reports whether this field participates in the byName lookup
// table. Skipped, remainder, and inline fields are excluded because they
// either have no name (skipped) or don't map to a single source key
// (remainder absorbs unknowns; inline contributes its own children).
func (fi fieldInfo) isIndexable() bool {
	return !fi.tag.Skip && !fi.tag.Remainder && !fi.tag.Inline
}

type structPlan struct {
	fields    []fieldInfo
	byName    map[string]int // normalized name -> fields index
	remainder int            // -1 if none
}

// buildStructPlan indexes the struct type and returns the effective match
// mode, unknown policy, and tag-fallback order. No cache for v0.1.
func buildStructPlan(t reflect.Type, cfg *plugin.Config) (*structPlan, plugin.MatchMode, plugin.UnknownPolicy, []string) {
	sp := &structPlan{byName: map[string]int{}, remainder: -1}
	st := tag.ReadStructTag(t)

	match := cfg.Match
	if st.Match != nil {
		match = *st.Match
	}
	unknown := cfg.OnUnknown
	if st.OnUnknown != nil {
		unknown = *st.OnUnknown
	}
	tagFallback := cfg.TagFallback
	if len(st.TagFallback) > 0 {
		tagFallback = st.TagFallback
	}

	collectFields(t, nil, tagFallback, sp)

	for idx, fi := range sp.fields {
		if !fi.isIndexable() {
			continue
		}
		keys := append([]string{fi.name}, fi.tag.Aliases...)
		for _, k := range keys {
			norm := normalizeKey(k, match)
			sp.byName[norm] = idx
		}
	}
	return sp, match, unknown, tagFallback
}

// collectFields walks the struct type, producing fieldInfos. Anonymous
// (embedded) struct fields and inline-tagged fields are walked into the
// parent's namespace.
func collectFields(t reflect.Type, prefix []int, fallback []string, sp *structPlan) {
	for i := range t.NumField() {
		f := t.Field(i)
		if f.Name == "_" {
			continue
		}
		idx := append(append([]int{}, prefix...), i)

		ft := tag.ReadFieldTag(f, fallback)

		if ft.Skip {
			continue
		}

		// Anonymous without explicit unface name: inline.
		if f.Anonymous && ft.Name == f.Name {
			inner := f.Type
			if inner.Kind() == reflect.Pointer {
				inner = inner.Elem()
			}
			if inner.Kind() == reflect.Struct {
				collectFields(inner, idx, fallback, sp)
				continue
			}
		}
		if ft.Inline {
			inner := f.Type
			if inner.Kind() == reflect.Pointer {
				inner = inner.Elem()
			}
			if inner.Kind() == reflect.Struct {
				collectFields(inner, idx, fallback, sp)
				continue
			}
		}

		fi := fieldInfo{
			index:    idx,
			tag:      ft,
			name:     ft.Name,
			fieldTyp: f.Type,
		}
		sp.fields = append(sp.fields, fi)
		if ft.Remainder {
			sp.remainder = len(sp.fields) - 1
		}
	}
}

// normalizeKey folds a source or field name according to the active match
// mode so that lookup is a single map get.
func normalizeKey(s string, mode plugin.MatchMode) string {
	switch mode {
	case plugin.MatchExact:
		return s
	case plugin.MatchInsensitive:
		return strings.ToLower(s)
	default: // MatchFold
		return strcase.Normalize(s)
	}
}

// StructWalk walks the src Map into destVal (a settable struct
// reflect.Value). The dispatch closure is the caller's effective
// Facer.Unface; a second "strict" closure is used for fields tagged with
// the `strict` modifier (bypasses lenient plugin coercions).
//
// Most callers should use the public API (Facer.Unface with StructPlugin
// registered). This entry point exists so codegen can emit direct calls
// against generated code.
func StructWalk(src any, destVal reflect.Value, cfg *plugin.Config, lenient, strict DispatchFunc) error {
	m, ok := plugin.MapOf(src)
	if !ok {
		return fmt.Errorf("unface/struct: cannot walk %T into %s", src, destVal.Type())
	}
	sp, match, unknown, _ := buildStructPlan(destVal.Type(), cfg)

	seen := make(map[int]bool, len(sp.fields))
	unmatched := map[string]any{}

	if err := walkMapFields(m, match, sp, destVal, lenient, strict, seen, unmatched); err != nil {
		return err
	}

	for idx, fi := range sp.fields {
		if fi.tag.Required && !seen[idx] {
			return fmt.Errorf("%w: %s", plugin.ErrRequired, fi.name)
		}
	}

	if sp.remainder >= 0 && len(unmatched) > 0 {
		fi := sp.fields[sp.remainder]
		fv := destVal.FieldByIndex(fi.index)
		if err := lenient(unmatched, fv.Addr().Interface()); err != nil {
			return err
		}
		unmatched = nil
	}

	return applyUnknownPolicy(unmatched, unknown, cfg)
}

// walkMapFields iterates the source map once, dispatching each matched key
// to the corresponding struct field. Unmatched keys are collected into
// `unmatched` for later remainder-absorption or policy application.
func walkMapFields(
	m plugin.Map,
	match plugin.MatchMode,
	sp *structPlan,
	destVal reflect.Value,
	lenient, strict DispatchFunc,
	seen map[int]bool,
	unmatched map[string]any,
) error {
	var walkErr error
	m.Iter(func(k, v any) bool {
		ks, ok := k.(string)
		if !ok {
			ks = fmt.Sprint(k)
		}
		norm := normalizeKey(ks, match)
		idx, ok := sp.byName[norm]
		if !ok {
			unmatched[ks] = v
			return true
		}
		seen[idx] = true
		fi := sp.fields[idx]
		fv := destVal.FieldByIndex(fi.index)

		dispatch := lenient
		if fi.tag.Strict {
			dispatch = strict
		}
		if err := dispatch(v, fv.Addr().Interface()); err != nil {
			walkErr = wrapFieldError(fi.name, v, err)
			return false
		}
		return true
	})
	return walkErr
}

// applyUnknownPolicy enforces cfg.OnUnknown once the walker has exhausted
// the remainder-absorbing field (if any).
func applyUnknownPolicy(unmatched map[string]any, policy plugin.UnknownPolicy, cfg *plugin.Config) error {
	if len(unmatched) == 0 {
		return nil
	}
	switch policy {
	case plugin.UnknownIgnore:
		return nil
	case plugin.UnknownError:
		for k := range unmatched {
			return fmt.Errorf("%w: %s", plugin.ErrUnknownField, k)
		}
	case plugin.UnknownWarn:
		if cfg.UnknownHandler != nil {
			for k, v := range unmatched {
				cfg.UnknownHandler(k, v)
			}
		}
	}
	return nil
}

// wrapFieldError adds the given field name to the error's path trace. If
// the wrapped error is already a *plugin.Error, the path is prepended;
// otherwise a fresh *plugin.Error is minted.
func wrapFieldError(name string, src any, err error) error {
	var existing *plugin.Error
	if errors.As(err, &existing) {
		return existing.WithPath(name)
	}
	return &plugin.Error{Path: []string{name}, Src: src, Err: err}
}
