package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kiribu/jwt-practice/models"
	"github.com/kiribu/jwt-practice/storage"
	"github.com/kiribu/jwt-practice/utils"
)

// POST /register
// Body: {"username": "...", "password": "..."}
// Response: {"id": 1, "username": "...", "created_at": "..."}
func Register(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if creds.Username == "" || creds.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username и password обязательны")
		return
	}

	user, err := storage.Store.CreateUser(creds.Username, creds.Password)
	if err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

// POST /login
// Body: {"username": "...", "password": "..."}
// Response: {"access_token": "...", "refresh_token": "...", "token_type": "Bearer"}
func Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	user, err := storage.Store.ValidatePassword(creds.Username, creds.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Неверные учетные данные")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	// Сохраняем refresh token с временем истечения
	expiresAt := time.Now().Add(utils.RefreshTokenDuration)
	if err := storage.Store.SaveRefreshToken(refreshToken, user.ID, expiresAt); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка сохранения токена")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// POST /refresh
// Body: {"refresh_token": "..."}
// Response: {"access_token": "...", "refresh_token": "...", "token_type": "Bearer"}
func Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// Теперь возвращает userID вместо username
	userID, err := storage.Store.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Невалидный refresh token")
		return
	}

	// Получаем пользователя по ID
	user, err := storage.Store.GetUserByID(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения пользователя")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	// Удаляем старый и сохраняем новый токен
	storage.Store.DeleteRefreshToken(req.RefreshToken)
	expiresAt := time.Now().Add(utils.RefreshTokenDuration)
	if err := storage.Store.SaveRefreshToken(newRefreshToken, user.ID, expiresAt); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка сохранения токена")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}

	respondWithJSON(w, http.StatusOK, response)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.ErrorResponse{Error: message})
}
