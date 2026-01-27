package worker

import (
	"context"
	"log"
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

	log.Printf("Reminder worker started with interval %v", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping reminder worker...")
			return
		case <-ticker.C:
			w.processPending()
		}
	}
}

func (w *NotificationWorker) processPending() {
	reminders, err := w.storage.GetPending()
	if err != nil {
		log.Printf("Error fetching pending reminders: %v", err)
		return
	}

	if len(reminders) > 0 {
		log.Printf("Found %d pending reminders, creating outbox events", len(reminders))
	}

	for _, reminder := range reminders {
		if err := w.storage.CreateNotificationEventsAndMarkSent(reminder); err != nil {
			log.Printf("Error creating notification events for reminder %d: %v", reminder.ID, err)
		} else {
			log.Printf("Successfully created notification events for reminder %d", reminder.ID)
		}
	}
}
