package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"
)

// HeadcountService defines the interface for headcount calculations
type HeadcountService interface {
	GetTodayHeadcount() ([]*DailyHeadcountSummary, error)
	GetHeadcountByDate(date string) (*DailyHeadcountSummary, error)
	GetDetailedHeadcount(date, mealType string) (*DetailedHeadcount, error)
}

// MealHeadcount represents participation breakdown for a single meal
type MealHeadcount struct {
	Participating int `json:"participating"`
	OptedOut      int `json:"opted_out"`
}

// DailyHeadcountSummary represents the headcount summary for a day
type DailyHeadcountSummary struct {
	Date             string                   `json:"date"`
	DayStatus        models.DayStatus         `json:"day_status"`
	TotalActiveUsers int                      `json:"total_active_users"`
	Meals            map[string]MealHeadcount `json:"meals"`
}

// DetailedHeadcount represents detailed headcount for a specific meal
type DetailedHeadcount struct {
	Date            string            `json:"date"`
	MealType        string            `json:"meal_type"`
	Participants    []ParticipantInfo `json:"participants"`
	NonParticipants []ParticipantInfo `json:"non_participants"`
	TotalCount      int               `json:"total_count"`
}

// ParticipantInfo represents a user's participation info
type ParticipantInfo struct {
	UserID          string `json:"user_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsParticipating bool   `json:"is_participating"`
	Source          string `json:"source"`
}

// headcountService implements HeadcountService
type headcountService struct {
	userRepo     repository.UserRepository
	scheduleRepo repository.ScheduleRepository
	resolver     ParticipationResolver
}

// NewHeadcountService creates a new headcount service
func NewHeadcountService(
	userRepo repository.UserRepository,
	scheduleRepo repository.ScheduleRepository,
	resolver ParticipationResolver,
) HeadcountService {
	return &headcountService{
		userRepo:     userRepo,
		scheduleRepo: scheduleRepo,
		resolver:     resolver,
	}
}

// GetTodayHeadcount gets today's and tomorrow's headcount summary
func (s *headcountService) GetTodayHeadcount() ([]*DailyHeadcountSummary, error) {
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	
	todaySummary, err := s.GetHeadcountByDate(today)
	if err != nil {
		return nil, err
	}
	
	tomorrowSummary, err := s.GetHeadcountByDate(tomorrow)
	if err != nil {
		return nil, err
	}
	
	return []*DailyHeadcountSummary{todaySummary, tomorrowSummary}, nil
}

// GetHeadcountByDate gets headcount summary for a specific date
func (s *headcountService) GetHeadcountByDate(date string) (*DailyHeadcountSummary, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Get all active users
	filters := map[string]interface{}{
		"active": true,
	}
	users, err := s.userRepo.FindAll(filters)
	if err != nil {
		return nil, err
	}

	// Get day schedule
	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	dayStatus := models.DayStatusNormal
	availableMeals := []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}

	if schedule != nil {
		dayStatus = schedule.DayStatus
		if schedule.AvailableMeals != nil {
			availableMeals = parseMealTypes(*schedule.AvailableMeals)
		}
	}

	totalActiveUsers := len(users)

	// Calculate counts for each meal
	meals := make(map[string]MealHeadcount)

	for _, mealType := range availableMeals {
		participating := 0
		for _, user := range users {
			isParticipating, _, err := s.resolver.ResolveParticipation(user.ID.String(), date, string(mealType))
			if err != nil {
				return nil, err
			}
			if isParticipating {
				participating++
			}
		}
		meals[string(mealType)] = MealHeadcount{
			Participating: participating,
			OptedOut:      totalActiveUsers - participating,
		}
	}

	return &DailyHeadcountSummary{
		Date:             date,
		DayStatus:        dayStatus,
		TotalActiveUsers: totalActiveUsers,
		Meals:            meals,
	}, nil
}

// GetDetailedHeadcount gets detailed headcount for a specific date and meal
func (s *headcountService) GetDetailedHeadcount(date, mealType string) (*DetailedHeadcount, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Get all active users
	filters := map[string]interface{}{
		"active": true,
	}
	users, err := s.userRepo.FindAll(filters)
	if err != nil {
		return nil, err
	}

	participants := []ParticipantInfo{}
	nonParticipants := []ParticipantInfo{}
	totalCount := 0

	for _, user := range users {
		if !user.Active {
			continue
		}

		isParticipating, source, err := s.resolver.ResolveParticipation(user.ID.String(), date, mealType)
		if err != nil {
			return nil, err
		}

		info := ParticipantInfo{
			UserID:          user.ID.String(),
			Name:            user.Name,
			Email:           user.Email,
			IsParticipating: isParticipating,
			Source:          source,
		}

		if isParticipating {
			participants = append(participants, info)
			totalCount++
		} else {
			nonParticipants = append(nonParticipants, info)
		}
	}

	return &DetailedHeadcount{
		Date:            date,
		MealType:        mealType,
		Participants:    participants,
		NonParticipants: nonParticipants,
		TotalCount:      totalCount,
	}, nil
}
