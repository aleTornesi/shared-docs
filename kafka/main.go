package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	brokers := []string{getEnv("KAFKA_BROKERS", "kafka:9092")}
	topic := getEnv("KAFKA_TOPIC", "doc-edits")
	groupID := getEnv("KAFKA_GROUP_ID", "batch-processor")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@db:5432/postgres?sslmode=disable")

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer dbConn.Close()

	processor := NewProcessor(dbConn)
	consumer := NewConsumer(brokers, topic, groupID, processor)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("starting kafka batch processor")
	if err := consumer.Run(ctx); err != nil {
		log.Fatal("consumer:", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
