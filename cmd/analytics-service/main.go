package main

import (
	"log"
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

	log.Println("Analytics Service: Successfully connected to PostgreSQL")

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
		log.Fatalf("Failed to start listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(grpcServer, analyticsServer)

	log.Printf("Analytics Service (gRPC) started on port %s\n", grpcPort)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Analytics Service...")
	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
