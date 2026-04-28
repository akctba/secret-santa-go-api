package database

import (
	"database/sql"
	"log"
)

const (
	DbDriver = "sqlite3"
	DbName   = "secretsanta.db"
)

func CreateTables() {
	// Connect to the database
	// Open a connection to the SQLite database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create database tables
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS Users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_name TEXT,
		user_email TEXT,
		password TEXT,
		gender TEXT,
		date_of_birth TEXT
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS Groups (
		group_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		date_created DATETIME,
		date_draw DATETIME,
		creator_user_id INTEGER
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS Participants (
		group_id INTEGER,
		user_id INTEGER,
		joined_at TEXT,
		friend_user_id INTEGER,
		PRIMARY KEY (group_id, user_id)
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	if err := ensureParticipantFriendColumn(db); err != nil {
		log.Printf("ensure participant friend_user_id column: %v\n", err)
	}
}

func ensureParticipantFriendColumn(db *sql.DB) error {
	hasFriend, err := participantColumnExists(db, "friend_user_id")
	if err != nil {
		return err
	}
	if hasFriend {
		return nil
	}

	hasFried, err := participantColumnExists(db, "fried_user_id")
	if err != nil {
		return err
	}
	if !hasFried {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE Participants RENAME COLUMN fried_user_id TO friend_user_id;`); err != nil {
		return err
	}

	return nil
}

func participantColumnExists(db *sql.DB, columnName string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(Participants);`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func GetDb() (*sql.DB, error) {
	// Connect to the database
	db, err := sql.Open(DbDriver, DbName)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func CloseDb(db *sql.DB) {
	db.Close()
}

func DropTables(db *sql.DB) {

	sqlStmt := `DROP TABLE IF EXISTS Users;`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `DROP TABLE IF EXISTS Groups;`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `DROP TABLE IF EXISTS Participants;`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}
