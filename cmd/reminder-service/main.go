package main

import (
	"context"
	"log"
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
	"google.golang.org/grpc"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env file not found, using system environment variables")
	}
}

func main() {
	dbConfig := config.LoadDatabaseConfig()

	db, err := config.ConnectDatabase(dbConfig)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	log.Println("Reminder Service: Successfully connected to PostgreSQL")

	store := storage.NewPostgresStorage(db)

	// Kafka & Worker
	brokersEnv := getEnv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")

	// Producer for Notifications (for Worker)
	notificationTopic := getEnv("KAFKA_TOPIC_NOTIFICATIONS", "notifications")
	notificationProducer := kafka.NewProducer(brokers, notificationTopic)
	defer func() {
		if err := notificationProducer.Close(); err != nil {
			log.Printf("Failed to close notification producer: %v", err)
		}
	}()

	// Producer for Lifecycle Events (for Service)
	lifecycleTopic := getEnv("KAFKA_TOPIC_LIFECYCLE", "reminder_lifecycle")
	lifecycleProducer := kafka.NewProducer(brokers, lifecycleTopic)
	defer func() {
		if err := lifecycleProducer.Close(); err != nil {
			log.Printf("Failed to close lifecycle producer: %v", err)
		}
	}()

	reminderService := service.NewReminderService(store, lifecycleProducer)
	reminderServer := remindergrpc.NewReminderServer(reminderService)

	grpcServer := grpc.NewServer()
	pb.RegisterReminderServiceServer(grpcServer, reminderServer)

	port := getEnv("REMINDER_GRPC_PORT", "50052")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	log.Printf("Reminder Service (gRPC) started on port %s\n", port)

	intervalStr := getEnv("WORKER_INTERVAL", "5s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatalf("Invalid WORKER_INTERVAL: %v", err)
	}

	workerInstance := worker.NewWorker(store, notificationProducer, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workerInstance.Start(ctx)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Reminder Service...")

	cancel()

	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
