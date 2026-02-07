package service

import (
	"context"
	"database/sql"
	"log/slog"

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
	slog.Info("Processing event", "event_id", event.EventID, "type", event.EventType, "user_id", event.UserID)

	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	processed, err := s.storage.IsEventProcessed(ctx, tx, event.EventID)
	if err != nil {
		return err
	}
	if processed {
		slog.Info("Event already processed, skipping", "event_id", event.EventID)
		return nil
	}

	switch event.EventType {
	case "created":
		err = s.storage.IncrementCreated(ctx, tx, event.UserID, event.Timestamp)
	case "updated":
		err = nil // No-op for updated
	case "notification_sent":
		err = s.storage.IncrementCompleted(ctx, tx, event.UserID, event.Timestamp)
	case "deleted":
		err = s.storage.IncrementDeleted(ctx, tx, event.UserID, event.Timestamp)
	default:
		slog.Warn("Unknown event type", "type", event.EventType)
		err = nil
	}

	if err != nil {
		return err
	}

	if err := s.storage.MarkEventProcessed(ctx, tx, event.EventID); err != nil {
		return err
	}

	return tx.Commit()
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
