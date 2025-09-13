package handlers

import (
	"api/constants"
	"api/internal/entities"
	"api/internal/services"
	"api/pkg/errors"
	"api/pkg/request"
	"api/pkg/response"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService services.EventServiceInterface
	venueService services.VenueServiceInterface
}

func NewEventHandler(eventService services.EventServiceInterface, venueService services.VenueServiceInterface) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		venueService: venueService,
	}
}

// GetEvents returns a list of events with pagination and filters
func (h *EventHandler) GetEvents(c *gin.Context) {
	var req request.EventFilterRequest
	if err := request.BindQuery(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request parameters", err.Error())
		return
	}

	offset := (req.Page - 1) * req.Limit
	events, total, err := h.eventService.GetEvents(context.Background(), req.Limit, offset, req.EventType, req.City)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response format
	eventResponses := make([]response.EventResponse, len(events))
	for i, event := range events {
		// Calculate available seats using the service
		availableSeats, err := h.eventService.GetAvailableSeatsCount(context.Background(), event.ID)
		if err != nil {
			// Log error but don't fail the request, just set to 0
			availableSeats = 0
		}

		eventResponses[i] = response.EventResponse{
			ID:          event.ID,
			Name:        event.Name,
			Description: event.Description,
			Venue: response.VenueResponse{
				ID:          event.Venue.ID,
				Name:        event.Venue.Name,
				Address:     event.Venue.Address,
				City:        event.Venue.City,
				State:       event.Venue.State,
				Country:     event.Venue.Country,
				Rows:        event.Venue.Rows,
				Columns:     event.Venue.Columns,
				Capacity:    event.Venue.Rows * event.Venue.Columns,
				Description: event.Venue.Description,
			},
			StartTime:      event.StartTime,
			EndTime:        event.EndTime,
			Capacity:       event.Venue.Rows * event.Venue.Columns,
			AvailableSeats: int(availableSeats),
			Price:          event.Price,
			EventType:      event.EventType,
			Status:         event.Status,
			IsHighDemand:   event.IsHighDemand,
		}
	}

	response.Paginated(c, http.StatusOK, eventResponses, req.Page, req.Limit, total)
}

// GetEventByID returns a single event with details
func (h *EventHandler) GetEventByID(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	event, err := h.eventService.GetEventByID(context.Background(), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert seats to response format
	seatResponses := make([]response.SeatResponse, len(event.Seats))
	for i, seat := range event.Seats {
		seatResponses[i] = response.SeatResponse{
			ID:          seat.ID,
			Row:         seat.Row,
			Column:      seat.Column,
			SeatType:    seat.SeatType,
			Price:       seat.Price,
			IsAvailable: seat.IsAvailable,
			IsLocked:    seat.IsLocked,
		}
	}

	// Calculate available seats count using the service
	availableSeats, err := h.eventService.GetAvailableSeatsCount(context.Background(), event.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	eventResp := response.EventDetailResponse{
		EventResponse: response.EventResponse{
			ID:          event.ID,
			Name:        event.Name,
			Description: event.Description,
			Venue: response.VenueResponse{
				ID:          event.Venue.ID,
				Name:        event.Venue.Name,
				Address:     event.Venue.Address,
				City:        event.Venue.City,
				State:       event.Venue.State,
				Country:     event.Venue.Country,
				Rows:        event.Venue.Rows,
				Columns:     event.Venue.Columns,
				Capacity:    event.Venue.Rows * event.Venue.Columns,
				Description: event.Venue.Description,
			},
			StartTime:      event.StartTime,
			EndTime:        event.EndTime,
			Capacity:       event.Venue.Rows * event.Venue.Columns,
			AvailableSeats: int(availableSeats),
			Price:          event.Price,
			EventType:      event.EventType,
			Status:         event.Status,
			IsHighDemand:   event.IsHighDemand,
		},
		Seats: seatResponses,
	}

	response.JSON(c, http.StatusOK, eventResp)
}

// GetAvailableSeats returns available seats for an event
func (h *EventHandler) GetAvailableSeats(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Check if event exists
	_, err = h.eventService.GetEventByID(context.Background(), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	seats, err := h.eventService.GetAvailableSeats(context.Background(), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Check if no seats are available
	if len(seats) == 0 {
		response.Error(c, http.StatusNotFound, "no available seats found for this event")
		return
	}

	// Convert to response format
	seatResponses := make([]response.SeatResponse, len(seats))
	for i, seat := range seats {
		seatResponses[i] = response.SeatResponse{
			ID:          seat.ID,
			Row:         seat.Row,
			Column:      seat.Column,
			SeatType:    seat.SeatType,
			Price:       seat.Price,
			IsAvailable: seat.IsAvailable,
			IsLocked:    seat.IsLocked,
		}
	}

	response.JSON(c, http.StatusOK, seatResponses)
}

// CreateEvent creates a new event (admin only)
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req request.CreateEventRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	// Validate venue exists
	_, err := h.venueService.GetVenueByID(context.Background(), req.VenueID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "venue not found")
		return
	}

	// Create event entity
	event := &entities.Event{
		Name:         req.Name,
		Description:  req.Description,
		VenueID:      req.VenueID,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Price:        req.Price,
		EventType:    req.EventType,
		Status:       constants.EventStatusActive,
		IsHighDemand: req.IsHighDemand,
	}

	if err := h.eventService.CreateEvent(context.Background(), event); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "event created successfully", map[string]uint{"event_id": event.ID})
}

// UpdateEvent updates an existing event (admin only)
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	var req request.UpdateEventRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.VenueID != nil {
		updates["venue_id"] = *req.VenueID
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.EventType != nil {
		updates["event_type"] = *req.EventType
	}
	if req.IsHighDemand != nil {
		updates["is_high_demand"] = *req.IsHighDemand
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	event, err := h.eventService.UpdateEvent(context.Background(), uint(eventID), updates)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "event updated successfully", map[string]uint{"event_id": event.ID})
}

// DeleteEvent deletes an event (admin only)
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	if err := h.eventService.DeleteEvent(context.Background(), uint(eventID)); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "event cancelled successfully", nil)
}

// GetEventStats returns event statistics (admin only)
func (h *EventHandler) GetEventStats(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	stats, err := h.eventService.GetEventStats(context.Background(), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	statsResp := response.EventStatsResponse{
		EventID:             stats["event_id"].(uint),
		EventName:           stats["event_name"].(string),
		TotalSeats:          stats["total_seats"].(int64),
		BookedSeats:         stats["booked_seats"].(int64),
		LockedSeats:         stats["locked_seats"].(int64),
		AvailableSeats:      stats["available_seats"].(int64),
		CapacityUtilization: stats["capacity_utilization"].(float64),
		TotalRevenue:        stats["total_revenue"].(float64),
		BookingRate:         stats["booking_rate"].(float64),
	}

	response.JSON(c, http.StatusOK, statsResp)
}

// handleError converts application errors to appropriate HTTP responses
func (h *EventHandler) handleError(c *gin.Context, err error) {
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
