package services

import (
	"api/internal/entities"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// BookingServiceInterface defines the contract for booking operations
type BookingServiceInterface interface {
	CreateBookingIntent(ctx context.Context, userID, seatID uint) (*entities.BookingIntent, error)
	ConfirmBooking(ctx context.Context, bookingIntentID uint, paymentID string) (*entities.Booking, error)
	CancelBookingIntent(ctx context.Context, bookingIntentID uint, userID uint) error
	CancelBooking(ctx context.Context, bookingID uint, userID uint) error
	GetUserBookings(ctx context.Context, userID uint, limit, offset int) ([]entities.Booking, int64, error)
	GetBookingByID(ctx context.Context, bookingID, userID uint) (*entities.Booking, error)
	CleanupExpiredIntents(ctx context.Context) error
}

// EventServiceInterface defines the contract for event operations
type EventServiceInterface interface {
	GetEvents(ctx context.Context, limit, offset int, eventType, city string) ([]entities.Event, int64, error)
	GetEventByID(ctx context.Context, eventID uint) (*entities.Event, error)
	GetAvailableSeats(ctx context.Context, eventID uint) ([]entities.Seat, error)
	GetAvailableSeatsCount(ctx context.Context, eventID uint) (int64, error)
	CreateEvent(ctx context.Context, event *entities.Event) error
	UpdateEvent(ctx context.Context, eventID uint, updates map[string]interface{}) (*entities.Event, error)
	DeleteEvent(ctx context.Context, eventID uint) error
	GetEventStats(ctx context.Context, eventID uint) (map[string]interface{}, error)
}

// UserServiceInterface defines the contract for user operations
type UserServiceInterface interface {
	Register(ctx context.Context, email, password, firstName, lastName, phone string, isAdmin bool) (*entities.User, error)
	Login(ctx context.Context, email, password string) (*entities.User, error)
	GetByID(ctx context.Context, userID uint) (*entities.User, error)
}

// VenueServiceInterface defines the contract for venue operations
type VenueServiceInterface interface {
	GetVenues(ctx context.Context, limit, offset int, city string) ([]entities.Venue, int64, error)
	GetVenueByID(ctx context.Context, venueID uint) (*entities.Venue, error)
	CreateVenue(ctx context.Context, venue *entities.Venue) error
	UpdateVenue(ctx context.Context, venueID uint, updates map[string]interface{}) (*entities.Venue, error)
	DeleteVenue(ctx context.Context, venueID uint) error
}

// QueueServiceInterface defines the contract for queue operations
type QueueServiceInterface interface {
	JoinQueue(ctx context.Context, userID, eventID uint) (*entities.EventQueue, error)
	LeaveQueue(ctx context.Context, userID, eventID uint) error
	GetQueueStatus(ctx context.Context, userID, eventID uint) (*entities.EventQueue, error)
	GetQueueLength(ctx context.Context, eventID uint) (int64, error)
}

// WaitlistServiceInterface defines the contract for waitlist operations
type WaitlistServiceInterface interface {
	JoinWaitlist(ctx context.Context, userID, eventID uint) (*WaitlistEntry, error)
	GetWaitlistPosition(ctx context.Context, userID, eventID uint) (*WaitlistEntry, error)
	LeaveWaitlist(ctx context.Context, userID, eventID uint) error
	GetWaitlistSize(ctx context.Context, eventID uint) (int, error)
	ProcessSeatAvailability(ctx context.Context, eventID uint, availableSeats int) ([]*WaitlistEntry, error)
	CleanupExpiredWaitlist(ctx context.Context) error
	RemoveUserFromWaitlistAfterBooking(ctx context.Context, userID, eventID uint) error
}

type WaitlistEntry struct {
	UserID    uint      `json:"user_id"`
	EventID   uint      `json:"event_id"`
	JoinedAt  time.Time `json:"joined_at"`
	Position  int       `json:"position"`
	NotifiedAt *time.Time `json:"notified_at,omitempty"`
}

// JWTServiceInterface defines the contract for JWT operations
type JWTServiceInterface interface {
	GenerateToken(userID uint, isAdmin bool) (string, error)
	ValidateToken(tokenStr string) (*jwt.Token, error)
	GetClaimsFromToken(tokenStr string) (jwt.MapClaims, error)
}

// SeatLockServiceInterface defines the contract for seat locking operations
type SeatLockServiceInterface interface {
	LockSeat(ctx context.Context, seatID uint, userID uint, intentID string) error
	UnlockSeat(ctx context.Context, seatID uint, userID uint, intentID string) error
	IsLocked(ctx context.Context, seatID uint) (bool, string, error)
	ExtendLock(ctx context.Context, seatID uint, userID uint, intentID string) error
	GetLockTTL(ctx context.Context, seatID uint) (time.Duration, error)
	CleanupExpiredLocks(ctx context.Context) error
}
