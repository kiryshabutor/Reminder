package authgrpc

import (
	"context"

	"github.com/kiribu/jwt-practice/internal/auth/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/auth/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthServer(svc *service.AuthService) *AuthServer {
	return &AuthServer{service: svc}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	user, err := s.service.Register(req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	return &pb.RegisterResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokens, err := s.service.Login(req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	tokens, err := s.service.Refresh(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &pb.RefreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: "access_token is required",
		}, nil
	}

	username, userID, err := s.service.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		Username: username,
		UserId:   userID.String(),
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := s.service.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &pb.LogoutResponse{Success: true}, nil
}

func (s *AuthServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	user, err := s.service.GetProfile(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UserResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
