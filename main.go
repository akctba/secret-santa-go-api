package main

import (
	"net/http"

	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/routes"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
)

func main() {
	database.CreateTables()

	r := mux.NewRouter()
	routes.Register(r)

	http.ListenAndServe(":8080", r)
}
