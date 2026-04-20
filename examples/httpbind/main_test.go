package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBindJSONSuccess(t *testing.T) {
	req := httptest.NewRequest("POST", "/orders",
		strings.NewReader(`{"customer":"alice","items":["apples"],"total":"4.99","ccy":"USD"}`))
	var order CreateOrder
	if err := bindJSON(req, &order); err != nil {
		t.Fatal(err)
	}
	if order.Customer != "alice" || order.Currency != "USD" || order.Total != 4.99 {
		t.Fatalf("order=%+v", order)
	}
	if len(order.Items) != 1 || order.Items[0] != "apples" {
		t.Fatalf("items=%v", order.Items)
	}
}

func TestBindJSONMissingRequired(t *testing.T) {
	req := httptest.NewRequest("POST", "/orders",
		strings.NewReader(`{"customer":"bob"}`))
	var order CreateOrder
	err := bindJSON(req, &order)
	if err == nil {
		t.Fatal("expected required-field error for items")
	}
}

func TestBindJSONInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/orders",
		strings.NewReader(`{not json`))
	var order CreateOrder
	err := bindJSON(req, &order)
	if err == nil {
		t.Fatal("expected json error")
	}
}

func TestHandlerReturns422ForMissingRequired(t *testing.T) {
	req := httptest.NewRequest("POST", "/orders",
		strings.NewReader(`{"customer":"bob"}`))
	w := httptest.NewRecorder()
	createOrderHandler(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code=%d", w.Code)
	}
}
