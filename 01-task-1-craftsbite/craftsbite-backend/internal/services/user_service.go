package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/internal/utils"
	"fmt"

	"github.com/google/uuid"
)

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email                 string      `json:"email" validate:"required,email"`
	Name                  string      `json:"name" validate:"required"`
	Password              string      `json:"password" validate:"required,min=8"`
	Role                  models.Role `json:"role" validate:"required"`
	DefaultMealPreference string      `json:"default_meal_preference"`
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Name                  *string      `json:"name"`
	Role                  *models.Role `json:"role"`
	DefaultMealPreference *string      `json:"default_meal_preference"`
	Password              *string      `json:"password" validate:"omitempty,min=8"`
}

// UserService defines the interface for user management operations
type UserService interface {
	CreateUser(input CreateUserInput) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(id string, input UpdateUserInput) (*models.User, error)
	DeactivateUser(id string) error
	ListUsers(filters map[string]interface{}) ([]models.User, error)
}

// userService implements UserService
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// CreateUser creates a new user
func (s *userService) CreateUser(input CreateUserInput) (*models.User, error) {
	// Check if email already exists
	existingUser, _ := s.userRepo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default meal preference if not provided
	if input.DefaultMealPreference == "" {
		input.DefaultMealPreference = "opt_in"
	}

	// Create user
	user := &models.User{
		ID:                    uuid.New(),
		Email:                 input.Email,
		Name:                  input.Name,
		Password:              hashedPassword,
		Role:                  input.Role,
		Active:                true,
		DefaultMealPreference: input.DefaultMealPreference,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(id string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(id string, input UpdateUserInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.DefaultMealPreference != nil {
		user.DefaultMealPreference = *input.DefaultMealPreference
	}
	if input.Password != nil {
		hashedPassword, err := utils.HashPassword(*input.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeactivateUser deactivates a user
func (s *userService) DeactivateUser(id string) error {
	return s.userRepo.Delete(id)
}

// ListUsers lists all users with optional filters
func (s *userService) ListUsers(filters map[string]interface{}) ([]models.User, error) {
	return s.userRepo.FindAll(filters)
}
