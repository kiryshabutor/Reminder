package client

import (
	"context"
	"log"
	"time"

	"github.com/kiribu/jwt-practice/internal/reminder/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReminderClient struct {
	conn   *grpc.ClientConn
	client pb.ReminderServiceClient
}

func NewReminderClient(addr string) (*ReminderClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Connecting to Reminder Service at %s...", addr)
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to Reminder Service at %s", addr)

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

func (c *ReminderClient) GetAll(ctx context.Context, userID int64, status string) (*pb.GetRemindersResponse, error) {
	return c.client.GetReminders(ctx, &pb.GetRemindersRequest{
		UserId: userID,
		Status: status,
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
