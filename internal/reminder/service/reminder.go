package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kiribu/jwt-practice/internal/reminder/kafka"
	"github.com/kiribu/jwt-practice/internal/reminder/storage"
	"github.com/kiribu/jwt-practice/models"
)

type ReminderService struct {
	storage       storage.ReminderStorage
	eventProducer *kafka.Producer
}

func NewReminderService(storage storage.ReminderStorage, eventProducer *kafka.Producer) *ReminderService {
	return &ReminderService{
		storage:       storage,
		eventProducer: eventProducer,
	}
}

func (s *ReminderService) Create(userID int64, title, description, remindAtStr string) (*models.Reminder, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	remindAt, err := time.Parse(time.RFC3339, remindAtStr)
	if err != nil {
		return nil, errors.New("invalid remind_at format, use RFC3339: 2026-01-25T10:00:00+03:00")
	}

	if remindAt.Before(time.Now()) {
		return nil, errors.New("remind_at must be in the future")
	}

	reminder, err := s.storage.Create(userID, title, description, remindAt)
	if err != nil {
		return nil, err
	}

	event := models.LifecycleEvent{
		EventType:  "created",
		ReminderID: reminder.ID,
		UserID:     reminder.UserID,
		Timestamp:  time.Now(),
		Payload:    reminder,
	}
	if err := s.eventProducer.SendEvent(fmt.Sprintf("%d", reminder.ID), event); err != nil {
		log.Printf("Failed to send created event: %v", err)
	}

	return reminder, nil
}

func (s *ReminderService) GetByUserID(userID int64, status string) ([]models.Reminder, error) {
	return s.storage.GetByUserID(userID, status)
}

func (s *ReminderService) GetByID(userID, id int64) (*models.Reminder, error) {
	return s.storage.GetByID(userID, id)
}

func (s *ReminderService) Update(userID, id int64, title, description, remindAtStr string) (*models.Reminder, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	remindAt, err := time.Parse(time.RFC3339, remindAtStr)
	if err != nil {
		return nil, errors.New("invalid remind_at format, use RFC3339: 2026-01-25T10:00:00+03:00")
	}

	reminder, err := s.storage.Update(userID, id, title, description, remindAt)
	if err != nil {
		return nil, err
	}

	event := models.LifecycleEvent{
		EventType:  "updated",
		ReminderID: reminder.ID,
		UserID:     reminder.UserID,
		Timestamp:  time.Now(),
		Payload:    reminder,
	}
	if err := s.eventProducer.SendEvent(fmt.Sprintf("%d", reminder.ID), event); err != nil {
		log.Printf("Failed to send updated event: %v", err)
	}

	return reminder, nil
}

func (s *ReminderService) Delete(userID, id int64) error {
	err := s.storage.Delete(userID, id)
	if err != nil {
		return err
	}

	event := models.LifecycleEvent{
		EventType:  "deleted",
		ReminderID: id,
		UserID:     userID,
		Timestamp:  time.Now(),
		Payload:    nil,
	}
	if err := s.eventProducer.SendEvent(fmt.Sprintf("%d", id), event); err != nil {
		log.Printf("Failed to send deleted event: %v", err)
	}

	return nil
}
