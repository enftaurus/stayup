package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a 15-minute HS256 JWT
func GenerateAccessToken(userID string, secret string) (string, int64, error) {
	expiresIn := int64(15 * 60) // 15 minutes in seconds (900)
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, expiresIn, nil
}

// ValidateAccessToken parses and validates the access token
func ValidateAccessToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.UserID == "" {
			return nil, errors.New("token subject is missing")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
