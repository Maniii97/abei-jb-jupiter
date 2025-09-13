package repository

import (
	"api/internal/entities"
	"api/pkg/errors"
	"context"

	"gorm.io/gorm"
)

type VenueRepository struct {
	db *gorm.DB
}

func NewVenueRepository(db *gorm.DB) *VenueRepository {
	return &VenueRepository{db: db}
}

// GetVenues returns a paginated list of venues
func (s *VenueRepository) GetVenues(ctx context.Context, limit, offset int, city string) ([]entities.Venue, int64, error) {
	var venues []entities.Venue
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.Venue{})

	if city != "" {
		query = query.Where("city ILIKE ?", "%"+city+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalError("Failed to count venues", err)
	}

	// Get paginated results
	if err := query.Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&venues).Error; err != nil {
		return nil, 0, errors.NewInternalError("Failed to fetch venues", err)
	}

	return venues, total, nil
}

// GetVenueByID returns a single venue with details
func (s *VenueRepository) GetVenueByID(ctx context.Context, venueID uint) (*entities.Venue, error) {
	var venue entities.Venue

	if err := s.db.WithContext(ctx).
		Preload("Events", "status = ?", "active").
		First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Venue not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch venue", err)
	}

	return &venue, nil
}

// CreateVenue creates a new venue (admin only)
func (s *VenueRepository) CreateVenue(ctx context.Context, venue *entities.Venue) error {
	if err := s.db.WithContext(ctx).Create(venue).Error; err != nil {
		return errors.NewInternalError("Failed to create venue", err)
	}

	return nil
}

// UpdateVenue updates an existing venue (admin only)
func (s *VenueRepository) UpdateVenue(ctx context.Context, venueID uint, updates map[string]interface{}) (*entities.Venue, error) {
	var venue entities.Venue

	if err := s.db.WithContext(ctx).First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Venue not found", errors.ErrRecordNotFound)
		}
		return nil, errors.NewInternalError("Failed to fetch venue", err)
	}

	if err := s.db.WithContext(ctx).Model(&venue).Updates(updates).Error; err != nil {
		return nil, errors.NewInternalError("Failed to update venue", err)
	}

	return &venue, nil
}

// DeleteVenue soft deletes a venue (admin only)
func (s *VenueRepository) DeleteVenue(ctx context.Context, venueID uint) error {
	var venue entities.Venue

	if err := s.db.WithContext(ctx).First(&venue, venueID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Venue not found", errors.ErrRecordNotFound)
		}
		return errors.NewInternalError("Failed to fetch venue", err)
	}

	// Check if venue has active events
	var activeEventsCount int64
	if err := s.db.WithContext(ctx).Model(&entities.Event{}).
		Where("venue_id = ? AND status = ?", venueID, "active").
		Count(&activeEventsCount).Error; err != nil {
		return errors.NewInternalError("Failed to check active events", err)
	}

	if activeEventsCount > 0 {
		return errors.NewBadRequestError("Cannot delete venue with active events", nil)
	}

	if err := s.db.WithContext(ctx).Delete(&venue).Error; err != nil {
		return errors.NewInternalError("Failed to delete venue", err)
	}

	return nil
}
