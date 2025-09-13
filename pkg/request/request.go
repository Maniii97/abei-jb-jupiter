package request

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Auth requests
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	IsAdmin   bool   `json:"is_admin"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Venue requests
type CreateVenueRequest struct {
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state" binding:"required"`
	Country     string `json:"country" binding:"required"`
	Rows        int    `json:"rows" binding:"required,min=1"`
	Columns     int    `json:"columns" binding:"required,min=1"`
	Description string `json:"description"`
}

type UpdateVenueRequest struct {
	Name        *string `json:"name"`
	Address     *string `json:"address"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	Country     *string `json:"country"`
	Rows        *int    `json:"rows"`
	Columns     *int    `json:"columns"`
	Description *string `json:"description"`
}

// Event requests
type CreateEventRequest struct {
	Name         string    `json:"name" binding:"required"`
	Description  string    `json:"description"`
	VenueID      uint      `json:"venue_id" binding:"required"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	Price        float64   `json:"price" binding:"required,min=0"`
	EventType    string    `json:"event_type" binding:"required"`
	IsHighDemand bool      `json:"is_high_demand"`
}

type UpdateEventRequest struct {
	Name         *string    `json:"name"`
	Description  *string    `json:"description"`
	VenueID      *uint      `json:"venue_id"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Price        *float64   `json:"price"`
	EventType    *string    `json:"event_type"`
	IsHighDemand *bool      `json:"is_high_demand"`
	Status       *string    `json:"status"`
}

// Booking requests
type CreateBookingIntentRequest struct {
	SeatID uint `json:"seat_id" binding:"required"`
}

type ConfirmBookingRequest struct {
	BookingIntentID uint   `json:"booking_intent_id" binding:"required"`
	PaymentID       string `json:"payment_id" binding:"required"`
}

type CancelBookingIntentRequest struct {
	BookingIntentID uint `json:"booking_intent_id" binding:"required"`
}

// Queue requests
type JoinQueueRequest struct {
	EventID uint `json:"event_id" binding:"required"`
}

// Pagination and filtering
type PaginationRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}

type EventFilterRequest struct {
	PaginationRequest
	City      string `form:"city"`
	EventType string `form:"event_type"`
}

type VenueFilterRequest struct {
	PaginationRequest
	City string `form:"city"`
}

// Helper function to bind JSON request
func BindJSON(c *gin.Context, req interface{}) error {
	return c.ShouldBindJSON(req)
}

// Helper function to bind query parameters
func BindQuery(c *gin.Context, req interface{}) error {
	return c.ShouldBindQuery(req)
}
