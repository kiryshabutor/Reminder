package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kiribu/jwt-practice/models"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateUser(username, password string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	ValidatePassword(username, password string) (*models.User, error)
	SaveRefreshToken(token string, userID uuid.UUID, expiresAt time.Time) error
	ValidateRefreshToken(token string) (uuid.UUID, error)
	DeleteRefreshToken(token string) error
}

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) CreateUser(username, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := uuid.Must(uuid.NewV7())
	var user models.User
	err = s.db.QueryRowx(
		`INSERT INTO users (id, username, password_hash) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, username, password_hash, created_at`,
		userID, username, string(hashedPassword),
	).StructScan(&user)

	if err != nil {
		return nil, errors.New("user with this username already exists")
	}

	return &user, nil
}

func (s *PostgresStorage) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
func (s *PostgresStorage) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *PostgresStorage) ValidatePassword(username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

func (s *PostgresStorage) SaveRefreshToken(token string, userID uuid.UUID, expiresAt time.Time) error {
	tokenID := uuid.Must(uuid.NewV7())
	_, err := s.db.Exec(
		`INSERT INTO refresh_tokens (id, token, user_id, expires_at) VALUES ($1, $2, $3, $4)`,
		tokenID, token, userID, expiresAt,
	)
	return err
}

func (s *PostgresStorage) ValidateRefreshToken(token string) (uuid.UUID, error) {
	var rt struct {
		UserID    uuid.UUID `db:"user_id"`
		ExpiresAt time.Time `db:"expires_at"`
	}

	err := s.db.Get(&rt,
		`SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1`,
		token,
	)
	if err != nil {
		return uuid.Nil, errors.New("token not found")
	}

	if time.Now().After(rt.ExpiresAt) {
		s.DeleteRefreshToken(token)
		return uuid.Nil, errors.New("token expired")
	}

	return rt.UserID, nil
}

func (s *PostgresStorage) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}
