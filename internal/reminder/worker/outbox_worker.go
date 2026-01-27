package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kiribu/jwt-practice/internal/reminder/kafka"
	"github.com/kiribu/jwt-practice/internal/reminder/storage"
	"github.com/kiribu/jwt-practice/models"
)

type OutboxWorker struct {
	storage              storage.ReminderStorage
	lifecycleProducer    *kafka.Producer
	notificationProducer *kafka.Producer
	interval             time.Duration
	batchSize            int
}

func NewOutboxWorker(
	storage storage.ReminderStorage,
	lifecycleProducer *kafka.Producer,
	notificationProducer *kafka.Producer,
	interval time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		storage:              storage,
		lifecycleProducer:    lifecycleProducer,
		notificationProducer: notificationProducer,
		interval:             interval,
		batchSize:            50,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Outbox Worker started with interval %v", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Outbox Worker...")
			return
		case <-ticker.C:
			w.processOutbox()
		}
	}
}

func (w *OutboxWorker) processOutbox() {
	events, err := w.storage.GetPendingOutboxEvents(w.batchSize)
	if err != nil {
		log.Printf("Error fetching outbox events: %v", err)
		return
	}

	if len(events) > 0 {
		log.Printf("Processing %d outbox events", len(events))
	}

	for _, event := range events {
		if err := w.processEvent(event); err != nil {
			log.Printf("Error processing outbox event %d: %v", event.ID, err)

			if err := w.storage.IncrementOutboxRetryCount(event.ID, err.Error()); err != nil {
				log.Printf("Failed to update retry count for event %d: %v", event.ID, err)
			}
			continue
		}

		if err := w.storage.MarkOutboxEventAsSent(event.ID); err != nil {
			log.Printf("Failed to mark event %d as sent: %v", event.ID, err)
		}
	}
}

func (w *OutboxWorker) processEvent(event storage.OutboxEvent) error {
	key := fmt.Sprintf("%d", event.UserID)

	switch event.EventType {
	case "created", "updated", "deleted", "notification_sent":
		var lifecycleEvent models.LifecycleEvent
		if err := json.Unmarshal(event.Payload, &lifecycleEvent); err != nil {
			return fmt.Errorf("failed to unmarshal lifecycle event: %w", err)
		}

		if err := w.lifecycleProducer.SendEvent(key, lifecycleEvent); err != nil {
			return fmt.Errorf("failed to send to lifecycle topic: %w", err)
		}

		log.Printf("Sent %s event to reminder_lifecycle (event_id=%d, reminder_id=%d)",
			event.EventType, event.ID, event.AggregateID)

	case "notification_trigger":
		var reminder models.Reminder
		if err := json.Unmarshal(event.Payload, &reminder); err != nil {
			return fmt.Errorf("failed to unmarshal reminder: %w", err)
		}

		if err := w.notificationProducer.SendEvent(key, reminder); err != nil {
			return fmt.Errorf("failed to send to notifications topic: %w", err)
		}

		log.Printf("Sent notification_trigger event to notifications (event_id=%d, reminder_id=%d)",
			event.ID, event.AggregateID)

	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}

	return nil
}
