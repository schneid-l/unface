package unface_test

import (
	"encoding/json"
	"testing"

	"github.com/schneid-l/unface"
)

func TestJSONRawFromMap(t *testing.T) {
	var raw json.RawMessage
	f := unface.New(unface.With(unface.JSONRawPlugin))
	if err := f.Unface(map[string]any{"a": 1}, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"a":1}` {
		t.Fatalf("raw=%s", raw)
	}
}

func TestJSONRawFromString(t *testing.T) {
	var raw json.RawMessage
	f := unface.New(unface.With(unface.JSONRawPlugin))
	if err := f.Unface("hello", &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw) != `"hello"` {
		t.Fatalf("raw=%s", raw)
	}
}

func TestJSONRawFromRaw(t *testing.T) {
	var raw json.RawMessage
	in := json.RawMessage(`{"k":"v"}`)
	f := unface.New(unface.With(unface.JSONRawPlugin))
	if err := f.Unface(in, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"k":"v"}` {
		t.Fatalf("raw=%s", raw)
	}
}

func TestJSONRawFromNil(t *testing.T) {
	var raw json.RawMessage
	f := unface.New(unface.With(unface.JSONRawPlugin))
	if err := f.Unface(nil, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw) != "null" {
		t.Fatalf("raw=%s", raw)
	}
}
