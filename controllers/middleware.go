package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/akctba/secret-santa-go-api/auth"
)

type authContextKey string

const authenticatedUserIDKey authContextKey = "authenticatedUserID"

func authenticatedUserIDFromRequest(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(authenticatedUserIDKey).(int)
	if !ok {
		return 0, false
	}

	return userID, true
}

// BearerAuth is middleware that validates a Bearer token in the Authorization header.
func BearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), authenticatedUserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
