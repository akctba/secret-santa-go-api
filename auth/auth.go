package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 5 * time.Minute

// CreateToken generates a new bearer token for the given user ID.
func CreateToken(userID int) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey())
	if err != nil {
		return "", fmt.Errorf("sign jwt token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken checks that a token is valid and not expired.
// It returns the associated user ID on success.
func ValidateToken(token string) (int, error) {
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return signingKey(), nil
	})
	if err != nil {
		return 0, errors.New("invalid token")
	}

	if !parsedToken.Valid {
		return 0, errors.New("invalid token")
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, errors.New("invalid token")
	}

	return userID, nil
}

func signingKey() []byte {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return []byte(value)
	}

	return []byte("secret-santa-dev-secret")
}
