// Command k8sresource shows how to decode Kubernetes-style resources
// where the Spec field's concrete type depends on the Kind. Each Kind
// implements Unmapper to dispatch the spec to the right struct.
//
// Run with:
//
//	go run ./examples/k8sresource
package main

import (
	"fmt"
	"log"

	"github.com/schneid-l/unface"
)

// Spec is the polymorphic body. Each concrete Kind defines its own type.
type Spec interface{ kind() string }

type DeploymentSpec struct {
	Image    string `unface:"image,required"`
	Replicas int    `unface:"replicas"`
}

func (DeploymentSpec) kind() string { return "Deployment" }

type ServiceSpec struct {
	Port     int    `unface:"port,required"`
	Selector string `unface:"selector"`
}

func (ServiceSpec) kind() string { return "Service" }

// Resource wraps metadata + a polymorphic Spec. It implements Unmapper so
// that after looking at Kind, it picks the right Spec type.
type Resource struct {
	APIVersion string `unface:"apiVersion"`
	Kind       string `unface:"kind,required"`
	Name       string `unface:"metadata.name"` // flat alias omitted for clarity
	Spec       Spec   `unface:"-"`             // filled manually by Unmap
}

func (r *Resource) Unmap(m unface.Map) error {
	// First, populate the plain fields using the default walker path via
	// a throwaway anonymous struct that mirrors Resource's non-Spec fields.
	if v, ok := m.GetString("apiVersion"); ok {
		r.APIVersion = v
	}
	if v, ok := m.GetString("kind"); ok {
		r.Kind = v
	}
	if meta, ok := m.GetMap("metadata"); ok {
		if name, ok := meta.GetString("name"); ok {
			r.Name = name
		}
	}
	if r.Kind == "" {
		return fmt.Errorf("resource: kind is required")
	}

	spec, _ := m.Get("spec")
	switch r.Kind {
	case "Deployment":
		var s DeploymentSpec
		if err := unface.Unface(spec, &s); err != nil {
			return fmt.Errorf("deployment spec: %w", err)
		}
		r.Spec = s
	case "Service":
		var s ServiceSpec
		if err := unface.Unface(spec, &s); err != nil {
			return fmt.Errorf("service spec: %w", err)
		}
		r.Spec = s
	default:
		return fmt.Errorf("unknown kind %q", r.Kind)
	}
	return nil
}

func main() {
	docs := []map[string]any{
		{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]any{"name": "web"},
			"spec": map[string]any{
				"image":    "nginx:1.25",
				"replicas": "3",
			},
		},
		{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata":   map[string]any{"name": "web-svc"},
			"spec": map[string]any{
				"port":     "8080",
				"selector": "app=web",
			},
		},
	}
	for _, raw := range docs {
		var r Resource
		if err := unface.Unface(raw, &r); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-10s %-10s %+v\n", r.Kind, r.Name, r.Spec)
	}
}
