// Package tag parses the `unface:"..."` field tag and the struct-wide
// marker tag (on the zero-size `_ struct{}` field). It is an internal
// implementation detail of the struct walker.
package tag

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/schneid-l/unface/plugin"
)

// FieldTag captures the parsed `unface:"..."` tag for one struct field.
type FieldTag struct {
	Name      string
	Aliases   []string
	Required  bool
	Remainder bool
	Strict    bool
	Inline    bool
	Skip      bool
	Match     *plugin.MatchMode // nil means use struct/instance default
}

// StructTag captures struct-wide options parsed from the zero-size marker
// field (`_ struct{}` with an unface tag).
type StructTag struct {
	Match       *plugin.MatchMode
	OnUnknown   *plugin.UnknownPolicy
	TagFallback []string
}

// ParseFieldTag is a helper that accepts a full reflect.StructTag string.
func ParseFieldTag(raw string) (FieldTag, error) {
	st := reflect.StructTag(raw)
	return ParseFieldTagValue(st.Get("unface"))
}

// ParseFieldTagValue parses the value of the "unface" tag (the part between
// the quotes). Returns a zero FieldTag when value is empty.
func ParseFieldTagValue(value string) (FieldTag, error) {
	var ft FieldTag
	if value == "" {
		return ft, nil
	}
	parts := strings.Split(value, ",")
	ft.Name = parts[0]
	if ft.Name == "-" {
		ft.Skip = true
		ft.Name = ""
	}
	for _, p := range parts[1:] {
		kv := strings.SplitN(p, "=", 2)
		key := kv[0]
		val := ""
		if len(kv) == 2 {
			val = kv[1]
		}
		switch key {
		case "required":
			ft.Required = true
		case "remainder":
			ft.Remainder = true
		case "strict":
			ft.Strict = true
		case "inline":
			ft.Inline = true
		case "omitempty":
			// reserved for symmetry; no-op for now.
		case "alias":
			if val != "" {
				ft.Aliases = append(ft.Aliases, val)
			}
		case "match":
			m, err := ParseMatchMode(val)
			if err != nil {
				return ft, err
			}
			ft.Match = &m
		default:
			return ft, fmt.Errorf("unface: unknown field-tag modifier %q", key)
		}
	}
	return ft, nil
}

// ParseMatchMode parses a match-mode token from the tag grammar.
func ParseMatchMode(s string) (plugin.MatchMode, error) {
	switch s {
	case "fold":
		return plugin.MatchFold, nil
	case "insensitive":
		return plugin.MatchInsensitive, nil
	case "exact", "strict":
		return plugin.MatchExact, nil
	default:
		return 0, fmt.Errorf("unface: unknown match mode %q", s)
	}
}

// ParseUnknownPolicy parses an unknown-policy token from the tag grammar.
func ParseUnknownPolicy(s string) (plugin.UnknownPolicy, error) {
	switch s {
	case "ignore":
		return plugin.UnknownIgnore, nil
	case "error":
		return plugin.UnknownError, nil
	case "warn":
		return plugin.UnknownWarn, nil
	default:
		return 0, fmt.Errorf("unface: unknown unknown-policy %q", s)
	}
}

// ReadFieldTag resolves the effective field tag, falling back across the
// provided tag names. Name is taken from the first non-empty source; all
// modifiers are only honored from the "unface" tag.
func ReadFieldTag(f reflect.StructField, fallback []string) FieldTag {
	var ft FieldTag
	if raw, ok := f.Tag.Lookup("unface"); ok {
		if parsed, err := ParseFieldTagValue(raw); err == nil {
			ft = parsed
		}
	}
	if ft.Name == "" && !ft.Skip && !ft.Remainder && !ft.Inline {
		for _, tag := range fallback {
			if tag == "unface" {
				continue
			}
			if raw, ok := f.Tag.Lookup(tag); ok {
				first := strings.SplitN(raw, ",", 2)[0]
				if first == "-" {
					ft.Skip = true
					break
				}
				if first != "" {
					ft.Name = first
					break
				}
			}
		}
	}
	if ft.Name == "" && !ft.Skip && !ft.Remainder && !ft.Inline {
		ft.Name = f.Name
	}
	return ft
}

// ReadStructTag reads the zero-size marker field (`_ struct{}`) for
// struct-wide options.
func ReadStructTag(t reflect.Type) StructTag {
	var st StructTag
	if t.Kind() != reflect.Struct {
		return st
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name != "_" || f.Type.Kind() != reflect.Struct {
			continue
		}
		raw, ok := f.Tag.Lookup("unface")
		if !ok {
			continue
		}
		parts := strings.Split(raw, ",")
		for _, p := range parts[1:] { // skip positional (always empty for marker)
			kv := strings.SplitN(p, "=", 2)
			key := kv[0]
			val := ""
			if len(kv) == 2 {
				val = kv[1]
			}
			switch key {
			case "match":
				if m, err := ParseMatchMode(val); err == nil {
					st.Match = &m
				}
			case "unknown":
				if u, err := ParseUnknownPolicy(val); err == nil {
					st.OnUnknown = &u
				}
			case "tags":
				st.TagFallback = strings.Split(val, "+")
			}
		}
		return st
	}
	return st
}
