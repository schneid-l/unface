# unface — Examples

Each directory below is a standalone runnable program demonstrating a different facet of the library.

| Example | What it shows | Run |
|---|---|---|
| [`url/`](./url) | One field type accepts either a shorthand string or a structured map. The struct implements `Unstringer`; the map form is handled by the default walker. | `go run ./examples/url` |
| [`httpbind/`](./httpbind) | Bind a JSON request body to a typed domain struct. Two-step flow: `encoding/json` → `map[string]any` → `unface.Unface`. Shows tag validation (`required`, `alias=`) returning HTTP 422 cleanly. | `go run ./examples/httpbind` |
| [`k8sresource/`](./k8sresource) | Polymorphic Kubernetes-style resources: `Kind` determines the concrete `Spec` type. Resource implements `Unmapper` to dispatch to the right struct. | `go run ./examples/k8sresource` |

Every example has a `main_test.go` that asserts the expected behavior — run `go test ./examples/...` to verify.
