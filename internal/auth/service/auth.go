package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/internal/auth/storage"
	"github.com/kiribu/jwt-practice/models"
	"github.com/kiribu/jwt-practice/utils"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	store storage.Storage
	redis *redis.Client
}

func NewAuthService(store storage.Storage, redisClient *redis.Client) *AuthService {
	return &AuthService{
		store: store,
		redis: redisClient,
	}
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

func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, uuid.UUID, error) {
	// Check Blacklist
	val, err := s.redis.Get(ctx, "blacklist:"+token).Result()
	if err == nil && val == "revoked" {
		log.Printf("Blacklist hit for token %s", token)
		return "", uuid.Nil, errors.New("token revoked")
	}

	claims, err := utils.ValidateAccessToken(token)
	if err != nil {
		return "", uuid.Nil, err
	}

	// Check User Cache
	cacheKey := "user:" + claims.Username
	val, err = s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit
		log.Printf("Cache hit for user %s", claims.Username)
		var user models.User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return user.Username, user.ID, nil
		}
	}

	// Cache Miss
	log.Printf("Cache miss for user %s", claims.Username)
	user, err := s.store.GetUserByUsername(claims.Username)
	if err != nil {
		return "", uuid.Nil, err
	}

	// Set Cache
	if userJSON, err := json.Marshal(user); err == nil {
		s.redis.Set(ctx, cacheKey, userJSON, utils.AccessTokenDuration)
	}

	return claims.Username, user.ID, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.redis.Set(ctx, "blacklist:"+token, "revoked", utils.AccessTokenDuration).Err()
}
