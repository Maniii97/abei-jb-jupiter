package services

import (
	"api/internal/entities"
	"api/internal/repository"
	"context"
)

type EventService struct {
	eventRepo *repository.EventRepository
}

// GetAvailableSeatsCount implements EventServiceInterface.
func (s *EventService) GetAvailableSeatsCount(ctx context.Context, eventID uint) (int64, error) {
	return s.eventRepo.CountAvailableSeats(ctx, eventID)
}

// Ensure EventService implements EventServiceInterface
var _ EventServiceInterface = (*EventService)(nil)

func NewEventService(eventRepo *repository.EventRepository) *EventService {
	return &EventService{eventRepo: eventRepo}
}

// GetEvents returns a paginated list of events
func (s *EventService) GetEvents(ctx context.Context, limit, offset int, eventType, city string) ([]entities.Event, int64, error) {
	return s.eventRepo.GetEvents(ctx, limit, offset, eventType, city)
}

func (s *EventService) GetEventByID(ctx context.Context, eventID uint) (*entities.Event, error) {
	return s.eventRepo.GetEventByID(ctx, eventID)
}

func (s *EventService) GetAvailableSeats(ctx context.Context, eventID uint) ([]entities.Seat, error) {
	return s.eventRepo.GetAvailableSeats(ctx, eventID)
}

func (s *EventService) CreateEvent(ctx context.Context, event *entities.Event) error {
	return s.eventRepo.CreateEvent(ctx, event)
}

func (s *EventService) UpdateEvent(ctx context.Context, eventID uint, updates map[string]interface{}) (*entities.Event, error) {
	return s.eventRepo.UpdateEvent(ctx, eventID, updates)
}

func (s *EventService) DeleteEvent(ctx context.Context, eventID uint) error {
	return s.eventRepo.DeleteEvent(ctx, eventID)
}

func (s *EventService) GetEventStats(ctx context.Context, eventID uint) (map[string]interface{}, error) {
	return s.eventRepo.GetEventStats(ctx, eventID)
}
