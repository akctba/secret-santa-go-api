package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/akctba/secret-santa-go-api/models"
	_ "github.com/mattn/go-sqlite3"
)

func openParticipantTestDB(t *testing.T) *sql.DB {
	t.Helper()

	t.Chdir(t.TempDir())
	CreateTables()

	db, err := GetDb()
	if err != nil {
		t.Fatalf("open participant test db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func insertParticipantTestUser(t *testing.T, db *sql.DB, userID int, name string, email string) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO Users (user_id, user_name, user_email, password, gender, date_of_birth)
	VALUES (?, ?, ?, ?, ?, ?);`, userID, name, email, "secret", "", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert user %d: %v", userID, err)
	}
}

func TestParticipantRepoFreshDBFlow(t *testing.T) {
	db := openParticipantTestDB(t)

	insertParticipantTestUser(t, db, 1, "Alice", "alice@example.com")
	insertParticipantTestUser(t, db, 2, "Bob", "bob@example.com")

	if err := InsertParticipant(db, models.ParticipantRequest{GroupID: "1", UserID: 1}); err != nil {
		t.Fatalf("InsertParticipant first user returned error: %v", err)
	}
	if err := InsertParticipant(db, models.ParticipantRequest{GroupID: "1", UserID: 2}); err != nil {
		t.Fatalf("InsertParticipant second user returned error: %v", err)
	}

	byGroup, err := GetParticipantsByGroupID(db, "1")
	if err != nil {
		t.Fatalf("GetParticipantsByGroupID returned error: %v", err)
	}
	if len(byGroup) != 2 {
		t.Fatalf("expected 2 participants by group, got %d", len(byGroup))
	}

	byUser, err := GetParticipantByUserID(db, "1")
	if err != nil {
		t.Fatalf("GetParticipantByUserID returned error: %v", err)
	}
	if len(byUser) != 1 {
		t.Fatalf("expected 1 participant by user, got %d", len(byUser))
	}
	if byUser[0].UserName != "Alice" {
		t.Fatalf("expected participant user name Alice, got %q", byUser[0].UserName)
	}

	toDraw, err := GetParticipantsToDraw(db, "1")
	if err != nil {
		t.Fatalf("GetParticipantsToDraw before update returned error: %v", err)
	}
	if len(toDraw) != 2 {
		t.Fatalf("expected 2 participants to draw, got %d", len(toDraw))
	}

	updatedParticipant := toDraw[0]
	updatedParticipant.FriendUserID = 2
	if err := UpdateParticipant(db, updatedParticipant); err != nil {
		t.Fatalf("UpdateParticipant returned error: %v", err)
	}

	persisted, err := GetUserParticipant(db, updatedParticipant.UserID, 1)
	if err != nil {
		t.Fatalf("GetUserParticipant returned error: %v", err)
	}
	if persisted.FriendUserID != 2 {
		t.Fatalf("expected friend_user_id 2, got %d", persisted.FriendUserID)
	}

	remaining, err := GetParticipantsToDraw(db, "1")
	if err != nil {
		t.Fatalf("GetParticipantsToDraw after update returned error: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 participant remaining to draw, got %d", len(remaining))
	}
}

func TestCreateTablesMigratesParticipantFriendColumn(t *testing.T) {
	t.Chdir(t.TempDir())

	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	_, err = db.Exec(`CREATE TABLE Participants (
		group_id INTEGER,
		user_id INTEGER,
		joined_at TEXT,
		fried_user_id INTEGER,
		PRIMARY KEY (group_id, user_id)
	);`)
	if err != nil {
		t.Fatalf("create legacy Participants table: %v", err)
	}

	CreateTables()

	hasFriend, err := participantColumnExists(db, "friend_user_id")
	if err != nil {
		t.Fatalf("check friend_user_id column: %v", err)
	}
	if !hasFriend {
		t.Fatal("expected friend_user_id column to exist after migration")
	}

	hasFried, err := participantColumnExists(db, "fried_user_id")
	if err != nil {
		t.Fatalf("check fried_user_id column: %v", err)
	}
	if hasFried {
		t.Fatal("expected fried_user_id column to be removed after migration")
	}

	_, err = db.Exec(`SELECT friend_user_id FROM Participants LIMIT 1;`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("select friend_user_id after migration: %v", err)
	}
}
