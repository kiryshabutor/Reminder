package logger

import (
	"log/slog"
	"os"
)

// env: "local" (text handler) or "prod" (json handler)
func Setup(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case "local":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	case "prod":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
