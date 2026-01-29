package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func SlogLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		duration := time.Since(start)
		status := c.Response().Status

		attrs := []any{
			"method", c.Request().Method,
			"path", c.Request().URL.Path,
			"status", status,
			"duration", duration.String(),
			"ip", c.RealIP(),
		}

		if err != nil {
			attrs = append(attrs, "error", err.Error())
			// Log as Error if status is 5xx or there is an error
			if status >= 500 {
				slog.Error("HTTP Request Failed", attrs...)
			} else {
				slog.Warn("HTTP Request Error", attrs...)
			}
		} else {
			// Log as Info for successful requests
			slog.Info("HTTP Request", attrs...)
		}

		return err
	}
}
