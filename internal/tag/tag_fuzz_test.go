package tag

import "testing"

// FuzzParseFieldTagValue drives the internal parser directly so any panic
// in modifier handling surfaces immediately.
func FuzzParseFieldTagValue(f *testing.F) {
	for _, s := range []string{
		"", "name", "name,required", "-", ",remainder",
		"n,alias=a", "n,alias=a,alias=b",
		"n,match=exact", "n,match=fold", "n,match=insensitive",
		"n,bogus",
		",,,,", "=", "alias=", "match=", "n,alias=,required",
	} {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		_, _ = ParseFieldTagValue(s)
	})
}

// FuzzParseMatchMode ensures no panic on arbitrary match-mode strings.
func FuzzParseMatchMode(f *testing.F) {
	for _, s := range []string{
		"fold", "insensitive", "exact", "strict",
		"", "bogus", "FOLD",
	} {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		_, _ = ParseMatchMode(s)
	})
}

// FuzzParseUnknownPolicy ensures no panic on arbitrary unknown-policy strings.
func FuzzParseUnknownPolicy(f *testing.F) {
	for _, s := range []string{"ignore", "error", "warn", "", "bogus"} {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		_, _ = ParseUnknownPolicy(s)
	})
}
