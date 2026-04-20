# Guide: HTTP body binding

**What this page covers.** Binding a JSON request body to a typed struct, with required-field validation, aliases, and type coercion. Runnable reference: [`examples/httpbind`](../../examples/httpbind/main.go).

## The problem

`encoding/json` deserializes cleanly only when the wire format matches your Go types exactly. Clients in the wild send:

- Numeric fields as strings (`"total": "4.99"`).
- Alias names (`"ccy"` instead of `"currency"`).
- Missing fields you need to reject with 422.

You want `encoding/json` for the JSON parse and unface for the coercion-plus-validation step.

## The two-step pattern

1. Decode JSON into `any` (i.e. `map[string]any` / `[]any` / primitives).
2. Use `unface.Unface(raw, &dst)` to coerce and validate against your typed struct.

```go
type CreateOrder struct {
    Customer string   `unface:"customer,required"`
    Items    []string `unface:"items,required"`
    Total    float64  `unface:"total"`
    Currency string   `unface:"currency,alias=ccy"`
}

func bindJSON(r *http.Request, dst any) error {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return err
    }
    defer func() { _ = r.Body.Close() }()

    var raw any
    if err := json.Unmarshal(body, &raw); err != nil {
        return fmt.Errorf("invalid json: %w", err)
    }
    if err := unface.Unface(raw, dst); err != nil {
        return fmt.Errorf("bind: %w", err)
    }
    return nil
}
```

## Mapping status codes to error kinds

```go
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
    var order CreateOrder
    if err := bindJSON(r, &order); err != nil {
        status := http.StatusBadRequest
        if errors.Is(err, unface.ErrRequired) {
            status = http.StatusUnprocessableEntity
        }
        http.Error(w, err.Error(), status)
        return
    }
    _ = json.NewEncoder(w).Encode(order)
}
```

- `ErrRequired` → 422 (valid JSON, missing business field).
- Anything else → 400 (malformed JSON, unparseable number, ...).

## What happens with `"total": "4.99"`

1. JSON parse yields `string`.
2. Struct walker routes `"total"` to the `Total float64` field.
3. Float plugin's adapter sees a string source and calls `strconv.ParseFloat`.
4. Result: `Total == 4.99`, no error.

If the client sent `"total": "not-a-number"`, the adapter returns a hard error wrapped in `*unface.Error` with path `total`.

## Aliases

`alias=ccy` on `Currency` means the walker accepts either `"currency"` or `"ccy"` from the source map. Repeat the modifier for multiple aliases: `unface:"currency,alias=ccy,alias=curr"`.

## Strictness knobs

- Reject unknown body keys with `unface.OnUnknown(unface.UnknownError)` — pairs well with a 400.
- Catch typos in your own tags in CI using `go vet` or a custom linter; `unface` itself treats unknown tag modifiers as soft failures.
- Need byte-exact key matching (no fold)? `unface.WithFieldMatch(unface.MatchExact)` globally, or `match=exact` on the struct marker.

## Running the example

```bash
go run ./examples/httpbind
```

Output:

```
POST ok  → 200 {"customer":"alice", ...}
POST 422 → 422 bind: ...customer: unface: required field missing
```
