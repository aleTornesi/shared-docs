package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	r := gin.Default()
	r.RedirectTrailingSlash = false

	r.POST("/login", Login)
	r.POST("/login/", Login)
	r.GET("/documents", AuthMiddleware, GetDocuments)
	r.GET("/documents/", AuthMiddleware, GetDocuments)
	r.GET("/documents/:id", AuthMiddleware, GetDocument)
	r.GET("/documents/:id/", AuthMiddleware, GetDocument)
	r.PATCH("/documents/:id", AuthMiddleware, PatchDocument)
	r.PATCH("/documents/:id/", AuthMiddleware, PatchDocument)
	r.POST("/documents/:id/pages", AuthMiddleware, AddPage)
	r.POST("/documents/:id/pages/", AuthMiddleware, AddPage)

	r.Run(":8080")
}

func GetClaims(c *gin.Context) (*Claims, error) {
	rawClaims, ok := c.Get("claims")
	if !ok {
		return nil, fmt.Errorf("no claims found")
	}

	claims, ok := rawClaims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("claims is not of type *Claims")
	}
	return claims, nil
}
