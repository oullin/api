package auth

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// JWTHandler manages creation and validation of JSON Web Tokens.
type JWTHandler struct {
	// SecretKey is used to sign tokens.
	SecretKey []byte
	// TTL defines how long generated tokens remain valid.
	TTL time.Duration
}

// Claims represents application specific JWT claims.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// MakeJWTHandler validates the provided secret and returns a configured handler.
func MakeJWTHandler(secret []byte, ttl time.Duration) (JWTHandler, error) {
	if len(secret) < 16 {
		return JWTHandler{}, errors.New("secret key too short")
	}

	return JWTHandler{SecretKey: secret, TTL: ttl}, nil
}

// Generate creates a signed JWT for the provided username.
func (j JWTHandler) Generate(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// Validate parses the token string and returns the Claims if valid.
func (j JWTHandler) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return j.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
