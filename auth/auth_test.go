package auth

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCreateAndValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := CreateToken(42)
	if err != nil {
		t.Fatalf("CreateToken returned error: %v", err)
	}

	userID, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if userID != 42 {
		t.Fatalf("ValidateToken returned %d, want 42", userID)
	}
}

func TestValidateJWTConfigRequiresSecretInDev(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("APP_ENV", "DEV")

	if err := ValidateJWTConfig(); err == nil {
		t.Fatal("ValidateJWTConfig expected error when JWT_SECRET is missing")
	}
}

func TestValidateJWTConfigAllowsFallbackInLocal(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("APP_ENV", "LOCAL")

	if err := ValidateJWTConfig(); err != nil {
		t.Fatalf("ValidateJWTConfig returned error: %v", err)
	}
}

func TestValidateJWTConfigRejectsShortSecretOutsideLocal(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("APP_ENV", "PROD")

	if err := ValidateJWTConfig(); err == nil {
		t.Fatal("ValidateJWTConfig expected error for short JWT_SECRET")
	}
}

func TestValidateJWTConfigAcceptsStrongSecretInProd(t *testing.T) {
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("APP_ENV", "PROD")

	if err := ValidateJWTConfig(); err != nil {
		t.Fatalf("ValidateJWTConfig returned error: %v", err)
	}
}

func TestValidateJWTConfigRejectsInvalidEnvironment(t *testing.T) {
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("APP_ENV", "STAGING")

	if err := ValidateJWTConfig(); err == nil {
		t.Fatal("ValidateJWTConfig expected error for invalid APP_ENV")
	}
}

func TestValidateTokenInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	if _, err := ValidateToken("not-a-token"); err == nil {
		t.Fatal("ValidateToken expected error for malformed token")
	}
}

func TestValidateTokenExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(7),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
		TokenType: accessTokenType,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey())
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := ValidateToken(tokenString); err == nil {
		t.Fatal("ValidateToken expected error for expired token")
	}
}

func TestValidateTokenRejectsRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	refreshToken, err := CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("CreateRefreshToken returned error: %v", err)
	}

	if _, err := ValidateToken(refreshToken); err == nil {
		t.Fatal("ValidateToken expected error for refresh token")
	}
}

func TestCreateAndValidateRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	refreshToken, err := CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("CreateRefreshToken returned error: %v", err)
	}

	userID, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}

	if userID != 42 {
		t.Fatalf("ValidateRefreshToken returned %d, want 42", userID)
	}
}

func TestCreateAndValidateTokenConcurrent(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)

	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		userID := i + 1
		go func() {
			defer wg.Done()

			token, err := CreateToken(userID)
			if err != nil {
				errCh <- err
				return
			}

			validatedUserID, err := ValidateToken(token)
			if err != nil {
				errCh <- err
				return
			}

			if validatedUserID != userID {
				errCh <- &mismatchError{got: validatedUserID, want: userID}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent auth failed: %v", err)
	}
}

type mismatchError struct {
	got  int
	want int
}

func (e *mismatchError) Error() string {
	return "mismatched user ID"
}
