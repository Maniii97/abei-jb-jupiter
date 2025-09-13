package mocks

import (
	"api/internal/entities"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockBookingService struct {
	mock.Mock
}

func (m *MockBookingService) CreateBookingIntent(ctx context.Context, userID, seatID uint) (*entities.BookingIntent, error) {
	args := m.Called(ctx, userID, seatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.BookingIntent), args.Error(1)
}

func (m *MockBookingService) ConfirmBooking(ctx context.Context, bookingIntentID uint, paymentID string) (*entities.Booking, error) {
	args := m.Called(ctx, bookingIntentID, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Booking), args.Error(1)
}

func (m *MockBookingService) CancelBookingIntent(ctx context.Context, bookingIntentID uint, userID uint) error {
	args := m.Called(ctx, bookingIntentID, userID)
	return args.Error(0)
}

func (m *MockBookingService) CancelBooking(ctx context.Context, bookingID uint, userID uint) error {
	args := m.Called(ctx, bookingID, userID)
	return args.Error(0)
}

func (m *MockBookingService) GetUserBookings(ctx context.Context, userID uint, limit, offset int) ([]entities.Booking, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]entities.Booking), args.Get(1).(int64), args.Error(2)
}

func (m *MockBookingService) GetBookingByID(ctx context.Context, bookingID, userID uint) (*entities.Booking, error) {
	args := m.Called(ctx, bookingID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Booking), args.Error(1)
}

func (m *MockBookingService) CleanupExpiredIntents(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
