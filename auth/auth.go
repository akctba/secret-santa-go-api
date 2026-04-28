package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 30 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	accessTokenType  = "access"
	refreshTokenType = "refresh"

	minJWTSecretLength = 32
	appEnvVar          = "APP_ENV"

	envLocal = "LOCAL"
	envDev   = "DEV"
	envProd  = "PROD"
)

type tokenClaims struct {
	jwt.RegisteredClaims
	TokenType string `json:"token_type"`
}

// CreateToken generates a new bearer token for the given user ID.
func CreateToken(userID int) (string, error) {
	return CreateAccessToken(userID)
}

// CreateAccessToken generates a new access token for the given user ID.
func CreateAccessToken(userID int) (string, error) {
	return createToken(userID, accessTokenTTL, accessTokenType)
}

// CreateRefreshToken generates a new refresh token for the given user ID.
func CreateRefreshToken(userID int) (string, error) {
	return createToken(userID, refreshTokenTTL, refreshTokenType)
}

func createToken(userID int, ttl time.Duration, tokenType string) (string, error) {
	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TokenType: tokenType,
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
	return validateTokenType(token, accessTokenType)
}

// ValidateRefreshToken checks that a refresh token is valid and not expired.
// It returns the associated user ID on success.
func ValidateRefreshToken(token string) (int, error) {
	return validateTokenType(token, refreshTokenType)
}

func validateTokenType(token string, expectedType string) (int, error) {
	claims := &tokenClaims{}
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

	if claims.TokenType != expectedType {
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

	if currentEnvironment() == envLocal {
		return []byte("secret-santa-dev-secret")
	}

	return nil
}

// ValidateJWTConfig checks whether JWT signing configuration is safe to use.
func ValidateJWTConfig() error {
	env := currentEnvironment()
	if env != envLocal && env != envDev && env != envProd {
		return errors.New("APP_ENV must be one of LOCAL, DEV, PROD")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if env == envLocal {
			return nil
		}

		return errors.New("JWT_SECRET must be set")
	}

	if env != envLocal && len(jwtSecret) < minJWTSecretLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLength)
	}

	return nil
}

// ResolvedEnvironment returns the effective application environment.
func ResolvedEnvironment() string {
	return currentEnvironment()
}

func currentEnvironment() string {
	envValue := strings.TrimSpace(strings.ToUpper(os.Getenv(appEnvVar)))
	if envValue == "" {
		return envProd
	}

	return envValue
}
