package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateBulkOptOutInput represents input for creating a bulk opt-out
type CreateBulkOptOutInput struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	MealType  string `json:"meal_type" binding:"required"`
	Reason    string `json:"reason"`
}

// BulkOptOutService defines the interface for bulk opt-out management
type BulkOptOutService interface {
	GetBulkOptOuts(userID string) ([]models.BulkOptOut, error)
	CreateBulkOptOut(userID string, input CreateBulkOptOutInput, createdByUserID string) (*models.BulkOptOut, error)
	CreateBatchBulkOptOut(userIDs []string, input CreateBulkOptOutInput, createdByUserID string) ([]models.BulkOptOut, error)
	DeleteBulkOptOut(userID, id string) error
}

// bulkOptOutService implements BulkOptOutService
type bulkOptOutService struct {
	bulkOptOutRepo repository.BulkOptOutRepository
	historyRepo    repository.HistoryRepository
}

// NewBulkOptOutService creates a new bulk opt-out service
func NewBulkOptOutService(bulkOptOutRepo repository.BulkOptOutRepository, historyRepo repository.HistoryRepository) BulkOptOutService {
	return &bulkOptOutService{
		bulkOptOutRepo: bulkOptOutRepo,
		historyRepo:    historyRepo,
	}
}

// GetBulkOptOuts retrieves all bulk opt-outs for a user
func (s *bulkOptOutService) GetBulkOptOuts(userID string) ([]models.BulkOptOut, error) {
	optOuts, err := s.bulkOptOutRepo.FindByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk opt-outs: %w", err)
	}
	return optOuts, nil
}

// CreateBulkOptOut creates a new bulk opt-out for a user
func (s *bulkOptOutService) CreateBulkOptOut(userID string, input CreateBulkOptOutInput, createdByUserID string) (*models.BulkOptOut, error) {
	// Validate date format
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: must be YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: must be YYYY-MM-DD")
	}

	// Validate end_date >= start_date
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}

	// Validate meal type
	mealType := models.MealType(input.MealType)
	if !mealType.IsValid() {
		return nil, fmt.Errorf("invalid meal_type: must be one of lunch, snacks, iftar, event_dinner, optional_dinner")
	}

	// Parse user UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Create bulk opt-out
	bulkOptOut := &models.BulkOptOut{
		UserID:          userUUID,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		MealType:        mealType,
		IsActive:        true,
		CreatedByUserID: uuid.MustParse(createdByUserID),
		Reason:          input.Reason,
	}

	if err := s.bulkOptOutRepo.Create(bulkOptOut); err != nil {
		return nil, fmt.Errorf("failed to create bulk opt-out: %w", err)
	}

	// Record in history
	historyRecord := &models.MealParticipationHistory{
		UserID:          userUUID,
		Date:            input.StartDate, // Use start date as reference
		MealType:        mealType,
		Action:          models.HistoryActionOptedOut,
		ChangedByUserID: &userUUID,
		Reason:          strPtr(fmt.Sprintf("Bulk opt-out from %s to %s. Reason: %s", input.StartDate, input.EndDate, input.Reason)),
	}

	if err := s.historyRepo.Create(historyRecord); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to record bulk opt-out in history: %v\n", err)
	}

	return bulkOptOut, nil
}

// CreateBatchBulkOptOut creates bulk opt-outs for multiple users
func (s *bulkOptOutService) CreateBatchBulkOptOut(userIDs []string, input CreateBulkOptOutInput, createdByUserID string) ([]models.BulkOptOut, error) {
	var results []models.BulkOptOut

	// Basic validation of the input once
	_, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: must be YYYY-MM-DD")
	}
	_, err = time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: must be YYYY-MM-DD")
	}
	// Note: Detailed validation happens inside CreateBulkOptOut

	for _, userID := range userIDs {
		optOut, err := s.CreateBulkOptOut(userID, input, createdByUserID)
		if err != nil {
			// for now, if one fails, we return error. In future we might want partial success.
			return nil, fmt.Errorf("failed to create opt-out for user %s: %w", userID, err)
		}
		results = append(results, *optOut)
	}

	return results, nil
}

// DeleteBulkOptOut deletes a bulk opt-out if it belongs to the user
func (s *bulkOptOutService) DeleteBulkOptOut(userID, id string) error {
	// Get all user's bulk opt-outs to verify ownership
	optOuts, err := s.bulkOptOutRepo.FindByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to verify bulk opt-out ownership: %w", err)
	}

	// Check if the opt-out belongs to the user
	found := false
	for _, optOut := range optOuts {
		if optOut.ID.String() == id {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("bulk opt-out not found or does not belong to user")
	}

	// Delete the opt-out
	if err := s.bulkOptOutRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete bulk opt-out: %w", err)
	}

	return nil
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}
