package main

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	batchSize    = 100
	batchTimeout = 500 * time.Millisecond
)

type Consumer struct {
	reader    *kafka.Reader
	processor *Processor
}

func NewConsumer(brokers []string, topic, groupID string, processor *Processor) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	return &Consumer{reader: reader, processor: processor}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()

	batch := make([]kafka.Message, 0, batchSize)
	timer := time.NewTimer(batchTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				c.flush(ctx, batch)
			}
			return nil
		case <-timer.C:
			if len(batch) > 0 {
				c.flush(ctx, batch)
				batch = batch[:0]
			}
			timer.Reset(batchTimeout)
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				log.Println("fetch:", err)
				continue
			}

			batch = append(batch, msg)

			if len(batch) >= batchSize {
				c.flush(ctx, batch)
				if err := c.reader.CommitMessages(ctx, batch...); err != nil {
					log.Println("commit:", err)
				}
				batch = batch[:0]
				timer.Reset(batchTimeout)
			}
		}
	}
}

func (c *Consumer) flush(ctx context.Context, batch []kafka.Message) {
	if err := c.processor.ProcessBatch(ctx, batch); err != nil {
		log.Println("process batch:", err)
	}
}
