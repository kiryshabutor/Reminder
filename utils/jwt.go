package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessSecret []byte
	once         sync.Once
)

const (
	AccessTokenDuration  = 2 * time.Minute
	RefreshTokenDuration = 2 * time.Hour
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// getAccessSecret возвращает JWT секрет с ленивой инициализацией
func getAccessSecret() []byte {
	once.Do(func() {
		accessSecret = getJWTSecret()
	})
	return accessSecret
}

func GenerateAccessToken(username string) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getAccessSecret())
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи")
		}
		return getAccessSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("невалидный токен")
	}

	return claims, nil
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required but not set")
	}
	return []byte(secret)
}
