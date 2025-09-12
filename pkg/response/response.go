package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Auth responses
type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	IsAdmin   bool   `json:"is_admin"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// Venue responses
type VenueResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Capacity    int    `json:"capacity"`
	Description string `json:"description"`
}

type VenueDetailResponse struct {
	VenueResponse
	Events []EventResponse `json:"events,omitempty"`
}

// Event responses
type EventResponse struct {
	ID             uint          `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Venue          VenueResponse `json:"venue"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Capacity       int           `json:"capacity"`
	AvailableSeats int           `json:"available_seats"`
	Price          float64       `json:"price"`
	EventType      string        `json:"event_type"`
	Status         string        `json:"status"`
	IsHighDemand   bool          `json:"is_high_demand"`
}

type EventDetailResponse struct {
	EventResponse
	Seats []SeatResponse `json:"seats,omitempty"`
}

// Seat responses
type SeatResponse struct {
	ID          uint    `json:"id"`
	SeatNumber  string  `json:"seat_number"`
	Row         string  `json:"row"`
	Section     string  `json:"section"`
	SeatType    string  `json:"seat_type"`
	Price       float64 `json:"price"`
	IsAvailable bool    `json:"is_available"`
	IsLocked    bool    `json:"is_locked"`
}

// Booking responses
type BookingIntentResponse struct {
	ID            uint          `json:"id"`
	IntentID      string        `json:"intent_id"`
	Event         EventResponse `json:"event"`
	Seat          SeatResponse  `json:"seat"`
	Status        string        `json:"status"`
	LockExpiresAt time.Time     `json:"lock_expires_at"`
}

type BookingResponse struct {
	ID            uint          `json:"id"`
	BookingNumber string        `json:"booking_number"`
	Event         EventResponse `json:"event"`
	Seat          SeatResponse  `json:"seat"`
	Status        string        `json:"status"`
	PaymentStatus string        `json:"payment_status"`
	TotalAmount   float64       `json:"total_amount"`
	BookedAt      time.Time     `json:"booked_at"`
	CancelledAt   *time.Time    `json:"cancelled_at,omitempty"`
}

// Queue responses
type QueueResponse struct {
	ID            uint       `json:"id"`
	QueuePosition int        `json:"queue_position"`
	Status        string     `json:"status"`
	JoinedAt      time.Time  `json:"joined_at"`
	ActiveAt      *time.Time `json:"active_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

// Pagination responses
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// Analytics responses
type EventStatsResponse struct {
	EventID             uint    `json:"event_id"`
	EventName           string  `json:"event_name"`
	TotalSeats          int64   `json:"total_seats"`
	BookedSeats         int64   `json:"booked_seats"`
	LockedSeats         int64   `json:"locked_seats"`
	AvailableSeats      int64   `json:"available_seats"`
	CapacityUtilization float64 `json:"capacity_utilization"`
	TotalRevenue        float64 `json:"total_revenue"`
	BookingRate         float64 `json:"booking_rate"`
}

// Generic responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Gin response helpers
func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, err string, message ...string) {
	response := ErrorResponse{Error: err}
	if len(message) > 0 {
		response.Message = message[0]
	}
	c.JSON(status, response)
}

func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func Paginated(c *gin.Context, status int, data interface{}, page, limit int, total int64) {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	c.JSON(status, PaginatedResponse{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}
