package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/internal/notification/kafka"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env file not found, using system environment variables")
	}
}

func main() {
	log.Println("Starting Notification Service...")

	brokersEnv := getEnv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")
	topic := getEnv("KAFKA_TOPIC", "notifications")
	groupID := getEnv("KAFKA_GROUP_ID", "notification-workers")

	consumer := kafka.NewConsumer(brokers, topic, groupID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.Start(ctx)

	log.Printf("Notification Service started. Listening on topic: %s", topic)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Notification Service...")
	cancel()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
