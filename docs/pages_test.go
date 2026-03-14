package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/aleTornesi/shared-docs/db"
)

func TestAddPage(t *testing.T) {
	conn := setupTestDB(t)
	docID := seedTestData(t, conn)

	r := testRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, authedRequest(t, "POST", fmt.Sprintf("/documents/%d/pages", docID), AddPageBody{PageIndex: 0}))

	if w.Code != 201 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	// Verify pages shifted
	rows, err := db.New(conn).GetDocument(context.Background(), db.GetDocumentParams{ID: docID, OwnerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d pages, want 2", len(rows))
	}
	// Original page 0 should now be page 1, new page at 1
	for i, r := range rows {
		if r.PageNumber != int16(i) {
			t.Fatalf("page %d has number %d", i, r.PageNumber)
		}
	}
}

func TestAddPage_Unauthorized(t *testing.T) {
	conn := setupTestDB(t)
	_ = seedTestData(t, conn)

	// Create a second user with no access
	_, _ = conn.ExecContext(context.Background(), "INSERT INTO users (username) VALUES ($1)", "stranger")

	r := testRouter()
	w := httptest.NewRecorder()

	// Use user 2's token
	req := authedRequest(t, "POST", fmt.Sprintf("/documents/%d/pages", 1), AddPageBody{PageIndex: 0})
	// Override with user 2 token
	req.Header.Set("Authorization", "Bearer "+validToken(t, 2))
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Should be 403 or 500 (ValidateAccess returns error for no rows)
	if w.Code == 201 {
		t.Fatal("expected non-201 for unauthorized user")
	}
}
