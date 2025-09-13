package services

import (
	"api/internal/entities"
	"api/internal/repository"
	"context"
)

type BookingService struct {
	bookingRepo     *repository.BookingRepository
	seatLockService *SeatLockService
}

// Ensure BookingService implements BookingServiceInterface
var _ BookingServiceInterface = (*BookingService)(nil)

func NewBookingService(bookingRepo *repository.BookingRepository, seatLockService *SeatLockService) *BookingService {
	return &BookingService{
		bookingRepo:     bookingRepo,
		seatLockService: seatLockService,
	}
}

// CreateBookingIntent creates a booking intent and locks the seat
func (s *BookingService) CreateBookingIntent(ctx context.Context, userID, seatID uint) (*entities.BookingIntent, error) {
	return s.bookingRepo.CreateBookingIntent(ctx, userID, seatID)
}

func (s *BookingService) ConfirmBooking(ctx context.Context, bookingIntentID uint, paymentID string) (*entities.Booking, error) {
	return s.bookingRepo.ConfirmBooking(ctx, bookingIntentID, paymentID)
}

func (s *BookingService) CancelBookingIntent(ctx context.Context, bookingIntentID uint, userID uint) error {
	return s.bookingRepo.CancelBookingIntent(ctx, bookingIntentID, userID)
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID uint, userID uint) error {
	return s.bookingRepo.CancelBooking(ctx, bookingID, userID)
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uint, limit, offset int) ([]entities.Booking, int64, error) {
	return s.bookingRepo.GetUserBookings(ctx, userID, limit, offset)
}

func (s *BookingService) GetBookingByID(ctx context.Context, bookingID, userID uint) (*entities.Booking, error) {
	return s.bookingRepo.GetBookingByID(ctx, bookingID, userID)
}

func (s *BookingService) CleanupExpiredIntents(ctx context.Context) error {
	return s.bookingRepo.CleanupExpiredIntents(ctx)
}
