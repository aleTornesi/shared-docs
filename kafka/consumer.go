package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	batchSize       = 100
	batchTimeout    = 500 * time.Millisecond
	pipelineSize    = 2
	maxFetchRetries = 5
)

type Consumer struct {
	reader    *kafka.Reader
	processor *Processor
}

func NewConsumer(brokers []string, topic, groupID string, processor *Processor) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: reader, processor: processor}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()

	pipeline := make(chan []kafka.Message, pipelineSize)
	fetchCh := make(chan kafka.Message)

	// fetch goroutine: keeps reading so the main select stays responsive to timer/ctx
	go func() {
		defer close(fetchCh)
		retries := 0
		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				retries++
				if retries >= maxFetchRetries {
					log.Printf("fetch: max retries (%d) reached: %v", maxFetchRetries, err)
					return
				}
				backoff := time.Duration(1<<retries) * 100 * time.Millisecond
				log.Printf("fetch (retry %d/%d): %v", retries, maxFetchRetries, err)
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return
				}
				continue
			}
			retries = 0
			select {
			case fetchCh <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	var processorWg sync.WaitGroup
	processorWg.Go(func() {
		for batch := range pipeline {
			if err := c.processor.ProcessBatch(ctx, batch); err != nil {
				log.Println("process batch:", err)
				continue
			}
			if err := c.reader.CommitMessages(ctx, batch...); err != nil {
				log.Println("commit:", err)
			}
		}
	})

	batch := make([]kafka.Message, 0, batchSize)
	timer := time.NewTimer(batchTimeout)
	defer timer.Stop()

	sendBatch := func() {
		b := make([]kafka.Message, len(batch))
		copy(b, batch)
		select {
		case pipeline <- b:
		case <-ctx.Done():
		}
		batch = batch[:0]
		timer.Reset(batchTimeout)
	}

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				b := make([]kafka.Message, len(batch))
				copy(b, batch)
				pipeline <- b
			}
			close(pipeline)
			processorWg.Wait()
			return nil
		case <-timer.C:
			if len(batch) > 0 {
				sendBatch()
			} else {
				timer.Reset(batchTimeout)
			}
		case msg, ok := <-fetchCh:
			if !ok {
				if len(batch) > 0 {
					b := make([]kafka.Message, len(batch))
					copy(b, batch)
					pipeline <- b
				}
				close(pipeline)
				processorWg.Wait()
				return nil
			}
			batch = append(batch, msg)
			if len(batch) >= batchSize {
				sendBatch()
			}
		}
	}
}
