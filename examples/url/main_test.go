package main

import "testing"

func TestURLFromString(t *testing.T) {
	cfg, err := Load(map[string]any{
		"name": "demo",
		"api":  "https://api.example.com:8080/v1",
		"web":  "http://example.com/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.API.Scheme != "https" || cfg.API.Host != "api.example.com" ||
		cfg.API.Port != 8080 || cfg.API.Path != "/v1" {
		t.Fatalf("API=%+v", cfg.API)
	}
	if cfg.Web.Scheme != "http" || cfg.Web.Host != "example.com" ||
		cfg.Web.Path != "/" {
		t.Fatalf("Web=%+v", cfg.Web)
	}
}

func TestURLFromMap(t *testing.T) {
	cfg, err := Load(map[string]any{
		"name": "demo",
		"api":  map[string]any{"scheme": "https", "host": "api.example.com", "port": 8080, "path": "/v1"},
		"web":  map[string]any{"scheme": "https", "host": "example.com", "path": "/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.API.Host != "api.example.com" || cfg.API.Port != 8080 {
		t.Fatalf("API=%+v", cfg.API)
	}
	if cfg.Web.Host != "example.com" {
		t.Fatalf("Web=%+v", cfg.Web)
	}
}

func TestURLHostIsRequired(t *testing.T) {
	_, err := Load(map[string]any{
		"name": "demo",
		"api":  map[string]any{"scheme": "https"},
		"web":  "http://x/",
	})
	if err == nil {
		t.Fatal("expected required-host error")
	}
}
