# Guide: polymorphic Kubernetes-style resources

**What this page covers.** Decoding documents where the concrete type of one field depends on another field's value — the classic Kubernetes `kind` + `spec` pattern. Runnable reference: [`examples/k8sresource`](../../examples/k8sresource/main.go).

## The problem

```yaml
kind: Deployment
spec: { image: nginx, replicas: 3 }
---
kind: Service
spec: { port: 8080, selector: app=web }
```

The `spec` field is a different struct per `kind`. `encoding/json` can't express this without a second pass.

## The Unmapper pattern

Implement `Unmap(m unface.Map) error` on the outer type. The dispatcher calls it whenever the source is a map; inside, you read `kind` first and dispatch `spec` to the right sub-struct.

```go
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

type Resource struct {
    APIVersion string `unface:"apiVersion"`
    Kind       string `unface:"kind,required"`
    Name       string
    Spec       Spec `unface:"-"` // filled manually in Unmap
}

func (r *Resource) Unmap(m unface.Map) error {
    if v, ok := m.GetString("apiVersion"); ok { r.APIVersion = v }
    if v, ok := m.GetString("kind"); ok { r.Kind = v }
    if meta, ok := m.GetMap("metadata"); ok {
        if name, ok := meta.GetString("name"); ok { r.Name = name }
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
```

## Key points

- `Unmap` wins over the default struct walker whenever the source is a map.
- `Spec Spec \`unface:"-"\`` tells the walker "hands off this field" — safe even if the outer type didn't override the walker.
- The recursive call `unface.Unface(spec, &s)` reuses the full pipeline: required-field checks, coercion, nested `Un*er` methods.
- Use the `Map` accessors (`GetString`, `GetMap`, `GetList`, `GetInt64`, `Get`) instead of reaching for raw `reflect`.

## Alternative: type registry

If you have many kinds, switch from a Go-style switch to a registry:

```go
var specRegistry = map[string]func() Spec{
    "Deployment": func() Spec { return &DeploymentSpec{} },
    "Service":    func() Spec { return &ServiceSpec{} },
}

func (r *Resource) Unmap(m unface.Map) error {
    // ... read kind ...
    ctor, ok := specRegistry[r.Kind]
    if !ok { return fmt.Errorf("unknown kind %q", r.Kind) }
    s := ctor()
    raw, _ := m.Get("spec")
    if err := unface.Unface(raw, s); err != nil {
        return err
    }
    r.Spec = s
    return nil
}
```

## Running

```bash
go run ./examples/k8sresource
```
