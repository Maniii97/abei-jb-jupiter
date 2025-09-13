package handlers

import (
	"api/internal/services"
	"api/pkg/errors"
	"api/pkg/request"
	"api/pkg/response"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService services.BookingServiceInterface
}

func NewBookingHandler(bookingService services.BookingServiceInterface) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBookingIntent creates a booking intent and locks the seat
func (h *BookingHandler) CreateBookingIntent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req request.CreateBookingIntentRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	intent, err := h.bookingService.CreateBookingIntent(context.Background(), userID.(uint), req.SeatID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	intentResp := response.BookingIntentResponse{
		ID: intent.ID,
		Event: response.EventResponse{
			ID:          intent.Event.ID,
			Name:        intent.Event.Name,
			Description: intent.Event.Description,
			Venue: response.VenueResponse{
				ID:          intent.Event.Venue.ID,
				Name:        intent.Event.Venue.Name,
				Address:     intent.Event.Venue.Address,
				City:        intent.Event.Venue.City,
				State:       intent.Event.Venue.State,
				Country:     intent.Event.Venue.Country,
				Rows:        intent.Event.Venue.Rows,
				Columns:     intent.Event.Venue.Columns,
				Capacity:    intent.Event.Venue.Rows * intent.Event.Venue.Columns,
				Description: intent.Event.Venue.Description,
			},
			StartTime:      intent.Event.StartTime,
			EndTime:        intent.Event.EndTime,
			Capacity:       intent.Event.Venue.Rows * intent.Event.Venue.Columns,
			AvailableSeats: intent.Event.AvailableSeats,
			Price:          intent.Event.Price,
			EventType:      intent.Event.EventType,
			Status:         intent.Event.Status,
			IsHighDemand:   intent.Event.IsHighDemand,
		},
		Seat: response.SeatResponse{
			ID:          intent.Seat.ID,
			Row:         intent.Seat.Row,
			Column:      intent.Seat.Column,
			SeatType:    intent.Seat.SeatType,
			Price:       intent.Seat.Price,
			IsAvailable: intent.Seat.IsAvailable,
			IsLocked:    intent.Seat.IsLocked,
		},
		Status: intent.Status,
	}

	response.Success(c, http.StatusCreated, "booking intent created successfully", intentResp)
}

// ConfirmBooking confirms a booking intent after successful payment
func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req request.ConfirmBookingRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	booking, err := h.bookingService.ConfirmBooking(context.Background(), req.BookingIntentID, req.PaymentID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Verify the booking belongs to the authenticated user
	if booking.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, "unauthorized access to booking")
		return
	}

	bookingResp := response.BookingResponse{
		ID: booking.ID,
		Event: response.EventResponse{
			ID:          booking.Event.ID,
			Name:        booking.Event.Name,
			Description: booking.Event.Description,
			Venue: response.VenueResponse{
				ID:          booking.Event.Venue.ID,
				Name:        booking.Event.Venue.Name,
				Address:     booking.Event.Venue.Address,
				City:        booking.Event.Venue.City,
				State:       booking.Event.Venue.State,
				Country:     booking.Event.Venue.Country,
				Rows:        booking.Event.Venue.Rows,
				Columns:     booking.Event.Venue.Columns,
				Capacity:    booking.Event.Venue.Rows * booking.Event.Venue.Columns,
				Description: booking.Event.Venue.Description,
			},
			StartTime:      booking.Event.StartTime,
			EndTime:        booking.Event.EndTime,
			Capacity:       booking.Event.Venue.Rows * booking.Event.Venue.Columns,
			AvailableSeats: booking.Event.AvailableSeats,
			Price:          booking.Event.Price,
			EventType:      booking.Event.EventType,
			Status:         booking.Event.Status,
			IsHighDemand:   booking.Event.IsHighDemand,
		},
		Seat: response.SeatResponse{
			ID:          booking.Seat.ID,
			Row:         booking.Seat.Row,
			Column:      booking.Seat.Column,
			SeatType:    booking.Seat.SeatType,
			Price:       booking.Seat.Price,
			IsAvailable: booking.Seat.IsAvailable,
			IsLocked:    booking.Seat.IsLocked,
		},
		Status:        booking.Status,
		PaymentStatus: booking.PaymentStatus,
		TotalAmount:   booking.TotalAmount,
		BookedAt:      booking.BookedAt,
		CancelledAt:   booking.CancelledAt,
	}

	response.Success(c, http.StatusOK, "booking confirmed successfully", bookingResp)
}

// CancelBookingIntent cancels a booking intent and unlocks the seat
func (h *BookingHandler) CancelBookingIntent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req request.CancelBookingIntentRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	if err := h.bookingService.CancelBookingIntent(context.Background(), req.BookingIntentID, userID.(uint)); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "booking intent cancelled successfully", nil)
}

// CancelBooking cancels a confirmed booking
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	bookingIDStr := c.Param("id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid booking ID")
		return
	}

	if err := h.bookingService.CancelBooking(context.Background(), uint(bookingID), userID.(uint)); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "booking cancelled successfully", nil)
}

