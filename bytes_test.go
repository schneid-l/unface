package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestBytesPluginFromString(t *testing.T) {
	var b []byte
	f := unface.New(unface.With(unface.BytesPlugin))
	if err := f.Unface("hi", &b); err != nil {
		t.Fatal(err)
	}
	if string(b) != "hi" {
		t.Fatalf("b=%q", b)
	}
}

func TestBytesPluginFromBytes(t *testing.T) {
	var b []byte
	f := unface.New(unface.With(unface.BytesPlugin))
	if err := f.Unface([]byte{1, 2, 3}, &b); err != nil {
		t.Fatal(err)
	}
	if len(b) != 3 || b[1] != 2 {
		t.Fatalf("b=%v", b)
	}
}
