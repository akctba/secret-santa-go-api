package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty value",
			input: "",
			want:  nil,
		},
		{
			name:  "single origin",
			input: "http://localhost:3000",
			want:  []string{"http://localhost:3000"},
		},
		{
			name:  "multiple origins with spaces",
			input: "http://localhost:3000, https://example.com  ,",
			want:  []string{"http://localhost:3000", "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAllowedOrigins(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseAllowedOrigins(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCorsHandlerAllowsConfiguredOriginPreflight(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/user", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Result().Header.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("expected Access-Control-Allow-Origin header to be set to allowed origin; status=%d headers=%v", rr.Code, rr.Result().Header)
	}
}

func TestCorsHandlerRejectsUnknownOriginPreflight(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/user", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Result().Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected Access-Control-Allow-Origin header to be empty for disallowed origin")
	}
}
