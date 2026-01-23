package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/internal/gateway/client"
	"github.com/kiribu/jwt-practice/internal/gateway/handlers"
)

func init() {
	if err := godotenv.Load(); err != nil {
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

	authHandler := handlers.NewAuthHandler(authClient)

	r := mux.NewRouter()

	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/refresh", authHandler.Refresh).Methods("POST")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(authHandler.AuthMiddleware)
	protected.HandleFunc("/profile", authHandler.Profile).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	port := getEnv("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("API Gateway (HTTP) started on http://localhost:%s\n", port)
	log.Println("Available endpoints:")
	log.Println("   POST   /register  - Registration")
	log.Println("   POST   /login     - Login")
	log.Println("   POST   /refresh   - Token refresh")
	log.Println("   GET    /profile   - Profile (protected)")
	log.Println("   GET    /health    - Health check")

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
