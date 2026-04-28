package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akctba/secret-santa-go-api/auth"
	"github.com/akctba/secret-santa-go-api/database"
	"github.com/akctba/secret-santa-go-api/routes"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	if err := auth.ValidateJWTConfig(); err != nil {
		log.Fatalf("invalid JWT configuration: %v", err)
	}
	log.Printf("application starting in %s environment", auth.ResolvedEnvironment())

	database.CreateTables()

	r := mux.NewRouter()
	routes.Register(r)
	handler := corsHandler(r)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func corsHandler(next http.Handler) http.Handler {
	allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(allowedOrigins) == 0 {
		log.Print("CORS_ALLOWED_ORIGINS not set; cross-origin browser requests are disabled")
	}

	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(next)
}

func parseAllowedOrigins(value string) []string {
	if value == "" {
		return nil
	}

	origins := strings.Split(value, ",")
	allowedOrigins := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmedOrigin := strings.TrimSpace(origin)
		if trimmedOrigin == "" {
			continue
		}
		allowedOrigins = append(allowedOrigins, trimmedOrigin)
	}

	if len(allowedOrigins) == 0 {
		return nil
	}

	return allowedOrigins
}
