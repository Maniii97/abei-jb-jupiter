package services

import (
	"api/internal/entities"
	"api/internal/repository"
)

type AnalyticsServiceInterface interface {
	GetBookingAnalytics() (*entities.BookingAnalytics, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsServiceInterface {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// GetBookingAnalytics returns comprehensive booking analytics for admin dashboard
func (s *analyticsService) GetBookingAnalytics() (*entities.BookingAnalytics, error) {
	// Get total booking counts
	confirmedCount, cancelledCount, err := s.analyticsRepo.GetTotalBookingCounts()
	if err != nil {
		return nil, err
	}

	// Calculate total bookings and cancellation rate
	totalBookings := confirmedCount + cancelledCount
	var cancellationRate float64
	if totalBookings > 0 {
		cancellationRate = float64(cancelledCount) / float64(totalBookings) * 100
	}

	// Get total revenue
	totalRevenue, err := s.analyticsRepo.GetTotalRevenue()
	if err != nil {
		return nil, err
	}

	// Get most popular events (by total bookings)
	popularEventsData, err := s.analyticsRepo.GetMostPopularEvents(10)
	if err != nil {
		return nil, err
	}

	// Get most booked events (by confirmed bookings)
	bookedEventsData, err := s.analyticsRepo.GetMostBookedEvents(10)
	if err != nil {
		return nil, err
	}

	// Get capacity utilization
	capacityData, err := s.analyticsRepo.GetCapacityUtilization()
	if err != nil {
		return nil, err
	}

	// Get daily booking stats for last 30 days
	dailyStatsData, err := s.analyticsRepo.GetDailyBookingStats(30)
	if err != nil {
		return nil, err
	}

	// Convert data to response format
	analytics := &entities.BookingAnalytics{
		TotalBookings:       totalBookings,
		ConfirmedBookings:   confirmedCount,
		CancelledBookings:   cancelledCount,
		CancellationRate:    cancellationRate,
		TotalRevenue:        totalRevenue,
		MostPopularEvents:   convertToPopularEvents(popularEventsData),
		MostBookedEvents:    convertToBookedEvents(bookedEventsData),
		CapacityUtilization: convertToCapacityUtilization(capacityData),
		DailyBookingStats:   convertToDailyBookingStats(dailyStatsData),
	}

	return analytics, nil
}

// Helper functions to convert database results to response format

func convertToPopularEvents(data []entities.EventBookingStats) []entities.PopularEvent {
	events := make([]entities.PopularEvent, len(data))
	for i, item := range data {
		events[i] = entities.PopularEvent{
			EventID:      item.EventID,
			EventName:    item.EventName,
			VenueName:    item.VenueName,
			BookingCount: item.BookingCount,
			Revenue:      item.Revenue,
		}
	}
	return events
}

func convertToBookedEvents(data []entities.EventBookingStats) []entities.BookedEvent {
	events := make([]entities.BookedEvent, len(data))
	for i, item := range data {
		var utilizationRate float64
		if item.TotalSeats > 0 {
			utilizationRate = float64(item.BookedSeats) / float64(item.TotalSeats) * 100
		}

		events[i] = entities.BookedEvent{
			EventID:         item.EventID,
			EventName:       item.EventName,
			VenueName:       item.VenueName,
			TotalSeats:      item.TotalSeats,
			BookedSeats:     item.BookedSeats,
			UtilizationRate: utilizationRate,
			Revenue:         item.Revenue,
		}
	}
	return events
}

func convertToCapacityUtilization(data []entities.EventBookingStats) []entities.CapacityUtilization {
	utilization := make([]entities.CapacityUtilization, len(data))
	for i, item := range data {
		var utilizationRate float64
		if item.TotalSeats > 0 {
			utilizationRate = float64(item.BookedSeats) / float64(item.TotalSeats) * 100
		}

		utilization[i] = entities.CapacityUtilization{
			EventID:         item.EventID,
			EventName:       item.EventName,
			VenueName:       item.VenueName,
			StartTime:       item.StartTime,
			TotalSeats:      item.TotalSeats,
			BookedSeats:     item.BookedSeats,
			UtilizationRate: utilizationRate,
			Status:          item.Status,
		}
	}
	return utilization
}

func convertToDailyBookingStats(data []entities.DailyStats) []entities.DailyBookingStat {
	stats := make([]entities.DailyBookingStat, len(data))
	for i, item := range data {
		var cancellationRate float64
		if item.TotalBookings > 0 {
			cancellationRate = float64(item.CancelledCount) / float64(item.TotalBookings) * 100
		}

		stats[i] = entities.DailyBookingStat{
			Date:             item.Date,
			TotalBookings:    item.TotalBookings,
			ConfirmedCount:   item.ConfirmedCount,
			CancelledCount:   item.CancelledCount,
			Revenue:          item.Revenue,
			CancellationRate: cancellationRate,
		}
	}
	return stats
}
