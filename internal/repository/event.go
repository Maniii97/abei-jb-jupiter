package repository

import (
	"api/constants"
	"api/internal/entities"
	"api/pkg/errors"
	"context"
	"time"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// GetEvents returns a paginated list of events
func (s *EventRepository) GetEvents(ctx context.Context, limit, offset int, eventType, city string) ([]entities.Event, int64, error) {
	var events []entities.Event
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.Event{}).
		Where("status = ? AND start_time > ?", constants.EventStatusActive, time.Now()).
		Preload("Venue")

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	if city != "" {
		query = query.Joins("JOIN venues ON events.venue_id = venues.id").
			Where("venues.city ILIKE ?", "%"+city+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalError("Failed to count events", err)
	}

	// Get paginated results
	if err := query.Order("start_time ASC").
		Limit(limit).Offset(offset).
		Find(&events).Error; err != nil {
		return nil, 0, errors.NewInternalError("Failed to fetch events", err)
	}

	return events, total, nil
}

// GetEventByID returns a single event with all details
func (s *EventRepository) GetEventByID(ctx context.Context, eventID uint) (*entities.Event, error) {
	var event entities.Event

	if err := s.db.WithContext(ctx).
		Preload("Venue").
		Preload("Seats", "is_available = true").
		First(&event, eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Event not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch event", err)
	}

	return &event, nil
}

// GetAvailableSeats returns available seats for an event
func (s *EventRepository) GetAvailableSeats(ctx context.Context, eventID uint) ([]entities.Seat, error) {
	var seats []entities.Seat

	if err := s.db.WithContext(ctx).
		Where("event_id = ? AND is_available = true AND is_locked = false", eventID).
		Order("\"row\" ASC, \"column\" ASC").
		Find(&seats).Error; err != nil {
		return nil, errors.NewInternalError("Failed to fetch available seats", err)
	}

	return seats, nil
}

// CountAvailableSeats returns the count of available seats for an event
func (s *EventRepository) CountAvailableSeats(ctx context.Context, eventID uint) (int64, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&entities.Seat{}).
		Where("event_id = ? AND is_available = true AND is_locked = false", eventID).
		Count(&count).Error; err != nil {
		return 0, errors.NewInternalError("Failed to count available seats", err)
	}

	return count, nil
}

// CreateEvent creates a new event (admin only)
func (s *EventRepository) CreateEvent(ctx context.Context, event *entities.Event) error {
	// First, verify the venue exists and get its information
	var venue entities.Venue
	if err := s.db.WithContext(ctx).First(&venue, event.VenueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Venue not found", errors.ErrRecordNotFound)
		}
		return errors.NewInternalError("Failed to fetch venue", err)
	}

	// Check for venue time conflicts
	if err := s.checkVenueTimeConflict(ctx, event.VenueID, event.StartTime, event.EndTime, 0); err != nil {
		return err
	}

	// Validate event times
	if err := s.validateEventTimes(event.StartTime, event.EndTime); err != nil {
		return err
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Set initial available seats to venue capacity
	event.AvailableSeats = venue.Rows * venue.Columns

	// Create the event
	if err := tx.Create(event).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to create event", err)
	}

	// Create seats for the event using venue rows and columns
	if err := s.createSeatsForEvent(tx, event, venue.Rows, venue.Columns); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateEvent updates an existing event (admin only)
func (s *EventRepository) UpdateEvent(ctx context.Context, eventID uint, updates map[string]interface{}) (*entities.Event, error) {
	var event entities.Event

	if err := s.db.WithContext(ctx).First(&event, eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Event not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch event", err)
	}

	// Check for venue time conflicts if venue_id, start_time, or end_time are being updated
	venueID := event.VenueID
	startTime := event.StartTime
	endTime := event.EndTime

	if newVenueID, ok := updates["venue_id"]; ok {
		venueID = newVenueID.(uint)
	}
	if newStartTime, ok := updates["start_time"]; ok {
		startTime = newStartTime.(time.Time)
	}
	if newEndTime, ok := updates["end_time"]; ok {
		endTime = newEndTime.(time.Time)
	}

	// Only check for conflicts if venue, start_time, or end_time are being changed
	if _, hasVenue := updates["venue_id"]; hasVenue {
		if err := s.validateEventTimes(startTime, endTime); err != nil {
			return nil, err
		}
		if err := s.checkVenueTimeConflict(ctx, venueID, startTime, endTime, eventID); err != nil {
			return nil, err
		}
	} else if _, hasStartTime := updates["start_time"]; hasStartTime {
		if err := s.validateEventTimes(startTime, endTime); err != nil {
			return nil, err
		}
		if err := s.checkVenueTimeConflict(ctx, venueID, startTime, endTime, eventID); err != nil {
			return nil, err
		}
	} else if _, hasEndTime := updates["end_time"]; hasEndTime {
		if err := s.validateEventTimes(startTime, endTime); err != nil {
			return nil, err
		}
		if err := s.checkVenueTimeConflict(ctx, venueID, startTime, endTime, eventID); err != nil {
			return nil, err
		}
	}

	if err := s.db.WithContext(ctx).Model(&event).Updates(updates).Error; err != nil {
		return nil, errors.NewInternalError("Failed to update event", err)
	}

	return &event, nil
}

