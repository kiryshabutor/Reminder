package models

import (
	"time"

	"github.com/google/uuid"
)

type LifecycleEvent struct {
	EventType  string      `json:"event_type"` // "created", "updated", "deleted", "notification_sent"
	ReminderID uuid.UUID   `json:"reminder_id"`
	UserID     uuid.UUID   `json:"user_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Payload    interface{} `json:"payload,omitempty"` // Reminder snapshot or nil
}
