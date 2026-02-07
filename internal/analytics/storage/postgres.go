package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kiribu/jwt-practice/models"
)

type AnalyticsStorage interface {
	GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error)

	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	IsEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID) (bool, error)
	MarkEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID) error
	IncrementCreated(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, timestamp time.Time) error
	IncrementCompleted(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, timestamp time.Time) error
	IncrementDeleted(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, timestamp time.Time) error
}

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error) {
	var stats models.UserStatistics
	query := `SELECT * FROM analytics.user_statistics WHERE user_id = $1`
	err := s.db.GetContext(ctx, &stats, query, userID)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *PostgresStorage) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return s.db.BeginTxx(ctx, nil)
}

func (s *PostgresStorage) IsEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM analytics.processed_events WHERE event_id = $1)`
	err := tx.GetContext(ctx, &exists, query, eventID)
	return exists, err
}

func (s *PostgresStorage) MarkEventProcessed(ctx context.Context, tx *sqlx.Tx, eventID uuid.UUID) error {
	query := `INSERT INTO analytics.processed_events (event_id) VALUES ($1)`
	_, err := tx.ExecContext(ctx, query, eventID)
	return err
}

func (s *PostgresStorage) IncrementCreated(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, timestamp time.Time) error {
	query := `
		INSERT INTO analytics.user_statistics (user_id, total_reminders_created, active_reminders, first_reminder_at, last_activity_at)
		VALUES ($1, 1, 1, $2, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			total_reminders_created = user_statistics.total_reminders_created + 1,
			active_reminders = user_statistics.active_reminders + 1,
			last_activity_at = $2,
			completion_rate = CASE 
				WHEN user_statistics.total_reminders_created + 1 > 0 
				THEN ROUND((user_statistics.total_reminders_completed::DECIMAL / (user_statistics.total_reminders_created + 1)) * 100, 2)
				ELSE 0 
			END,
			updated_at = NOW()
	`
	_, err := tx.ExecContext(ctx, query, userID, timestamp)
	return err
}

func (s *PostgresStorage) IncrementCompleted(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, timestamp time.Time) error {
	query := `
		UPDATE analytics.user_statistics SET
			total_reminders_completed = total_reminders_completed + 1,
			active_reminders = GREATEST(active_reminders - 1, 0),
			last_activity_at = $2,
			completion_rate = CASE 
				WHEN total_reminders_created > 0 
				THEN ROUND(((total_reminders_completed + 1)::DECIMAL / total_reminders_created) * 100, 2)
				ELSE 0 
			END,
			updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := tx.ExecContext(ctx, query, userID, timestamp)
	return err
}

func (s *PostgresStorage) IncrementDeleted(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, timestamp time.Time) error {
	query := `
		UPDATE analytics.user_statistics SET
			total_reminders_deleted = total_reminders_deleted + 1,
			active_reminders = GREATEST(active_reminders - 1, 0),
			last_activity_at = $2,
			updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := tx.ExecContext(ctx, query, userID, timestamp)
	return err
}
