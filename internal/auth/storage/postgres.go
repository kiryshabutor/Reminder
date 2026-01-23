package storage

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kiribu/jwt-practice/models"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateUser(username, password string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	ValidatePassword(username, password string) (*models.User, error)
	SaveRefreshToken(token string, userID int, expiresAt time.Time) error
	ValidateRefreshToken(token string) (int, error)
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

	var user models.User
	err = s.db.QueryRowx(
		`INSERT INTO users (username, password_hash) 
		 VALUES ($1, $2) 
		 RETURNING id, username, password_hash, created_at`,
		username, string(hashedPassword),
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
func (s *PostgresStorage) GetUserByID(id int) (*models.User, error) {
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

func (s *PostgresStorage) SaveRefreshToken(token string, userID int, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	return err
}

func (s *PostgresStorage) ValidateRefreshToken(token string) (int, error) {
	var rt struct {
		UserID    int       `db:"user_id"`
		ExpiresAt time.Time `db:"expires_at"`
	}

	err := s.db.Get(&rt,
		`SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1`,
		token,
	)
	if err != nil {
		return 0, errors.New("token not found")
	}

	if time.Now().After(rt.ExpiresAt) {
		s.DeleteRefreshToken(token)
		return 0, errors.New("token expired")
	}

	return rt.UserID, nil
}

func (s *PostgresStorage) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}
