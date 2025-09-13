package constants

// HTTP Status Codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusAccepted            = 202
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
	StatusServiceUnavailable  = 503
)

// Booking Status
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusRefunded  = "refunded"
)

// Payment Status
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// Booking Intent Status
const (
	IntentStatusPending   = "pending"
	IntentStatusExpired   = "expired"
	IntentStatusConfirmed = "confirmed"
	IntentStatusCancelled = "cancelled"
)

// Event Status
const (
	EventStatusActive    = "active"
	EventStatusCancelled = "cancelled"
	EventStatusCompleted = "completed"
	EventStatusSoldOut   = "sold_out"
)

// Queue Status
const (
	QueueStatusWaiting   = "waiting"
	QueueStatusActive    = "active"
	QueueStatusExpired   = "expired"
	QueueStatusCompleted = "completed"
)

// Seat Types
const (
	SeatTypeStandard = "standard"
	SeatTypePremium  = "premium"
	SeatTypeVIP      = "vip"
)

// Event Types
const (
	EventTypeConcert    = "concert"
	EventTypeTheater    = "theater"
	EventTypeSports     = "sports"
	EventTypeConference = "conference"
	EventTypeOther      = "other"
)

// Redis Keys
const (
	SeatLockPrefix    = "seat_lock:"
	QueuePrefix       = "queue:"
	UserSessionPrefix = "user_session:"
)

// Lock Durations (in minutes)
const (
	SeatLockDuration    = 8
	QueueActiveDuration = 10
)

// Error Messages
const (
	ErrSeatNotAvailable    = "seat is not available"
	ErrSeatAlreadyLocked   = "seat is already locked by another user"
	ErrPaymentFailed       = "payment processing failed"
	ErrBookingExpired      = "booking intent has expired"
	ErrInsufficientSeats   = "not enough seats available"
	ErrEventSoldOut        = "event is sold out"
	ErrEventNotFound       = "event not found"
	ErrUnauthorizedAccess  = "unauthorized access"
	ErrInvalidBookingState = "invalid booking state"
	ErrVenueTimeConflict   = "venue is already booked for another event during this time period"
)
