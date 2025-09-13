package handlers

import (
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

type VenueHandler struct {
	venueService services.VenueServiceInterface
}

func NewVenueHandler(venueService services.VenueServiceInterface) *VenueHandler {
	return &VenueHandler{
		venueService: venueService,
	}
}

// GetVenues returns a list of venues with pagination and filters
func (h *VenueHandler) GetVenues(c *gin.Context) {
	var req request.VenueFilterRequest
	if err := request.BindQuery(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request parameters", err.Error())
		return
	}

	offset := (req.Page - 1) * req.Limit
	venues, total, err := h.venueService.GetVenues(context.Background(), req.Limit, offset, req.City)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert to response format
	venueResponses := make([]response.VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = response.VenueResponse{
			ID:          venue.ID,
			Name:        venue.Name,
			Address:     venue.Address,
			City:        venue.City,
			State:       venue.State,
			Country:     venue.Country,
			Rows:        venue.Rows,
			Columns:     venue.Columns,
			Capacity:    venue.Rows * venue.Columns,
			Description: venue.Description,
		}
	}

	response.Paginated(c, http.StatusOK, venueResponses, req.Page, req.Limit, total)
}

// GetVenueByID returns a single venue with details
func (h *VenueHandler) GetVenueByID(c *gin.Context) {
	venueIDStr := c.Param("id")
	venueID, err := strconv.ParseUint(venueIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid venue ID")
		return
	}

	venue, err := h.venueService.GetVenueByID(context.Background(), uint(venueID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert events to response format
	eventResponses := make([]response.EventResponse, len(venue.Events))
	for i, event := range venue.Events {
		eventResponses[i] = response.EventResponse{
			ID:           event.ID,
			Name:         event.Name,
			Description:  event.Description,
			StartTime:    event.StartTime,
			EndTime:      event.EndTime,
			Price:        event.Price,
			EventType:    event.EventType,
			Status:       event.Status,
			IsHighDemand: event.IsHighDemand,
		}
	}

	venueResp := response.VenueDetailResponse{
		VenueResponse: response.VenueResponse{
			ID:          venue.ID,
			Name:        venue.Name,
			Address:     venue.Address,
			City:        venue.City,
			State:       venue.State,
			Country:     venue.Country,
			Rows:        venue.Rows,
			Columns:     venue.Columns,
			Capacity:    venue.Rows * venue.Columns,
			Description: venue.Description,
		},
		Events: eventResponses,
	}

	response.JSON(c, http.StatusOK, venueResp)
}

// CreateVenue creates a new venue (admin only)
func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var req request.CreateVenueRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	venue := &entities.Venue{
		Name:        req.Name,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		Rows:        req.Rows,
		Columns:     req.Columns,
		Description: req.Description,
	}

	if err := h.venueService.CreateVenue(context.Background(), venue); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "venue created successfully", map[string]uint{"venue_id": venue.ID})
}

// UpdateVenue updates an existing venue (admin only)
func (h *VenueHandler) UpdateVenue(c *gin.Context) {
	venueIDStr := c.Param("id")
	venueID, err := strconv.ParseUint(venueIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid venue ID")
		return
	}

	var req request.UpdateVenueRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.State != nil {
		updates["state"] = *req.State
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Rows != nil {
		updates["rows"] = *req.Rows
	}
	if req.Columns != nil {
		updates["columns"] = *req.Columns
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	venue, err := h.venueService.UpdateVenue(context.Background(), uint(venueID), updates)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "venue updated successfully", map[string]uint{"venue_id": venue.ID})
}

// DeleteVenue deletes a venue (admin only)
func (h *VenueHandler) DeleteVenue(c *gin.Context) {
	venueIDStr := c.Param("id")
	venueID, err := strconv.ParseUint(venueIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid venue ID")
		return
	}

	if err := h.venueService.DeleteVenue(context.Background(), uint(venueID)); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "venue deleted successfully", nil)
}

// handleError converts application errors to appropriate HTTP responses
func (h *VenueHandler) handleError(c *gin.Context, err error) {
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
