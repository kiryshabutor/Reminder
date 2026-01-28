package remindergrpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/internal/reminder/grpc/pb"
	"github.com/kiribu/jwt-practice/internal/reminder/service"
	"github.com/kiribu/jwt-practice/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReminderServer struct {
	pb.UnimplementedReminderServiceServer
	service *service.ReminderService
}

func NewReminderServer(svc *service.ReminderService) *ReminderServer {
	return &ReminderServer{service: svc}
}

func (s *ReminderServer) CreateReminder(ctx context.Context, req *pb.CreateReminderRequest) (*pb.ReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	reminder, err := s.service.Create(userID, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) GetReminders(ctx context.Context, req *pb.GetRemindersRequest) (*pb.GetRemindersResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	reminders, err := s.service.GetByUserID(userID, req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoReminders []*pb.ReminderResponse
	for _, r := range reminders {
		protoReminders = append(protoReminders, toProtoReminder(&r))
	}

	return &pb.GetRemindersResponse{Reminders: protoReminders}, nil
}

func (s *ReminderServer) GetReminder(ctx context.Context, req *pb.GetReminderRequest) (*pb.ReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	reminder, err := s.service.GetByID(userID, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) UpdateReminder(ctx context.Context, req *pb.UpdateReminderRequest) (*pb.ReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	reminder, err := s.service.Update(userID, id, req.Title, req.Description, req.RemindAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoReminder(reminder), nil
}

func (s *ReminderServer) DeleteReminder(ctx context.Context, req *pb.DeleteReminderRequest) (*pb.DeleteReminderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.DeleteReminderResponse{
			Success: false,
			Message: "invalid user_id: " + err.Error(),
		}, nil
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeleteReminderResponse{
			Success: false,
			Message: "invalid id: " + err.Error(),
		}, nil
	}

	err = s.service.Delete(userID, id)
	if err != nil {
		return &pb.DeleteReminderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteReminderResponse{
		Success: true,
		Message: "Reminder deleted successfully",
	}, nil
}

func toProtoReminder(r *models.Reminder) *pb.ReminderResponse {
	return &pb.ReminderResponse{
		Id:          r.ID.String(),     // UUID to string
		UserId:      r.UserID.String(), // UUID to string
		Title:       r.Title,
		Description: r.Description,
		RemindAt:    r.RemindAt.Format("2006-01-02T15:04:05Z07:00"),
		IsSent:      r.IsSent,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
