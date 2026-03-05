package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"frontend/templates"

	"github.com/gin-gonic/gin"
)

func apiBase() string {
	if base := os.Getenv("API_BASE_URL"); base != "" {
		return base
	}
	return "http://localhost:8081"
}

func DocumentsPage(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	templates.DocumentsPage().Render(c.Request.Context(), c.Writer)
}

func DocumentsList(c *gin.Context) {
	req, _ := http.NewRequest("GET", apiBase()+"/documents", nil)
	token, _ := c.Get("token")
	if t, ok := token.(string); ok {
		req.Header.Set("Authorization", "Bearer "+t)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer resp.Body.Close()

	var docs []templates.Document
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse documents")
		return
	}

	c.Header("Content-Type", "text/html")
	templates.DocumentList(docs).Render(c.Request.Context(), c.Writer)
}
