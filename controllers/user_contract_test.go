package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
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
