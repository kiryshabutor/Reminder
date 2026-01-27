package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kiribu/jwt-practice/models"
)

type ReminderStorage interface {
	Create(userID int64, title, description string, remindAt time.Time) (*models.Reminder, error)
	GetByUserID(userID int64, status string) ([]models.Reminder, error)
	GetByID(userID, id int64) (*models.Reminder, error)
	Update(userID, id int64, title, description string, remindAt time.Time) (*models.Reminder, error)
	Delete(userID, id int64) error
	GetPending() ([]models.Reminder, error)
	MarkAsSent(id int64) error
	// Outbox methods
	GetPendingOutboxEvents(limit int) ([]OutboxEvent, error)
	MarkOutboxEventAsSent(id int64) error
	IncrementOutboxRetryCount(id int64, errMsg string) error
	CreateNotificationEventsAndMarkSent(reminder models.Reminder) error
}

// OutboxEvent represents an event waiting to be sent to Kafka
type OutboxEvent struct {
	ID          int64           `db:"id"`
	EventType   string          `db:"event_type"`
	AggregateID int64           `db:"aggregate_id"`
	UserID      int64           `db:"user_id"`
	Payload     json.RawMessage `db:"payload"`
	RetryCount  int             `db:"retry_count"`
}

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// createOutboxEvent inserts an event into the outbox table within a transaction
func (s *PostgresStorage) createOutboxEvent(tx *sqlx.Tx, eventType string, userID, aggregateID int64, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO reminders_outbox (event_type, aggregate_id, user_id, payload)
		VALUES ($1, $2, $3, $4)`,
		eventType, aggregateID, userID, payloadJSON,
	)
	return err
}

func (s *PostgresStorage) Create(userID int64, title, description string, remindAt time.Time) (*models.Reminder, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var reminder models.Reminder
	err = tx.QueryRowx(`
		INSERT INTO reminders (user_id, title, description, remind_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, description, remind_at, is_sent, created_at, updated_at`,
		userID, title, description, remindAt,
	).StructScan(&reminder)
	if err != nil {
		return nil, fmt.Errorf("failed to insert reminder: %w", err)
	}

	event := models.LifecycleEvent{
		EventType:  "created",
		ReminderID: reminder.ID,
		UserID:     reminder.UserID,
		Timestamp:  time.Now(),
		Payload:    reminder,
	}
	if err := s.createOutboxEvent(tx, "created", reminder.UserID, reminder.ID, event); err != nil {
		return nil, fmt.Errorf("failed to create outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &reminder, nil
}

func (s *PostgresStorage) GetByUserID(userID int64, status string) ([]models.Reminder, error) {
	var reminders []models.Reminder
	var query string
	var args []interface{}

	baseQuery := `SELECT id, user_id, title, description, remind_at, is_sent, created_at, updated_at
		 FROM reminders WHERE user_id = $1`

	switch status {
	case "pending":
		query = baseQuery + " AND is_sent = FALSE ORDER BY remind_at ASC"
		args = []interface{}{userID}
	case "sent":
		query = baseQuery + " AND is_sent = TRUE ORDER BY remind_at DESC"
		args = []interface{}{userID}
	default:
		query = baseQuery + " ORDER BY remind_at ASC"
		args = []interface{}{userID}
	}

	err := s.db.Select(&reminders, query, args...)
	if err != nil {
		return nil, err
	}

	return reminders, nil
}

func (s *PostgresStorage) GetByID(userID, id int64) (*models.Reminder, error) {
	var reminder models.Reminder
	err := s.db.Get(&reminder,
		`SELECT id, user_id, title, description, remind_at, is_sent, created_at, updated_at
		 FROM reminders WHERE user_id = $1 AND id = $2`,
		userID, id,
	)
	if err != nil {
		return nil, errors.New("reminder not found")
	}

	return &reminder, nil
}

func (s *PostgresStorage) Update(userID, id int64, title, description string, remindAt time.Time) (*models.Reminder, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var reminder models.Reminder
	err = tx.QueryRowx(`
		UPDATE reminders
		 SET title = $1, description = $2, remind_at = $3, updated_at = NOW()
		 WHERE user_id = $4 AND id = $5 AND is_sent = FALSE
		 RETURNING id, user_id, title, description, remind_at, is_sent, created_at, updated_at`,
		title, description, remindAt, userID, id,
	).StructScan(&reminder)
	if err != nil {
		return nil, errors.New("reminder not found or already sent")
	}

	event := models.LifecycleEvent{
		EventType:  "updated",
		ReminderID: reminder.ID,
		UserID:     reminder.UserID,
		Timestamp:  time.Now(),
		Payload:    reminder,
	}
	if err := s.createOutboxEvent(tx, "updated", reminder.UserID, reminder.ID, event); err != nil {
		return nil, fmt.Errorf("failed to create outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &reminder, nil
}

func (s *PostgresStorage) Delete(userID, id int64) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`DELETE FROM reminders WHERE user_id = $1 AND id = $2 AND is_sent = FALSE`,
		userID, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("reminder not found or already sent")
	}

	event := models.LifecycleEvent{
		EventType:  "deleted",
		ReminderID: id,
		UserID:     userID,
		Timestamp:  time.Now(),
		Payload:    nil,
	}
	if err := s.createOutboxEvent(tx, "deleted", userID, id, event); err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetPending() ([]models.Reminder, error) {
	var reminders []models.Reminder
	err := s.db.Select(&reminders,
		`SELECT id, user_id, title, description, remind_at, is_sent, created_at, updated_at
		 FROM reminders 
		 WHERE is_sent = FALSE AND remind_at <= NOW()`,
	)
	if err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *PostgresStorage) MarkAsSent(id int64) error {
	_, err := s.db.Exec(
		`UPDATE reminders SET is_sent = TRUE, updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

func (s *PostgresStorage) GetPendingOutboxEvents(limit int) ([]OutboxEvent, error) {
	var events []OutboxEvent
	err := s.db.Select(&events, `
		SELECT id, event_type, aggregate_id, user_id, payload, retry_count
		FROM reminders_outbox
		WHERE status = 'PENDING' AND retry_count < 5
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`,
		limit,
	)
	return events, err
}

func (s *PostgresStorage) MarkOutboxEventAsSent(id int64) error {
	_, err := s.db.Exec(`
		UPDATE reminders_outbox
		SET status = 'SENT', processed_at = NOW()
		WHERE id = $1`,
		id,
	)
	return err
}

func (s *PostgresStorage) IncrementOutboxRetryCount(id int64, errMsg string) error {
	_, err := s.db.Exec(`
		UPDATE reminders_outbox
		SET retry_count = retry_count + 1,
		    error_message = $2,
		    status = CASE WHEN retry_count + 1 >= 5 THEN 'FAILED' ELSE status END
		WHERE id = $1`,
		id, errMsg,
	)
	return err
}

func (s *PostgresStorage) CreateNotificationEventsAndMarkSent(reminder models.Reminder) error {
	// Begin transaction
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	payloadJSON, err := json.Marshal(reminder)
	if err != nil {
		return fmt.Errorf("failed to marshal reminder: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO reminders_outbox (event_type, aggregate_id, user_id, payload)
		VALUES ('notification_trigger', $1, $2, $3)`,
		reminder.ID, reminder.UserID, payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification_trigger event: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO reminders_outbox (event_type, aggregate_id, user_id, payload)
		VALUES ('notification_sent', $1, $2, $3)`,
		reminder.ID, reminder.UserID, payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification_sent event: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE reminders 
		SET is_sent = TRUE, updated_at = NOW() 
		WHERE id = $1`,
		reminder.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark reminder as sent: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
