package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/labstack/echo/v4"
)

type ReminderHandler struct {
	reminderClient *client.ReminderClient
}

func NewReminderHandler(reminderClient *client.ReminderClient) *ReminderHandler {
	return &ReminderHandler{
		reminderClient: reminderClient,
	}
}

type CreateReminderRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	RemindAt    string `json:"remind_at"`
}

type UpdateReminderRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	RemindAt    string `json:"remind_at"`
}

func (h *ReminderHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(int64)

	var req CreateReminderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.reminderClient.Create(ctx, userID, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *ReminderHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(int64)

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.reminderClient.GetAll(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, resp.Reminders)
}

func (h *ReminderHandler) Get(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid reminder ID"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.reminderClient.GetByID(ctx, userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Reminder not found"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *ReminderHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid reminder ID"})
	}

	var req UpdateReminderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.reminderClient.Update(ctx, userID, id, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *ReminderHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(int64)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid reminder ID"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.reminderClient.Delete(ctx, userID, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	if !resp.Success {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: resp.Message})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": resp.Message})
}
