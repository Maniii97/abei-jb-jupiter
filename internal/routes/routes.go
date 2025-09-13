package routes

import (
	"api/internal/container"
	"api/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(deps *container.Container) *gin.Engine {
	userHandler := handlers.NewUserHandler(deps.UserService, deps.JWTService)
	eventHandler := handlers.NewEventHandler(deps.EventService, deps.VenueService)
	venueHandler := handlers.NewVenueHandler(deps.VenueService)
	bookingHandler := handlers.NewBookingHandler(deps.BookingService)

	r := gin.Default()

	// Public API routes
	api := r.Group("/api")
	{
		// Authentication
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)

		// Events
		api.GET("/events", eventHandler.GetEvents)
		api.GET("/events/:id", eventHandler.GetEventByID)
		api.GET("/events/:id/seats", eventHandler.GetAvailableSeats)

		// Venues
		api.GET("/venues", venueHandler.GetVenues)
		api.GET("/venues/:id", venueHandler.GetVenueByID)
	}

	// Protected API routes
	protected := api.Group("/")
	protected.Use(deps.JWTMiddleware.AuthRequired())
	{
		// User profile
		protected.GET("/profile", userHandler.GetProfile)

		// Booking management
		protected.POST("/booking-intents", bookingHandler.CreateBookingIntent)
		protected.POST("/bookings/confirm", bookingHandler.ConfirmBooking)
		protected.POST("/booking-intents/cancel", bookingHandler.CancelBookingIntent)
		protected.DELETE("/bookings/:id", bookingHandler.CancelBooking)
		protected.GET("/bookings", bookingHandler.GetUserBookings)
		protected.GET("/bookings/:id", bookingHandler.GetBookingByID)
	}

	// Admin only routes
	admin := protected.Group("/admin")
	admin.Use(deps.JWTMiddleware.AdminRequired())
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
	}

	return r
}
