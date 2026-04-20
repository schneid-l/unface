// Command httpbind shows how to bind a JSON request body to a typed
// domain object using unface. The two-step flow is:
//
//  1. Use encoding/json to decode the body into map[string]any.
//  2. Use unface.Unface(raw, &dst) to coerce into your typed struct
//     — allowing shorthand strings, type coercion (e.g. "8080" → int),
//     and tag-driven validation (`required`, aliases, etc.).
//
// Run with:
//
//	go run ./examples/httpbind
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/schneid-l/unface"
)

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
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"order":  order,
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", createOrderHandler)

	// Demo in-process: fire two requests against a test server.
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// 1. Normal payload — Total is sent as string but coerced to float64.
	body := `{"customer":"alice","items":["apples","bread"],"total":"4.99","ccy":"USD"}`
	res, _ := http.Post(srv.URL+"/orders", "application/json", strings.NewReader(body))
	b, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	fmt.Printf("POST ok  → %d %s", res.StatusCode, string(b))

	// 2. Missing required field.
	res, _ = http.Post(srv.URL+"/orders", "application/json",
		strings.NewReader(`{"customer":"bob"}`))
	b, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	fmt.Printf("POST 422 → %d %s\n", res.StatusCode, string(b))
	log.SetFlags(0)
}
