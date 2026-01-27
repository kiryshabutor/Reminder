package models

import "time"

type LifecycleEvent struct {
	EventType  string      `json:"event_type"` // "created", "updated", "deleted", "notification_sent"
	ReminderID int64       `json:"reminder_id"`
	UserID     int64       `json:"user_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Payload    interface{} `json:"payload,omitempty"` // Reminder snapshot or nil
}
