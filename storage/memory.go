package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/kiribu/jwt-practice/models"
	"golang.org/x/crypto/bcrypt"
)

type MemoryStorage struct {
	users         map[string]*models.User // username -> User
	refreshTokens map[string]string       // refreshToken -> username
	nextID        int
	mu            sync.RWMutex
}

var Store *MemoryStorage

func init() {
	Store = &MemoryStorage{
		users:         make(map[string]*models.User),
		refreshTokens: make(map[string]string),
		nextID:        1,
	}
}

func (s *MemoryStorage) CreateUser(username, password string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[username]; exists {
		return nil, errors.New("пользователь уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           s.nextID,
		Username:     username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	s.users[username] = user
	s.nextID++

	return user, nil
}

func (s *MemoryStorage) GetUser(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("пользователь не найден")
	}

	return user, nil
}

func (s *MemoryStorage) ValidatePassword(username, password string) (*models.User, error) {
	user, err := s.GetUser(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	return user, nil
}

func (s *MemoryStorage) SaveRefreshToken(token, username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[token] = username
}

func (s *MemoryStorage) ValidateRefreshToken(token string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	username, exists := s.refreshTokens[token]
	if !exists {
		return "", errors.New("невалидный refresh token")
	}

	return username, nil
}

func (s *MemoryStorage) DeleteRefreshToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, token)
}
