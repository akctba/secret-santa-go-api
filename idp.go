package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

type Token struct {
	UserId    int
	Token     string
	ExpiresAt time.Time
}

var tokenMap = make(map[int]Token)

func CreateToken(userId int) (string, error) {
	token := Token{
		UserId:    userId,
		Token:     createRandomToken(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	tokenMap[userId] = token
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

func ValidateToken(token string) (int, error) {
	for _, t := range tokenMap {
		if t.Token == token {
			if time.Now().After(t.ExpiresAt) {
				delete(tokenMap, t.UserId)
				return 0, errors.New("token expired")
			}
			return t.UserId, nil
		}
	}
	return 0, errors.New("invalid token")
}
