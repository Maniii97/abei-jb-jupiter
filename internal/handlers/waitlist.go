package handlers

import (
	"api/internal/services"
	"api/pkg/errors"
	"api/pkg/response"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WaitlistHandler struct {
	waitlistService services.WaitlistServiceInterface
}

func NewWaitlistHandler(waitlistService services.WaitlistServiceInterface) *WaitlistHandler {
	return &WaitlistHandler{
		waitlistService: waitlistService,
	}
}

// JoinWaitlist adds a user to the event waitlist
func (h *WaitlistHandler) JoinWaitlist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	entry, err := h.waitlistService.JoinWaitlist(context.Background(), userID.(uint), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	waitlistResp := response.WaitlistResponse{
		EventID:  entry.EventID,
		UserID:   entry.UserID,
		Position: entry.Position,
		JoinedAt: entry.JoinedAt,
		Status:   "waiting",
	}

	response.Success(c, http.StatusCreated, "Successfully joined waitlist", waitlistResp)
}

// GetWaitlistPosition returns the current position of a user in the waitlist
func (h *WaitlistHandler) GetWaitlistPosition(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	entry, err := h.waitlistService.GetWaitlistPosition(context.Background(), userID.(uint), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	var status string
	if entry.NotifiedAt != nil {
		status = "notified"
	} else {
		status = "waiting"
	}

	waitlistResp := response.WaitlistResponse{
		EventID:    entry.EventID,
		UserID:     entry.UserID,
		Position:   entry.Position,
		Status:     status,
		JoinedAt:   entry.JoinedAt,
		NotifiedAt: entry.NotifiedAt,
	}

	response.Success(c, http.StatusOK, "Waitlist position retrieved", waitlistResp)
}

// LeaveWaitlist removes a user from the waitlist
func (h *WaitlistHandler) LeaveWaitlist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	err = h.waitlistService.LeaveWaitlist(context.Background(), userID.(uint), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Successfully left waitlist", nil)
}

// GetWaitlistStats returns waitlist statistics for an event
func (h *WaitlistHandler) GetWaitlistStats(c *gin.Context) {
	eventIDStr := c.Param("eventId")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event ID")
		return
	}

	size, err := h.waitlistService.GetWaitlistSize(context.Background(), uint(eventID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	sizeResp := map[string]interface{}{
		"event_id":      eventID,
		"waitlist_size": size,
	}

	response.Success(c, http.StatusOK, "Waitlist size retrieved", sizeResp)
}

// handleError handles different types of errors and sends appropriate responses
func (h *WaitlistHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case "BAD_REQUEST":
			response.Error(c, http.StatusBadRequest, appErr.Message)
		case "NOT_FOUND":
			response.Error(c, http.StatusNotFound, appErr.Message)
		case "CONFLICT":
			response.Error(c, http.StatusConflict, appErr.Message)
		case "UNAUTHORIZED":
			response.Error(c, http.StatusUnauthorized, appErr.Message)
		case "INTERNAL_ERROR":
			response.Error(c, http.StatusInternalServerError, appErr.Message)
		default:
			response.Error(c, http.StatusInternalServerError, appErr.Message)
		}
	} else {
		response.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
	}
}
