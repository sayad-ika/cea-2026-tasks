package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ScheduleService defines the interface for day schedule business logic
type ScheduleService interface {
	GetSchedule(date string) (*models.DaySchedule, error)
	GetScheduleRange(startDate, endDate string) ([]models.DaySchedule, error)
	CreateSchedule(adminID string, input CreateScheduleInput) (*models.DaySchedule, error)
	UpdateSchedule(id string, input UpdateScheduleInput) (*models.DaySchedule, error)
	DeleteSchedule(id string) error
}

// CreateScheduleInput represents input for creating a day schedule
type CreateScheduleInput struct {
	Date           string            `json:"date" binding:"required"`
	DayStatus      models.DayStatus  `json:"day_status" binding:"required"`
	Reason         string            `json:"reason"`
	AvailableMeals []models.MealType `json:"available_meals"`
}

// UpdateScheduleInput represents input for updating a day schedule
type UpdateScheduleInput struct {
	DayStatus      *models.DayStatus  `json:"day_status"`
	Reason         *string            `json:"reason"`
	AvailableMeals *[]models.MealType `json:"available_meals"`
}

// scheduleService implements ScheduleService
type scheduleService struct {
	scheduleRepo repository.ScheduleRepository
}

// NewScheduleService creates a new schedule service
func NewScheduleService(scheduleRepo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{
		scheduleRepo: scheduleRepo,
	}
}

// GetSchedule gets a day schedule by date; auto-creates a default entry if none exists
func (s *scheduleService) GetSchedule(date string) (*models.DaySchedule, error) {
	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	existing, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	// If schedule exists, return it
	if existing != nil {
		return existing, nil
	}

	// Auto-create default schedule entry and save to DB
	schedule := s.buildDefaultSchedule(parsedDate)
	if err := s.scheduleRepo.Create(schedule); err != nil {
		return nil, fmt.Errorf("failed to create default schedule: %w", err)
	}

	return schedule, nil
}

// GetScheduleRange gets day schedules within a date range; auto-creates defaults for missing dates
func (s *scheduleService) GetScheduleRange(startDate, endDate string) ([]models.DaySchedule, error) {
	// Validate date formats
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format, expected YYYY-MM-DD: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format, expected YYYY-MM-DD: %w", err)
	}

	// Fetch existing schedules for the range
	existing, err := s.scheduleRepo.FindByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Build a map of existing dates for quick lookup
	// Normalize the date key to YYYY-MM-DD format to handle PostgreSQL date format variations
	existingMap := make(map[string]models.DaySchedule)
	for _, sched := range existing {
		normalizedDate := normalizeDateStr(sched.Date)
		existingMap[normalizedDate] = sched
	}

	// Iterate through all dates in the range and fill in missing ones
	var result []models.DaySchedule
	var toCreate []*models.DaySchedule
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if sched, ok := existingMap[dateStr]; ok {
			result = append(result, sched)
		} else {
			// Build default and queue for batch insert
			defaultSched := s.buildDefaultSchedule(d)
			result = append(result, *defaultSched)
			toCreate = append(toCreate, defaultSched)
		}
	}

	// Batch create all missing schedules
	for _, sched := range toCreate {
		if err := s.scheduleRepo.Create(sched); err != nil {
			return nil, fmt.Errorf("failed to create default schedule for %s: %w", sched.Date, err)
		}
	}

	return result, nil
}

// normalizeDateStr converts date strings to YYYY-MM-DD format
// Handles both "2026-02-20" and "2026-02-20T00:00:00Z" formats from PostgreSQL
func normalizeDateStr(dateStr string) string {
	// Try parsing as full timestamp first
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t.Format("2006-01-02")
	}
	// Try parsing as date with time
	if t, err := time.Parse("2006-01-02T15:04:05Z07:00", dateStr); err == nil {
		return t.Format("2006-01-02")
	}
	// Already in YYYY-MM-DD format or just return first 10 chars
	if len(dateStr) >= 10 {
		return dateStr[:10]
	}
	return dateStr
}

