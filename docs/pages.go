package main

import (
	"context"
	"database/sql"
	"docs/db"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type AddPageBody struct {
	PageIndex int16 `json:"page_index"`
}

func AddPage(c *gin.Context) {
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

	tx, err := conn.BeginTx(context.Background(), &sql.TxOptions{})

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer tx.Rollback()

	dal := db.New(tx)

	row, err := dal.ValidateAccess(context.Background(), db.ValidateAccessParams{
		OwnerID: user_id,
		ID:      id,
	})

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if row == 0 {
		c.JSON(403, gin.H{
			"error": "not allowed",
		})
		return
	}

	var body AddPageBody

	if err = c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = dal.UpdatePageNumbers(context.Background(), db.UpdatePageNumbersParams{
		DocumentID: id,
		PageNumber: body.PageIndex + 1,
	})

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = dal.CreatePage(context.Background(), db.CreatePageParams{
		DocumentID: id,
		PageNumber: body.PageIndex + 1,
	})

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = tx.Commit()

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{})
}
