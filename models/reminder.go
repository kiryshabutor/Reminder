package models

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	RemindAt    time.Time `db:"remind_at" json:"remind_at"`
	IsSent      bool      `db:"is_sent" json:"is_sent"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
