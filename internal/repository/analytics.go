package repository

import (
	"api/internal/entities"
	"time"

	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	GetTotalBookingCounts() (confirmed int64, cancelled int64, err error)
	GetTotalRevenue() (float64, error)
	GetMostPopularEvents(limit int) ([]entities.EventBookingStats, error)
	GetMostBookedEvents(limit int) ([]entities.EventBookingStats, error)
	GetCapacityUtilization() ([]entities.EventBookingStats, error)
	GetDailyBookingStats(days int) ([]entities.DailyStats, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// GetTotalBookingCounts returns the count of confirmed and cancelled bookings
func (r *analyticsRepository) GetTotalBookingCounts() (confirmed int64, cancelled int64, err error) {
	err = r.db.Model(&entities.Booking{}).
		Select("COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed, COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled").
		Row().Scan(&confirmed, &cancelled)
	return
}

// GetTotalRevenue returns the total revenue from confirmed bookings
func (r *analyticsRepository) GetTotalRevenue() (float64, error) {
	var revenue float64
	err := r.db.Model(&entities.Booking{}).
		Where("status = ?", "confirmed").
		Select("COALESCE(SUM(total_amount), 0)").
		Row().Scan(&revenue)
	return revenue, err
}

// GetMostPopularEvents returns events with highest booking counts
func (r *analyticsRepository) GetMostPopularEvents(limit int) ([]entities.EventBookingStats, error) {
	var results []entities.EventBookingStats

	err := r.db.Table("bookings b").
		Select(`
			e.id as event_id,
			e.name as event_name,
			v.name as venue_name,
			COUNT(b.id) as booking_count,
			COALESCE(SUM(CASE WHEN b.status = 'confirmed' THEN b.total_amount ELSE 0 END), 0) as revenue,
			(v.rows * v.columns) as total_seats,
			COUNT(CASE WHEN b.status = 'confirmed' THEN b.id END) as booked_seats,
			e.start_time,
			e.status
		`).
		Joins("JOIN events e ON b.event_id = e.id").
		Joins("JOIN venues v ON e.venue_id = v.id").
		Group("e.id, e.name, v.name, v.rows, v.columns, e.start_time, e.status").
		Order("booking_count DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

// GetMostBookedEvents returns events with highest confirmed bookings
func (r *analyticsRepository) GetMostBookedEvents(limit int) ([]entities.EventBookingStats, error) {
	var results []entities.EventBookingStats

	err := r.db.Table("bookings b").
		Select(`
			e.id as event_id,
			e.name as event_name,
			v.name as venue_name,
			COUNT(CASE WHEN b.status = 'confirmed' THEN b.id END) as booking_count,
			COALESCE(SUM(CASE WHEN b.status = 'confirmed' THEN b.total_amount ELSE 0 END), 0) as revenue,
			(v.rows * v.columns) as total_seats,
			COUNT(CASE WHEN b.status = 'confirmed' THEN b.id END) as booked_seats,
			e.start_time,
			e.status
		`).
		Joins("JOIN events e ON b.event_id = e.id").
		Joins("JOIN venues v ON e.venue_id = v.id").
		Where("b.status = ?", "confirmed").
		Group("e.id, e.name, v.name, v.rows, v.columns, e.start_time, e.status").
		Order("booked_seats DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

// GetCapacityUtilization returns capacity utilization for all events
func (r *analyticsRepository) GetCapacityUtilization() ([]entities.EventBookingStats, error) {
	var results []entities.EventBookingStats

	err := r.db.Table("events e").
		Select(`
			e.id as event_id,
			e.name as event_name,
			v.name as venue_name,
			COALESCE(COUNT(b.id), 0) as booking_count,
			COALESCE(SUM(CASE WHEN b.status = 'confirmed' THEN b.total_amount ELSE 0 END), 0) as revenue,
			(v.rows * v.columns) as total_seats,
			COALESCE(COUNT(CASE WHEN b.status = 'confirmed' THEN b.id END), 0) as booked_seats,
			e.start_time,
			e.status
		`).
		Joins("JOIN venues v ON e.venue_id = v.id").
		Joins("LEFT JOIN bookings b ON e.id = b.event_id").
		Group("e.id, e.name, v.name, v.rows, v.columns, e.start_time, e.status").
		Order("e.start_time DESC").
		Scan(&results).Error

	return results, err
}

// GetDailyBookingStats returns daily booking statistics for the last N days
func (r *analyticsRepository) GetDailyBookingStats(days int) ([]entities.DailyStats, error) {
	var results []entities.DailyStats

	err := r.db.Table("bookings").
		Select(`
			DATE(booked_at) as date,
			COUNT(*) as total_bookings,
			COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed_count,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_count,
			COALESCE(SUM(CASE WHEN status = 'confirmed' THEN total_amount ELSE 0 END), 0) as revenue
		`).
		Where("booked_at >= ?", time.Now().AddDate(0, 0, -days)).
		Group("DATE(booked_at)").
		Order("date DESC").
		Scan(&results).Error

	return results, err
}
