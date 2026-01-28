package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	analyticsClient *client.AnalyticsClient
}

func NewAnalyticsHandler(analyticsClient *client.AnalyticsClient) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsClient: analyticsClient}
}

func (h *AnalyticsHandler) GetStats(c echo.Context) error {
	userID := c.Get("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	resp, err := h.analyticsClient.GetUserStats(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
