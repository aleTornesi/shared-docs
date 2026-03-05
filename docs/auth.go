package main

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(c *gin.Context) {
	auth_header := c.GetHeader("Authorization")

	split_header := strings.Split(auth_header, " ")

	if len(split_header) != 2 {
		c.JSON(401, gin.H{
			"error": "Invalid Authorization Header",
		})
		return
	}

	if split_header[0] != "Bearer" {
		c.JSON(401, gin.H{
			"error": "Invalid Token Type",
		})
		return
	}

	token := split_header[1]

	claims := new(Claims)

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		return []byte("secret"), nil
	})

	if err != nil {
		c.JSON(401, gin.H{
			"error": "Invalid Token",
		})
		return
	}

	c.Set("claims", claims)

	c.Next()

}

func Login(c *gin.Context) {
	user_id := int64(1)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: user_id,
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"token": tokenString,
	})
}
