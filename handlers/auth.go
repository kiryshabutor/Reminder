package handlers

import (
	"encoding/json"
	"net/http"

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

	storage.Store.SaveRefreshToken(refreshToken, user.Username)

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

	username, err := storage.Store.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Невалидный refresh token")
		return
	}

	accessToken, err := utils.GenerateAccessToken(username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	storage.Store.DeleteRefreshToken(req.RefreshToken)
	storage.Store.SaveRefreshToken(newRefreshToken, username)

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
