package test

import (
	"api/internal/entities"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupTestGin sets up a Gin router for testing
func SetupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// CreateTestRequest creates an HTTP request for testing
func CreateTestRequest(method, url string, body interface{}) (*http.Request, error) {
	var bodyReader *bytes.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// ExecuteRequest executes a test request and returns the response recorder
func ExecuteRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// AssertJSONResponse asserts that the response contains expected JSON
func AssertJSONResponse(t assert.TestingT, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	assert.Equal(t, expectedStatus, w.Code)

	if expectedMessage == "" {
		return
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// For success responses
	if response["message"] != nil {
		assert.Equal(t, expectedMessage, response["message"])
	} else if response["error"] != nil {
		// For error responses, check if the error message contains expected text
		errorMsg := response["error"].(string)
		assert.Equal(t, expectedMessage, errorMsg)
	}
}

// MockEntities provides sample entities for testing
type MockEntities struct{}

func (m *MockEntities) GetMockUser() *entities.User {
	return &entities.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
		IsAdmin:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (m *MockEntities) GetMockAdminUser() *entities.User {
	return &entities.User{
		ID:        2,
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Phone:     "+1234567891",
		IsAdmin:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (m *MockEntities) GetMockVenue() *entities.Venue {
	return &entities.Venue{
		ID:          1,
		Name:        "Test Arena",
		Address:     "123 Main St",
		City:        "New York",
		State:       "NY",
		Country:     "USA",
		Rows:        10,
		Columns:     20,
		Description: "A test arena for events",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (m *MockEntities) GetMockEvent() *entities.Event {
	venue := m.GetMockVenue()
	return &entities.Event{
		ID:             1,
		Name:           "Test Concert",
		Description:    "A test concert event",
		VenueID:        venue.ID,
		Venue:          *venue,
		StartTime:      time.Now().Add(24 * time.Hour),
		EndTime:        time.Now().Add(26 * time.Hour),
		Price:          100.0,
		EventType:      "concert",
		Status:         "active",
		IsHighDemand:   false,
		AvailableSeats: 200,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (m *MockEntities) GetMockSeat() *entities.Seat {
	event := m.GetMockEvent()
	return &entities.Seat{
		ID:          1,
		EventID:     event.ID,
		Event:       *event,
		Row:         1,
		Column:      1,
		SeatType:    "Standard",
		Price:       100.0,
		IsAvailable: true,
		IsLocked:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (m *MockEntities) GetMockBookingIntent() *entities.BookingIntent {
	user := m.GetMockUser()
	event := m.GetMockEvent()
	seat := m.GetMockSeat()

	return &entities.BookingIntent{
		ID:              1,
		UserID:          user.ID,
		User:            *user,
		EventID:         event.ID,
		Event:           *event,
		SeatID:          seat.ID,
		Seat:            *seat,
		Status:          "pending",
		PaymentIntentID: "pi_test123",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func (m *MockEntities) GetMockBooking() *entities.Booking {
	user := m.GetMockUser()
	event := m.GetMockEvent()
	seat := m.GetMockSeat()
	intentID := uint(1)

	return &entities.Booking{
		ID:              1,
		UserID:          user.ID,
		User:            *user,
		EventID:         event.ID,
		Event:           *event,
		SeatID:          seat.ID,
		Seat:            *seat,
		BookingIntentID: &intentID,
		Status:          "confirmed",
		PaymentStatus:   "paid",
		PaymentID:       "pay_test123",
		TotalAmount:     100.0,
		BookedAt:        time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
