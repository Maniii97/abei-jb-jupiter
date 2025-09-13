package services

import (
	"api/internal/entities"
	"api/internal/repository"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type WaitlistService struct {
	waitlistRepo *repository.WaitlistRepository
	eventRepo    *repository.EventRepository
	db           *gorm.DB
}

func NewWaitlistService(waitlistRepo *repository.WaitlistRepository, eventRepo *repository.EventRepository, db *gorm.DB) *WaitlistService {
	return &WaitlistService{
		waitlistRepo: waitlistRepo,
		eventRepo:    eventRepo,
		db:           db,
	}
}

// JoinWaitlist adds a user to the event waitlist if the event is full
func (s *WaitlistService) JoinWaitlist(ctx context.Context, userID, eventID uint) (*WaitlistEntry, error) {
	// First check if the event exists and is active
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if event.Status != "active" {
		return nil, fmt.Errorf("event is not active")
	}

	// Check if there are available seats
	if event.AvailableSeats > 0 {
		return nil, fmt.Errorf("seats are still available for this event, please book directly instead of joining waitlist")
	}

	// Join the waitlist
	repoEntry, err := s.waitlistRepo.JoinWaitlist(ctx, userID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to join waitlist: %w", err)
	}

	// Convert to service WaitlistEntry
	entry := &WaitlistEntry{
		UserID:     repoEntry.UserID,
		EventID:    repoEntry.EventID,
		JoinedAt:   repoEntry.JoinedAt,
		Position:   repoEntry.Position,
		NotifiedAt: repoEntry.NotifiedAt,
	}

	// Also store in database for persistence
	dbEntry := &entities.EventQueue{
		EventID:       eventID,
		UserID:        userID,
		QueuePosition: repoEntry.Position,
		Status:        "waiting",
		JoinedAt:      repoEntry.JoinedAt,
	}

	if err := s.db.WithContext(ctx).Create(dbEntry).Error; err != nil {
		// If DB fails, try to remove from Redis to maintain consistency
		s.waitlistRepo.RemoveFromWaitlist(ctx, userID, eventID)
		return nil, fmt.Errorf("failed to save waitlist entry to database: %w", err)
	}

	return entry, nil
}

// GetWaitlistPosition returns the current position of a user in the waitlist
func (s *WaitlistService) GetWaitlistPosition(ctx context.Context, userID, eventID uint) (*WaitlistEntry, error) {
	repoEntry, err := s.waitlistRepo.GetWaitlistPosition(ctx, userID, eventID)
	if err != nil {
		return nil, err
	}

	// Convert to service WaitlistEntry
	entry := &WaitlistEntry{
		UserID:     repoEntry.UserID,
		EventID:    repoEntry.EventID,
		JoinedAt:   repoEntry.JoinedAt,
		Position:   repoEntry.Position,
		NotifiedAt: repoEntry.NotifiedAt,
	}

	return entry, nil
}

// LeaveWaitlist removes a user from the waitlist
func (s *WaitlistService) LeaveWaitlist(ctx context.Context, userID, eventID uint) error {
	// Remove from Redis
	err := s.waitlistRepo.RemoveFromWaitlist(ctx, userID, eventID)
	if err != nil {
		return fmt.Errorf("failed to remove from Redis waitlist: %w", err)
	}

	// Update database entry status
	result := s.db.WithContext(ctx).
		Model(&entities.EventQueue{}).
		Where("user_id = ? AND event_id = ? AND status = ?", userID, eventID, "waiting").
		Update("status", "cancelled")

	if result.Error != nil {
		return fmt.Errorf("failed to update database waitlist entry: %w", result.Error)
	}

	return nil
}

// GetWaitlistSize returns the number of people waiting for an event
func (s *WaitlistService) GetWaitlistSize(ctx context.Context, eventID uint) (int, error) {
	return s.waitlistRepo.GetWaitlistSize(ctx, eventID)
}

// ProcessSeatAvailability marks seats as available for waitlisted users
func (s *WaitlistService) ProcessSeatAvailability(ctx context.Context, eventID uint, availableSeats int) ([]*WaitlistEntry, error) {
	if availableSeats <= 0 {
		return nil, nil
	}

	// Mark the first N users in the waitlist as having seats available
	// They can check their status and book
	availableUsers := make([]*WaitlistEntry, 0)
	
	for i := 0; i < availableSeats; i++ {
		// Get the next user in queue
		nextUser, err := s.waitlistRepo.GetNextInWaitlist(ctx, eventID)
		if err != nil {
			break // No more users in queue
		}
		if nextUser == nil {
			break // Empty queue
		}

		// Update database entry to mark as active with expiration
		now := time.Now()
		expiresAt := now.Add(10 * time.Minute) // Give users 10 minutes to book

		err = s.db.WithContext(ctx).
			Model(&entities.EventQueue{}).
			Where("user_id = ? AND event_id = ? AND status = ?", nextUser.UserID, eventID, "waiting").
			Updates(map[string]interface{}{
				"status":     "active",
				"active_at":  &now,
				"expires_at": &expiresAt,
			}).Error

		if err != nil {
			fmt.Printf("Failed to update database for user %d: %v\n", nextUser.UserID, err)
			continue
		}

		// Convert to service WaitlistEntry
		serviceEntry := &WaitlistEntry{
			UserID:   nextUser.UserID,
			EventID:  nextUser.EventID,
			JoinedAt: nextUser.JoinedAt,
			Position: nextUser.Position,
		}

		availableUsers = append(availableUsers, serviceEntry)
	}

	return availableUsers, nil
}

// CleanupExpiredWaitlist removes users who were notified but didn't book within the time limit
func (s *WaitlistService) CleanupExpiredWaitlist(ctx context.Context) error {
	// Clean up expired notifications from Redis (5 minutes default)
	events, err := s.getActiveEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active events: %w", err)
	}

	for _, event := range events {
		err := s.waitlistRepo.CleanupExpiredNotifications(ctx, event.ID, 10*time.Minute)
		if err != nil {
			fmt.Printf("Failed to cleanup expired notifications for event %d: %v\n", event.ID, err)
		}
	}

	// Update database entries that have expired
	now := time.Now()
	err = s.db.WithContext(ctx).
		Model(&entities.EventQueue{}).
		Where("status = ? AND expires_at < ?", "active", now).
		Update("status", "expired").Error

	if err != nil {
		return fmt.Errorf("failed to update expired waitlist entries: %w", err)
	}

	return nil
}

// getActiveEvents helper function to get all active events
func (s *WaitlistService) getActiveEvents(ctx context.Context) ([]entities.Event, error) {
	var events []entities.Event
	err := s.db.WithContext(ctx).
		Where("status = ?", "active").
		Find(&events).Error
	
	return events, err
}

// RemoveUserFromWaitlistAfterBooking removes user from waitlist after successful booking
func (s *WaitlistService) RemoveUserFromWaitlistAfterBooking(ctx context.Context, userID, eventID uint) error {
	// Remove from Redis
	err := s.waitlistRepo.RemoveFromWaitlist(ctx, userID, eventID)
	if err != nil {
		// Log error but don't fail the booking process
		fmt.Printf("Failed to remove user %d from Redis waitlist for event %d: %v\n", userID, eventID, err)
	}

	// Update database entry status
	result := s.db.WithContext(ctx).
		Model(&entities.EventQueue{}).
		Where("user_id = ? AND event_id = ? AND status IN (?)", userID, eventID, []string{"waiting", "active"}).
		Update("status", "completed")

	if result.Error != nil {
		fmt.Printf("Failed to update database waitlist entry for user %d, event %d: %v\n", userID, eventID, result.Error)
	}

	return nil
}
