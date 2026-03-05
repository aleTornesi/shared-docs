package handlers

import (
	"encoding/json"
	"frontend/templates"
	"net/http"

	"github.com/gin-gonic/gin"
)

const tokenCookieName = "auth_token"

func LoginPage(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	templates.LoginPage().Render(c.Request.Context(), c.Writer)
}

func Login(c *gin.Context) {
	resp, err := http.Post(apiBase()+"/login", "application/json", nil)
	if err != nil {
		c.String(http.StatusBadGateway, "Failed to reach API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.String(resp.StatusCode, "Login failed")
		return
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse login response")
		return
	}

	c.SetCookie(tokenCookieName, result.Token, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(tokenCookieName)
		if err != nil || token == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}
}
