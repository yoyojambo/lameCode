package session

import (
	"errors"
	"fmt"
	"lameCode/internal/platform/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenExpTimeOffset = time.Hour * 8

var (
	ErrInvalidAccess     = errors.New("invalid access level")
	ErrInvalidToken      = errors.New("token is invalid")
	ErrInvalidClaimsType = errors.New("invalid claims type")
	ErrEmptyUsername     = errors.New("username cannot be empty")
	ErrEmptyToken        = errors.New("token string cannot be empty")
	ErrMissingClaims     = errors.New("missing required claims")
)

func createJwtToken(username, access string) (*jwt.Token, error) {
	if username == "" {
		return nil, ErrEmptyUsername
	}
	if access != "user" && access != "admin" {
		return nil, fmt.Errorf("%w: got %q, expected 'user' or 'admin'", ErrInvalidAccess, access)
	}

	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"aud": access,
		"iss": "lamecode",
		"iat": jwt.NewNumericDate(time.Now()),
		"nbf": jwt.NewNumericDate(time.Now()),
		"exp": jwt.NewNumericDate(time.Now().Add(TokenExpTimeOffset)),
	})

	return claim, nil
}

func CreateSignedJwtToken(username, access string) (string, error) {
	tok, err := createJwtToken(username, access)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT token: %w", err)
	}

	signedToken, err := tok.SignedString([]byte(config.JwtSecret()))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return signedToken, nil
}

func keyFunc(t *jwt.Token) (any, error) {
	return []byte(config.JwtSecret()), nil
}

func VerifyJwtToken(tokenStr string) (*jwt.Token, error) {
	if tokenStr == "" {
		return nil, ErrEmptyToken
	}

	token, err := jwt.Parse(tokenStr, keyFunc,
		jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaimsType
	}

	// Validate required claims exist
	if claims["sub"] == nil || claims["aud"] == nil {
		return nil, ErrMissingClaims
	}

	return token, nil
}
