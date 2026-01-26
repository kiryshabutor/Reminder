package kafka

import (
	"context"
	"encoding/json"
	"log"

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
	log.Printf("Starting Kafka consumer for topic: %s", c.reader.Config().Topic)
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
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
		log.Printf("Error reading message: %v", err)
		return
	}

	c.handlePayload(m.Value)
}

func (c *Consumer) handlePayload(data []byte) {
	var reminder models.Reminder
	if err := json.Unmarshal(data, &reminder); err != nil {
		log.Printf("Failed to unmarshal reminder: %v | Data: %s", err, string(data))
		return
	}

	log.Printf("[NOTIFICATION] Sending reminder to UserID %d: %s (Desc: %s)",
		reminder.UserID, reminder.Title, reminder.Description)
}
