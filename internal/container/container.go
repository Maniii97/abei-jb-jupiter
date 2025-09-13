package container

import (
	"api/internal/config"
	"api/internal/db"
	"api/internal/entities"
	"api/internal/middleware"
	redisconn "api/internal/redis"
	"api/internal/repository"
	"api/internal/services"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container holds all application dependencies
type Container struct {
	Config           *config.Config
	DB               *gorm.DB
	Redis            *redis.Client
	UserService      *services.UserService
	JWTService       *services.JWTService
	EventService     *services.EventService
	VenueService     *services.VenueService
	BookingService   *services.BookingService
	SeatLockService  *services.SeatLockService
	WaitlistService  *services.WaitlistService
	AnalyticsService services.AnalyticsServiceInterface
	JWTMiddleware    *middleware.JWTMiddleware
	RateLimiter      *middleware.RateLimiter
}

// NewContainer creates a new dependency container
func NewContainer() (*Container, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Connect to database
	database, err := db.Connect(cfg.DBUrl)
	if err != nil {
		return nil, err
	}

	// Connect to Redis
	redisWrapper := redisconn.NewRedisClient(cfg.RedisUrl)
	redisClient := redisWrapper.Client

	// Run migrations
	if err := database.AutoMigrate(
		&entities.User{},
		&entities.Venue{},
		&entities.Event{},
		&entities.Seat{},
		&entities.BookingIntent{},
		&entities.Booking{},
		&entities.EventQueue{},
	); err != nil {
		return nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	venueRepo := repository.NewVenueRepository(database)
	eventRepo := repository.NewEventRepository(database)
	analyticsRepo := repository.NewAnalyticsRepository(database)

	// Initialize services
	jwtService := services.NewJWTService(cfg.JwtSecret)
	userService := services.NewUserService(userRepo)
	venueService := services.NewVenueService(venueRepo)
	eventService := services.NewEventService(eventRepo)
	seatLockService := services.NewSeatLockService(redisClient)
	analyticsService := services.NewAnalyticsService(analyticsRepo)

	// BookingRepository needs SeatLockRepository as dependency
	seatLockRepo := repository.NewSeatLockRepository(redisClient)
	bookingRepo := repository.NewBookingRepository(database, seatLockRepo)
	
	// Initialize waitlist services
	waitlistRepo := repository.NewWaitlistRepository(redisClient)
	waitlistService := services.NewWaitlistService(waitlistRepo, eventRepo, database)
	
	// BookingService needs WaitlistService as dependency
	bookingService := services.NewBookingService(bookingRepo, seatLockService, waitlistService)

	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	return &Container{
		Config:           cfg,
		DB:               database,
		Redis:            redisClient,
		UserService:      userService,
		JWTService:       jwtService,
		EventService:     eventService,
		VenueService:     venueService,
		BookingService:   bookingService,
		SeatLockService:  seatLockService,
		WaitlistService:  waitlistService,
		AnalyticsService: analyticsService,
		JWTMiddleware:    jwtMiddleware,
		RateLimiter:      rateLimiter,
	}, nil
}

// Close cleans up all resources
func (c *Container) Close() error {
	// Close Redis connection
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			return err
		}
	}

	// Close database connection
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
