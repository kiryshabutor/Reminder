package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/kiribu/jwt-practice/models"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{reader: reader}
}

func (c *Consumer) Start(ctx context.Context) {
	slog.Info("Starting Kafka consumer", "topic", c.reader.Config().Topic)
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Kafka consumer...")
			return
		default:
			c.processMessage(ctx)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		slog.Error("Error reading message", "error", err)
		return
	}

	c.handlePayload(m.Value)
}

func (c *Consumer) handlePayload(data []byte) {
	var reminder models.Reminder
	if err := json.Unmarshal(data, &reminder); err != nil {
		slog.Error("Failed to unmarshal reminder", "error", err, "data", string(data))
		return
	}

	slog.Info("[NOTIFICATION] Sending reminder",
		"user_id", reminder.UserID,
		"title", reminder.Title,
		"desc", reminder.Description)
}