// DeleteEvent soft deletes an event (admin only)
func (s *EventRepository) DeleteEvent(ctx context.Context, eventID uint) error {
	var event entities.Event

	if err := s.db.WithContext(ctx).First(&event, eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Event not found", errors.ErrRecordNotFound)
		}
		return errors.NewInternalError("Failed to fetch event", err)
	}

	// Update status to cancelled instead of hard delete
	if err := s.db.WithContext(ctx).Model(&event).
		Update("status", constants.EventStatusCancelled).Error; err != nil {
		return errors.NewInternalError("Failed to cancel event", err)
	}

	return nil
}

// createSeatsForEvent creates seats for a new event using venue's row/column configuration
func (s *EventRepository) createSeatsForEvent(tx *gorm.DB, event *entities.Event, rows, columns int) error {
	var seats []entities.Seat

	for row := 1; row <= rows; row++ {
		for col := 1; col <= columns; col++ {
			// All seats are standard type with the same price as the event
			seat := entities.Seat{
				EventID:     event.ID,
				Row:         row,
				Column:      col,
				SeatType:    constants.SeatTypeStandard,
				Price:       event.Price,
				IsAvailable: true,
				IsLocked:    false,
			}
			seats = append(seats, seat)
		}
	}

	if err := tx.CreateInBatches(seats, 100).Error; err != nil {
		return errors.NewInternalError("Failed to create seats", err)
	}

	return nil
}

// GetEventStats returns statistics for an event (admin only)
func (s *EventRepository) GetEventStats(ctx context.Context, eventID uint) (map[string]interface{}, error) {
	var event entities.Event
	var totalSeats int64
	var bookedSeats int64
	var lockedSeats int64
	var revenue float64

	// Check if event exists
	if err := s.db.WithContext(ctx).First(&event, eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Event not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch event", err)
	}

	// Total seats
	if err := s.db.WithContext(ctx).Model(&entities.Seat{}).
		Where("event_id = ?", eventID).Count(&totalSeats).Error; err != nil {
		return nil, errors.NewInternalError("Failed to count total seats", err)
	}

	// Booked seats
	if err := s.db.WithContext(ctx).Model(&entities.Booking{}).
		Where("event_id = ? AND status = ?", eventID, constants.BookingStatusConfirmed).
		Count(&bookedSeats).Error; err != nil {
		return nil, errors.NewInternalError("Failed to count booked seats", err)
	}

	// Locked seats
	if err := s.db.WithContext(ctx).Model(&entities.Seat{}).
		Where("event_id = ? AND is_locked = true", eventID).
		Count(&lockedSeats).Error; err != nil {
		return nil, errors.NewInternalError("Failed to count locked seats", err)
	}

	// Total revenue
	if err := s.db.WithContext(ctx).Model(&entities.Booking{}).
		Where("event_id = ? AND status = ? AND payment_status = ?",
			eventID, constants.BookingStatusConfirmed, constants.PaymentStatusPaid).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&revenue).Error; err != nil {
		return nil, errors.NewInternalError("Failed to calculate revenue", err)
	}

	stats := map[string]interface{}{
		"event_id":             eventID,
		"event_name":           event.Name,
		"total_seats":          totalSeats,
		"booked_seats":         bookedSeats,
		"locked_seats":         lockedSeats,
		"available_seats":      totalSeats - bookedSeats - lockedSeats,
		"capacity_utilization": float64(bookedSeats) / float64(totalSeats) * 100,
		"total_revenue":        revenue,
		"booking_rate":         float64(bookedSeats) / float64(totalSeats) * 100,
	}

	return stats, nil
}

// checkVenueTimeConflict checks if there's a time conflict for events at the same venue
func (s *EventRepository) checkVenueTimeConflict(ctx context.Context, venueID uint, startTime, endTime time.Time, excludeEventID uint) error {
	var conflictingEvent entities.Event

	query := s.db.WithContext(ctx).
		Where("venue_id = ? AND status = ?", venueID, constants.EventStatusActive).
		Where("NOT (end_time <= ? OR start_time >= ?)", startTime, endTime)

	// Exclude current event when updating
	if excludeEventID > 0 {
		query = query.Where("id != ?", excludeEventID)
	}

	if err := query.First(&conflictingEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No conflict found
			return nil
		}
		return errors.NewInternalError("Failed to check venue time conflicts", err)
	}

	// Conflict found
	return errors.NewConflictError(constants.ErrVenueTimeConflict, nil)
}

// validateEventTimes validates event start and end times
func (s *EventRepository) validateEventTimes(startTime, endTime time.Time) error {
	// Check if end time is after start time
	if !endTime.After(startTime) {
		return errors.NewBadRequestError("End time must be after start time", nil)
	}

	// Check if start time is in the future
	if startTime.Before(time.Now()) {
		return errors.NewBadRequestError("Start time must be in the future", nil)
	}

	return nil
}
