package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/internal/analytics/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/analytics/service"
	"github.com/kiribu/jwt-practice/models"
)

type AnalyticsServer struct {
	pb.UnimplementedAnalyticsServiceServer
	service *service.AnalyticsService
}

func NewAnalyticsServer(service *service.AnalyticsService) *AnalyticsServer {
	return &AnalyticsServer{service: service}
}

func (s *AnalyticsServer) GetUserStats(ctx context.Context, req *pb.GetUserStatsRequest) (*pb.UserStatsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	stats, err := s.service.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return convertToProto(stats), nil
}

func convertToProto(s *models.UserStatistics) *pb.UserStatsResponse {
	resp := &pb.UserStatsResponse{
		UserId:                  s.UserID.String(),
		TotalRemindersCreated:   s.TotalRemindersCreated,
		TotalRemindersCompleted: s.TotalRemindersCompleted,
		TotalRemindersDeleted:   s.TotalRemindersDeleted,
		ActiveReminders:         s.ActiveReminders,
		CompletionRate:          s.CompletionRate,
	}
	if s.FirstReminderAt != nil {
		resp.FirstReminderAt = s.FirstReminderAt.Format(time.RFC3339)
	}
	if s.LastActivityAt != nil {
		resp.LastActivityAt = s.LastActivityAt.Format(time.RFC3339)
	}
	return resp
}
