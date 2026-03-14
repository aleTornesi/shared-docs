package db_test

import (
	"context"
	"testing"

	"github.com/aleTornesi/shared-docs/db"
)

func TestCreateDocumentAndGetDocuments(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	// Seed user
	_, err := conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "alice")
	if err != nil {
		t.Fatal(err)
	}

	docID, err := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Test Doc", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if docID == 0 {
		t.Fatal("expected non-zero doc ID")
	}

	// Create a page so GetDocuments lateral count works
	err = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})
	if err != nil {
		t.Fatal(err)
	}

	docs, err := q.GetDocuments(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("got %d docs, want 1", len(docs))
	}
	if docs[0].Title != "Test Doc" {
		t.Fatalf("title %q, want %q", docs[0].Title, "Test Doc")
	}
	if docs[0].C != 1 {
		t.Fatalf("page count %d, want 1", docs[0].C)
	}
}

func TestGetDocumentWithPages(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, err := conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "alice")
	if err != nil {
		t.Fatal(err)
	}

	docID, err := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "My Doc", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}

	for i := int16(0); i < 3; i++ {
		err = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: i})
		if err != nil {
			t.Fatal(err)
		}
	}

	rows, err := q.GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if rows[0].PageNumber != 0 || rows[2].PageNumber != 2 {
		t.Fatalf("unexpected page ordering: %d, %d", rows[0].PageNumber, rows[2].PageNumber)
	}
}

func TestCreateDocumentAccessAndGetDocument(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, err := conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "owner")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "shared")
	if err != nil {
		t.Fatal(err)
	}

	docID, err := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Shared Doc", OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	err = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})
	if err != nil {
		t.Fatal(err)
	}

	// User 2 can't access yet
	rows, err := q.GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no rows for unauthorized user, got %d", len(rows))
	}

	// Grant access
	err = q.CreateDocumentAccess(ctx, db.CreateDocumentAccessParams{UserID: 2, DocumentID: docID})
	if err != nil {
		t.Fatal(err)
	}

	// Now user 2 can access
	rows, err = q.GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestValidateAccess(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "owner")
	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "shared")
	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "stranger")

	docID, _ := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Doc", OwnerID: 1})
	_ = q.CreateDocumentAccess(ctx, db.CreateDocumentAccessParams{UserID: 2, DocumentID: docID})

	// Owner
	ok, err := q.ValidateAccess(ctx, db.ValidateAccessParams{OwnerID: 1, ID: docID})
	if err != nil || !ok {
		t.Fatalf("owner should have access: ok=%v err=%v", ok, err)
	}

	// Shared user
	ok, err = q.ValidateAccess(ctx, db.ValidateAccessParams{OwnerID: 2, ID: docID})
	if err != nil || !ok {
		t.Fatalf("shared user should have access: ok=%v err=%v", ok, err)
	}

	// Unauthorized
	_, err = q.ValidateAccess(ctx, db.ValidateAccessParams{OwnerID: 3, ID: docID})
	if err == nil {
		t.Fatal("stranger should not have access")
	}
}

func TestInsertProcessedEvent(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	rows, err := q.InsertProcessedEvent(ctx, "evt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("first insert: rows=%d, want 1", rows)
	}

	// Duplicate
	rows, err = q.InsertProcessedEvent(ctx, "evt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("duplicate insert: rows=%d, want 0", rows)
	}
}

func TestUpdatePageContent(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "alice")
	docID, _ := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Doc", OwnerID: 1})
	_ = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})

	err := q.UpdatePageContent(ctx, db.UpdatePageContentParams{
		DocumentID: docID, PageNumber: 0, Content: "hello world",
	})
	if err != nil {
		t.Fatal(err)
	}

	rows, _ := q.GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: 1})
	if len(rows) != 1 || rows[0].Content != "hello world" {
		t.Fatalf("content %q, want %q", rows[0].Content, "hello world")
	}
}

func TestUpdatePageNumbersAndCreatePage(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "alice")
	docID, _ := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Doc", OwnerID: 1})
	_ = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})
	_ = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 1})

	// Insert page at index 1 (shift existing page 1 → 2)
	err := q.UpdatePageNumbers(ctx, db.UpdatePageNumbersParams{DocumentID: docID, PageNumber: 1})
	if err != nil {
		t.Fatal(err)
	}
	err = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 1})
	if err != nil {
		t.Fatal(err)
	}

	rows, _ := q.GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: 1})
	if len(rows) != 3 {
		t.Fatalf("got %d pages, want 3", len(rows))
	}
	for i, r := range rows {
		if r.PageNumber != int16(i) {
			t.Fatalf("page %d has number %d", i, r.PageNumber)
		}
	}
}

func TestPutTitle(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "alice")
	docID, _ := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Old", OwnerID: 1})
	_ = q.CreatePage(ctx, db.CreatePageParams{DocumentID: docID, PageNumber: 0})

	err := q.PutTitle(ctx, db.PutTitleParams{Title: "New", ID: docID, OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}

	rows, _ := q.GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: 1})
	if rows[0].Title != "New" {
		t.Fatalf("title %q, want %q", rows[0].Title, "New")
	}
}

func TestDeleteDocumentAccess(t *testing.T) {
	conn := SetupTestDB(t)
	ctx := context.Background()
	q := db.New(conn)

	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "owner")
	_, _ = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ($1)", "shared")

	docID, _ := q.CreateDocument(ctx, db.CreateDocumentParams{Title: "Doc", OwnerID: 1})
	_ = q.CreateDocumentAccess(ctx, db.CreateDocumentAccessParams{UserID: 2, DocumentID: docID})

	// Verify access
	ok, _ := q.ValidateAccess(ctx, db.ValidateAccessParams{OwnerID: 2, ID: docID})
	if !ok {
		t.Fatal("expected access before delete")
	}

	// Revoke
	err := q.DeleteDocumentAccess(ctx, db.DeleteDocumentAccessParams{UserID: 2, DocumentID: docID})
	if err != nil {
		t.Fatal(err)
	}

	// Verify revoked
	_, err = q.ValidateAccess(ctx, db.ValidateAccessParams{OwnerID: 2, ID: docID})
	if err == nil {
		t.Fatal("expected no access after delete")
	}
}
