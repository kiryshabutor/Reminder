package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/kiribu/jwt-practice/internal/analytics/service"
	"github.com/kiribu/jwt-practice/models"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	service *service.AnalyticsService
}

func NewConsumer(brokers []string, topic string, service *service.AnalyticsService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "analytics-service-group", // Unique group for analytics
		MinBytes: 10e3,                      // 10KB
		MaxBytes: 10e6,                      // 10MB
	})

	return &Consumer{
		reader:  reader,
		service: service,
	}
}

func (c *Consumer) Start() {
	slog.Info("Kafka Consumer started...")

	for {
		m, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			slog.Error("Error reading message", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var event models.LifecycleEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			slog.Error("Error unmarshalling event", "error", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = c.service.ProcessEvent(ctx, event)
		cancel()

		if err != nil {
			slog.Error("Error processing event", "error", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
