package main

import (
	"frontend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/login", handlers.Login)
	r.GET("/login", handlers.LoginPage)

	auth := r.Group("/", handlers.RequireAuth())
	auth.GET("/", handlers.DocumentsPage)
	auth.GET("/documents", handlers.DocumentsList)

	r.Run(":8080")
}
