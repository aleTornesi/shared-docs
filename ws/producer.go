package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/segmentio/kafka-go"
)

// EditEvent mirrors the struct consumed by the batch-processor service.
type EditEvent struct {
	DocumentID int64  `json:"document_id"`
	UserID     int64  `json:"user_id"`
	Type       string `json:"type"`
	Position   int    `json:"position"`
	Character  string `json:"character"`
}

var kafkaWriter = &kafka.Writer{
	Addr:     kafka.TCP(getEnv("KAFKA_BROKERS", "kafka:9092")),
	Topic:    getEnv("KAFKA_TOPIC", "doc-edits"),
	Balancer: &kafka.Hash{}, // routes by key hash → same document_id → same partition
}

func publishEditEvent(ctx context.Context, event EditEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(event.DocumentID, 10)),
		Value: value,
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
