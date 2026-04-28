package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

// Token holds an authentication token and its metadata.
type Token struct {
	UserID    int
	Token     string
	ExpiresAt time.Time
}

var tokenMap = make(map[int]Token)

// CreateToken generates a new bearer token for the given user ID.
func CreateToken(userID int) (string, error) {
	token := Token{
		UserID:    userID,
		Token:     createRandomToken(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	tokenMap[userID] = token
	return token.Token, nil
}

func createRandomToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// ValidateToken checks that a token is valid and not expired.
// It returns the associated user ID on success.
func ValidateToken(token string) (int, error) {
	for _, t := range tokenMap {
		if t.Token == token {
			if time.Now().After(t.ExpiresAt) {
				delete(tokenMap, t.UserID)
				return 0, errors.New("token expired")
			}
			return t.UserID, nil
		}
	}
	return 0, errors.New("invalid token")
}
