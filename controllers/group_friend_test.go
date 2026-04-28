package controllers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/akctba/secret-santa-go-api/auth"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func setupGroupFriendTestDB(t *testing.T) *sql.DB {
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

	createParticipantsTable := `
	CREATE TABLE Participants (
		group_id INTEGER,
		user_id INTEGER,
		joined_at TEXT,
		friend_user_id INTEGER,
		PRIMARY KEY (group_id, user_id)
	);`
	if _, err := db.Exec(createParticipantsTable); err != nil {
		db.Close()
		t.Fatalf("create Participants table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestGetSecretFriendReturnsAssignedFriend(t *testing.T) {
	db := setupGroupFriendTestDB(t)
	withTestDB(t, db)

	_, err := db.Exec(`INSERT INTO Users (user_id, user_name, user_email, password) VALUES
		(1, 'Alice', 'alice@example.com', 'secret'),
		(2, 'Bob', 'bob@example.com', 'secret')`)
	if err != nil {
		t.Fatalf("insert users: %v", err)
	}

	_, err = db.Exec(`INSERT INTO Participants (group_id, user_id, joined_at, friend_user_id) VALUES (?, ?, ?, ?)`, 1, 1, time.Now().UTC().Format(time.RFC3339), 2)
	if err != nil {
		t.Fatalf("insert participant assignment: %v", err)
	}

	token, err := auth.CreateToken(1)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/group/1/friend", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	BearerAuth(GetSecretFriend)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	payload := decodeJSONBody(t, rr.Body.String())
	if payload["user_name"] != "Bob" {
		t.Fatalf("expected friend user_name Bob, got payload: %s", rr.Body.String())
	}
	if _, ok := payload["password"]; ok {
		t.Fatalf("expected response to omit password, got payload: %s", rr.Body.String())
	}
}

func TestGetSecretFriendReturnsForbiddenForNonParticipant(t *testing.T) {
	db := setupGroupFriendTestDB(t)
	withTestDB(t, db)

	_, err := db.Exec(`INSERT INTO Users (user_id, user_name, user_email, password) VALUES (1, 'Alice', 'alice@example.com', 'secret')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	token, err := auth.CreateToken(1)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/group/1/friend", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	BearerAuth(GetSecretFriend)(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusForbidden, rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "User is not a participant of this group") {
		t.Fatalf("expected not-participant error, got: %s", rr.Body.String())
	}
}

func TestGetSecretFriendReturnsConflictWhenNotDrawn(t *testing.T) {
	db := setupGroupFriendTestDB(t)
	withTestDB(t, db)

	_, err := db.Exec(`INSERT INTO Users (user_id, user_name, user_email, password) VALUES (1, 'Alice', 'alice@example.com', 'secret')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO Participants (group_id, user_id, joined_at) VALUES (?, ?, ?)`, 1, 1, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert participant without draw: %v", err)
	}

	token, err := auth.CreateToken(1)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/group/1/friend", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	BearerAuth(GetSecretFriend)(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusConflict, rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Secret friend has not been drawn yet") {
		t.Fatalf("expected not-drawn error, got: %s", rr.Body.String())
	}
}
