package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/config"
	authgrpc "github.com/kiribu/jwt-practice/internal/auth/grpc"
	"github.com/kiribu/jwt-practice/internal/auth/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/auth/service"
	"github.com/kiribu/jwt-practice/internal/auth/storage"
	"github.com/kiribu/jwt-practice/pkg/logger"
	"github.com/kiribu/jwt-practice/pkg/redis"
	"google.golang.org/grpc"
)

func init() {
	_ = godotenv.Load(".env")
}

func main() {
	env := getEnv("APP_ENV", "local")
	logger.Setup(env)

	dbConfig := config.LoadDatabaseConfig()

	db, err := config.ConnectDatabase(dbConfig)
	if err != nil {
		slog.Error("DB connection error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("Auth Service: Successfully connected to PostgreSQL")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisClient, err := redis.NewRedisClient(redisAddr, redisPassword)
	if err != nil {
		slog.Error("Redis connection error", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("Auth Service: Successfully connected to Redis")

	store := storage.NewPostgresStorage(db)
	authService := service.NewAuthService(store, redisClient)
	authServer := authgrpc.NewAuthServer(authService)
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	port := getEnv("GRPC_PORT", "50051")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}

	slog.Info("Auth Service (gRPC) started", "port", port)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Auth Service...")
	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
