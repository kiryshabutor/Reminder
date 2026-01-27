package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/kiribu/jwt-practice/internal/gateway/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env file not found, using system environment variables")
	}
}

func main() {
	// Connect to Auth Service
	authServiceAddr := getEnv("AUTH_SERVICE_ADDR", "auth-service:50051")
	authClient, err := client.NewAuthClient(authServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	defer authClient.Close()
	log.Printf("API Gateway: Connected to Auth Service (%s)\n", authServiceAddr)

	// Connect to Reminder Service
	reminderServiceAddr := getEnv("REMINDER_SERVICE_ADDR", "reminder-service:50052")
	reminderClient, err := client.NewReminderClient(reminderServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Reminder Service: %v", err)
	}
	defer reminderClient.Close()
	log.Printf("API Gateway: Connected to Reminder Service (%s)\n", reminderServiceAddr)

	// Connect to Analytics Service
	analyticsServiceAddr := getEnv("ANALYTICS_SERVICE_ADDR", "analytics-service:50053")
	analyticsClient, err := client.NewAnalyticsClient(analyticsServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Analytics Service: %v", err)
	}
	defer analyticsClient.Close()
	log.Printf("API Gateway: Connected to Analytics Service (%s)\n", analyticsServiceAddr)

	authHandler := handlers.NewAuthHandler(authClient)
	reminderHandler := handlers.NewReminderHandler(reminderClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsClient)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/register", authHandler.Register)
	e.POST("/login", authHandler.Login)
	e.POST("/refresh", authHandler.Refresh)

	protected := e.Group("")
	protected.Use(authHandler.AuthMiddleware)
	protected.GET("/profile", authHandler.Profile)

	protected.POST("/reminders", reminderHandler.Create)
	protected.GET("/reminders", reminderHandler.List)
	protected.GET("/reminders/:id", reminderHandler.Get)
	protected.PUT("/reminders/:id", reminderHandler.Update)
	protected.DELETE("/reminders/:id", reminderHandler.Delete)

	protected.GET("/analytics/me", analyticsHandler.GetMyStats)

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	port := getEnv("HTTP_PORT", "8080")

	log.Printf("API Gateway (HTTP) started on http://localhost:%s\n", port)
	log.Println("Available endpoints:")
	log.Println("   POST   /register        - Registration")
	log.Println("   POST   /login           - Login")
	log.Println("   POST   /refresh         - Token refresh")
	log.Println("   GET    /profile         - Profile (protected)")
	log.Println("   POST   /reminders       - Create reminder (protected)")
	log.Println("   GET    /reminders       - List reminders (protected)")
	log.Println("   GET    /reminders/:id   - Get reminder (protected)")
	log.Println("   PUT    /reminders/:id   - Update reminder (protected)")
	log.Println("   DELETE /reminders/:id   - Delete reminder (protected)")
	log.Println("   GET    /analytics/me    - Get user stats (protected)")
	log.Println("   GET    /health          - Health check")

	go func() {
		if err := e.Start(":" + port); err != nil {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")
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
