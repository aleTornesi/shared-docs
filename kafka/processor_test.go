package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aleTornesi/shared-docs/db"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupProcessorDB(t *testing.T) *sql.DB {
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
		t.Fatal("start postgres:", err)
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

	return conn
}

func seedProcessorData(t *testing.T, conn *sql.DB, content string) int64 {
	t.Helper()
	ctx := context.Background()
	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "testuser")
	q := db.New(conn)
	docID, err := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Test", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	err = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})
	if err != nil {
		t.Fatal(err)
	}
	if content != "" {
		err = q.UpdatePageContent(ctx, db.UpdatePageContentParams{DocumentID: docID, PageNumber: 0, Content: content})
		if err != nil {
			t.Fatal(err)
		}
	}
	return docID
}

func makeMsg(t *testing.T, e EditEvent) kafka.Message {
	t.Helper()
	v, _ := json.Marshal(e)
	return kafka.Message{Value: v}
}

func getPageContent(t *testing.T, conn *sql.DB, docID int64) string {
	t.Helper()
	rows, err := db.New(conn).GetDocument(context.Background(), db.GetDocumentParams{ID: docID, OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("no pages found")
	}
	return rows[0].Content
}

func TestProcessBatch_SingleInsert(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "")
	p := NewProcessor(conn)

	msgs := []kafka.Message{
		makeMsg(t, EditEvent{
			EventID: "e1", DocumentID: docID, UserID: 1,
			PageNumber: 0, Type: "insert_character", Position: 0, Character: "H",
		}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	content := getPageContent(t, conn, docID)
	if content != "H" {
		t.Fatalf("content %q, want %q", content, "H")
	}
}

func TestProcessBatch_MultipleEventsSameDoc(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "")
	p := NewProcessor(conn)

	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "H"}),
		makeMsg(t, EditEvent{EventID: "e2", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "i"}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	content := getPageContent(t, conn, docID)
	if content != "Hi" {
		t.Fatalf("content %q, want %q", content, "Hi")
	}
}

func TestProcessBatch_DuplicateEvent(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "")
	p := NewProcessor(conn)

	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "A"}),
	}
	_ = p.ProcessBatch(context.Background(), msgs)

	// Process same event again
	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	content := getPageContent(t, conn, docID)
	if content != "A" {
		t.Fatalf("content %q, want %q (duplicate should be skipped)", content, "A")
	}
}

func TestProcessBatch_DeleteCharacter(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "Hi")
	p := NewProcessor(conn)

	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "delete_character", Position: 2}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	content := getPageContent(t, conn, docID)
	if content != "H" {
		t.Fatalf("content %q, want %q", content, "H")
	}
}

func TestProcessBatch_MixedInsertDelete(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "")
	p := NewProcessor(conn)

	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "a"}),
		makeMsg(t, EditEvent{EventID: "e2", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "b"}),
		makeMsg(t, EditEvent{EventID: "e3", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "c"}),
		makeMsg(t, EditEvent{EventID: "e4", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "delete_character", Position: 0}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	content := getPageContent(t, conn, docID)
	if content != "ab" {
		t.Fatalf("content %q, want %q", content, "ab")
	}
}

func TestProcessBatch_InvalidPageNumber(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "")
	p := NewProcessor(conn)

	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID, UserID: 1, PageNumber: 5, Type: "insert_character", Position: 0, Character: "X"}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err == nil {
		t.Fatal("expected error for invalid page number")
	}
}

func TestProcessBatch_CursorMovement(t *testing.T) {
	conn := setupProcessorDB(t)
	docID := seedProcessorData(t, conn, "hello")
	p := NewProcessor(conn)

	// Insert at position 0, then move to position 5 (original coords) and insert
	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: ">"}),
		makeMsg(t, EditEvent{EventID: "e2", DocumentID: docID, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 5, Character: "!"}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	content := getPageContent(t, conn, docID)
	if content != ">hello!" {
		t.Fatalf("content %q, want %q", content, ">hello!")
	}
}

func TestProcessBatch_MultipleDocs(t *testing.T) {
	conn := setupProcessorDB(t)
	docID1 := seedProcessorData(t, conn, "")

	// Create second doc
	ctx := context.Background()
	q := db.New(conn)
	docID2, _ := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Doc2", OwnerID: 1})
	_ = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID2, PageNumber: 0})

	p := NewProcessor(conn)
	msgs := []kafka.Message{
		makeMsg(t, EditEvent{EventID: "e1", DocumentID: docID1, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "A"}),
		makeMsg(t, EditEvent{EventID: "e2", DocumentID: docID2, UserID: 1, PageNumber: 0, Type: "insert_character", Position: 0, Character: "B"}),
	}

	err := p.ProcessBatch(context.Background(), msgs)
	if err != nil {
		t.Fatal(err)
	}

	if c := getPageContent(t, conn, docID1); c != "A" {
		t.Fatalf("doc1 content %q", c)
	}
	if c := getPageContent(t, conn, docID2); c != "B" {
		t.Fatalf("doc2 content %q", c)
	}
}
