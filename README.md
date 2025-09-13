# Evently API

A comprehensive event booking system API built with Go, Gin, GORM, Redis, and PostgreSQL. The API provides functionality for event management, seat booking with temporary locking, waitlist management, and real-time analytics.

## 🚀 Features

- **User Management**: Registration, authentication, and profile management
- **Event & Venue Management**: CRUD operations for events and venues
- **Seat Booking System**: Temporary seat locking with booking intents
- **Waitlist Management**: Queue system for sold-out events
- **Real-time Analytics**: Booking statistics and event insights
- **Admin Panel**: Administrative controls for managing the platform
- **Rate Limiting**: IP and user-based rate limiting
- **JWT Authentication**: Secure token-based authentication
- **Database Migration**: Automated database schema management

## 🏗️ Architecture

### Tech Stack

- **Backend**: Go 1.23.2
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Authentication**: JWT tokens
- **Configuration**: Viper
- **Logging**: Custom logging package

### Project Structure

```
evently_api/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── constants/
│   └── status.go                   # Application constants
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── container/
│   │   └── container.go           # Dependency injection container
│   ├── db/
│   │   └── db.go                  # Database connection and migration
│   ├── entities/
│   │   ├── analytics.go           # Analytics entities
│   │   └── models.go              # Database models
│   ├── handlers/
│   │   ├── analytics.go           # Analytics HTTP handlers
│   │   ├── booking.go             # Booking HTTP handlers
│   │   ├── event.go               # Event HTTP handlers
│   │   ├── user.go                # User HTTP handlers
│   │   ├── venue.go               # Venue HTTP handlers
│   │   └── waitlist.go            # Waitlist HTTP handlers
│   ├── middleware/
│   │   ├── jwt.go                 # JWT authentication middleware
│   │   └── rate_limiter.go        # Rate limiting middleware
│   ├── repository/
│   │   ├── analytics.go           # Analytics data access layer
│   │   ├── booking.go             # Booking data access layer
│   │   ├── event.go               # Event data access layer
│   │   ├── jwt.go                 # JWT data access layer
│   │   ├── lock_seat.go           # Seat locking data access layer
│   │   ├── user.go                # User data access layer
│   │   ├── venue.go               # Venue data access layer
│   │   └── waitlist.go            # Waitlist data access layer
│   ├── routes/
│   │   └── routes.go              # API route definitions
│   └── services/
│       ├── analytics.go           # Analytics business logic
│       ├── booking.go             # Booking business logic
│       ├── event.go               # Event business logic
│       ├── interfaces.go          # Service interfaces
│       ├── jwt.go                 # JWT service
│       ├── seat_lock.go           # Seat locking service
│       ├── user.go                # User business logic
│       ├── venue.go               # Venue business logic
│       └── waitlist.go            # Waitlist business logic
├── pkg/
│   ├── errors/
│   │   └── errors.go              # Custom error types
│   ├── logging/
│   │   └── logger.go              # Logging utilities
│   ├── request/
│   │   └── request.go             # Request DTOs
│   └── response/
│       └── response.go            # Response DTOs
├── test/
│   ├── mocks/
│   │   └── booking_service_mock.go # Test mocks
│   └── test_utils.go              # Test utilities
├── Dockerfile                      # Docker configuration
├── go.mod                         # Go module dependencies
├── go.sum                         # Go module checksums
└── README.md                      # This file
```

## 🛠️ Installation

### Prerequisites

- Go 1.23.2 or higher
- PostgreSQL 12+
- Redis 6+
- Docker (optional)

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Maniii97/abei-jb-jupiter.git
   cd abei-jb-jupiter
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   Create a `.env` file in the root directory:
   ```env
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=evently_db
   DB_SSLMODE=disable

   # Redis
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0

   # JWT
   JWT_SECRET=your-super-secret-jwt-key
   JWT_EXPIRY=24h

   # Server
   SERVER_HOST=localhost
   SERVER_PORT=8080

   # Logging
   LOG_LEVEL=debug
   ```

4. **Set up PostgreSQL database**
   ```sql
   CREATE DATABASE evently_db;
   CREATE USER evently_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE evently_db TO evently_user;
   ```

5. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

   The API will be available at `http://localhost:8080`

### Docker Setup

1. **Build and run with Docker Compose**
   ```bash
   docker-compose up -d
   ```

   This will start:
   - PostgreSQL database
   - Redis cache
   - The API server

## 📚 API Documentation

### OpenAPI Specification

The complete API documentation is available in the `openapi.yaml` file. You can view it using:

