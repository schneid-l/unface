package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestRunePluginFromRune(t *testing.T) {
	var r rune
	f := unface.New(unface.With(unface.RunePlugin))
	if err := f.Unface('A', &r); err != nil {
		t.Fatal(err)
	}
	if r != 'A' {
		t.Fatalf("r=%q", r)
	}
}

func TestRunePluginFromSingleCharString(t *testing.T) {
	var r rune
	f := unface.New(unface.With(unface.RunePlugin))
	if err := f.Unface("Z", &r); err != nil {
		t.Fatal(err)
	}
	if r != 'Z' {
		t.Fatalf("r=%q", r)
	}
}

func TestRunePluginFromNumber(t *testing.T) {
	var r rune
	f := unface.New(unface.With(unface.RunePlugin))
	if err := f.Unface(65, &r); err != nil {
		t.Fatal(err)
	}
	if r != 'A' {
		t.Fatalf("r=%q", r)
	}
}
