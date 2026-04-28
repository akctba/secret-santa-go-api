package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akctba/secret-santa-go-api/auth"
)

func TestBearerAuthRejectsNonBearerScheme(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := auth.CreateToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	handlerCalled := false
	handler := BearerAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+token)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if handlerCalled {
		t.Fatal("expected next handler not to be called")
	}

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}