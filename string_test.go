package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestStringPluginAcceptsString(t *testing.T) {
	var s string
	f := unface.New(unface.With(unface.StringPlugin))
	if err := f.Unface("hello", &s); err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Fatalf("s=%q", s)
	}
}

func TestStringPluginAcceptsNumber(t *testing.T) {
	var s string
	f := unface.New(unface.With(unface.StringPlugin))
	if err := f.Unface(42, &s); err != nil {
		t.Fatal(err)
	}
	if s != "42" {
		t.Fatalf("s=%q", s)
	}
}

func TestStringPluginAcceptsBool(t *testing.T) {
	var s string
	f := unface.New(unface.With(unface.StringPlugin))
	if err := f.Unface(true, &s); err != nil {
		t.Fatal(err)
	}
	if s != "true" {
		t.Fatalf("s=%q", s)
	}
}

func TestStringPluginBytes(t *testing.T) {
	var s string
	f := unface.New(unface.With(unface.StringPlugin))
	if err := f.Unface([]byte("hi"), &s); err != nil {
		t.Fatal(err)
	}
	if s != "hi" {
		t.Fatalf("s=%q", s)
	}
}
