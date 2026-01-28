package main

import (
	"log"
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

	log.Println("Auth Service: Successfully connected to PostgreSQL")

	store := storage.NewPostgresStorage(db)
	authService := service.NewAuthService(store)
	authServer := authgrpc.NewAuthServer(authService)
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	port := getEnv("GRPC_PORT", "50051")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	log.Printf("Auth Service (gRPC) started on port %s\n", port)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Auth Service...")
	grpcServer.GracefulStop()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
