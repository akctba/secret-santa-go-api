package controllers

import (
	"database/sql"

	"github.com/akctba/secret-santa-go-api/database"
)

var getDB = func() (*sql.DB, error) {
	return database.GetDb()
}