- [Swagger UI](https://app.swaggerhub.com/apis-docs/mani-bcb/evently-api/1.0.0)


### Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Rate Limiting

The API implements rate limiting:
- **Global**: 1000 requests per minute per IP
- **Authentication**: 10 requests per minute per IP
- **Public endpoints**: 200 requests per minute per IP
- **Protected endpoints**: 100 requests per minute per user
- **Booking operations**: 50 requests per minute per user
- **Waitlist operations**: 30 requests per minute per user
- **Admin operations**: 200 requests per minute per user

## 🔧 API Endpoints

### Authentication
- `POST /register` - Register a new user
- `POST /login` - User login

### User Profile
- `GET /profile` - Get user profile (authenticated)

### Events
- `GET /events` - List events with pagination and filtering
- `GET /events/{id}` - Get event details
- `GET /events/{id}/seats` - Get available seats for an event

### Venues
- `GET /venues` - List venues with pagination and filtering
- `GET /venues/{id}` - Get venue details

### Bookings
- `POST /booking-intents` - Create a booking intent (lock seat temporarily)
- `POST /bookings/confirm` - Confirm a booking
- `POST /booking-intents/cancel` - Cancel a booking intent
- `GET /bookings` - Get user's bookings
- `GET /bookings/{id}` - Get booking details
- `DELETE /bookings/{id}` - Cancel a booking

### Waitlist
- `POST /waitlist/events/{eventId}/join` - Join event waitlist
- `GET /waitlist/events/{eventId}/position` - Get waitlist position
- `DELETE /waitlist/events/{eventId}/leave` - Leave waitlist
- `GET /waitlist/events/{eventId}/stats` - Get waitlist statistics

### Admin Endpoints
- `GET /admin/users` - List all users
- `POST /admin/venues` - Create venue
- `PUT /admin/venues/{id}` - Update venue
- `DELETE /admin/venues/{id}` - Delete venue
- `POST /admin/events` - Create event
- `PUT /admin/events/{id}` - Update event
- `DELETE /admin/events/{id}` - Delete event
- `GET /admin/events/{id}/stats` - Get event statistics
- `GET /admin/analytics/bookings` - Get booking analytics

## 🎫 Booking Flow

The API implements a robust booking system with temporary seat locking:

1. **Browse Events**: Users can browse available events and see seat availability
2. **Create Booking Intent**: Lock a seat temporarily (15 minutes default)
3. **Payment Processing**: Process payment through external payment gateway
4. **Confirm Booking**: Convert the intent to a confirmed booking
5. **Automatic Cleanup**: Expired intents are automatically cleaned up

### Seat Locking Mechanism

- Seats are temporarily locked when a booking intent is created
- Lock duration is configurable (default: 15 minutes)
- Locked seats are not available to other users
- Automatic cleanup releases expired locks

## 📊 Waitlist System

For high-demand events, users can join a waitlist:

1. **Join Waitlist**: Add user to event waitlist when sold out
2. **Position Tracking**: Users can check their position in the queue
3. **Automatic Notifications**: Users are notified when seats become available
4. **Time-based Expiry**: Notifications expire if not acted upon

## 🔍 Analytics

The API provides comprehensive analytics for administrators:

- **Event Statistics**: Capacity utilization, booking rates, revenue
- **Booking Analytics**: Trends, popular events, user behavior
- **Real-time Metrics**: Current seat availability, waitlist status

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/handlers/tests/
```

### Test Structure

- Unit tests for individual components
- Integration tests for API endpoints
- Mock services for isolated testing

## 🚀 Deployment

### Environment Configuration

Set the following environment variables for production:

```env
# Database
DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-secure-password
DB_NAME=evently_production
DB_SSLMODE=require

# Redis
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# JWT
JWT_SECRET=your-super-secure-jwt-secret-min-32-chars
JWT_EXPIRY=24h

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Logging
LOG_LEVEL=info
```

### Docker Deployment

1. **Build the image**
   ```bash
   docker build -t evently-api .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 \
     -e DB_HOST=your-db-host \
     -e DB_PASSWORD=your-password \
     -e REDIS_HOST=your-redis-host \
     -e JWT_SECRET=your-jwt-secret \
     evently-api
   ```

### Production Considerations

- Use a reverse proxy (nginx) for SSL termination
- Implement proper database backup strategies
- Set up monitoring and alerting
- Configure log aggregation
- Use container orchestration (Kubernetes) for scaling

## 📝 Configuration

The application uses Viper for configuration management. Configuration can be provided via:

- Environment variables
- Configuration files (JSON, YAML, TOML)
- Command line flags

### Key Configuration Options

- **Database**: Connection parameters and pool settings
- **Redis**: Cache configuration and connection pooling
- **JWT**: Secret key and token expiry
- **Rate Limiting**: Request limits and time windows
- **Logging**: Log level and output format


## 📊 API Usage Examples

### Register a new user

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get events

```bash
curl -X GET "http://localhost:8080/api/events?page=1&limit=10&city=New York"
```

### Create booking intent

```bash
curl -X POST http://localhost:8080/api/booking-intents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "seat_id": 123
  }'
```

### Join waitlist

```bash
curl -X POST http://localhost:8080/api/waitlist/events/1/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

**Built with ❤️ by Mani (me, lol)**
