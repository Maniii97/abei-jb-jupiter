package entities

import "time"

type BookingAnalytics struct {
	TotalBookings       int64                 `json:"total_bookings"`
	ConfirmedBookings   int64                 `json:"confirmed_bookings"`
	CancelledBookings   int64                 `json:"cancelled_bookings"`
	CancellationRate    float64               `json:"cancellation_rate"`
	TotalRevenue        float64               `json:"total_revenue"`
	MostPopularEvents   []PopularEvent        `json:"most_popular_events"`
	MostBookedEvents    []BookedEvent         `json:"most_booked_events"`
	CapacityUtilization []CapacityUtilization `json:"capacity_utilization"`
	DailyBookingStats   []DailyBookingStat    `json:"daily_booking_stats"`
}

type PopularEvent struct {
	EventID      uint    `json:"event_id"`
	EventName    string  `json:"event_name"`
	VenueName    string  `json:"venue_name"`
	BookingCount int64   `json:"booking_count"`
	Revenue      float64 `json:"revenue"`
}

type BookedEvent struct {
	EventID         uint    `json:"event_id"`
	EventName       string  `json:"event_name"`
	VenueName       string  `json:"venue_name"`
	TotalSeats      int64   `json:"total_seats"`
	BookedSeats     int64   `json:"booked_seats"`
	UtilizationRate float64 `json:"utilization_rate"`
	Revenue         float64 `json:"revenue"`
}

type CapacityUtilization struct {
	EventID         uint      `json:"event_id"`
	EventName       string    `json:"event_name"`
	VenueName       string    `json:"venue_name"`
	StartTime       time.Time `json:"start_time"`
	TotalSeats      int64     `json:"total_seats"`
	BookedSeats     int64     `json:"booked_seats"`
	UtilizationRate float64   `json:"utilization_rate"`
	Status          string    `json:"status"`
}

type DailyBookingStat struct {
	Date             time.Time `json:"date"`
	TotalBookings    int64     `json:"total_bookings"`
	ConfirmedCount   int64     `json:"confirmed_count"`
	CancelledCount   int64     `json:"cancelled_count"`
	Revenue          float64   `json:"revenue"`
	CancellationRate float64   `json:"cancellation_rate"`
}

// Database query result structures
type EventBookingStats struct {
	EventID      uint      `json:"event_id"`
	EventName    string    `json:"event_name"`
	VenueName    string    `json:"venue_name"`
	BookingCount int64     `json:"booking_count"`
	Revenue      float64   `json:"revenue"`
	TotalSeats   int64     `json:"total_seats"`
	BookedSeats  int64     `json:"booked_seats"`
	StartTime    time.Time `json:"start_time"`
	Status       string    `json:"status"`
}

type DailyStats struct {
	Date           time.Time `json:"date"`
	TotalBookings  int64     `json:"total_bookings"`
	ConfirmedCount int64     `json:"confirmed_count"`
	CancelledCount int64     `json:"cancelled_count"`
	Revenue        float64   `json:"revenue"`
}
