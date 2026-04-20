package strcase_test

import (
	"testing"

	"github.com/schneid-l/unface/internal/strcase"
)

func TestFold(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Port", "port", true},
		{"HTTPPort", "http_port", true},
		{"HTTPPort", "http-port", true},
		{"HTTPPort", "httpPort", true},
		{"HTTPPort", "HTTP_PORT", true},
		{"userID", "user_id", true},
		{"userID", "USERID", true},
		{"APIv2", "api_v2", true},
		{"Port", "hort", false},
		{"Host", "HostName", false},
		{"", "", true},
	}
	for _, tc := range cases {
		if got := strcase.Fold(tc.a, tc.b); got != tc.want {
			t.Errorf("Fold(%q,%q)=%v want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"Port":      "port",
		"HTTPPort":  "httpport",
		"http_port": "httpport",
		"http-port": "httpport",
		"httpPort":  "httpport",
		"user_id":   "userid",
		"APIv2":     "apiv2",
		"":          "",
	}
	for in, want := range cases {
		if got := strcase.Normalize(in); got != want {
			t.Errorf("Normalize(%q)=%q want %q", in, got, want)
		}
	}
}

func TestEqualFold(t *testing.T) {
	if !strcase.EqualFold("Port", "port") {
		t.Fatal("EqualFold case-insensitive")
	}
	if strcase.EqualFold("http_port", "httpport") {
		t.Fatal("EqualFold must not fold separators")
	}
}
