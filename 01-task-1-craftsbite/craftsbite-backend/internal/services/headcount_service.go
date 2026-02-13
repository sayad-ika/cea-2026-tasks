package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"sort"
	"strings"
	"time"
)

// HeadcountService defines the interface for headcount calculations
type HeadcountService interface {
	GetTodayHeadcount() ([]*DailyHeadcountSummary, error)
	GetHeadcountByDate(date string) (*DailyHeadcountSummary, error)
	GetDetailedHeadcount(date, mealType string) (*DetailedHeadcount, error)
	GenerateDailyAnnouncement(date string) (*DailyAnnouncementResponse, error)
	GetEnhancedHeadcountReport(date ...string) ([]*EnhancedHeadcountReportResponse, error)
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

// AnnouncementMealTotal represents meal-wise totals in the daily announcement
type AnnouncementMealTotal struct {
	MealType      string `json:"meal_type"`
	Participating int    `json:"participating"`
	OptedOut      int    `json:"opted_out"`
}

// DailyAnnouncementResponse represents copy/paste-friendly daily announcement content
type DailyAnnouncementResponse struct {
	Date             string                  `json:"date"`
	DayStatus        models.DayStatus        `json:"day_status"`
	SpecialDayNote   string                  `json:"special_day_note,omitempty"`
	TotalActiveUsers int                     `json:"total_active_users"`
	MealTotals       []AnnouncementMealTotal `json:"meal_totals"`
	Message          string                  `json:"message"`
}

// OfficeWFHSplit represents effective office vs WFH counts
type OfficeWFHSplit struct {
	Office int `json:"office"`
	WFH    int `json:"wfh"`
}

// TeamHeadcountTotals represents totals for a team on a specific date
type TeamHeadcountTotals struct {
	TeamID         string                   `json:"team_id"`
	TeamName       string                   `json:"team_name"`
	TotalMembers   int                      `json:"total_members"`
	OfficeWFHSplit OfficeWFHSplit           `json:"office_wfh_split"`
	MealTypeTotals map[string]MealHeadcount `json:"meal_type_totals"`
}

// EnhancedHeadcountReportResponse represents improved headcount reporting for a date
type EnhancedHeadcountReportResponse struct {
	Date            string                   `json:"date"`
	DayStatus       models.DayStatus         `json:"day_status"`
	SpecialDayNote  string                   `json:"special_day_note,omitempty"`
	OverallTotal    int                      `json:"overall_total"`
	MealTypeTotals  map[string]MealHeadcount `json:"meal_type_totals"`
	TeamTotals      []TeamHeadcountTotals    `json:"team_totals"`
	OfficeWFHSplit  OfficeWFHSplit           `json:"office_wfh_split"`
	UnassignedUsers int                      `json:"unassigned_users"`
}

// headcountService implements HeadcountService
type headcountService struct {
	userRepo         repository.UserRepository
	scheduleRepo     repository.ScheduleRepository
	teamRepo         repository.TeamRepository
	workLocationRepo repository.WorkLocationRepository
	resolver         ParticipationResolver
	weekendDays      map[string]bool
}

// NewHeadcountService creates a new headcount service
func NewHeadcountService(
	userRepo repository.UserRepository,
	scheduleRepo repository.ScheduleRepository,
	teamRepo repository.TeamRepository,
	workLocationRepo repository.WorkLocationRepository,
	resolver ParticipationResolver,
	cfg *config.Config,
) HeadcountService {
	weekendDays := make(map[string]bool)
	for _, day := range cfg.Meal.WeekendDays {
		weekendDays[strings.ToLower(strings.TrimSpace(day))] = true
	}
	return &headcountService{
		userRepo:         userRepo,
		scheduleRepo:     scheduleRepo,
		teamRepo:         teamRepo,
		workLocationRepo: workLocationRepo,
		resolver:         resolver,
		weekendDays:      weekendDays,
	}
}

// isWeekend checks if a date falls on a configured weekend day
func (s *headcountService) isWeekend(date string) bool {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	return s.weekendDays[strings.ToLower(parsedDate.Weekday().String())]
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

	// Weekend with no schedule override → zero headcount
	if schedule == nil && s.isWeekend(date) {
		return &DailyHeadcountSummary{
			Date:             date,
			DayStatus:        models.DayStatusWeekend,
			TotalActiveUsers: len(users),
			Meals:            make(map[string]MealHeadcount),
		}, nil
	}

	dayStatus := models.DayStatusNormal
	availableMeals := []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}

	if schedule != nil {
		dayStatus = schedule.DayStatus
		switch dayStatus {
		case models.DayStatusWeekend, models.DayStatusOfficeClosed, models.DayStatusGovtHoliday:
			// Non-normal days: no meals unless explicitly configured
			if schedule.AvailableMeals != nil && *schedule.AvailableMeals != "" {
				availableMeals = parseMealTypes(*schedule.AvailableMeals)
			} else {
				availableMeals = []models.MealType{}
			}
		default:
			// Normal/celebration days: use explicit meals or keep defaults
			if schedule.AvailableMeals != nil && *schedule.AvailableMeals != "" {
				availableMeals = parseMealTypes(*schedule.AvailableMeals)
			}
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

// GenerateDailyAnnouncement creates a copy/paste-friendly message for a selected date
func (s *headcountService) GenerateDailyAnnouncement(date string) (*DailyAnnouncementResponse, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	summary, err := s.GetHeadcountByDate(date)
	if err != nil {
		return nil, err
	}

	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	specialDayNote := buildSpecialDayNote(summary.DayStatus, schedule)
	mealTotals := buildAnnouncementMealTotals(summary.Meals)
	message := buildAnnouncementMessage(summary, specialDayNote, mealTotals)

	return &DailyAnnouncementResponse{
		Date:             summary.Date,
		DayStatus:        summary.DayStatus,
		SpecialDayNote:   specialDayNote,
		TotalActiveUsers: summary.TotalActiveUsers,
		MealTotals:       mealTotals,
		Message:          message,
	}, nil
}

// GetEnhancedHeadcountReport returns improved headcount reporting with meal/team/overall/location splits
func (s *headcountService) GetEnhancedHeadcountReport(dates ...string) ([]*EnhancedHeadcountReportResponse, error) {
	var reports []*EnhancedHeadcountReportResponse

	for _, date := range dates {
		report, err := s.getEnhancedHeadcountReportForDate(date)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (s *headcountService) getEnhancedHeadcountReportForDate(date string) (*EnhancedHeadcountReportResponse, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Get all active users
	users, err := s.userRepo.FindAll(map[string]interface{}{"active": true})
	if err != nil {
		return nil, err
	}

	activeUsersByID := make(map[string]models.User, len(users))
	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		id := user.ID.String()
		activeUsersByID[id] = user
		userIDs = append(userIDs, id)
	}

	// Resolve schedule context
	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	// Weekend with no schedule override → zero headcount
	if schedule == nil && s.isWeekend(date) {
		return &EnhancedHeadcountReportResponse{
			Date:            date,
			DayStatus:       models.DayStatusWeekend,
			OverallTotal:    len(users),
			MealTypeTotals:  make(map[string]MealHeadcount),
			TeamTotals:      []TeamHeadcountTotals{},
			OfficeWFHSplit:  OfficeWFHSplit{},
			UnassignedUsers: 0,
		}, nil
	}

	dayStatus := models.DayStatusNormal
	availableMeals := []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}
	if schedule != nil {
		dayStatus = schedule.DayStatus
		switch dayStatus {
		case models.DayStatusWeekend, models.DayStatusOfficeClosed, models.DayStatusGovtHoliday:
			// Non-normal days: no meals unless explicitly configured
			if schedule.AvailableMeals != nil && *schedule.AvailableMeals != "" {
				availableMeals = parseMealTypes(*schedule.AvailableMeals)
			} else {
				availableMeals = []models.MealType{}
			}
		default:
			// Normal/celebration days: use explicit meals or keep defaults
			if schedule.AvailableMeals != nil && *schedule.AvailableMeals != "" {
				availableMeals = parseMealTypes(*schedule.AvailableMeals)
			}
		}
	}

	// Overall meal-wise totals
	mealTypeTotals, err := s.computeMealTotalsForUsers(users, date, availableMeals)
	if err != nil {
		return nil, err
	}

	// Effective work locations and org-level office/WFH split
	effectiveLocations, err := s.resolveEffectiveLocationsForUsers(date, userIDs)
	if err != nil {
		return nil, err
	}
	overallSplit := buildOfficeWFHSplitForUserIDs(userIDs, effectiveLocations)

	// Team totals
	teamTotals, assignedUsers, err := s.computeTeamTotals(date, availableMeals, activeUsersByID, effectiveLocations)
	if err != nil {
		return nil, err
	}

	unassignedUsers := len(users) - len(assignedUsers)
	if unassignedUsers < 0 {
		unassignedUsers = 0
	}

	return &EnhancedHeadcountReportResponse{
		Date:            date,
		DayStatus:       dayStatus,
		SpecialDayNote:  buildSpecialDayNote(dayStatus, schedule),
		OverallTotal:    len(users),
		MealTypeTotals:  mealTypeTotals,
		TeamTotals:      teamTotals,
		OfficeWFHSplit:  overallSplit,
		UnassignedUsers: unassignedUsers,
	}, nil
}

func (s *headcountService) computeMealTotalsForUsers(users []models.User, date string, availableMeals []models.MealType) (map[string]MealHeadcount, error) {
	mealTotals := make(map[string]MealHeadcount, len(availableMeals))
	totalUsers := len(users)

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

		mealTotals[string(mealType)] = MealHeadcount{
			Participating: participating,
			OptedOut:      totalUsers - participating,
		}
	}

	return mealTotals, nil
}

func (s *headcountService) resolveEffectiveLocationsForUsers(date string, userIDs []string) (map[string]models.WorkLocation, error) {
	result := make(map[string]models.WorkLocation, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	globalPolicy, err := s.workLocationRepo.FindActiveGlobalPolicyByDate(date)
	if err != nil {
		return nil, err
	}

	if globalPolicy != nil {
		for _, userID := range userIDs {
			result[userID] = globalPolicy.Location
		}
		return result, nil
	}

	explicitStatuses, err := s.workLocationRepo.FindStatusesByDateForUsers(date, userIDs)
	if err != nil {
		return nil, err
	}

	explicitByUser := make(map[string]models.WorkLocation, len(explicitStatuses))
	for _, status := range explicitStatuses {
		explicitByUser[status.UserID.String()] = status.Location
	}

	for _, userID := range userIDs {
		if location, ok := explicitByUser[userID]; ok {
			result[userID] = location
		} else {
			result[userID] = models.WorkLocationOffice
		}
	}

	return result, nil
}

func (s *headcountService) computeTeamTotals(
	date string,
	availableMeals []models.MealType,
	activeUsersByID map[string]models.User,
	effectiveLocations map[string]models.WorkLocation,
) ([]TeamHeadcountTotals, map[string]bool, error) {
	teams, err := s.teamRepo.FindAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch teams: %w", err)
	}

	assignedUsers := make(map[string]bool)
	teamTotals := make([]TeamHeadcountTotals, 0, len(teams))

	for _, team := range teams {
		members, err := s.teamRepo.GetTeamMembers(team.ID.String())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch members for team %s: %w", team.ID.String(), err)
		}

		filteredMembers := make([]models.User, 0, len(members))
		memberIDs := make([]string, 0, len(members))
		for _, member := range members {
			memberID := member.ID.String()
			if _, exists := activeUsersByID[memberID]; !exists {
				continue
			}
			filteredMembers = append(filteredMembers, member)
			memberIDs = append(memberIDs, memberID)
			assignedUsers[memberID] = true
		}

		mealTotals, err := s.computeMealTotalsForUsers(filteredMembers, date, availableMeals)
		if err != nil {
			return nil, nil, err
		}

		teamTotals = append(teamTotals, TeamHeadcountTotals{
			TeamID:         team.ID.String(),
			TeamName:       team.Name,
			TotalMembers:   len(filteredMembers),
			OfficeWFHSplit: buildOfficeWFHSplitForUserIDs(memberIDs, effectiveLocations),
			MealTypeTotals: mealTotals,
		})
	}

	sort.Slice(teamTotals, func(i, j int) bool {
		return strings.ToLower(teamTotals[i].TeamName) < strings.ToLower(teamTotals[j].TeamName)
	})

	return teamTotals, assignedUsers, nil
}

func buildOfficeWFHSplitForUserIDs(userIDs []string, effectiveLocations map[string]models.WorkLocation) OfficeWFHSplit {
	split := OfficeWFHSplit{}
	for _, userID := range userIDs {
		location, exists := effectiveLocations[userID]
		if !exists {
			location = models.WorkLocationOffice
		}

		if location == models.WorkLocationWFH {
			split.WFH++
		} else {
			split.Office++
		}
	}
	return split
}

func buildSpecialDayNote(dayStatus models.DayStatus, schedule *models.DaySchedule) string {
	baseNote := ""
	switch dayStatus {
	case models.DayStatusOfficeClosed:
		baseNote = "Office is closed for this date."
	case models.DayStatusGovtHoliday:
		baseNote = "Government holiday."
	case models.DayStatusCelebration:
		baseNote = "Celebration day."
	case models.DayStatusWeekend:
		baseNote = "Weekend schedule applies."
	default:
		return ""
	}

	if schedule != nil && schedule.Reason != nil {
		reason := strings.TrimSpace(*schedule.Reason)
		if reason != "" {
			return fmt.Sprintf("%s Reason: %s", baseNote, reason)
		}
	}

	return baseNote
}

func buildAnnouncementMealTotals(meals map[string]MealHeadcount) []AnnouncementMealTotal {
	if len(meals) == 0 {
		return []AnnouncementMealTotal{}
	}

	preferredOrder := []string{"lunch", "snacks", "iftar", "event_dinner", "optional_dinner"}
	orderedKeys := make([]string, 0, len(meals))
	seen := make(map[string]bool, len(meals))

	for _, key := range preferredOrder {
		if _, ok := meals[key]; ok {
			orderedKeys = append(orderedKeys, key)
			seen[key] = true
		}
	}

	var extraKeys []string
	for key := range meals {
		if !seen[key] {
			extraKeys = append(extraKeys, key)
		}
	}
	sort.Strings(extraKeys)
	orderedKeys = append(orderedKeys, extraKeys...)

	totals := make([]AnnouncementMealTotal, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		meal := meals[key]
		totals = append(totals, AnnouncementMealTotal{
			MealType:      key,
			Participating: meal.Participating,
			OptedOut:      meal.OptedOut,
		})
	}

	return totals
}

func buildAnnouncementMessage(summary *DailyHeadcountSummary, specialDayNote string, totals []AnnouncementMealTotal) string {
	lines := []string{
		fmt.Sprintf("Daily Meal Participation Update - %s", summary.Date),
		fmt.Sprintf("Total Active Employees: %d", summary.TotalActiveUsers),
	}

	if specialDayNote != "" {
		lines = append(lines, fmt.Sprintf("Special Day Note: %s", specialDayNote))
	}

	lines = append(lines, "")
	lines = append(lines, "Meal-wise totals:")

	if len(totals) == 0 {
		lines = append(lines, "- No meals are scheduled for this date.")
	} else {
		for _, meal := range totals {
			lines = append(lines, fmt.Sprintf("- %s: %d confirmed, %d opted out",
				formatMealTypeForAnnouncement(meal.MealType), meal.Participating, meal.OptedOut))
		}
	}

	return strings.Join(lines, "\n")
}

func formatMealTypeForAnnouncement(mealType string) string {
	replaced := strings.ReplaceAll(mealType, "_", " ")
	parts := strings.Fields(replaced)
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}
