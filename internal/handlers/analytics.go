package handlers

import (
	"api/internal/services"
	"api/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService services.AnalyticsServiceInterface
}

func NewAnalyticsHandler(analyticsService services.AnalyticsServiceInterface) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetBookingAnalytics handles GET /admin/analytics/bookings
// @Summary Get booking analytics for admin dashboard
// @Description Retrieve comprehensive booking analytics including total bookings, popular events, capacity utilization, cancellation rates, and daily stats
// @Tags Admin Analytics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} entities.BookingAnalytics
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/analytics/bookings [get]
func (h *AnalyticsHandler) GetBookingAnalytics(c *gin.Context) {
	analytics, err := h.analyticsService.GetBookingAnalytics()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to retrieve booking analytics")
		return
	}

	response.Success(c, http.StatusOK, "booking analytics retrieved successfully", analytics)
}
