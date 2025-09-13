package repository

import (
	"api/internal/entities"
	"api/pkg/errors"
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (s *UserRepository) Register(ctx context.Context, email, password, firstName, lastName, phone string, isAdmin bool) (*entities.User, error) {
	// Check if user already exists
	var existingUser entities.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.NewConflictError("User already exists", errors.ErrUserAlreadyExists)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError("Failed to hash password", err)
	}

	// Create user
	user := &entities.User{
		Email:     strings.ToLower(email),
		Password:  string(hash),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		IsAdmin:   isAdmin,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.NewConflictError("User already exists", errors.ErrUserAlreadyExists)
		}
		return nil, errors.NewInternalError("Failed to create user", err)
	}

	// Clear password from response
	user.Password = ""
	return user, nil
}

func (s *UserRepository) Login(ctx context.Context, email, password string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("email = ?", strings.ToLower(email)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewUnauthorizedError("Invalid credentials", errors.ErrInvalidCredentials)
		}
		return nil, errors.NewInternalError("Database error", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.NewUnauthorizedError("Invalid credentials", errors.ErrInvalidCredentials)
	}

	// Clear password from response
	user.Password = ""
	return &user, nil
}

func (s *UserRepository) GetByID(ctx context.Context, userID uint) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User not found", errors.ErrUserNotFound)
		}
		return nil, errors.NewInternalError("Database error", err)
	}

	user.Password = ""
	return &user, nil
}
