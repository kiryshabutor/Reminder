package client

import (
	"context"

	"github.com/kiribu/jwt-practice/internal/reminder/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReminderClient struct {
	conn   *grpc.ClientConn
	client pb.ReminderServiceClient
}

func NewReminderClient(addr string) (*ReminderClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ReminderClient{
		conn:   conn,
		client: pb.NewReminderServiceClient(conn),
	}, nil
}

func (c *ReminderClient) Close() error {
	return c.conn.Close()
}

func (c *ReminderClient) Create(ctx context.Context, userID int64, title, description, remindAt string) (*pb.ReminderResponse, error) {
	return c.client.CreateReminder(ctx, &pb.CreateReminderRequest{
		UserId:      userID,
		Title:       title,
		Description: description,
		RemindAt:    remindAt,
	})
}

func (c *ReminderClient) GetAll(ctx context.Context, userID int64) (*pb.GetRemindersResponse, error) {
	return c.client.GetReminders(ctx, &pb.GetRemindersRequest{
		UserId: userID,
	})
}

func (c *ReminderClient) GetByID(ctx context.Context, userID, id int64) (*pb.ReminderResponse, error) {
	return c.client.GetReminder(ctx, &pb.GetReminderRequest{
		UserId: userID,
		Id:     id,
	})
}

func (c *ReminderClient) Update(ctx context.Context, userID, id int64, title, description, remindAt string) (*pb.ReminderResponse, error) {
	return c.client.UpdateReminder(ctx, &pb.UpdateReminderRequest{
		UserId:      userID,
		Id:          id,
		Title:       title,
		Description: description,
		RemindAt:    remindAt,
	})
}

func (c *ReminderClient) Delete(ctx context.Context, userID, id int64) (*pb.DeleteReminderResponse, error) {
	return c.client.DeleteReminder(ctx, &pb.DeleteReminderRequest{
		UserId: userID,
		Id:     id,
	})
}
