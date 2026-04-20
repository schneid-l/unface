package tag

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface/plugin"
)

func TestParseFieldTagName(t *testing.T) {
	ft, err := ParseFieldTag(`unface:"port"`)
	if err != nil {
		t.Fatal(err)
	}
	if ft.Name != "port" {
		t.Fatalf("Name=%q", ft.Name)
	}
}

func TestParseFieldTagModifiers(t *testing.T) {
	ft, err := ParseFieldTag(`unface:"host,required,alias=hostname,alias=addr,strict"`)
	if err != nil {
		t.Fatal(err)
	}
	if !ft.Required || !ft.Strict {
		t.Fatalf("required=%v strict=%v", ft.Required, ft.Strict)
	}
	if len(ft.Aliases) != 2 || ft.Aliases[0] != "hostname" {
		t.Fatalf("aliases=%v", ft.Aliases)
	}
}

func TestParseFieldTagSkip(t *testing.T) {
	ft, err := ParseFieldTag(`unface:"-"`)
	if err != nil {
		t.Fatal(err)
	}
	if !ft.Skip {
		t.Fatal("Skip should be true")
	}
}

func TestParseFieldTagRemainder(t *testing.T) {
	ft, err := ParseFieldTag(`unface:",remainder"`)
	if err != nil {
		t.Fatal(err)
	}
	if !ft.Remainder {
		t.Fatal("Remainder should be true")
	}
}

func TestParseFieldTagInline(t *testing.T) {
	ft, _ := ParseFieldTag(`unface:",inline"`)
	if !ft.Inline {
		t.Fatal("Inline should be true")
	}
}

func TestParseFieldTagMatchOverride(t *testing.T) {
	ft, _ := ParseFieldTag(`unface:"name,match=exact"`)
	if ft.Match == nil || *ft.Match != plugin.MatchExact {
		t.Fatalf("Match=%v", ft.Match)
	}
}

func TestParseFieldTagUnknownModifier(t *testing.T) {
	if _, err := ParseFieldTag(`unface:"name,nope"`); err == nil {
		t.Fatal("unknown modifier should fail")
	}
}

func TestFieldTagFallbackTags(t *testing.T) {
	type S struct {
		A int `yaml:"alpha"`
		B int `json:"beta"`
		C int `unface:"charlie" json:"bad"`
	}
	typ := reflect.TypeOf(S{})

	ft := ReadFieldTag(typ.Field(0), []string{"unface", "yaml", "json"})
	if ft.Name != "alpha" {
		t.Fatalf("A name=%q", ft.Name)
	}

	ft = ReadFieldTag(typ.Field(1), []string{"unface", "yaml", "json"})
	if ft.Name != "beta" {
		t.Fatalf("B name=%q", ft.Name)
	}

	ft = ReadFieldTag(typ.Field(2), []string{"unface", "yaml", "json"})
	if ft.Name != "charlie" {
		t.Fatalf("C name=%q (unface wins)", ft.Name)
	}
}

func TestFieldTagDefaultToFieldName(t *testing.T) {
	type S struct {
		Port int
	}
	ft := ReadFieldTag(reflect.TypeOf(S{}).Field(0), []string{"unface"})
	if ft.Name != "Port" {
		t.Fatalf("Name=%q want Port", ft.Name)
	}
}

func TestFieldTagFallbackSkip(t *testing.T) {
	type S struct {
		X int `yaml:"-"`
	}
	ft := ReadFieldTag(reflect.TypeOf(S{}).Field(0), []string{"unface", "yaml"})
	if !ft.Skip {
		t.Fatal("yaml:- should cause Skip")
	}
}

func TestReadStructTag(t *testing.T) {
	type S struct {
		_ struct{} `unface:",match=exact,unknown=error,tags=unface+json"`
		X int
	}
	st := ReadStructTag(reflect.TypeOf(S{}))
	if st.Match == nil || *st.Match != plugin.MatchExact {
		t.Fatalf("Match=%v", st.Match)
	}
	if st.OnUnknown == nil || *st.OnUnknown != plugin.UnknownError {
		t.Fatalf("OnUnknown=%v", st.OnUnknown)
	}
	if len(st.TagFallback) != 2 || st.TagFallback[1] != "json" {
		t.Fatalf("TagFallback=%v", st.TagFallback)
	}
}

func TestReadStructTagAbsent(t *testing.T) {
	type S struct {
		X int
	}
	st := ReadStructTag(reflect.TypeOf(S{}))
	if st.Match != nil || st.OnUnknown != nil || len(st.TagFallback) != 0 {
		t.Fatalf("expected zero-valued structTag, got %+v", st)
	}
}

func TestParseMatchMode(t *testing.T) {
	for s, want := range map[string]plugin.MatchMode{
		"fold": plugin.MatchFold, "insensitive": plugin.MatchInsensitive, "exact": plugin.MatchExact, "strict": plugin.MatchExact,
	} {
		m, err := ParseMatchMode(s)
		if err != nil || m != want {
			t.Fatalf("ParseMatchMode(%q)=%v err=%v want %v", s, m, err, want)
		}
	}
	if _, err := ParseMatchMode("bogus"); err == nil {
		t.Fatal("bogus should fail")
	}
}
