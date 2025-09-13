package services

import (
	"api/internal/entities"
	"api/internal/repository"
	"context"
)

type VenueService struct {
	venueRepo *repository.VenueRepository
}

// Ensure VenueService implements VenueServiceInterface
var _ VenueServiceInterface = (*VenueService)(nil)

func NewVenueService(venueRepo *repository.VenueRepository) *VenueService {
	return &VenueService{venueRepo: venueRepo}
}

func (s *VenueService) GetVenues(ctx context.Context, limit, offset int, city string) ([]entities.Venue, int64, error) {
	return s.venueRepo.GetVenues(ctx, limit, offset, city)
}

func (s *VenueService) GetVenueByID(ctx context.Context, venueID uint) (*entities.Venue, error) {
	return s.venueRepo.GetVenueByID(ctx, venueID)
}

func (s *VenueService) CreateVenue(ctx context.Context, venue *entities.Venue) error {
	return s.venueRepo.CreateVenue(ctx, venue)
}

func (s *VenueService) UpdateVenue(ctx context.Context, venueID uint, updates map[string]interface{}) (*entities.Venue, error) {
	return s.venueRepo.UpdateVenue(ctx, venueID, updates)
}

func (s *VenueService) DeleteVenue(ctx context.Context, venueID uint) error {
	return s.venueRepo.DeleteVenue(ctx, venueID)
}
