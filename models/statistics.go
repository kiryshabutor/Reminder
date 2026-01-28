package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatistics struct {
	UserID                  uuid.UUID  `db:"user_id" json:"user_id"`
	TotalRemindersCreated   int64      `db:"total_reminders_created" json:"total_reminders_created"`
	TotalRemindersCompleted int64      `db:"total_reminders_completed" json:"total_reminders_completed"`
	TotalRemindersDeleted   int64      `db:"total_reminders_deleted" json:"total_reminders_deleted"`
	ActiveReminders         int64      `db:"active_reminders" json:"active_reminders"`
	CompletionRate          float64    `db:"completion_rate" json:"completion_rate"`
	FirstReminderAt         *time.Time `db:"first_reminder_at" json:"first_reminder_at"`
	LastActivityAt          *time.Time `db:"last_activity_at" json:"last_activity_at"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}
