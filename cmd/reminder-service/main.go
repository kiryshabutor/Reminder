package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kiribu/jwt-practice/config"
	remindergrpc "github.com/kiribu/jwt-practice/internal/reminder/grpc"
	"github.com/kiribu/jwt-practice/internal/reminder/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/reminder/kafka"
	"github.com/kiribu/jwt-practice/internal/reminder/service"
	"github.com/kiribu/jwt-practice/internal/reminder/storage"
	"github.com/kiribu/jwt-practice/internal/reminder/worker"
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

	slog.Info("Reminder Service: Successfully connected to PostgreSQL")

	store := storage.NewPostgresStorage(db)

	brokersEnv := getEnv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")

	// Producer for Notifications (for NotificationWorker)
	notificationTopic := getEnv("KAFKA_TOPIC_NOTIFICATIONS", "notifications")
	notificationProducer := kafka.NewProducer(brokers, notificationTopic)
	defer func() {
		if err := notificationProducer.Close(); err != nil {
			slog.Error("Failed to close notification producer", "error", err)
		}
	}()

	// Producer for Lifecycle Events (for OutboxWorker)
	lifecycleTopic := getEnv("KAFKA_TOPIC_LIFECYCLE", "reminder_lifecycle")
	lifecycleProducer := kafka.NewProducer(brokers, lifecycleTopic)
	defer func() {
		if err := lifecycleProducer.Close(); err != nil {
			slog.Error("Failed to close lifecycle producer", "error", err)
		}
	}()

	reminderService := service.NewReminderService(store)
	reminderServer := remindergrpc.NewReminderServer(reminderService)

	grpcServer := grpc.NewServer()
	pb.RegisterReminderServiceServer(grpcServer, reminderServer)

	port := getEnv("REMINDER_GRPC_PORT", "50052")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Failed to start listener", "error", err)
		os.Exit(1)
	}

	slog.Info("Reminder Service (gRPC) started", "port", port)

	intervalStr := getEnv("WORKER_INTERVAL", "5s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		slog.Error("Invalid WORKER_INTERVAL", "error", err)
		os.Exit(1)
	}

	notificationWorker := worker.NewNotificationWorker(store, interval)
	outboxWorker := worker.NewOutboxWorker(store, lifecycleProducer, notificationProducer, 500*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go notificationWorker.Start(ctx)
	go outboxWorker.Start(ctx)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down Reminder Service...")

	cancel()

	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
