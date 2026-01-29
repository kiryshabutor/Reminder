package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/internal/notification/kafka"
	"github.com/kiribu/jwt-practice/pkg/logger"
)

func init() {
	_ = godotenv.Load(".env")
}

func main() {
	env := getEnv("APP_ENV", "local")
	logger.Setup(env)

	slog.Info("Starting Notification Service...")

	brokersEnv := getEnv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")

	// Consumer for Notifications (for NotificationWorker)
	topic := getEnv("KAFKA_TOPIC_NOTIFICATIONS", "notifications")
	groupID := getEnv("KAFKA_GROUP_ID", "notification-workers")

	consumer := kafka.NewConsumer(brokers, topic, groupID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.Start(ctx)

	slog.Info("Notification Service started", "topic", topic)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Notification Service...")
	cancel()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
