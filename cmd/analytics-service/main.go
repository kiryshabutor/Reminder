package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/config"
	analyticsgrpc "github.com/kiribu/jwt-practice/internal/analytics/grpc"
	"github.com/kiribu/jwt-practice/internal/analytics/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/analytics/kafka"
	"github.com/kiribu/jwt-practice/internal/analytics/service"
	"github.com/kiribu/jwt-practice/internal/analytics/storage"
	"github.com/kiribu/jwt-practice/pkg/logger"
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

	slog.Info("Analytics Service: Successfully connected to PostgreSQL")

	store := storage.NewPostgresStorage(db)
	analyticsService := service.NewAnalyticsService(store)
	analyticsServer := analyticsgrpc.NewAnalyticsServer(analyticsService)

	brokersEnv := getEnv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")
	lifecycleTopic := getEnv("KAFKA_TOPIC_LIFECYCLE", "reminder_lifecycle")

	consumer := kafka.NewConsumer(brokers, lifecycleTopic, analyticsService)
	go consumer.Start()
	defer consumer.Close()

	grpcPort := getEnv("ANALYTICS_GRPC_PORT", "50053")
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(grpcServer, analyticsServer)

	slog.Info("Analytics Service (gRPC) started", "port", grpcPort)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Analytics Service...")
	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
