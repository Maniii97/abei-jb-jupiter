package repository

import (
	"api/constants"
	"api/internal/entities"
	"api/pkg/errors"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db                 *gorm.DB
	seatLockRepository *SeatLockRepository
}

func NewBookingRepository(db *gorm.DB, seatLockRepository *SeatLockRepository) *BookingRepository {
	return &BookingRepository{
		db:                 db,
		seatLockRepository: seatLockRepository,
	}
}

// CreateBookingIntent creates a booking intent and locks the seat
func (s *BookingRepository) CreateBookingIntent(ctx context.Context, userID, seatID uint) (*entities.BookingIntent, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if seat exists and is available
	var seat entities.Seat
	if err := tx.Preload("Event").First(&seat, seatID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Seat not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch seat", err)
	}

	// Check if seat is available
	if !seat.IsAvailable {
		tx.Rollback()
		return nil, errors.NewBadRequestError(constants.ErrSeatNotAvailable, nil)
	}

	// Check if seat is already locked
	if seat.IsLocked {
		tx.Rollback()
		return nil, errors.NewConflictError(constants.ErrSeatAlreadyLocked, nil)
	}

	// Check if event is still active and in the future
	if seat.Event.Status != constants.EventStatusActive {
		tx.Rollback()
		return nil, errors.NewBadRequestError("Event is not active", nil)
	}

	if seat.Event.StartTime.Before(time.Now()) {
		tx.Rollback()
		return nil, errors.NewBadRequestError("Event has already started", nil)
	}

	// Check if event still has available capacity
	if seat.Event.AvailableSeats <= 0 {
		tx.Rollback()
		return nil, errors.NewBadRequestError(constants.ErrEventSoldOut, nil)
	}

	// Create booking intent
	intent := &entities.BookingIntent{
		UserID:  userID,
		EventID: seat.EventID,
		SeatID:  seatID,
		Status:  constants.IntentStatusPending,
	}

	if err := tx.Create(intent).Error; err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to create booking intent", err)
	}

	// Lock seat in database
	if err := tx.Model(&seat).Updates(map[string]interface{}{
		"is_locked": true,
		"locked_at": time.Now(),
		"locked_by": userID,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to lock seat", err)
	}

	// Lock seat in Redis using the booking intent ID (best effort)
	intentIDStr := fmt.Sprintf("%d", intent.ID)
	redisLockErr := s.seatLockRepository.LockSeat(ctx, seatID, userID, intentIDStr)
	if redisLockErr != nil {
		// Log Redis failure but continue with database-only locking
		fmt.Printf("Warning: Redis lock failed, falling back to database-only locking: %v\n", redisLockErr)
		// We don't rollback here - database lock is sufficient for consistency
	}

	// Commit transaction (database lock is the source of truth)
	if err := tx.Commit().Error; err != nil {
		// Try to cleanup Redis lock if it was successful but DB commit failed
		if redisLockErr == nil {
			s.seatLockRepository.UnlockSeat(ctx, seatID, userID, intentIDStr)
		}
		return nil, errors.NewInternalError("Failed to commit booking intent", err)
	}

	// Load the intent with relationships
	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Event.Venue").
		Preload("Event").
		Preload("Seat").
		First(intent, intent.ID).Error; err != nil {
		return nil, errors.NewInternalError("Failed to load booking intent", err)
	}

	return intent, nil
}

