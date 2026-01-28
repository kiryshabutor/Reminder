package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/internal/auth/storage"
	"github.com/kiribu/jwt-practice/utils"
)

type AuthService struct {
	store storage.Storage
}

func NewAuthService(store storage.Storage) *AuthService {
	return &AuthService{store: store}
}

type UserResponse struct {
	ID        uuid.UUID
	Username  string
	CreatedAt time.Time
}
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
}

func (s *AuthService) Register(username, password string) (*UserResponse, error) {
	user, err := s.store.CreateUser(username, password)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *AuthService) Login(username, password string) (*TokenResponse, error) {
	user, err := s.store.ValidatePassword(username, password)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(user.Username, user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(utils.RefreshTokenDuration)
	if err := s.store.SaveRefreshToken(refreshToken, user.ID, expiresAt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (*TokenResponse, error) {
	userID, err := s.store.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(user.Username, user.ID.String())
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	s.store.DeleteRefreshToken(refreshToken)
	expiresAt := time.Now().Add(utils.RefreshTokenDuration)
	if err := s.store.SaveRefreshToken(newRefreshToken, user.ID, expiresAt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) ValidateToken(token string) (string, uuid.UUID, error) {
	claims, err := utils.ValidateAccessToken(token)
	if err != nil {
		return "", uuid.Nil, err
	}

	user, err := s.store.GetUserByUsername(claims.Username)
	if err != nil {
		return "", uuid.Nil, err
	}

	return claims.Username, user.ID, nil
}
