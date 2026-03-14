package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func validToken(t *testing.T, userID int64) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:           userID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	})
	s, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAuthMiddleware_Valid(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthMiddleware, func(c *gin.Context) {
		claims, err := GetClaims(c)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"user_id": claims.UserID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, 42))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_Missing(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthMiddleware, func(c *gin.Context) {
		c.JSON(200, nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("status %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := gin.New()
	r.GET("/test", AuthMiddleware, func(c *gin.Context) {
		c.JSON(200, nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer bad.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("status %d, want 401", w.Code)
	}
}

func TestLogin(t *testing.T) {
	r := gin.New()
	r.POST("/login", Login)

	req := httptest.NewRequest("POST", "/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatal("empty body")
	}
}

func TestGetClaims_Valid(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	claims := &Claims{UserID: 99}
	c.Set("claims", claims)

	got, err := GetClaims(c)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != 99 {
		t.Fatalf("user_id %d, want 99", got.UserID)
	}
}

func TestGetClaims_NoClaims(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	_, err := GetClaims(c)
	if err == nil {
		t.Fatal("expected error")
	}
}
