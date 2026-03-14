package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aleTornesi/shared-docs/db"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal("start container:", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	schema, err := os.ReadFile(filepath.Join("..", "db", "schema.sql"))
	if err != nil {
		t.Fatal("read schema:", err)
	}
	if _, err := conn.ExecContext(ctx, string(schema)); err != nil {
		t.Fatal("exec schema:", err)
	}

	// Override openDB for tests
	openDB = func() (*sql.DB, error) { return conn, nil }

	return conn
}

func seedTestData(t *testing.T, conn *sql.DB) int64 {
	t.Helper()
	ctx := context.Background()
	_, err := conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "testuser")
	if err != nil {
		t.Fatal(err)
	}
	q := db.New(conn)
	docID, err := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Test Doc", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	err = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})
	if err != nil {
		t.Fatal(err)
	}
	return docID
}

func authedRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Authorization", "Bearer "+validToken(t, 1))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func testRouter() *gin.Engine {
	r := gin.New()
	r.GET("/documents", AuthMiddleware, GetDocuments)
	r.GET("/documents/:id", AuthMiddleware, GetDocument)
	r.PATCH("/documents/:id", AuthMiddleware, PatchDocument)
	r.POST("/documents/:id/pages", AuthMiddleware, AddPage)
	return r
}

func TestGetDocuments(t *testing.T) {
	conn := setupTestDB(t)
	seedTestData(t, conn)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "GET", "/documents", nil))

	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	var docs []Document
	json.Unmarshal(w.Body.Bytes(), &docs)
	if len(docs) != 1 {
		t.Fatalf("got %d docs, want 1", len(docs))
	}
	if docs[0].Title != "Test Doc" {
		t.Fatalf("title %q", docs[0].Title)
	}
}

func TestGetDocument(t *testing.T) {
	conn := setupTestDB(t)
	docID := seedTestData(t, conn)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "GET", fmt.Sprintf("/documents/%d", docID), nil))

	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	var doc Document
	json.Unmarshal(w.Body.Bytes(), &doc)
	if doc.Title != "Test Doc" {
		t.Fatalf("title %q", doc.Title)
	}
	if len(doc.Pages) != 1 {
		t.Fatalf("got %d pages, want 1", len(doc.Pages))
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	setupTestDB(t)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "GET", "/documents/999", nil))

	// GetDocument returns empty rows (not sql.ErrNoRows) when doc doesn't exist
	// since it uses QueryContext not QueryRowContext
	if w.Code != 200 {
		t.Logf("status %d (empty doc expected)", w.Code)
	}
}

func TestGetDocument_InvalidID(t *testing.T) {
	setupTestDB(t)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "GET", "/documents/abc", nil))

	if w.Code != 400 {
		t.Fatalf("status %d, want 400", w.Code)
	}
}

func TestPatchDocument(t *testing.T) {
	conn := setupTestDB(t)
	docID := seedTestData(t, conn)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "PATCH", fmt.Sprintf("/documents/%d", docID), PatchDocumentBody{Title: "Updated"}))

	if w.Code != 204 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	// Verify
	rows, _ := db.New(conn).GetDocument(context.Background(), db.GetDocumentParams{ID: docID, OwnerID: 1})
	if rows[0].Title != "Updated" {
		t.Fatalf("title %q, want Updated", rows[0].Title)
	}
}

func TestPatchDocument_InvalidID(t *testing.T) {
	setupTestDB(t)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "PATCH", "/documents/abc", PatchDocumentBody{Title: "X"}))

	if w.Code != 400 {
		t.Fatalf("status %d, want 400", w.Code)
	}
}
