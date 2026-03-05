package main

import (
	"context"
	"database/sql"
	"docs/db"
	"fmt"
	"slices"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.RedirectTrailingSlash = false

	r.POST("/login", Login)
	r.GET("/documents", AuthMiddleware, GetDocuments)
	r.GET("/documents/:id", AuthMiddleware, GetDocument)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})
	r.Run(":8080")
}

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type Page struct {
	Number  int16  `json:"number"`
	Content string `json:"content"`
}

type Document struct {
	ID     uint   `json:"id"`
	Title  string `json:"title"`
	Owner  User   `json:"owner"`
	Length uint   `json:"length,omitempty"`
	Pages  []Page `json:"pages,omitempty"`
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

	for _, row := range rows {

		i := slices.IndexFunc(documents, func(p Document) bool {
			return p.ID == uint(row.ID)
		})

		if i == -1 {
			documents = append(documents, Document{
				ID:     uint(row.ID),
				Title:  row.Title,
				Owner:  User{ID: uint(row.OwnerID), Username: row.Username},
				Length: uint(row.C),
			})
			i = len(documents) - 1
		}

	}

	c.JSON(200, documents)

}

func GetDocument(c *gin.Context) {
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid id",
		})
		return
	}
	rows, err := db.New(conn).GetDocument(context.Background(), db.GetDocumentParams{
		ID:      int64(id),
		OwnerID: user_id,
	})

	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{
			"error": "document not found",
		})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	var document Document
	for _, row := range rows {
		document.ID = uint(row.ID)
		document.Title = row.Title

		document.Owner = User{ID: uint(row.OwnerID), Username: row.Username}
		if row.PageNumber.Valid {
			document.Pages = append(document.Pages, Page{Number: row.PageNumber.Int16, Content: row.Content.String})
		}
	}

	c.JSON(200, document)
}
