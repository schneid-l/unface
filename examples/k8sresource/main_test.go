package main

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestDeploymentDispatch(t *testing.T) {
	raw := map[string]any{
		"kind":     "Deployment",
		"metadata": map[string]any{"name": "web"},
		"spec":     map[string]any{"image": "nginx:1.25", "replicas": "3"},
	}
	var r Resource
	if err := unface.Unface(raw, &r); err != nil {
		t.Fatal(err)
	}
	ds, ok := r.Spec.(DeploymentSpec)
	if !ok {
		t.Fatalf("Spec type=%T", r.Spec)
	}
	if ds.Image != "nginx:1.25" || ds.Replicas != 3 {
		t.Fatalf("ds=%+v", ds)
	}
	if r.Name != "web" {
		t.Fatalf("name=%q", r.Name)
	}
}

func TestServiceDispatch(t *testing.T) {
	raw := map[string]any{
		"kind":     "Service",
		"metadata": map[string]any{"name": "svc"},
		"spec":     map[string]any{"port": 8080, "selector": "app=x"},
	}
	var r Resource
	if err := unface.Unface(raw, &r); err != nil {
		t.Fatal(err)
	}
	ss, ok := r.Spec.(ServiceSpec)
	if !ok {
		t.Fatalf("Spec type=%T", r.Spec)
	}
	if ss.Port != 8080 || ss.Selector != "app=x" {
		t.Fatalf("ss=%+v", ss)
	}
}

func TestUnknownKindRejected(t *testing.T) {
	raw := map[string]any{
		"kind": "Unknown",
		"spec": map[string]any{},
	}
	var r Resource
	if err := unface.Unface(raw, &r); err == nil {
		t.Fatal("expected unknown kind error")
	}
}

func TestMissingKindRejected(t *testing.T) {
	raw := map[string]any{"spec": map[string]any{"image": "x"}}
	var r Resource
	if err := unface.Unface(raw, &r); err == nil {
		t.Fatal("expected missing kind error")
	}
}

func TestDeploymentImageRequired(t *testing.T) {
	raw := map[string]any{
		"kind": "Deployment",
		"spec": map[string]any{"replicas": 1},
	}
	var r Resource
	if err := unface.Unface(raw, &r); err == nil {
		t.Fatal("expected image-required error")
	}
}
