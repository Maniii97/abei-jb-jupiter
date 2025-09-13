package services

import (
	"api/internal/entities"
	"api/internal/repository"
	"context"
)

type UserService struct {
	userRepo *repository.UserRepository
}

// Ensure UserService implements UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Register(ctx context.Context, email, password, firstName, lastName, phone string, isAdmin bool) (*entities.User, error) {
	return s.userRepo.Register(ctx, email, password, firstName, lastName, phone, isAdmin)
}

func (s *UserService) Login(ctx context.Context, email, password string) (*entities.User, error) {
	return s.userRepo.Login(ctx, email, password)
}

func (s *UserService) GetByID(ctx context.Context, userID uint) (*entities.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
