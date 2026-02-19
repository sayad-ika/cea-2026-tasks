package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WorkLocationRepository interface {
	Upsert(wl *models.WorkLocation) error
	FindByUserAndDate(userID, date string) (*models.WorkLocation, error)
	FindByDate(date string) ([]models.WorkLocation, error)
	FindByDateAndUserIDs(date string, userIDs []string) ([]models.WorkLocation, error)
}

type workLocationRepository struct {
	db *gorm.DB
}

func NewWorkLocationRepository(db *gorm.DB) WorkLocationRepository {
	return &workLocationRepository{db: db}
}

func (r *workLocationRepository) Upsert(wl *models.WorkLocation) error {
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"location", "set_by", "reason", "updated_at"}),
	}).Create(wl)
	if result.Error != nil {
		return fmt.Errorf("failed to upsert work location: %w", result.Error)
	}
	return nil
}

func (r *workLocationRepository) FindByUserAndDate(userID, date string) (*models.WorkLocation, error) {
	var wl models.WorkLocation
	err := r.db.Where("user_id = ? AND date = ?", userID, date).First(&wl).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find work location: %w", err)
	}
	return &wl, nil
}

func (r *workLocationRepository) FindByDate(date string) ([]models.WorkLocation, error) {
	var wls []models.WorkLocation
	err := r.db.Preload("User").Where("date = ?", date).Find(&wls).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list work locations: %w", err)
	}
	return wls, nil
}

func (r *workLocationRepository) FindByDateAndUserIDs(date string, userIDs []string) ([]models.WorkLocation, error) {
	if len(userIDs) == 0 {
		return []models.WorkLocation{}, nil
	}
	var wls []models.WorkLocation
	err := r.db.Preload("User").Where("date = ? AND user_id IN ?", date, userIDs).Find(&wls).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list work locations for team: %w", err)
	}
	return wls, nil
}
