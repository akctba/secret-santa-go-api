package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akctba/secret-santa-go-api/auth"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func setupUserContractTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	createUsersTable := `
	CREATE TABLE Users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_name TEXT,
		user_email TEXT,
		password TEXT,
		gender TEXT,
		date_of_birth TEXT
	);`

	if _, err := db.Exec(createUsersTable); err != nil {
		db.Close()
		t.Fatalf("create Users table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func withTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	originalGetDB := getDB
	getDB = func() (*sql.DB, error) {
		return db, nil
	}

	t.Cleanup(func() {
		getDB = originalGetDB
	})
}

func decodeJSONBody(t *testing.T, body string) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode json body: %v, body: %s", err, body)
	}
	return payload
}

func TestCreateUserResponseOmitsPassword(t *testing.T) {
	db := setupUserContractTestDB(t)
	withTestDB(t, db)

	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"user_name":"Alice","email":"alice@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	payload := decodeJSONBody(t, rr.Body.String())
	if _, ok := payload["password"]; ok {
		t.Fatalf("expected response to omit password, got: %s", rr.Body.String())
	}
}

func TestGetUserResponseOmitsPassword(t *testing.T) {
	db := setupUserContractTestDB(t)
	withTestDB(t, db)

	if _, err := db.Exec(`INSERT INTO Users (user_id, user_name, user_email, password) VALUES (1, 'Alice', 'alice@example.com', 'hashed-password')`); err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/user/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	payload := decodeJSONBody(t, rr.Body.String())
	if _, ok := payload["password"]; ok {
		t.Fatalf("expected response to omit password, got: %s", rr.Body.String())
	}
}

func TestSigninResponseReturnsAccessAndRefreshTokens(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	db := setupUserContractTestDB(t)
	withTestDB(t, db)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO Users (user_id, user_name, user_email, password) VALUES (1, 'Alice', 'alice@example.com', ?)`, string(hashedPassword)); err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/user/signin", strings.NewReader(`{"email":"alice@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	Signin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	payload := decodeJSONBody(t, rr.Body.String())
	accessToken, ok := payload["access_token"].(string)
	if !ok || accessToken == "" {
		t.Fatalf("expected access_token in response, got: %s", rr.Body.String())
	}

	refreshToken, ok := payload["refresh_token"].(string)
	if !ok || refreshToken == "" {
		t.Fatalf("expected refresh_token in response, got: %s", rr.Body.String())
	}

	if _, err := auth.ValidateToken(accessToken); err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if _, err := auth.ValidateRefreshToken(refreshToken); err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}
}

func TestRefreshTokenReturnsNewAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	refreshToken, err := auth.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("CreateRefreshToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/user/refresh", strings.NewReader(`{"refresh_token":"`+refreshToken+`"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	payload := decodeJSONBody(t, rr.Body.String())
	accessToken, ok := payload["access_token"].(string)
	if !ok || accessToken == "" {
		t.Fatalf("expected access_token in response, got: %s", rr.Body.String())
	}

	userID, err := auth.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if userID != 42 {
		t.Fatalf("ValidateToken returned %d, want 42", userID)
	}
}

func TestRefreshTokenRejectsAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	accessToken, err := auth.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("CreateAccessToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/user/refresh", strings.NewReader(`{"refresh_token":"`+accessToken+`"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}
