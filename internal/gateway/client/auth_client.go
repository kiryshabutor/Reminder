package client

import (
	"context"
	"time"

	"github.com/kiribu/jwt-practice/internal/auth/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient
type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}
func (c *AuthClient) Register(ctx context.Context, username, password string) (*pb.RegisterResponse, error) {
	return c.client.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})
}

func (c *AuthClient) Login(ctx context.Context, username, password string) (*pb.LoginResponse, error) {
	return c.client.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})
}

func (c *AuthClient) Refresh(ctx context.Context, refreshToken string) (*pb.RefreshResponse, error) {
	return c.client.Refresh(ctx, &pb.RefreshRequest{
		RefreshToken: refreshToken,
	})
}

func (c *AuthClient) ValidateToken(ctx context.Context, accessToken string) (*pb.ValidateTokenResponse, error) {
	return c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		AccessToken: accessToken,
	})
}

func (c *AuthClient) Logout(ctx context.Context, token string) (*pb.LogoutResponse, error) {
	return c.client.Logout(ctx, &pb.LogoutRequest{
		Token: token,
	})
}

func (c *AuthClient) GetProfile(ctx context.Context, username string) (*pb.UserResponse, error) {
	return c.client.GetProfile(ctx, &pb.GetProfileRequest{
		Username: username,
	})
}