// buildDefaultSchedule creates a default schedule entry for a given date
func (s *scheduleService) buildDefaultSchedule(date time.Time) *models.DaySchedule {
	dateStr := date.Format("2006-01-02")
	now := time.Now()

	var dayStatus models.DayStatus
	var mealsStr string

	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		dayStatus = models.DayStatusWeekend
		mealsStr = "" // no meals on weekends
	} else {
		dayStatus = models.DayStatusNormal
		mealsStr = serializeMealTypes([]models.MealType{models.MealTypeLunch, models.MealTypeSnacks})
	}

	return &models.DaySchedule{
		ID:             uuid.New(),
		Date:           dateStr,
		DayStatus:      dayStatus,
		AvailableMeals: &mealsStr,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// CreateSchedule creates a new day schedule
func (s *scheduleService) CreateSchedule(adminID string, input CreateScheduleInput) (*models.DaySchedule, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Validate day status
	if !input.DayStatus.IsValid() {
		return nil, fmt.Errorf("invalid day status: %s", input.DayStatus)
	}

	// Requirement 3: Validate celebration requires a reason
	if input.DayStatus == models.DayStatusCelebration && input.Reason == "" {
		return nil, fmt.Errorf("reason is required for celebration day status")
	}

	// Parse admin UUID
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID: %w", err)
	}

	// Requirement 3: Auto-set available_meals based on day_status
	var finalMeals []models.MealType
	switch input.DayStatus {
	case models.DayStatusOfficeClosed, models.DayStatusGovtHoliday:
		finalMeals = []models.MealType{}
	case models.DayStatusCelebration, models.DayStatusNormal:
		if len(input.AvailableMeals) > 0 {
			finalMeals = input.AvailableMeals
		} else {
			finalMeals = []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}
		}
	default:
		if len(input.AvailableMeals) > 0 {
			finalMeals = input.AvailableMeals
		} else {
			finalMeals = []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}
		}
	}

	mealsStr := serializeMealTypes(finalMeals)

	// Create reason pointer if not empty
	var reasonPtr *string
	if input.Reason != "" {
		reasonPtr = &input.Reason
	}

	// Check if schedule already exists for this date
	existing, err := s.scheduleRepo.FindByDate(input.Date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// If the existing schedule was system-generated (no CreatedBy), allow overwriting it
		if existing.CreatedBy == nil {
			existing.DayStatus = input.DayStatus
			existing.CreatedBy = &adminUUID
			existing.Reason = reasonPtr
			existing.AvailableMeals = &mealsStr

			if err := s.scheduleRepo.Update(existing); err != nil {
				return nil, err
			}
			return existing, nil
		}
		return nil, fmt.Errorf("schedule already exists for date %s", input.Date)
	}

	schedule := &models.DaySchedule{
		ID:             uuid.New(),
		Date:           input.Date,
		DayStatus:      input.DayStatus,
		Reason:         reasonPtr,
		AvailableMeals: &mealsStr,
		CreatedBy:      &adminUUID,
	}

	if err := s.scheduleRepo.Create(schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// UpdateSchedule updates an existing day schedule
func (s *scheduleService) UpdateSchedule(id string, input UpdateScheduleInput) (*models.DaySchedule, error) {
	// Find existing schedule
	schedule, err := s.scheduleRepo.FindByDate(id)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	// Requirement 3: Validate if day_status is being updated to celebration, reason must be provided
	if input.DayStatus != nil {
		// Validate day status
		if !input.DayStatus.IsValid() {
			return nil, fmt.Errorf("invalid day status: %s", *input.DayStatus)
		}

		// Check if updating to celebration requires a reason
		if *input.DayStatus == models.DayStatusCelebration {
			// Check if reason is being updated or if existing reason is sufficient
			if input.Reason != nil && *input.Reason == "" {
				return nil, fmt.Errorf("reason is required for celebration day status")
			}
			if input.Reason == nil && (schedule.Reason == nil || *schedule.Reason == "") {
				return nil, fmt.Errorf("reason is required for celebration day status")
			}
		}
	}

	// Update fields if provided
	if input.DayStatus != nil {
		schedule.DayStatus = *input.DayStatus

		// Requirement 3: Auto-set available_meals based on new day_status
		switch *input.DayStatus {
		case models.DayStatusOfficeClosed, models.DayStatusGovtHoliday:
			// Force clear meals for closed/holiday days
			emptyMeals := serializeMealTypes([]models.MealType{})
			schedule.AvailableMeals = &emptyMeals
		case models.DayStatusCelebration, models.DayStatusNormal:
			// Only set default if not explicitly provided in input
			if input.AvailableMeals == nil {
				// Keep existing meals if they exist, otherwise set default
				if schedule.AvailableMeals == nil || *schedule.AvailableMeals == "" {
					defaultMeals := serializeMealTypes([]models.MealType{models.MealTypeLunch, models.MealTypeSnacks})
					schedule.AvailableMeals = &defaultMeals
				}
			}
		}
	}

	if input.Reason != nil {
		schedule.Reason = input.Reason
	}

	if input.AvailableMeals != nil {
		mealsStr := serializeMealTypes(*input.AvailableMeals)
		schedule.AvailableMeals = &mealsStr
	}

	if err := s.scheduleRepo.Update(schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// DeleteSchedule deletes a day schedule
func (s *scheduleService) DeleteSchedule(id string) error {
	return s.scheduleRepo.Delete(id)
}
