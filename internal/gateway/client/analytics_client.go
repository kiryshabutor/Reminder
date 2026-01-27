package client

import (
	"context"
	"log"
	"time"

	"github.com/kiribu/jwt-practice/internal/analytics/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AnalyticsClient struct {
	conn   *grpc.ClientConn
	client pb.AnalyticsServiceClient
}

func NewAnalyticsClient(addr string) (*AnalyticsClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Connecting to Analytics Service at %s...", addr)
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to Analytics Service at %s", addr)

	return &AnalyticsClient{conn: conn,
		client: pb.NewAnalyticsServiceClient(conn),
	}, nil
}

func (c *AnalyticsClient) Close() error {
	return c.conn.Close()
}

func (c *AnalyticsClient) GetUserStats(ctx context.Context, userID int64) (*pb.UserStatsResponse, error) {
	req := &pb.GetUserStatsRequest{
		UserId: userID,
	}
	return c.client.GetUserStats(ctx, req)
}
