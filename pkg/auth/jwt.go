package auth

import (
        "errors"
        "time"

        jwt "github.com/golang-jwt/jwt/v5"
        "github.com/oullin/database/repository"
)

// JWTHandler manages creation and validation of JSON Web Tokens.
type JWTHandler struct {
        Keys *repository.ApiKeys
        // TTL defines how long generated tokens remain valid.
        TTL time.Duration
}

// Claims represents application specific JWT claims.
type Claims struct {
	AccountName string `json:"account_name"`
	jwt.RegisteredClaims
}

// MakeJWTHandler returns a configured handler using the provided API key repository.
func MakeJWTHandler(keys *repository.ApiKeys, ttl time.Duration) (JWTHandler, error) {
        if keys == nil {
                return JWTHandler{}, errors.New("api key repository is nil")
        }
        return JWTHandler{Keys: keys, TTL: ttl}, nil
}

// Generate creates a signed JWT for the provided account name.
func (j JWTHandler) Generate(accountName string) (string, error) {
	apiKey := j.Keys.FindBy(accountName)
	if apiKey == nil {
		return "", errors.New("api key not found")
	}

	claims := Claims{
		AccountName: accountName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(apiKey.SecretKey)
}

// Validate parses the token string and returns the Claims if valid.
func (j JWTHandler) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		claims, ok := token.Claims.(*Claims)
		if !ok {
			return nil, errors.New("invalid token claims")
		}
		apiKey := j.Keys.FindBy(claims.AccountName)
		if apiKey == nil {
			return nil, errors.New("api key not found")
		}
		return apiKey.SecretKey, nil
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
