package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestBoolPluginFromBool(t *testing.T) {
	var b bool
	f := unface.New(unface.With(unface.BoolPlugin))
	if err := f.Unface(true, &b); err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("b should be true")
	}
}

func TestBoolPluginFromString(t *testing.T) {
	cases := map[string]bool{
		"true": true, "TRUE": true, "yes": true, "y": true, "on": true, "1": true, "enabled": true,
		"false": false, "FALSE": false, "no": false, "n": false, "off": false, "0": false, "disabled": false,
	}
	f := unface.New(unface.With(unface.BoolPlugin))
	for in, want := range cases {
		var b bool
		if err := f.Unface(in, &b); err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if b != want {
			t.Fatalf("%q got %v want %v", in, b, want)
		}
	}
}

func TestBoolPluginStringInvalid(t *testing.T) {
	var b bool
	f := unface.New(unface.With(unface.BoolPlugin))
	err := f.Unface("maybe", &b)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBoolPluginFromNumber(t *testing.T) {
	var b bool
	f := unface.New(unface.With(unface.BoolPlugin))
	if err := f.Unface(1, &b); err != nil || !b {
		t.Fatalf("got b=%v err=%v", b, err)
	}
	var b2 bool
	if err := f.Unface(0, &b2); err != nil || b2 {
		t.Fatalf("got b2=%v err=%v", b2, err)
	}
}
