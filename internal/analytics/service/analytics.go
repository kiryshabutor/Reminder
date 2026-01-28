package service

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/internal/analytics/storage"
	"github.com/kiribu/jwt-practice/models"
)

type AnalyticsService struct {
	storage storage.AnalyticsStorage
}

func NewAnalyticsService(storage storage.AnalyticsStorage) *AnalyticsService {
	return &AnalyticsService{storage: storage}
}

func (s *AnalyticsService) ProcessEvent(ctx context.Context, event models.LifecycleEvent) error {
	log.Printf("Processing event: %s for user %s", event.EventType, event.UserID)

	switch event.EventType {
	case "created":
		return s.storage.IncrementCreated(ctx, event.UserID, event.Timestamp)
	case "updated":
		return nil
	case "notification_sent":
		return s.storage.IncrementCompleted(ctx, event.UserID, event.Timestamp)
	case "deleted":
		return s.storage.IncrementDeleted(ctx, event.UserID, event.Timestamp)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}
}

func (s *AnalyticsService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error) {
	stats, err := s.storage.GetUserStats(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.UserStatistics{UserID: userID}, nil
		}
		return nil, err
	}
	return stats, nil
}
