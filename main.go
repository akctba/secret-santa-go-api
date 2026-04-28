package main

import (
	"log"
	"net/http"
	"time"

	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/routes"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
)

func main() {
	database.CreateTables()

	r := mux.NewRouter()
	routes.Register(r)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
