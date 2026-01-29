package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

	slog.Info("Outbox Worker started", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Outbox Worker...")
			return
		case <-ticker.C:
			w.processOutbox()
		}
	}
}

func (w *OutboxWorker) processOutbox() {
	events, err := w.storage.GetPendingOutboxEvents(w.batchSize)
	if err != nil {
		slog.Error("Error fetching outbox events", "error", err)
		return
	}

	if len(events) > 0 {
		slog.Info("Processing outbox events", "count", len(events))
	}

	for _, event := range events {
		if err := w.processEvent(event); err != nil {
			slog.Error("Error processing outbox event", "event_id", event.ID, "error", err)

			if err := w.storage.IncrementOutboxRetryCount(event.ID, err.Error()); err != nil {
				slog.Error("Failed to update retry count", "event_id", event.ID, "error", err)
			}
			continue
		}

		if err := w.storage.MarkOutboxEventAsSent(event.ID); err != nil {
			slog.Error("Failed to mark event as sent", "event_id", event.ID, "error", err)
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

		slog.Debug("Sent event to reminder_lifecycle", "type", event.EventType, "event_id", event.ID, "reminder_id", event.AggregateID)

	case "notification_trigger":
		var reminder models.Reminder
		if err := json.Unmarshal(event.Payload, &reminder); err != nil {
			return fmt.Errorf("failed to unmarshal reminder: %w", err)
		}

		if err := w.notificationProducer.SendEvent(key, reminder); err != nil {
			return fmt.Errorf("failed to send to notifications topic: %w", err)
		}

		slog.Debug("Sent notification_trigger event", "event_id", event.ID, "reminder_id", event.AggregateID)

	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}

	return nil
}
