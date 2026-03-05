package main

import (
	"context"
	"database/sql"
	"docs/db"
	"fmt"
	_ "github.com/lib/pq"
	"slices"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/login", Login)
	r.GET("/documents", AuthMiddleware, GetDocuments)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})
	r.Run(":8080")
}

type Document struct {
	ID    uint     `json:"id"`
	Title string   `json:"title"`
	Pages []string `json:"pages"`
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

func GetDocuments(c *gin.Context) {
	conn, err := sql.Open("postgres", "host=db port=5432 user=postgres password=postgres dbname=shared-docs sslmode=disable")

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	claims, err := GetClaims(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	user_id := claims.UserID
	rows, err := db.New(conn).GetDocuments(context.Background(), user_id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	var documents []Document

	println(user_id, len(rows))
	for _, row := range rows {

		i := slices.IndexFunc(documents, func(p Document) bool {
			return p.ID == uint(row.ID)
		})

		if i == -1 {
			documents = append(documents, Document{
				ID:    uint(row.ID),
				Title: row.Title,
			})
			i = len(documents) - 1
		}

		if row.PageNumber.Valid {
			documents[i].Pages = append(documents[i].Pages, row.Content.String)
		}

	}

	c.JSON(200, documents)

}
