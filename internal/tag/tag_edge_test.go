package tag

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface/plugin"
)

func TestParseFieldTagMatchInvalid(t *testing.T) {
	if _, err := ParseFieldTag(`unface:"n,match=bogus"`); err == nil {
		t.Fatal("bogus match should fail")
	}
}

func TestParseFieldTagEmptyValue(t *testing.T) {
	ft, err := ParseFieldTag(``)
	if err != nil {
		t.Fatal(err)
	}
	if ft.Name != "" || ft.Skip {
		t.Fatalf("ft=%+v", ft)
	}
}

func TestParseFieldTagAllModifiers(t *testing.T) {
	raw := `unface:"name,required,alias=a,alias=b,strict,inline,omitempty,match=exact"`
	ft, err := ParseFieldTag(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ft.Name != "name" || !ft.Required || !ft.Strict || !ft.Inline {
		t.Fatalf("ft=%+v", ft)
	}
	if len(ft.Aliases) != 2 {
		t.Fatalf("aliases=%v", ft.Aliases)
	}
	if ft.Match == nil || *ft.Match != plugin.MatchExact {
		t.Fatalf("match=%v", ft.Match)
	}
}

func TestParseFieldTagAliasEmpty(t *testing.T) {
	// alias= with empty value is silently dropped.
	ft, err := ParseFieldTag(`unface:"n,alias="`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ft.Aliases) != 0 {
		t.Fatalf("aliases=%v", ft.Aliases)
	}
}

func TestParseUnknownPolicyAll(t *testing.T) {
	for s, want := range map[string]plugin.UnknownPolicy{
		"ignore": plugin.UnknownIgnore,
		"error":  plugin.UnknownError,
		"warn":   plugin.UnknownWarn,
	} {
		got, err := ParseUnknownPolicy(s)
		if err != nil || got != want {
			t.Errorf("ParseUnknownPolicy(%q)=%v err=%v", s, got, err)
		}
	}
	if _, err := ParseUnknownPolicy("bogus"); err == nil {
		t.Fatal("bogus must fail")
	}
}

func TestReadStructTagNonStruct(t *testing.T) {
	st := ReadStructTag(reflect.TypeOf(42))
	if st.Match != nil || st.OnUnknown != nil || len(st.TagFallback) != 0 {
		t.Fatalf("non-struct returned %+v", st)
	}
}

func TestReadStructTagNoMarker(t *testing.T) {
	type T struct{ A int }
	st := ReadStructTag(reflect.TypeOf(T{}))
	if st.Match != nil {
		t.Fatal("no-marker should not populate Match")
	}
}

func TestReadStructTagInvalidMatchIgnored(t *testing.T) {
	type T struct {
		_ struct{} `unface:",match=bogus"`
		A int
	}
	st := ReadStructTag(reflect.TypeOf(T{}))
	if st.Match != nil {
		t.Fatal("bogus match should not populate Match")
	}
}

func TestReadStructTagInvalidUnknownIgnored(t *testing.T) {
	type T struct {
		_ struct{} `unface:",unknown=bogus"`
		A int
	}
	st := ReadStructTag(reflect.TypeOf(T{}))
	if st.OnUnknown != nil {
		t.Fatal("bogus unknown should not populate OnUnknown")
	}
}

func TestReadStructTagWarnPolicy(t *testing.T) {
	type T struct {
		_ struct{} `unface:",unknown=warn"`
		A int
	}
	st := ReadStructTag(reflect.TypeOf(T{}))
	if st.OnUnknown == nil || *st.OnUnknown != plugin.UnknownWarn {
		t.Fatalf("OnUnknown=%v", st.OnUnknown)
	}
}

func TestReadFieldTagBadUnfaceFallsBack(t *testing.T) {
	// When the unface tag has an invalid modifier, the parser error is
	// swallowed and we fall back to fallback tags.
	type T struct {
		A int `unface:"a,bogus" yaml:"alpha"`
	}
	ft := ReadFieldTag(reflect.TypeOf(T{}).Field(0), []string{"unface", "yaml"})
	if ft.Name != "alpha" {
		t.Fatalf("expected yaml fallback, got %q", ft.Name)
	}
}

func TestReadFieldTagFallbackEmpty(t *testing.T) {
	// yaml:"" is treated as "no name in fallback"; falls through to field name.
	type T struct {
		Port int `yaml:""`
	}
	ft := ReadFieldTag(reflect.TypeOf(T{}).Field(0), []string{"unface", "yaml"})
	if ft.Name != "Port" {
		t.Fatalf("Name=%q", ft.Name)
	}
}

func TestParseFieldTagSkipDash(t *testing.T) {
	ft, err := ParseFieldTag(`unface:"-"`)
	if err != nil {
		t.Fatal(err)
	}
	if !ft.Skip || ft.Name != "" {
		t.Fatalf("ft=%+v", ft)
	}
}

func TestParseMatchModeAllVariants(t *testing.T) {
	for s, want := range map[string]plugin.MatchMode{
		"fold":        plugin.MatchFold,
		"insensitive": plugin.MatchInsensitive,
		"exact":       plugin.MatchExact,
		"strict":      plugin.MatchExact, // alias
	} {
		got, err := ParseMatchMode(s)
		if err != nil || got != want {
			t.Errorf("ParseMatchMode(%q)=%v err=%v", s, got, err)
		}
	}
}
