package api

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingToken = errors.New("missing token")
)

const (
	AuthHeader     = "Authorization"
	expiryDuration = time.Hour
)

func createToken(userId uuid.UUID) (string, error) {
	// create a new token with symmetric signing (HMAC)
	key := []byte(os.Getenv("JWT_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": userId,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(expiryDuration).Unix(),
		})
	s, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return s, nil
}

func verifyToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_KEY")), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}
	if time.Now().Unix() > claims["exp"].(int64) {
		return uuid.Nil, errors.New("token expired")
	}

	return claims["sub"].(uuid.UUID), nil
}
