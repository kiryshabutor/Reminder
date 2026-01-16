package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kiribu/jwt-practice/middleware"
	"github.com/kiribu/jwt-practice/storage"
)

type ProtectedData struct {
	Message  string `json:"message"`
	Username string `json:"username"`
	Secret   string `json:"secret"`
}

// GET /protected
// Header: Authorization: Bearer <access_token>
// Response: {"message": "...", "username": "...", "secret": "..."}
func Protected(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetUsernameFromContext(r.Context())

	data := ProtectedData{
		Message:  "Это защищенные данные!",
		Username: username,
		Secret:   "Секретная информация, доступная только авторизованным пользователям",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GET /profile
// Header: Authorization: Bearer <access_token>
// Response: {"id": "...", "username": "...", "created_at": "..."}
func Profile(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetUsernameFromContext(r.Context())

	user, err := storage.Store.GetUser(username)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}
