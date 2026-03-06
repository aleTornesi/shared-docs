package main

import (
	"context"
	"database/sql"
	"docs/db"
	"slices"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

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

type PatchDocumentBody struct {
	Title string `json:"title"`
}

func PatchDocument(c *gin.Context) {
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

	var body PatchDocumentBody

	if err = c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if body.Title != "" {
		err := db.New(conn).PutTitle(context.Background(), db.PutTitleParams{
			ID:    int64(id),
			Title: body.Title,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(204, gin.H{})
}
