package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MealService defines the interface for meal participation business logic
type MealService interface {
	GetTodayMeals(userID string) (*TodayMealsResponse, error)
	GetParticipation(userID, date string) ([]ParticipationStatus, error)
	SetParticipation(userID, date, mealType string, participating bool) error
	OverrideParticipation(adminID, userID, date, mealType string, participating bool, reason string) error
}

// TodayMealsResponse represents the response for today's meals
type TodayMealsResponse struct {
	Date           string                `json:"date"`
	DayStatus      models.DayStatus      `json:"day_status"`
	AvailableMeals []models.MealType     `json:"available_meals"`
	Participations []ParticipationStatus `json:"participations"`
}

// ParticipationStatus represents a user's participation status for a meal
type ParticipationStatus struct {
	MealType        models.MealType `json:"meal_type"`
	IsParticipating bool            `json:"is_participating"`
	Source          string          `json:"source"`
}

// mealService implements MealService
type mealService struct {
	mealRepo       repository.MealRepository
	scheduleRepo   repository.ScheduleRepository
	historyRepo    repository.HistoryRepository
	userRepo       repository.UserRepository
	teamRepo       repository.TeamRepository
	resolver       ParticipationResolver
	cutoffTime     string
	cutoffTimezone string
}

// NewMealService creates a new meal service
func NewMealService(
	mealRepo repository.MealRepository,
	scheduleRepo repository.ScheduleRepository,
	historyRepo repository.HistoryRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	resolver ParticipationResolver,
	cfg *config.Config,
) MealService {
	return &mealService{
		mealRepo:       mealRepo,
		scheduleRepo:   scheduleRepo,
		historyRepo:    historyRepo,
		userRepo:       userRepo,
		teamRepo:       teamRepo,
		resolver:       resolver,
		cutoffTime:     cfg.Meal.CutoffTime,
		cutoffTimezone: cfg.Meal.CutoffTimezone,
	}
}

// GetTodayMeals gets today's meals and participation status for a user
func (s *mealService) GetTodayMeals(userID string) (*TodayMealsResponse, error) {
	today := time.Now().Format("2006-01-02")

	// Get day schedule
	schedule, err := s.scheduleRepo.FindByDate(today)
	if err != nil {
		return nil, err
	}

	response := &TodayMealsResponse{
		Date:           today,
		DayStatus:      models.DayStatusNormal,
		AvailableMeals: []models.MealType{models.MealTypeLunch, models.MealTypeSnacks},
		Participations: []ParticipationStatus{},
	}

	if schedule != nil {
		response.DayStatus = schedule.DayStatus
		if schedule.AvailableMeals != nil {
			response.AvailableMeals = parseMealTypes(*schedule.AvailableMeals)
		}
	}

	// Resolve participation for each available meal
	for _, mealType := range response.AvailableMeals {
		isParticipating, source, err := s.resolver.ResolveParticipation(userID, today, string(mealType))
		if err != nil {
			return nil, err
		}

		response.Participations = append(response.Participations, ParticipationStatus{
			MealType:        mealType,
			IsParticipating: isParticipating,
			Source:          source,
		})
	}

	return response, nil
}

// GetParticipation gets participation status for a user on a specific date
func (s *mealService) GetParticipation(userID, date string) ([]ParticipationStatus, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Get day schedule to know available meals
	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	availableMeals := []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}
	if schedule != nil && schedule.AvailableMeals != nil {
		availableMeals = parseMealTypes(*schedule.AvailableMeals)
	}

	participations := []ParticipationStatus{}
	for _, mealType := range availableMeals {
		isParticipating, source, err := s.resolver.ResolveParticipation(userID, date, string(mealType))
		if err != nil {
			return nil, err
		}

		participations = append(participations, ParticipationStatus{
			MealType:        mealType,
			IsParticipating: isParticipating,
			Source:          source,
		})
	}

	return participations, nil
}