// GetUserBookings returns user's booking history
func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req request.PaginationRequest
	if err := request.BindQuery(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request parameters", err.Error())
		return
	}

	offset := (req.Page - 1) * req.Limit
	bookings, total, err := h.bookingService.GetUserBookings(context.Background(), userID.(uint), req.Limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response format
	bookingResponses := make([]response.BookingResponse, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = response.BookingResponse{
			ID: booking.ID,
			Event: response.EventResponse{
				ID:          booking.Event.ID,
				Name:        booking.Event.Name,
				Description: booking.Event.Description,
				Venue: response.VenueResponse{
					ID:          booking.Event.Venue.ID,
					Name:        booking.Event.Venue.Name,
					Address:     booking.Event.Venue.Address,
					City:        booking.Event.Venue.City,
					State:       booking.Event.Venue.State,
					Country:     booking.Event.Venue.Country,
					Rows:        booking.Event.Venue.Rows,
					Columns:     booking.Event.Venue.Columns,
					Capacity:    booking.Event.Venue.Rows * booking.Event.Venue.Columns,
					Description: booking.Event.Venue.Description,
				},
				StartTime:      booking.Event.StartTime,
				EndTime:        booking.Event.EndTime,
				Capacity:       booking.Event.Venue.Rows * booking.Event.Venue.Columns,
				AvailableSeats: booking.Event.AvailableSeats,
				Price:          booking.Event.Price,
				EventType:      booking.Event.EventType,
				Status:         booking.Event.Status,
				IsHighDemand:   booking.Event.IsHighDemand,
			},
			Seat: response.SeatResponse{
				ID:          booking.Seat.ID,
				Row:         booking.Seat.Row,
				Column:      booking.Seat.Column,
				SeatType:    booking.Seat.SeatType,
				Price:       booking.Seat.Price,
				IsAvailable: booking.Seat.IsAvailable,
				IsLocked:    booking.Seat.IsLocked,
			},
			Status:        booking.Status,
			PaymentStatus: booking.PaymentStatus,
			TotalAmount:   booking.TotalAmount,
			BookedAt:      booking.BookedAt,
			CancelledAt:   booking.CancelledAt,
		}
	}

	response.Paginated(c, http.StatusOK, bookingResponses, req.Page, req.Limit, total)
}

// GetBookingByID returns a specific booking
func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	bookingIDStr := c.Param("id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid booking ID")
		return
	}

	booking, err := h.bookingService.GetBookingByID(context.Background(), uint(bookingID), userID.(uint))
	if err != nil {
		h.handleError(c, err)
		return
	}

	bookingResp := response.BookingResponse{
		ID: booking.ID,
		Event: response.EventResponse{
			ID:          booking.Event.ID,
			Name:        booking.Event.Name,
			Description: booking.Event.Description,
			Venue: response.VenueResponse{
				ID:          booking.Event.Venue.ID,
				Name:        booking.Event.Venue.Name,
				Address:     booking.Event.Venue.Address,
				City:        booking.Event.Venue.City,
				State:       booking.Event.Venue.State,
				Country:     booking.Event.Venue.Country,
				Rows:        booking.Event.Venue.Rows,
				Columns:     booking.Event.Venue.Columns,
				Capacity:    booking.Event.Venue.Rows * booking.Event.Venue.Columns,
				Description: booking.Event.Venue.Description,
			},
			StartTime:      booking.Event.StartTime,
			EndTime:        booking.Event.EndTime,
			Capacity:       booking.Event.Venue.Rows * booking.Event.Venue.Columns,
			AvailableSeats: booking.Event.AvailableSeats,
			Price:          booking.Event.Price,
			EventType:      booking.Event.EventType,
			Status:         booking.Event.Status,
			IsHighDemand:   booking.Event.IsHighDemand,
		},
		Seat: response.SeatResponse{
			ID:          booking.Seat.ID,
			Row:         booking.Seat.Row,
			Column:      booking.Seat.Column,
			SeatType:    booking.Seat.SeatType,
			Price:       booking.Seat.Price,
			IsAvailable: booking.Seat.IsAvailable,
			IsLocked:    booking.Seat.IsLocked,
		},
		Status:        booking.Status,
		PaymentStatus: booking.PaymentStatus,
		TotalAmount:   booking.TotalAmount,
		BookedAt:      booking.BookedAt,
		CancelledAt:   booking.CancelledAt,
	}

	response.JSON(c, http.StatusOK, bookingResp)
}

// handleError converts application errors to appropriate HTTP responses
func (h *BookingHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case "BAD_REQUEST":
			response.Error(c, http.StatusBadRequest, appErr.Message)
		case "UNAUTHORIZED":
			response.Error(c, http.StatusUnauthorized, appErr.Message)
		case "NOT_FOUND":
			response.Error(c, http.StatusNotFound, appErr.Message)
		case "CONFLICT":
			response.Error(c, http.StatusConflict, appErr.Message)
		case "INTERNAL_ERROR":
			response.Error(c, http.StatusInternalServerError, "internal server error")
		default:
			response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	} else {
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
