package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

type EditEvent struct {
	DocumentID int64  `json:"document_id"`
	UserID     int64  `json:"user_id"`
	Type       string `json:"type"`
	Position   int    `json:"position"`
	Character  string `json:"character"`
}

type Processor struct {
	db *sql.DB
}

func NewProcessor(db *sql.DB) *Processor {
	return &Processor{db: db}
}

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

	var wg sync.WaitGroup
	for docID, events := range byDoc {
		wg.Add(1)
		go func(docID int64, events []EditEvent) {
			defer wg.Done()
			if err := p.applyEdits(ctx, docID, events); err != nil {
				log.Printf("applyEdits doc=%d: %v", docID, err)
			}
		}(docID, events)
	}
	wg.Wait()
	return nil
}

func (p *Processor) applyEdits(ctx context.Context, docID int64, events []EditEvent) error {
	// TODO: apply events in order to page.content for docID
	log.Printf("applying %d edits to doc=%d", len(events), docID)
	return nil
}