// SetParticipation sets a user's participation for a specific date and meal
func (s *mealService) SetParticipation(userID, date, mealType string, participating bool) error {
	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Validate cutoff time
	if err := s.validateCutoffTime(parsedDate); err != nil {
		return err
	}

	// Check if existing record exists to get its ID for proper upsert
	existing, err := s.mealRepo.FindByUserDateMeal(userID, date, mealType)
	if err != nil {
		return fmt.Errorf("failed to check existing participation: %w", err)
	}

	// Create or update participation record
	var participationID uuid.UUID
	if existing != nil {
		participationID = existing.ID // Reuse existing ID for UPDATE
	} else {
		participationID = uuid.New() // New ID for INSERT
	}

	participation := &models.MealParticipation{
		ID:              participationID,
		UserID:          uuid.MustParse(userID),
		Date:            date,
		MealType:        models.MealType(mealType),
		IsParticipating: participating,
	}

	if participating {
		participation.OptedOutAt = nil
	} else {
		now := time.Now()
		participation.OptedOutAt = &now
	}

	if err := s.mealRepo.CreateOrUpdate(participation); err != nil {
		return err
	}

	// Record in history
	action := models.HistoryActionOptedOut
	if participating {
		action = models.HistoryActionOptedIn
	}

	history := &models.MealParticipationHistory{
		ID:       uuid.New(),
		UserID:   uuid.MustParse(userID),
		Date:     date,
		MealType: models.MealType(mealType),
		Action:   action,
	}

	return s.historyRepo.Create(history)
}

// OverrideParticipation allows an admin or team lead to override a user's participation
// Team leads can only override their own team members
func (s *mealService) OverrideParticipation(requesterID, userID, date, mealType string, participating bool, reason string) error {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Parse requester UUID
	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return fmt.Errorf("invalid requester ID: %w", err)
	}

	// Get requester's role to validate permissions
	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return fmt.Errorf("failed to find requester: %w", err)
	}

	// If requester is a team lead, verify they manage the target user
	if requester.Role == models.RoleTeamLead {
		isMember, err := s.teamRepo.IsUserInAnyTeamLedBy(requesterID, userID)
		if err != nil {
			return fmt.Errorf("failed to check team membership: %w", err)
		}
		if !isMember {
			return fmt.Errorf("team lead can only override participation for their own team members")
		}
	}
	// Admin and Logistics roles can override anyone (no additional check needed)

	// Check if existing record exists to get its ID for proper upsert
	existing, err := s.mealRepo.FindByUserDateMeal(userID, date, mealType)
	if err != nil {
		return fmt.Errorf("failed to check existing participation: %w", err)
	}

	// Create or update participation record with override info
	var participationID uuid.UUID
	if existing != nil {
		participationID = existing.ID // Reuse existing ID for UPDATE
	} else {
		participationID = uuid.New() // New ID for INSERT
	}

	participation := &models.MealParticipation{
		ID:              participationID,
		UserID:          uuid.MustParse(userID),
		Date:            date,
		MealType:        models.MealType(mealType),
		IsParticipating: participating,
		OverrideBy:      &requesterUUID,
		OverrideReason:  &reason,
	}

	if err := s.mealRepo.CreateOrUpdate(participation); err != nil {
		return err
	}

	// Record in history
	action := models.HistoryActionOverrideOut
	if participating {
		action = models.HistoryActionOverrideIn
	}

	history := &models.MealParticipationHistory{
		ID:              uuid.New(),
		UserID:          uuid.MustParse(userID),
		Date:            date,
		MealType:        models.MealType(mealType),
		Action:          action,
		ChangedByUserID: &requesterUUID,
	}

	return s.historyRepo.Create(history)
}

// validateCutoffTime checks if the current time is before the cutoff time for the given date
func (s *mealService) validateCutoffTime(targetDate time.Time) error {
	// Load timezone
	loc, err := time.LoadLocation(s.cutoffTimezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	// Parse cutoff time (e.g., "11:00")
	cutoffParts := s.cutoffTime
	cutoffTime, err := time.Parse("15:04", cutoffParts)
	if err != nil {
		return fmt.Errorf("invalid cutoff time format: %w", err)
	}

	// Build cutoff datetime for the target date
	cutoffDateTime := time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		cutoffTime.Hour(),
		cutoffTime.Minute(),
		0, 0, loc,
	)

	// Get current time in the configured timezone
	now := time.Now().In(loc)

	// Check if current time is past the cutoff
	if now.After(cutoffDateTime) {
		return fmt.Errorf("cutoff time (%s %s) has passed for date %s",
			s.cutoffTime, s.cutoffTimezone, targetDate.Format("2006-01-02"))
	}

	return nil
}
