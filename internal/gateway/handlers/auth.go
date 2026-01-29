package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authClient *client.AuthClient
}

func NewAuthHandler(authClient *client.AuthClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Register(ctx, creds.Username, creds.Password)
	if err != nil {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Login(ctx, creds.Username, creds.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Profile(c echo.Context) error {
	username := c.Get("username").(string)

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	userProfile, err := h.authClient.GetProfile(ctx, username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch profile"})
	}

	return c.JSON(http.StatusOK, userProfile)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header is required"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid Authorization header format"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	_, err := h.authClient.Logout(ctx, parts[1])
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to logout"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Successfully logged out"})
}

// AuthMiddleware validates the JWT token
func (h *AuthHandler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header is required"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid Authorization header format"})
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()

		resp, err := h.authClient.ValidateToken(ctx, parts[1])
		if err != nil || !resp.Valid {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		}

		// Add username and user_id to context
		c.Set("username", resp.Username)
		c.Set("user_id", resp.UserId)
		return next(c)
	}
}
