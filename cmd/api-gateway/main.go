package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/kiribu/jwt-practice/internal/gateway/handlers"
	customMiddleware "github.com/kiribu/jwt-practice/internal/gateway/middleware"
	"github.com/kiribu/jwt-practice/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	_ = godotenv.Load(".env")
}

func main() {
	env := getEnv("APP_ENV", "local")
	logger.Setup(env)

	// Connect to Auth Service
	authServiceAddr := getEnv("AUTH_SERVICE_ADDR", "auth-service:50051")
	authClient, err := client.NewAuthClient(authServiceAddr)
	if err != nil {
		slog.Error("Failed to connect to Auth Service", "error", err)
		os.Exit(1)
	}
	defer authClient.Close()
	slog.Info("API Gateway: Connected to Auth Service", "addr", authServiceAddr)

	// Connect to Reminder Service
	reminderServiceAddr := getEnv("REMINDER_SERVICE_ADDR", "reminder-service:50052")
	reminderClient, err := client.NewReminderClient(reminderServiceAddr)
	if err != nil {
		slog.Error("Failed to connect to Reminder Service", "error", err)
		os.Exit(1)
	}
	defer reminderClient.Close()
	slog.Info("API Gateway: Connected to Reminder Service", "addr", reminderServiceAddr)

	// Connect to Analytics Service
	analyticsServiceAddr := getEnv("ANALYTICS_SERVICE_ADDR", "analytics-service:50053")
	analyticsClient, err := client.NewAnalyticsClient(analyticsServiceAddr)
	if err != nil {
		slog.Error("Failed to connect to Analytics Service", "error", err)
		os.Exit(1)
	}
	defer analyticsClient.Close()
	slog.Info("API Gateway: Connected to Analytics Service", "addr", analyticsServiceAddr)

	authHandler := handlers.NewAuthHandler(authClient)
	reminderHandler := handlers.NewReminderHandler(reminderClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsClient)

	e := echo.New()
	e.HideBanner = true

	e.Use(customMiddleware.SlogLogger)
	e.Use(middleware.Recover())

	e.POST("/auth/register", authHandler.Register)
	e.POST("/auth/login", authHandler.Login)
	e.POST("/auth/refresh", authHandler.Refresh)

	protected := e.Group("")
	protected.Use(authHandler.AuthMiddleware)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/profile", authHandler.Profile)

	protected.POST("/reminders", reminderHandler.Create)
	protected.GET("/reminders", reminderHandler.List)
	protected.GET("/reminders/:id", reminderHandler.Get)
	protected.PUT("/reminders/:id", reminderHandler.Update)
	protected.DELETE("/reminders/:id", reminderHandler.Delete)

	protected.GET("/analytics/me", analyticsHandler.GetStats)

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	port := getEnv("HTTP_PORT", "8080")

	slog.Info("API Gateway (HTTP) started", "port", port)
	slog.Info("Available endpoints:", "endpoints", []string{
		"POST   /register",
		"POST   /login",
		"POST   /refresh",
		"POST   /auth/logout",
		"GET    /profile",
		"POST   /reminders",
		"GET    /reminders",
		"GET    /reminders/:id",
		"PUT    /reminders/:id",
		"DELETE /reminders/:id",
		"GET    /analytics/me",
		"GET    /health",
	})

	go func() {
		if err := e.Start(":" + port); err != nil {
			slog.Info("HTTP server stopped", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(ctx)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
