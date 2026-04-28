package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestSigninReturnsBadRequestForMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/signin", strings.NewReader(`{"email":"alice@example.com",`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Signin(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Fatalf("expected invalid body error, got: %s", rr.Body.String())
	}
}

func TestSigninReturnsBadRequestForMissingFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/signin", strings.NewReader(`{"email":"","password":""}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Signin(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "email and password are required") {
		t.Fatalf("expected missing fields error, got: %s", rr.Body.String())
	}
}

func TestRefreshTokenReturnsBadRequestForMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/refresh", strings.NewReader(`{"refresh_token":"abc"`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Fatalf("expected invalid body error, got: %s", rr.Body.String())
	}
}

func TestRefreshTokenReturnsBadRequestForMissingField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/refresh", strings.NewReader(`{"refresh_token":""}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "refresh_token is required") {
		t.Fatalf("expected missing field error, got: %s", rr.Body.String())
	}
}

func TestCreateGroupReturnsBadRequestForMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/group", strings.NewReader(`{"name":"xmas"`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	CreateGroup(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Fatalf("expected invalid body error, got: %s", rr.Body.String())
	}
}

func TestCreateUserReturnsBadRequestForMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"user_name":"alice",`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Fatalf("expected invalid body error, got: %s", rr.Body.String())
	}
}

func TestAddParticipantReturnsBadRequestForMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/group/1/participant", strings.NewReader(`{"group_id":"1","user_id":`))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	AddParticipant(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Fatalf("expected invalid body error, got: %s", rr.Body.String())
	}
}

func TestCreateGroupReturnsBadRequestForUnknownField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/group", strings.NewReader(`{"name":"xmas","unexpected":true}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	CreateGroup(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid request body") {
		t.Fatalf("expected invalid body error, got: %s", rr.Body.String())
	}
}

func TestCreateGroupReturnsBadRequestWhenGroupIDProvided(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/group", strings.NewReader(`{"group_id":"123","name":"xmas"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	CreateGroup(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "group_id must not be provided") {
		t.Fatalf("expected group_id validation error, got: %s", rr.Body.String())
	}
}

func TestAddParticipantReturnsBadRequestForMissingFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/group/1/participant", strings.NewReader(`{"group_id":"","user_id":0}`))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	AddParticipant(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "group_id and user_id are required") {
		t.Fatalf("expected missing fields error, got: %s", rr.Body.String())
	}
}