// ConfirmBooking confirms a booking intent after successful payment
func (s *BookingRepository) ConfirmBooking(ctx context.Context, bookingIntentID uint, paymentID string) (*entities.Booking, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get booking intent with optimized query
	var intent entities.BookingIntent
	if err := tx.Select("id, user_id, event_id, seat_id, status, created_at").
		Where("id = ? AND status = ?", bookingIntentID, constants.IntentStatusPending).
		First(&intent).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Booking intent not found or already processed", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch booking intent", err)
	}

	// Check if intent is still valid
	// logic is Now - lockCreatedAt > duration -> expired
	if time.Now().After(intent.CreatedAt.Add(time.Duration(constants.SeatLockDuration) * time.Minute)) {
		tx.Rollback()
		return nil, errors.NewBadRequestError(constants.ErrBookingExpired, nil)
	}

	// Get seat price efficiently
	var seatPrice float64
	if err := tx.Model(&entities.Seat{}).Select("price").Where("id = ?", intent.SeatID).Scan(&seatPrice).Error; err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to fetch seat price", err)
	}

	// Create booking
	booking := &entities.Booking{
		UserID:          intent.UserID,
		EventID:         intent.EventID,
		SeatID:          intent.SeatID,
		BookingIntentID: &intent.ID,
		Status:          constants.BookingStatusConfirmed,
		PaymentStatus:   constants.PaymentStatusPaid,
		PaymentID:       paymentID,
		TotalAmount:     seatPrice,
		BookedAt:        time.Now(),
	}

	if err := tx.Create(booking).Error; err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to create booking", err)
	}

	// Batch update booking intent and seat in a single operation each
	if err := tx.Model(&entities.BookingIntent{}).Where("id = ?", intent.ID).
		Updates(map[string]interface{}{
			"status":            constants.IntentStatusConfirmed,
			"payment_intent_id": paymentID,
			"updated_at":        time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to update booking intent", err)
	}

	// Update seat availability efficiently
	if err := tx.Model(&entities.Seat{}).Where("id = ?", intent.SeatID).
		Updates(map[string]interface{}{
			"is_available": false,
			"is_locked":    false,
			"locked_at":    nil,
			"locked_by":    nil,
			"updated_at":   time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to update seat", err)
	}

	// Update event available seats count using atomic operation with capacity check
	result := tx.Model(&entities.Event{}).
		Where("id = ? AND available_seats > 0", intent.EventID).
		Update("available_seats", gorm.Expr("available_seats - ?", 1))

	if result.Error != nil {
		tx.Rollback()
		return nil, errors.NewInternalError("Failed to update event capacity", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.NewBadRequestError(constants.ErrEventSoldOut, nil)
	}

	// Unlock seat in Redis (don't fail transaction if this fails)
	intentIDStr := fmt.Sprintf("%d", intent.ID)
	if err := s.seatLockRepository.UnlockSeat(ctx, intent.SeatID, intent.UserID, intentIDStr); err != nil {
		// Log this error but don't fail the transaction as the booking is already confirmed
		fmt.Printf("Warning: Failed to unlock seat in Redis: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.NewInternalError("Failed to commit booking", err)
	}

	// Load the booking with relationships using optimized query
	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Event.Venue").
		Preload("Event").
		Preload("Seat").
		First(booking, booking.ID).Error; err != nil {
		return nil, errors.NewInternalError("Failed to load booking", err)
	}

	return booking, nil
}

// CancelBookingIntent cancels a booking intent and unlocks the seat
func (s *BookingRepository) CancelBookingIntent(ctx context.Context, bookingIntentID uint, userID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get booking intent
	var intent entities.BookingIntent
	if err := tx.Where("id = ? AND user_id = ? AND status = ?",
		bookingIntentID, userID, constants.IntentStatusPending).
		First(&intent).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Booking intent not found", errors.ErrRecordNotFound)
		}
		return errors.NewInternalError("Failed to fetch booking intent", err)
	}

	// Update intent status
	if err := tx.Model(&intent).Update("status", constants.IntentStatusCancelled).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to update booking intent", err)
	}

	// Unlock seat in database
	if err := tx.Model(&entities.Seat{}).Where("id = ?", intent.SeatID).
		Updates(map[string]interface{}{
			"is_locked": false,
			"locked_at": nil,
			"locked_by": nil,
		}).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to unlock seat", err)
	}

	// Unlock seat in Redis (don't fail transaction if this fails)
	intentIDStr := fmt.Sprintf("%d", intent.ID)
	if err := s.seatLockRepository.UnlockSeat(ctx, intent.SeatID, userID, intentIDStr); err != nil {
		// Log this error but don't fail the transaction as the database unlock is sufficient
		fmt.Printf("Warning: Failed to unlock seat in Redis: %v\n", err)
	}

	return tx.Commit().Error
}

