package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/kiribu/jwt-practice/internal/reminder/storage"
)

type NotificationWorker struct {
	storage  storage.ReminderStorage
	interval time.Duration
}

func NewNotificationWorker(storage storage.ReminderStorage, interval time.Duration) *NotificationWorker {
	return &NotificationWorker{
		storage:  storage,
		interval: interval,
	}
}

func (w *NotificationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("Reminder worker started", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping reminder worker...")
			return
		case <-ticker.C:
			w.processPending()
		}
	}
}

func (w *NotificationWorker) processPending() {
	reminders, err := w.storage.GetPending()
	if err != nil {
		slog.Error("Error fetching pending reminders", "error", err)
		return
	}

	if len(reminders) > 0 {
		slog.Info("Found pending reminders, creating outbox events", "count", len(reminders))
	}

	for _, reminder := range reminders {
		if err := w.storage.CreateNotificationEventsAndMarkSent(reminder); err != nil {
			slog.Error("Error creating notification events", "reminder_id", reminder.ID, "error", err)
		} else {
			slog.Debug("Successfully created notification events", "reminder_id", reminder.ID)
		}
	}
}
