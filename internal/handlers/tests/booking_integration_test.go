package tests

import (
	"api/internal/entities"
	"api/internal/handlers"
	"api/pkg/request"
	"api/test"
	"api/test/mocks"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestBookingFlowIntegration tests the complete booking flow
func TestBookingFlowIntegration(t *testing.T) {
	// Setup
	router := test.SetupTestGin()
	bookingService := &mocks.MockBookingService{}
	handler := handlers.NewBookingHandler(bookingService)
	mockEntities := &test.MockEntities{}

	// Setup routes with auth middleware
	api := router.Group("/api")
	protected := api.Group("/")
	protected.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	{
		protected.POST("/booking-intents", handler.CreateBookingIntent)
		protected.POST("/bookings/confirm", handler.ConfirmBooking)
		protected.GET("/bookings", handler.GetUserBookings)
	}

	// Test data
	mockIntent := mockEntities.GetMockBookingIntent()
	mockBooking := mockEntities.GetMockBooking()

	// Step 1: Create booking intent
	bookingService.On("CreateBookingIntent",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(mockIntent, nil).Once()

	createReq := request.CreateBookingIntentRequest{SeatID: 1}
	req1, _ := test.CreateTestRequest("POST", "/api/booking-intents", createReq)
	w1 := test.ExecuteRequest(router, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &createResponse)
	assert.NoError(t, err)
	assert.Equal(t, "booking intent created successfully", createResponse["message"])

	// Step 2: Confirm booking
	bookingService.On("ConfirmBooking",
		mock.Anything,
		uint(1),
		"pay_test123",
	).Return(mockBooking, nil).Once()

	confirmReq := request.ConfirmBookingRequest{
		BookingIntentID: 1,
		PaymentID:       "pay_test123",
	}
	req2, _ := test.CreateTestRequest("POST", "/api/bookings/confirm", confirmReq)
	w2 := test.ExecuteRequest(router, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var confirmResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &confirmResponse)
	assert.NoError(t, err)
	assert.Equal(t, "booking confirmed successfully", confirmResponse["message"])

	// Step 3: Get user bookings
	bookingService.On("GetUserBookings",
		mock.Anything,
		uint(1),
		10,
		0,
	).Return([]entities.Booking{*mockBooking}, int64(1), nil).Once()

	req3, _ := test.CreateTestRequest("GET", "/api/bookings?page=1&limit=10", nil)
	w3 := test.ExecuteRequest(router, req3)

	assert.Equal(t, http.StatusOK, w3.Code)

	var listResponse map[string]interface{}
	err = json.Unmarshal(w3.Body.Bytes(), &listResponse)
	assert.NoError(t, err)

	data := listResponse["data"].([]interface{})
	assert.Equal(t, 1, len(data))
	assert.Equal(t, float64(1), listResponse["total"])

	// Verify all expectations were met
	bookingService.AssertExpectations(t)
}
