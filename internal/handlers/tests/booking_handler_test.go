package tests

import (
	"api/internal/entities"
	"api/internal/handlers"
	"api/pkg/errors"
	"api/pkg/request"
	"api/test"
	"api/test/mocks"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BookingHandlerTestSuite struct {
	suite.Suite
	router         *gin.Engine
	bookingService *mocks.MockBookingService
	handler        *handlers.BookingHandler
	mockEntities   *test.MockEntities
}

func (suite *BookingHandlerTestSuite) SetupTest() {
	suite.router = test.SetupTestGin()
	suite.bookingService = &mocks.MockBookingService{}
	suite.handler = handlers.NewBookingHandler(suite.bookingService)
	suite.mockEntities = &test.MockEntities{}

	// Setup routes
	api := suite.router.Group("/api")
	protected := api.Group("/")
	protected.Use(suite.mockAuthMiddleware())
	{
		protected.POST("/booking-intents", suite.handler.CreateBookingIntent)
		protected.POST("/bookings/confirm", suite.handler.ConfirmBooking)
		protected.POST("/booking-intents/cancel", suite.handler.CancelBookingIntent)
		protected.DELETE("/bookings/:id", suite.handler.CancelBooking)
		protected.GET("/bookings", suite.handler.GetUserBookings)
		protected.GET("/bookings/:id", suite.handler.GetBookingByID)
	}
}

func (suite *BookingHandlerTestSuite) TearDownTest() {
	suite.bookingService.AssertExpectations(suite.T())
}

// Mock middleware to simulate authentication
func (suite *BookingHandlerTestSuite) mockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	}
}

