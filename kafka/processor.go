package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/aleTornesi/shared-docs/db"
	"github.com/segmentio/kafka-go"
)

type EditEvent struct {
	EventID    string `json:"event_id"`
	DocumentID int64  `json:"document_id"`
	UserID     int64  `json:"user_id"`
	PageNumber int16  `json:"page_number"`
	Type       string `json:"type"`
	Position   int    `json:"position"`
	Character  string `json:"character"`
}

type Processor struct {
	db *sql.DB
}

type DocumentProcessor struct {
	p      *Processor
	doc    []db.GetDocumentRow
	pages  []*GapBuffer
	events []EditEvent
}

func NewProcessor(db *sql.DB) *Processor {
	return &Processor{db: db}
}

const maxConcurrentDocs = 10

func (p *Processor) ProcessBatch(ctx context.Context, msgs []kafka.Message) error {
	byDoc := map[int64][]EditEvent{}
	for _, msg := range msgs {
		var e EditEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			log.Printf("unmarshal offset=%d: %v", msg.Offset, err)
			continue
		}
		byDoc[e.DocumentID] = append(byDoc[e.DocumentID], e)
	}

	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
		sem  = make(chan struct{}, maxConcurrentDocs)
	)

	for docID, events := range byDoc {
		wg.Add(1)
		sem <- struct{}{}
		go func(docID int64, events []EditEvent) {
			defer wg.Done()
			defer func() { <-sem }()
			docProcessor := &DocumentProcessor{p: p, events: events}
			if localErr := docProcessor.applyEdits(ctx, docID); localErr != nil {
				log.Printf("applyEdits doc=%d: %v", docID, localErr)
				mu.Lock()
				errs = append(errs, localErr)
				mu.Unlock()
			}
		}(docID, events)
	}
	wg.Wait()
	return errors.Join(errs...)
}

func (dp *DocumentProcessor) applyEdits(ctx context.Context, docID int64) error {
	tx, err := dp.p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dp.doc, err = db.New(tx).GetDocument(ctx, db.GetDocumentParams{ID: docID, OwnerID: dp.events[0].UserID})
	if err != nil {
		return err
	}
	dp.pages = make([]*GapBuffer, len(dp.doc))

	queries := db.New(tx)
	for _, event := range dp.events {
		rows, err := queries.InsertProcessedEvent(ctx, event.EventID)
		if err != nil {
			return err
		}
		if rows == 0 {
			continue // already processed
		}

		if event.PageNumber >= int16(len(dp.pages)) {
			return fmt.Errorf("page number %d out of range", event.PageNumber)
		}
		if err := dp.editProcessingSetup(event); err != nil {
			return err
		}
		page := dp.pages[event.PageNumber]
		switch event.Type {
		case "insert_character":
			page.InsertString(event.Character)
		case "delete_character":
			page.Delete()
		}
	}

	err = dp.WriteEdits(ctx, tx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (dp *DocumentProcessor) editProcessingSetup(event EditEvent) error {
	if dp.pages[event.PageNumber] == nil {
		err := dp.populateGapBuffer(event)
		return err
	}

	if event.Position == dp.pages[event.PageNumber].OriginalCursorPos() {
		return nil
	}

	return dp.pages[event.PageNumber].MoveCursorOriginal(event.Position)
}

func (dp *DocumentProcessor) populateGapBuffer(event EditEvent) error {
	pageIndex, ok := slices.BinarySearchFunc(dp.doc, event.PageNumber, func(p db.GetDocumentRow, target int16) int {
		return int(p.PageNumber - target)
	})

	if !ok {
		return fmt.Errorf("page %d not found", event.PageNumber)
	}

	page := dp.doc[pageIndex]

	var err error
	dp.pages[event.PageNumber], err = NewGapBuffer(
		page.Content,
		event.Position,
	)

	return err
}

func (dp *DocumentProcessor) WriteEdits(ctx context.Context, tx *sql.Tx) error {
	for i, page := range dp.pages {
		if page == nil {
			continue
		}
		err := db.New(tx).UpdatePageContent(ctx, db.UpdatePageContentParams{
			DocumentID: dp.doc[0].ID,
			PageNumber: int16(i),
			Content:    page.String(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
