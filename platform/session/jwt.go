package session

import (
	"fmt"
	"lameCode/platform/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


const TokenExpTimeOffset = time.Hour * 8

// CreateJwtToken returns a new jwt.Token object, set with the `sub`
// to the username parameter and `aud` as the level of access of the
// user.
// That level of acces can be one of "user" and "admin", but
// can be further extended later.
func createJwtToken(username, access string) (*jwt.Token, error) {
	if access != "user" && access != "admin"{
		return nil, fmt.Errorf("Not a valid access value")
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
		return "", err
	}

	return tok.SignedString(config.JwtSecret)
}

func keyFunc(t *jwt.Token) (interface{}, error) {
	return config.JwtSecret, nil
}

func VerifyJwtToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, keyFunc, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}

	// Still not sure what would trigger this, does a mismatch on the
	// signature not set err in jwt.Parse? 
	if !token.Valid {
		return nil, fmt.Errorf("Token found invalid after parse")
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return nil, fmt.Errorf("Invalid claims type, expected jwt.MapClaims")
	}

	return token, nil
}