// CancelBooking cancels a confirmed booking
func (s *BookingRepository) CancelBooking(ctx context.Context, bookingID uint, userID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get booking
	var booking entities.Booking
	if err := tx.Preload("Event").
		Where("id = ? AND user_id = ? AND status = ?",
			bookingID, userID, constants.BookingStatusConfirmed).
		First(&booking).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Booking not found", errors.ErrRecordNotFound)
		}
		return errors.NewInternalError("Failed to fetch booking", err)
	}

	// Check if event hasn't started yet (allow cancellation only before event starts)
	if booking.Event.StartTime.Before(time.Now()) {
		tx.Rollback()
		return errors.NewBadRequestError("Cannot cancel booking after event has started", nil)
	}

	// Update booking status
	if err := tx.Model(&booking).Updates(map[string]interface{}{
		"status":       constants.BookingStatusCancelled,
		"cancelled_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to cancel booking", err)
	}

	// Make seat available again
	if err := tx.Model(&entities.Seat{}).Where("id = ?", booking.SeatID).
		Update("is_available", true).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to update seat availability", err)
	}

	// Update event available seats count
	if err := tx.Model(&booking.Event).UpdateColumn("available_seats", gorm.Expr("available_seats + ?", 1)).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to update event capacity", err)
	}

	return tx.Commit().Error
}

// GetUserBookings returns user's booking history
func (s *BookingRepository) GetUserBookings(ctx context.Context, userID uint, limit, offset int) ([]entities.Booking, int64, error) {
	var bookings []entities.Booking
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.Booking{}).Where("user_id = ?", userID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalError("Failed to count bookings", err)
	}

	// Get paginated results
	if err := query.Preload("Event.Venue").Preload("Event").Preload("Seat").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, 0, errors.NewInternalError("Failed to fetch bookings", err)
	}

	return bookings, total, nil
}

// GetBookingByID returns a specific booking
func (s *BookingRepository) GetBookingByID(ctx context.Context, bookingID, userID uint) (*entities.Booking, error) {
	var booking entities.Booking

	if err := s.db.WithContext(ctx).
		Preload("Event.Venue").
		Preload("Event").
		Preload("Seat").
		Where("id = ? AND user_id = ?", bookingID, userID).
		First(&booking).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Booking not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch booking", err)
	}

	return &booking, nil
}

// CleanupExpiredIntents removes expired booking intents and unlocks seats
func (s *BookingRepository) CleanupExpiredIntents(ctx context.Context) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find expired intents
	var expiredIntents []entities.BookingIntent
	if err := tx.Where("status = ? AND lock_expires_at < ?",
		constants.IntentStatusPending, time.Now()).
		Find(&expiredIntents).Error; err != nil {
		tx.Rollback()
		return errors.NewInternalError("Failed to fetch expired intents", err)
	}

	// Update expired intents
	if len(expiredIntents) > 0 {
		intentIDs := make([]uint, len(expiredIntents))
		seatIDs := make([]uint, len(expiredIntents))

		for i, intent := range expiredIntents {
			intentIDs[i] = intent.ID
			seatIDs[i] = intent.SeatID

			// Unlock in Redis using intent ID
			intentIDStr := fmt.Sprintf("%d", intent.ID)
			s.seatLockRepository.UnlockSeat(ctx, intent.SeatID, intent.UserID, intentIDStr)
		}

		// Update intent statuses
		if err := tx.Model(&entities.BookingIntent{}).
			Where("id IN ?", intentIDs).
			Update("status", constants.IntentStatusExpired).Error; err != nil {
			tx.Rollback()
			return errors.NewInternalError("Failed to update expired intents", err)
		}

		// Unlock seats
		if err := tx.Model(&entities.Seat{}).
			Where("id IN ?", seatIDs).
			Updates(map[string]interface{}{
				"is_locked": false,
				"locked_at": nil,
				"locked_by": nil,
			}).Error; err != nil {
			tx.Rollback()
			return errors.NewInternalError("Failed to unlock seats", err)
		}
	}

	return tx.Commit().Error
}
