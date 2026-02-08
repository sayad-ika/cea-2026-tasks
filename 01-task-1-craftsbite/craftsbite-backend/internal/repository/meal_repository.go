package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MealRepository defines the interface for meal participation data access
type MealRepository interface {
	CreateOrUpdate(participation *models.MealParticipation) error
	FindByUserAndDate(userID, date string) ([]models.MealParticipation, error)
	FindByUserDateMeal(userID, date, mealType string) (*models.MealParticipation, error)
	FindByDate(date string) ([]models.MealParticipation, error)
	FindByDateAndMeal(date, mealType string) ([]models.MealParticipation, error)
}

// mealRepository implements MealRepository
type mealRepository struct {
	db *gorm.DB
}

// NewMealRepository creates a new meal repository
func NewMealRepository(db *gorm.DB) MealRepository {
	return &mealRepository{db: db}
}

// CreateOrUpdate creates or updates a meal participation (upsert)
func (r *mealRepository) CreateOrUpdate(participation *models.MealParticipation) error {
	// Use GORM's Clauses with OnConflict to handle upsert
	// The unique constraint is on (user_id, date, meal_type)
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "date"},
			{Name: "meal_type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"is_participating",
			"opted_out_at",
			"override_by",
			"override_reason",
			"updated_at",
		}),
	}).Create(participation).Error

	if err != nil {
		return fmt.Errorf("failed to create or update meal participation: %w", err)
	}
	return nil
}

// FindByUserAndDate finds all meal participations for a user on a specific date
func (r *mealRepository) FindByUserAndDate(userID, date string) ([]models.MealParticipation, error) {
	var participations []models.MealParticipation
	err := r.db.Where("user_id = ? AND date = ?", userID, date).Find(&participations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find meal participations: %w", err)
	}
	return participations, nil
}

// FindByUserDateMeal finds a specific meal participation
func (r *mealRepository) FindByUserDateMeal(userID, date, mealType string) (*models.MealParticipation, error) {
	var participation models.MealParticipation
	err := r.db.Where("user_id = ? AND date = ? AND meal_type = ?", userID, date, mealType).First(&participation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error, return nil
		}
		return nil, fmt.Errorf("failed to find meal participation: %w", err)
	}
	return &participation, nil
}

// FindByDate finds all meal participations for a specific date
func (r *mealRepository) FindByDate(date string) ([]models.MealParticipation, error) {
	var participations []models.MealParticipation
	err := r.db.Where("date = ?", date).Find(&participations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find meal participations by date: %w", err)
	}
	return participations, nil
}

// FindByDateAndMeal finds all participations for a specific date and meal type
func (r *mealRepository) FindByDateAndMeal(date, mealType string) ([]models.MealParticipation, error) {
	var participations []models.MealParticipation
	err := r.db.Where("date = ? AND meal_type = ?", date, mealType).Find(&participations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find meal participations by date and meal: %w", err)
	}
	return participations, nil
}
