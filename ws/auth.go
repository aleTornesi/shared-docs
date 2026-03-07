package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func ParseToken(r *http.Request) (*Claims, error) {
	auth_header := r.Header.Get("Authorization")
	split_header := strings.Split(auth_header, " ")

	if len(split_header) != 2 {
		return nil, fmt.Errorf("invalid authorization header")
	}

	if split_header[0] != "Bearer" {
		return nil, fmt.Errorf("invalid token type")
	}

	claims := new(Claims)
	_, err := jwt.ParseWithClaims(split_header[1], claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		return []byte("secret"), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
