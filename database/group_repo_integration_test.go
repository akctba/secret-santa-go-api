package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/akctba/secret-santa-go-api/models"
	_ "github.com/mattn/go-sqlite3"
)

func newGroupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	createStmt := `
	CREATE TABLE Groups (
		group_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		date_created DATETIME,
		date_draw DATETIME,
		creator_user_id INTEGER
	);`

	if _, err = db.Exec(createStmt); err != nil {
		db.Close()
		t.Fatalf("create Groups table: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestGroupRepoCRUD(t *testing.T) {
	db := newGroupTestDB(t)

	now := time.Now().UTC().Truncate(time.Second)
	group := models.Group{
		GroupID:       "1",
		Name:          "Holiday Crew",
		DateCreated:   now,
		DateDraw:      now.Add(24 * time.Hour),
		CreatorUserID: 42,
	}

	if err := InsertGroup(db, group); err != nil {
		t.Fatalf("InsertGroup returned error: %v", err)
	}

	got, err := GetGroupByID(db, group.GroupID)
	if err != nil {
		t.Fatalf("GetGroupByID returned error: %v", err)
	}

	if got.GroupID != group.GroupID {
		t.Fatalf("expected group id %q, got %q", group.GroupID, got.GroupID)
	}
	if got.Name != group.Name {
		t.Fatalf("expected name %q, got %q", group.Name, got.Name)
	}
	if got.CreatorUserID != group.CreatorUserID {
		t.Fatalf("expected creator_user_id %d, got %d", group.CreatorUserID, got.CreatorUserID)
	}

	group.Name = "Updated Holiday Crew"
	if err := UpdateGroup(db, group); err != nil {
		t.Fatalf("UpdateGroup returned error: %v", err)
	}

	updated, err := GetGroupByID(db, group.GroupID)
	if err != nil {
		t.Fatalf("GetGroupByID after update returned error: %v", err)
	}
	if updated.Name != group.Name {
		t.Fatalf("expected updated name %q, got %q", group.Name, updated.Name)
	}

	if err := DeleteGroup(db, group.GroupID); err != nil {
		t.Fatalf("DeleteGroup returned error: %v", err)
	}

	_, err = GetGroupByID(db, group.GroupID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}
}