// Test CreateBookingIntent - Success case
func (suite *BookingHandlerTestSuite) TestCreateBookingIntent_Success() {
	mockIntent := suite.mockEntities.GetMockBookingIntent()

	suite.bookingService.On("CreateBookingIntent",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(mockIntent, nil)

	reqBody := request.CreateBookingIntentRequest{
		SeatID: 1,
	}

	req, _ := test.CreateTestRequest("POST", "/api/booking-intents", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "booking intent created successfully", response["message"])
	assert.NotNil(suite.T(), response["data"])
}

// Test CreateBookingIntent - Seat not available
func (suite *BookingHandlerTestSuite) TestCreateBookingIntent_SeatNotAvailable() {
	suite.bookingService.On("CreateBookingIntent",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(nil, errors.NewConflictError("Seat is not available", nil))

	reqBody := request.CreateBookingIntentRequest{
		SeatID: 1,
	}

	req, _ := test.CreateTestRequest("POST", "/api/booking-intents", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Seat is not available", response["error"])
}

// Test CreateBookingIntent - Seat not found
func (suite *BookingHandlerTestSuite) TestCreateBookingIntent_SeatNotFound() {
	suite.bookingService.On("CreateBookingIntent",
		mock.Anything,
		uint(1),
		uint(999),
	).Return(nil, errors.NewNotFoundError("Seat not found", nil))

	reqBody := request.CreateBookingIntentRequest{
		SeatID: 999,
	}

	req, _ := test.CreateTestRequest("POST", "/api/booking-intents", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Seat not found", response["error"])
}

// Test ConfirmBooking - Success case
func (suite *BookingHandlerTestSuite) TestConfirmBooking_Success() {
	mockBooking := suite.mockEntities.GetMockBooking()

	suite.bookingService.On("ConfirmBooking",
		mock.Anything,
		uint(1),
		"pay_test123",
	).Return(mockBooking, nil)

	reqBody := request.ConfirmBookingRequest{
		BookingIntentID: 1,
		PaymentID:       "pay_test123",
	}

	req, _ := test.CreateTestRequest("POST", "/api/bookings/confirm", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "booking confirmed successfully", response["message"])
	assert.NotNil(suite.T(), response["data"])
}

// Test ConfirmBooking - Intent not found
func (suite *BookingHandlerTestSuite) TestConfirmBooking_IntentNotFound() {
	suite.bookingService.On("ConfirmBooking",
		mock.Anything,
		uint(999),
		"pay_test123",
	).Return(nil, errors.NewNotFoundError("Booking intent not found", nil))

	reqBody := request.ConfirmBookingRequest{
		BookingIntentID: 999,
		PaymentID:       "pay_test123",
	}

	req, _ := test.CreateTestRequest("POST", "/api/bookings/confirm", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Booking intent not found", response["error"])
}

// Test ConfirmBooking - Intent expired
func (suite *BookingHandlerTestSuite) TestConfirmBooking_IntentExpired() {
	suite.bookingService.On("ConfirmBooking",
		mock.Anything,
		uint(1),
		"pay_test123",
	).Return(nil, errors.NewBadRequestError("Booking intent has expired", nil))

	reqBody := request.ConfirmBookingRequest{
		BookingIntentID: 1,
		PaymentID:       "pay_test123",
	}

	req, _ := test.CreateTestRequest("POST", "/api/bookings/confirm", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Booking intent has expired", response["error"])
}

// Test CancelBookingIntent - Success
func (suite *BookingHandlerTestSuite) TestCancelBookingIntent_Success() {
	suite.bookingService.On("CancelBookingIntent",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(nil)

	reqBody := request.CancelBookingIntentRequest{
		BookingIntentID: 1,
	}

	req, _ := test.CreateTestRequest("POST", "/api/booking-intents/cancel", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "booking intent cancelled successfully", response["message"])
}

// Test CancelBookingIntent - Not found
func (suite *BookingHandlerTestSuite) TestCancelBookingIntent_NotFound() {
	suite.bookingService.On("CancelBookingIntent",
		mock.Anything,
		uint(999),
		uint(1),
	).Return(errors.NewNotFoundError("Booking intent not found", nil))

	reqBody := request.CancelBookingIntentRequest{
		BookingIntentID: 999,
	}

	req, _ := test.CreateTestRequest("POST", "/api/booking-intents/cancel", reqBody)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Booking intent not found", response["error"])
}

// Test CancelBooking - Success
func (suite *BookingHandlerTestSuite) TestCancelBooking_Success() {
	suite.bookingService.On("CancelBooking",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(nil)

	req, _ := test.CreateTestRequest("DELETE", "/api/bookings/1", nil)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "booking cancelled successfully", response["message"])
}

// Test CancelBooking - Not found
func (suite *BookingHandlerTestSuite) TestCancelBooking_NotFound() {
	suite.bookingService.On("CancelBooking",
		mock.Anything,
		uint(999),
		uint(1),
	).Return(errors.NewNotFoundError("Booking not found", nil))

	req, _ := test.CreateTestRequest("DELETE", "/api/bookings/999", nil)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Booking not found", response["error"])
}

// Test GetUserBookings - Success
func (suite *BookingHandlerTestSuite) TestGetUserBookings_Success() {
	mockBookings := []entities.Booking{*suite.mockEntities.GetMockBooking()}

	suite.bookingService.On("GetUserBookings",
		mock.Anything,
		uint(1),
		10,
		0,
	).Return(mockBookings, int64(1), nil)

	req, _ := test.CreateTestRequest("GET", "/api/bookings?page=1&limit=10", nil)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["data"])
	assert.Equal(suite.T(), float64(1), response["total"])
}

// Test GetUserBookings - Empty result
func (suite *BookingHandlerTestSuite) TestGetUserBookings_EmptyResult() {
	suite.bookingService.On("GetUserBookings",
		mock.Anything,
		uint(1),
		10,
		0,
	).Return([]entities.Booking{}, int64(0), nil)

	req, _ := test.CreateTestRequest("GET", "/api/bookings?page=1&limit=10", nil)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].([]interface{})
	assert.Equal(suite.T(), 0, len(data))
}

// Test GetBookingByID - Success
func (suite *BookingHandlerTestSuite) TestGetBookingByID_Success() {
	mockBooking := suite.mockEntities.GetMockBooking()

	suite.bookingService.On("GetBookingByID",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(mockBooking, nil)

	req, _ := test.CreateTestRequest("GET", "/api/bookings/1", nil)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(mockBooking.ID), response["id"])
}

// Test GetBookingByID - Not found
func (suite *BookingHandlerTestSuite) TestGetBookingByID_NotFound() {
	suite.bookingService.On("GetBookingByID",
		mock.Anything,
		uint(999),
		uint(1),
	).Return(nil, errors.NewNotFoundError("Booking not found", nil))

	req, _ := test.CreateTestRequest("GET", "/api/bookings/999", nil)
	w := test.ExecuteRequest(suite.router, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Booking not found", response["error"])
}

// Test authentication scenarios
func (suite *BookingHandlerTestSuite) TestCreateBookingIntent_NoAuth() {
	// Create router without auth middleware
	router := test.SetupTestGin()
	api := router.Group("/api")
	api.POST("/booking-intents", suite.handler.CreateBookingIntent)

	reqBody := request.CreateBookingIntentRequest{
		SeatID: 1,
	}

	req, _ := test.CreateTestRequest("POST", "/api/booking-intents", reqBody)
	w := test.ExecuteRequest(router, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "user not authenticated", response["error"])
}

// Test concurrent booking scenario
func (suite *BookingHandlerTestSuite) TestConcurrentBookingIntents() {
	mockIntent := suite.mockEntities.GetMockBookingIntent()

	// First request succeeds
	suite.bookingService.On("CreateBookingIntent",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(mockIntent, nil).Once()

	// Second request fails due to seat being locked
	suite.bookingService.On("CreateBookingIntent",
		mock.Anything,
		uint(1),
		uint(1),
	).Return(nil, errors.NewConflictError("Seat is already locked by another user", nil)).Once()

	reqBody := request.CreateBookingIntentRequest{
		SeatID: 1,
	}

	// First request - should succeed
	req1, _ := test.CreateTestRequest("POST", "/api/booking-intents", reqBody)
	w1 := test.ExecuteRequest(suite.router, req1)
	assert.Equal(suite.T(), http.StatusCreated, w1.Code)

	// Second request - should fail with conflict
	req2, _ := test.CreateTestRequest("POST", "/api/booking-intents", reqBody)
	w2 := test.ExecuteRequest(suite.router, req2)
	assert.Equal(suite.T(), http.StatusConflict, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Seat is already locked by another user", response["error"])
}

func TestBookingHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BookingHandlerTestSuite))
}
