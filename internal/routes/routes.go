package routes

import (
	"api/internal/container"
	"api/internal/handlers"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(deps *container.Container) *gin.Engine {
	userHandler := handlers.NewUserHandler(deps.UserService, deps.JWTService)
	eventHandler := handlers.NewEventHandler(deps.EventService, deps.VenueService)
	venueHandler := handlers.NewVenueHandler(deps.VenueService)
	bookingHandler := handlers.NewBookingHandler(deps.BookingService)
	analyticsHandler := handlers.NewAnalyticsHandler(deps.AnalyticsService)
	waitlistHandler := handlers.NewWaitlistHandler(deps.WaitlistService)

	r := gin.Default()

	// global rate limiting - 1000 requests per minute per IP
	r.Use(deps.RateLimiter.RateLimit(1000, time.Minute))

	// Public API routes
	api := r.Group("/api")
	{
		// Authentication
		auth := api.Group("/")
		auth.Use(deps.RateLimiter.RateLimit(10, time.Minute)) // 10 auth attempts per minute
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// Events
		events := api.Group("/events")
		events.Use(deps.RateLimiter.RateLimit(200, time.Minute)) // 200 requests per minute
		{
			events.GET("", eventHandler.GetEvents)
			events.GET("/:id", eventHandler.GetEventByID)
			events.GET("/:id/seats", eventHandler.GetAvailableSeats)
		}

		// Venues
		venues := api.Group("/venues")
		venues.Use(deps.RateLimiter.RateLimit(200, time.Minute)) // 200 requests per minute
		{
			venues.GET("", venueHandler.GetVenues)
			venues.GET("/:id", venueHandler.GetVenueByID)
		}
	}

	// Protected API routes
	protected := api.Group("/")
	protected.Use(deps.JWTMiddleware.AuthRequired())
	{
		// User profile
		profile := protected.Group("/")
		profile.Use(deps.RateLimiter.UserRateLimit(100, time.Minute)) // 100 requests per user per minute
		{
			profile.GET("/profile", userHandler.GetProfile)
		}

		// Booking management
		bookings := protected.Group("/")
		bookings.Use(deps.RateLimiter.UserRateLimit(50, time.Minute)) // 50 booking ops per user per minute
		{
			bookings.POST("/booking-intents", bookingHandler.CreateBookingIntent)
			bookings.POST("/bookings/confirm", bookingHandler.ConfirmBooking)
			bookings.POST("/booking-intents/cancel", bookingHandler.CancelBookingIntent)
			bookings.DELETE("/bookings/:id", bookingHandler.CancelBooking)
			bookings.GET("/bookings", bookingHandler.GetUserBookings)
			bookings.GET("/bookings/:id", bookingHandler.GetBookingByID)
		}

		// Waitlist management
		waitlist := protected.Group("/waitlist")
		waitlist.Use(deps.RateLimiter.UserRateLimit(30, time.Minute)) // 30 waitlist ops per user per minute
		{
			waitlist.POST("/events/:eventId/join", waitlistHandler.JoinWaitlist)
			waitlist.GET("/events/:eventId/position", waitlistHandler.GetWaitlistPosition)
			waitlist.DELETE("/events/:eventId/leave", waitlistHandler.LeaveWaitlist)
			waitlist.GET("/events/:eventId/stats", waitlistHandler.GetWaitlistStats)
		}
	}

	// Admin only routes
	admin := protected.Group("/admin")
	admin.Use(deps.JWTMiddleware.AdminRequired())
	admin.Use(deps.RateLimiter.UserRateLimit(200, time.Minute)) // 200 admin ops per minute
	{
		// User management
		admin.GET("/users", userHandler.ListUsers)

		// Venue management
		admin.POST("/venues", venueHandler.CreateVenue)
		admin.PUT("/venues/:id", venueHandler.UpdateVenue)
		admin.DELETE("/venues/:id", venueHandler.DeleteVenue)

		// Event management
		admin.POST("/events", eventHandler.CreateEvent)
		admin.PUT("/events/:id", eventHandler.UpdateEvent)
		admin.DELETE("/events/:id", eventHandler.DeleteEvent)
		admin.GET("/events/:id/stats", eventHandler.GetEventStats)

		// Analytics
		admin.GET("/analytics/bookings", analyticsHandler.GetBookingAnalytics)
	}

	return r
}
