package storage

import (
	"errors"
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
}

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Create(userID int64, title, description string, remindAt time.Time) (*models.Reminder, error) {
	var reminder models.Reminder
	err := s.db.QueryRowx(
		`INSERT INTO reminders (user_id, title, description, remind_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, title, description, remind_at, is_sent, created_at, updated_at`,
		userID, title, description, remindAt,
	).StructScan(&reminder)

	if err != nil {
		return nil, err
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
	var reminder models.Reminder
	err := s.db.QueryRowx(
		`UPDATE reminders
		 SET title = $1, description = $2, remind_at = $3, updated_at = NOW()
		 WHERE user_id = $4 AND id = $5 AND is_sent = FALSE
		 RETURNING id, user_id, title, description, remind_at, is_sent, created_at, updated_at`,
		title, description, remindAt, userID, id,
	).StructScan(&reminder)

	if err != nil {
		return nil, errors.New("reminder not found or already sent")
	}

	return &reminder, nil
}

func (s *PostgresStorage) Delete(userID, id int64) error {
	result, err := s.db.Exec(
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

	return nil
}
