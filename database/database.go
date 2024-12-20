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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_name TEXT,
		date_created TEXT,
		date_draw TEXT,
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
		fried_user_id INTEGER,
		PRIMARY KEY (group_id, user_id)
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
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
