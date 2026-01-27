package handlers

import (
	"fmt"
	"net/http"

	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	analyticsClient *client.AnalyticsClient
}

func NewAnalyticsHandler(analyticsClient *client.AnalyticsClient) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsClient: analyticsClient}
}

func (h *AnalyticsHandler) GetMyStats(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	stats, err := h.analyticsClient.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to fetch stats: %v", err)})
	}

	return c.JSON(http.StatusOK, stats)
}
