package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WorkLocationRepository defines the interface for work location data access
type WorkLocationRepository interface {
	UpsertStatus(status *models.WorkLocationStatus) error
	FindStatusByUserAndDate(userID, date string) (*models.WorkLocationStatus, error)
	FindStatusesByDateForUsers(date string, userIDs []string) ([]models.WorkLocationStatus, error)

	CreateGlobalPolicy(policy *models.GlobalWorkLocationPolicy) error
	FindActiveGlobalPolicyByDate(date string) (*models.GlobalWorkLocationPolicy, error)
	FindGlobalPolicies(startDate, endDate string) ([]models.GlobalWorkLocationPolicy, error)
	FindGlobalPolicyByID(id string) (*models.GlobalWorkLocationPolicy, error)
	DeactivateGlobalPolicy(id string) error

	CreateHistory(history *models.WorkLocationHistory) error
	FindHistoryByUser(userID string, limit int) ([]models.WorkLocationHistory, error)
}

// workLocationRepository implements WorkLocationRepository
type workLocationRepository struct {
	db *gorm.DB
}

// NewWorkLocationRepository creates a new work location repository
func NewWorkLocationRepository(db *gorm.DB) WorkLocationRepository {
	return &workLocationRepository{db: db}
}

// UpsertStatus creates or updates a work location status by (user_id, date)
func (r *workLocationRepository) UpsertStatus(status *models.WorkLocationStatus) error {
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"location":   status.Location,
			"updated_by": status.UpdatedBy,
			"reason":     status.Reason,
			"updated_at": time.Now(),
		}),
	}).Create(status).Error
	if err != nil {
		return fmt.Errorf("failed to upsert work location status: %w", err)
	}
	return nil
}

// FindStatusByUserAndDate finds a specific work location status
func (r *workLocationRepository) FindStatusByUserAndDate(userID, date string) (*models.WorkLocationStatus, error) {
	var status models.WorkLocationStatus
	err := r.db.Where("user_id = ? AND date = ?", userID, date).First(&status).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find work location status: %w", err)
	}
	return &status, nil
}

// FindStatusesByDateForUsers finds work location statuses for multiple users on a date
func (r *workLocationRepository) FindStatusesByDateForUsers(date string, userIDs []string) ([]models.WorkLocationStatus, error) {
	if len(userIDs) == 0 {
		return []models.WorkLocationStatus{}, nil
	}

	var statuses []models.WorkLocationStatus
	err := r.db.Where("date = ? AND user_id IN ?", date, userIDs).Find(&statuses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find work location statuses by date/users: %w", err)
	}
	return statuses, nil
}

// CreateGlobalPolicy creates a global work location policy
func (r *workLocationRepository) CreateGlobalPolicy(policy *models.GlobalWorkLocationPolicy) error {
	if err := r.db.Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create global work location policy: %w", err)
	}
	return nil
}

// FindActiveGlobalPolicyByDate finds an active global policy that covers a date
func (r *workLocationRepository) FindActiveGlobalPolicyByDate(date string) (*models.GlobalWorkLocationPolicy, error) {
	var policy models.GlobalWorkLocationPolicy
	err := r.db.Where("is_active = ? AND ? BETWEEN start_date AND end_date", true, date).
		Order("created_at DESC").
		First(&policy).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find active global policy: %w", err)
	}
	return &policy, nil
}

// FindGlobalPolicies finds global policies with optional overlapping date-range filter
func (r *workLocationRepository) FindGlobalPolicies(startDate, endDate string) ([]models.GlobalWorkLocationPolicy, error) {
	var policies []models.GlobalWorkLocationPolicy
	query := r.db.Model(&models.GlobalWorkLocationPolicy{})

	if startDate != "" && endDate != "" {
		query = query.Where("start_date <= ? AND end_date >= ?", endDate, startDate)
	}

	if err := query.Order("created_at DESC").Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to find global work location policies: %w", err)
	}
	return policies, nil
}

// FindGlobalPolicyByID finds a global policy by ID
func (r *workLocationRepository) FindGlobalPolicyByID(id string) (*models.GlobalWorkLocationPolicy, error) {
	var policy models.GlobalWorkLocationPolicy
	err := r.db.Where("id = ?", id).First(&policy).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find global work location policy: %w", err)
	}
	return &policy, nil
}

// DeactivateGlobalPolicy deactivates a global policy
func (r *workLocationRepository) DeactivateGlobalPolicy(id string) error {
	result := r.db.Model(&models.GlobalWorkLocationPolicy{}).
		Where("id = ? AND is_active = ?", id, true).
		Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to deactivate global work location policy: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("global work location policy not found or already inactive")
	}
	return nil
}

// CreateHistory creates a new work location history record
func (r *workLocationRepository) CreateHistory(history *models.WorkLocationHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create work location history record: %w", err)
	}
	return nil
}

// FindHistoryByUser finds work location history records for a user, ordered by created_at DESC
func (r *workLocationRepository) FindHistoryByUser(userID string, limit int) ([]models.WorkLocationHistory, error) {
	var history []models.WorkLocationHistory
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to find work location history records: %w", err)
	}
	return history, nil
}
