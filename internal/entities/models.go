package entities

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	IsAdmin   bool   `gorm:"default:false"`
	FirstName string `gorm:"size:100"`
	LastName  string `gorm:"size:100"`
	Phone     string `gorm:"size:20"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Bookings  []Booking `gorm:"foreignKey:UserID"`
}

type Venue struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;size:255"`
	Address     string `gorm:"not null;size:500"`
	City        string `gorm:"not null;size:100"`
	State       string `gorm:"not null;size:100"`
	Country     string `gorm:"not null;size:100"`
	Capacity    int    `gorm:"not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Events      []Event `gorm:"foreignKey:VenueID"`
}

type Event struct {
	ID             uint      `gorm:"primaryKey"`
	Name           string    `gorm:"not null;size:255;index"`
	Description    string    `gorm:"type:text"`
	VenueID        uint      `gorm:"index;not null"`
	Venue          Venue     `gorm:"foreignKey:VenueID;references:ID"`
	StartTime      time.Time `gorm:"not null;index"`
	EndTime        time.Time `gorm:"not null;index"`
	Price          float64   `gorm:"not null"`
	EventType      string    `gorm:"not null;size:50;index"`                  // concert, theater, sports, etc. - add index
	Status         string    `gorm:"not null;size:20;default:'active';index"` // active, cancelled, completed - add index
	IsHighDemand   bool      `gorm:"default:false;index"`                     // for queue system - add index
	AvailableSeats int       `gorm:"default:0;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Seats          []Seat          `gorm:"foreignKey:EventID"`
	Bookings       []Booking       `gorm:"foreignKey:EventID"`
	BookingIntents []BookingIntent `gorm:"foreignKey:EventID"`
}

type Seat struct {
	ID             uint       `gorm:"primaryKey"`
	EventID        uint       `gorm:"index;not null"`
	Event          Event      `gorm:"foreignKey:EventID"`
	SeatNumber     string     `gorm:"not null;size:20;index:idx_seat_lookup"`
	Row            string     `gorm:"not null;size:10;index:idx_seat_lookup"`
	Section        string     `gorm:"size:50;index:idx_seat_lookup"`
	SeatType       string     `gorm:"not null;size:50;index"` // VIP, Premium, Standard - add index
	Price          float64    `gorm:"not null"`
	IsAvailable    bool       `gorm:"default:true;index"`
	IsLocked       bool       `gorm:"default:false;index"`
	LockedAt       *time.Time `gorm:"index"`
	LockedBy       *uint      `gorm:"index"` // UserID who locked it - add index
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Bookings       []Booking       `gorm:"foreignKey:SeatID"`
	BookingIntents []BookingIntent `gorm:"foreignKey:SeatID"`
}

type BookingIntent struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"index;not null"`
	User            User      `gorm:"foreignKey:UserID"`
	EventID         uint      `gorm:"index;not null"`
	Event           Event     `gorm:"foreignKey:EventID"`
	SeatID          uint      `gorm:"index;not null"`
	Seat            Seat      `gorm:"foreignKey:SeatID"`
	IntentID        string    `gorm:"unique;not null;size:100;index"` // UUID for tracking - add index
	Status          string    `gorm:"not null;size:20;index"`         // pending, expired, confirmed, cancelled - add index
	LockExpiresAt   time.Time `gorm:"not null;index"`
	PaymentIntentID string    `gorm:"size:255;index"` // from payment gateway - add index
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Booking struct {
	ID              uint       `gorm:"primaryKey"`
	BookingNumber   string     `gorm:"unique;not null;size:50;index"`
	UserID          uint       `gorm:"index;not null"`
	User            User       `gorm:"foreignKey:UserID"`
	EventID         uint       `gorm:"index;not null"`
	Event           Event      `gorm:"foreignKey:EventID"`
	SeatID          uint       `gorm:"index;not null"`
	Seat            Seat       `gorm:"foreignKey:SeatID"`
	BookingIntentID *uint      `gorm:"index"`                  // reference to the intent that created this booking
	Status          string     `gorm:"not null;size:20;index"` // confirmed, cancelled, refunded - add index
	PaymentStatus   string     `gorm:"not null;size:20;index"` // paid, pending, failed, refunded - add index
	PaymentID       string     `gorm:"size:255;index"`         // from payment gateway - add index
	TotalAmount     float64    `gorm:"not null"`
	BookedAt        time.Time  `gorm:"not null;index"`
	CancelledAt     *time.Time `gorm:"index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type EventQueue struct {
	ID            uint       `gorm:"primaryKey"`
	EventID       uint       `gorm:"index;not null"`
	Event         Event      `gorm:"foreignKey:EventID"`
	UserID        uint       `gorm:"index;not null"`
	User          User       `gorm:"foreignKey:UserID"`
	QueuePosition int        `gorm:"not null;index"`         // Add index for position-based queries
	Status        string     `gorm:"not null;size:20;index"` // waiting, active, expired, completed - add index
	JoinedAt      time.Time  `gorm:"not null;index"`
	ActiveAt      *time.Time `gorm:"index"`
	ExpiresAt     *time.Time `gorm:"index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
