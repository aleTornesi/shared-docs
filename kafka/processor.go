package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// EditEvent mirrors the edit messages published by the ws service.
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
	events := make([]EditEvent, 0, len(msgs))
	for _, msg := range msgs {
		var e EditEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			log.Printf("unmarshal msg offset=%d: %v", msg.Offset, err)
			continue
		}
		events = append(events, e)
	}

	// TODO: apply events to document content in DB.
	// Group by document_id, apply edits in order, then persist in a single transaction.
	log.Printf("processed batch of %d events", len(events))
	return nil
}
