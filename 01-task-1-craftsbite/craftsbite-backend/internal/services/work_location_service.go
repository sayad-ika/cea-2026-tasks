package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// WorkLocationService defines business logic for work location management
type WorkLocationService interface {
	GetMyLocation(userID, date string) (*WorkLocationEffectiveResponse, error)
	SetMyLocation(userID string, input SetMyLocationInput) error
	OverrideUserLocation(requesterID string, input OverrideUserLocationInput) error
	GetTeamLocations(teamLeadID, date string) (*TeamWorkLocationResponse, error)
	CreateGlobalWFHPolicy(requesterID string, input CreateGlobalWFHPolicyInput) (*models.GlobalWorkLocationPolicy, error)
	GetGlobalWFHPolicies(startDate, endDate string) ([]models.GlobalWorkLocationPolicy, error)
	DeleteGlobalWFHPolicy(requesterID, policyID string) error
}

// SetMyLocationInput represents input for setting own work location
type SetMyLocationInput struct {
	Date     string              `json:"date" binding:"required"`
	Location models.WorkLocation `json:"location" binding:"required"`
}

// OverrideUserLocationInput represents input for overriding another user's work location
type OverrideUserLocationInput struct {
	UserID   string              `json:"user_id" binding:"required"`
	Date     string              `json:"date" binding:"required"`
	Location models.WorkLocation `json:"location" binding:"required"`
	Reason   string              `json:"reason" binding:"required"`
}

// CreateGlobalWFHPolicyInput represents input for creating a global WFH policy
type CreateGlobalWFHPolicyInput struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason"`
}

// WorkLocationEffectiveResponse represents a user's effective work location for a date
type WorkLocationEffectiveResponse struct {
	UserID   string              `json:"user_id"`
	Date     string              `json:"date"`
	Location models.WorkLocation `json:"location"`
	Source   string              `json:"source"`
}

// TeamWorkLocationResponse represents team members' effective work locations for a date
type TeamWorkLocationResponse struct {
	Date  string             `json:"date"`
	Teams []TeamWorkLocation `json:"teams"`
}

// TeamWorkLocation represents work location view for a single team
type TeamWorkLocation struct {
	TeamID   string               `json:"team_id"`
	TeamName string               `json:"team_name"`
	Members  []MemberWorkLocation `json:"members"`
}

// MemberWorkLocation represents a team member's effective location
type MemberWorkLocation struct {
	UserID   string              `json:"user_id"`
	Name     string              `json:"name"`
	Email    string              `json:"email"`
	Location models.WorkLocation `json:"location"`
	Source   string              `json:"source"`
}

type workLocationService struct {
	workLocationRepo repository.WorkLocationRepository
	userRepo         repository.UserRepository
	teamRepo         repository.TeamRepository
}

// NewWorkLocationService creates a new work location service
func NewWorkLocationService(
	workLocationRepo repository.WorkLocationRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
) WorkLocationService {
	return &workLocationService{
		workLocationRepo: workLocationRepo,
		userRepo:         userRepo,
		teamRepo:         teamRepo,
	}
}

// GetMyLocation gets effective location for the authenticated user for a specific date
func (s *workLocationService) GetMyLocation(userID, date string) (*WorkLocationEffectiveResponse, error) {
	if err := validateDate(date); err != nil {
		return nil, err
	}

	location, source, err := s.resolveEffectiveLocation(userID, date)
	if err != nil {
		return nil, err
	}

	return &WorkLocationEffectiveResponse{
		UserID:   userID,
		Date:     date,
		Location: location,
		Source:   source,
	}, nil
}

// SetMyLocation sets explicit work location for the authenticated user
func (s *workLocationService) SetMyLocation(userID string, input SetMyLocationInput) error {
	if err := validateDate(input.Date); err != nil {
		return err
	}
	if !input.Location.IsValid() {
		return fmt.Errorf("invalid location: must be one of office, wfh")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	existing, err := s.workLocationRepo.FindStatusByUserAndDate(userID, input.Date)
	if err != nil {
		return err
	}

	status := &models.WorkLocationStatus{
		ID:        uuid.New(),
		UserID:    userUUID,
		Date:      input.Date,
		Location:  input.Location,
		UpdatedBy: &userUUID,
	}
	if existing != nil {
		status.ID = existing.ID
	}

	if err := s.workLocationRepo.UpsertStatus(status); err != nil {
		return err
	}

	var previous *models.WorkLocation
	if existing != nil {
		prev := existing.Location
		previous = &prev
	}

	history := &models.WorkLocationHistory{
		ID:               uuid.New(),
		UserID:           userUUID,
		Date:             input.Date,
		PreviousLocation: previous,
		NewLocation:      input.Location,
		ChangedByUserID:  &userUUID,
		Action:           models.WorkLocationHistoryActionSelfSet,
	}
	if err := s.workLocationRepo.CreateHistory(history); err != nil {
		return err
	}

	return nil
}

// OverrideUserLocation sets explicit work location for another user by an authorized requester
func (s *workLocationService) OverrideUserLocation(requesterID string, input OverrideUserLocationInput) error {
	if err := validateDate(input.Date); err != nil {
		return err
	}
	if !input.Location.IsValid() {
		return fmt.Errorf("invalid location: must be one of office, wfh")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("reason is required for override")
	}

	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return fmt.Errorf("failed to find requester: %w", err)
	}

	action := models.WorkLocationHistoryActionAdminOverride
	switch requester.Role {
	case models.RoleTeamLead:
		isMember, err := s.teamRepo.IsUserInAnyTeamLedBy(requesterID, input.UserID)
		if err != nil {
			return fmt.Errorf("failed to validate team membership: %w", err)
		}
		if !isMember {
			return fmt.Errorf("team lead can only update work location for their own team members")
		}
		action = models.WorkLocationHistoryActionLeadOverride
	case models.RoleAdmin:
		action = models.WorkLocationHistoryActionAdminOverride
	default:
		return fmt.Errorf("insufficient permissions for override")
	}

	// Validate target user exists
	if _, err := s.userRepo.FindByID(input.UserID); err != nil {
		return fmt.Errorf("failed to find target user: %w", err)
	}

	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return fmt.Errorf("invalid requester ID: %w", err)
	}
	targetUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return fmt.Errorf("invalid target user ID: %w", err)
	}

	existing, err := s.workLocationRepo.FindStatusByUserAndDate(input.UserID, input.Date)
	if err != nil {
		return err
	}

	reason := strings.TrimSpace(input.Reason)
	status := &models.WorkLocationStatus{
		ID:        uuid.New(),
		UserID:    targetUUID,
		Date:      input.Date,
		Location:  input.Location,
		UpdatedBy: &requesterUUID,
		Reason:    &reason,
	}
	if existing != nil {
		status.ID = existing.ID
	}

	if err := s.workLocationRepo.UpsertStatus(status); err != nil {
		return err
	}

	var previous *models.WorkLocation
	if existing != nil {
		prev := existing.Location
		previous = &prev
	}

	history := &models.WorkLocationHistory{
		ID:               uuid.New(),
		UserID:           targetUUID,
		Date:             input.Date,
		PreviousLocation: previous,
		NewLocation:      input.Location,
		ChangedByUserID:  &requesterUUID,
		Reason:           &reason,
		Action:           action,
	}
	if err := s.workLocationRepo.CreateHistory(history); err != nil {
		return err
	}

	return nil
}

// GetTeamLocations returns effective locations for members of all teams led by the requester
func (s *workLocationService) GetTeamLocations(teamLeadID, date string) (*TeamWorkLocationResponse, error) {
	if err := validateDate(date); err != nil {
		return nil, err
	}

	requester, err := s.userRepo.FindByID(teamLeadID)
	if err != nil {
		return nil, fmt.Errorf("failed to find requester: %w", err)
	}
	if requester.Role != models.RoleTeamLead {
		return nil, fmt.Errorf("user is not a team lead")
	}

	teams, err := s.teamRepo.FindByTeamLeadID(teamLeadID)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams: %w", err)
	}

	var userIDs []string
	seen := make(map[string]bool)
	for _, team := range teams {
		for _, member := range team.Members {
			memberID := member.ID.String()
			if seen[memberID] {
				continue
			}
			seen[memberID] = true
			userIDs = append(userIDs, memberID)
		}
	}

	explicitStatuses, err := s.workLocationRepo.FindStatusesByDateForUsers(date, userIDs)
	if err != nil {
		return nil, err
	}
	statusByUser := make(map[string]models.WorkLocationStatus, len(explicitStatuses))
	for _, status := range explicitStatuses {
		statusByUser[status.UserID.String()] = status
	}

	globalPolicy, err := s.workLocationRepo.FindActiveGlobalPolicyByDate(date)
	if err != nil {
		return nil, err
	}

	var teamViews []TeamWorkLocation
	for _, team := range teams {
		var members []MemberWorkLocation
		for _, member := range team.Members {
			location := models.WorkLocationOffice
			source := "default"

			if globalPolicy != nil {
				location = models.WorkLocationWFH
				source = "global_policy"
			} else if explicit, ok := statusByUser[member.ID.String()]; ok {
				location = explicit.Location
				source = "explicit"
			}

			members = append(members, MemberWorkLocation{
				UserID:   member.ID.String(),
				Name:     member.Name,
				Email:    member.Email,
				Location: location,
				Source:   source,
			})
		}

		teamViews = append(teamViews, TeamWorkLocation{
			TeamID:   team.ID.String(),
			TeamName: team.Name,
			Members:  members,
		})
	}

	return &TeamWorkLocationResponse{
		Date:  date,
		Teams: teamViews,
	}, nil
}

// CreateGlobalWFHPolicy creates a global WFH date-range policy (Admin/Logistics)
func (s *workLocationService) CreateGlobalWFHPolicy(requesterID string, input CreateGlobalWFHPolicyInput) (*models.GlobalWorkLocationPolicy, error) {
	start, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}

	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to find requester: %w", err)
	}
	if requester.Role != models.RoleAdmin && requester.Role != models.RoleLogistics {
		return nil, fmt.Errorf("only admin and logistics can declare global WFH policy")
	}

	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return nil, fmt.Errorf("invalid requester ID: %w", err)
	}

	var reasonPtr *string
	reason := strings.TrimSpace(input.Reason)
	if reason != "" {
		reasonPtr = &reason
	}

	policy := &models.GlobalWorkLocationPolicy{
		ID:         uuid.New(),
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		Location:   models.WorkLocationWFH,
		IsActive:   true,
		Reason:     reasonPtr,
		DeclaredBy: requesterUUID,
	}

	if err := s.workLocationRepo.CreateGlobalPolicy(policy); err != nil {
		return nil, err
	}

	// Best-effort audit marker for policy creation
	history := &models.WorkLocationHistory{
		ID:              uuid.New(),
		UserID:          requesterUUID,
		Date:            input.StartDate,
		NewLocation:     models.WorkLocationWFH,
		ChangedByUserID: &requesterUUID,
		Reason:          reasonPtr,
		Action:          models.WorkLocationHistoryActionGlobalPolicyCreated,
	}
	if err := s.workLocationRepo.CreateHistory(history); err != nil {
		fmt.Printf("Warning: failed to record global policy creation in history: %v\n", err)
	}

	return policy, nil
}

// GetGlobalWFHPolicies returns global WFH policies with optional date-range overlap filter
func (s *workLocationService) GetGlobalWFHPolicies(startDate, endDate string) ([]models.GlobalWorkLocationPolicy, error) {
	if (startDate == "" && endDate != "") || (startDate != "" && endDate == "") {
		return nil, fmt.Errorf("start_date and end_date must be provided together")
	}

	if startDate != "" && endDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
		}
		if end.Before(start) {
			return nil, fmt.Errorf("end_date must be on or after start_date")
		}
	}

	return s.workLocationRepo.FindGlobalPolicies(startDate, endDate)
}

// DeleteGlobalWFHPolicy deactivates a global WFH policy (Admin/Logistics)
func (s *workLocationService) DeleteGlobalWFHPolicy(requesterID, policyID string) error {
	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return fmt.Errorf("failed to find requester: %w", err)
	}
	if requester.Role != models.RoleAdmin && requester.Role != models.RoleLogistics {
		return fmt.Errorf("only admin and logistics can remove global WFH policy")
	}

	policy, err := s.workLocationRepo.FindGlobalPolicyByID(policyID)
	if err != nil {
		return err
	}
	if policy == nil {
		return fmt.Errorf("global work location policy not found")
	}

	if err := s.workLocationRepo.DeactivateGlobalPolicy(policyID); err != nil {
		return err
	}

	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return fmt.Errorf("invalid requester ID: %w", err)
	}

	history := &models.WorkLocationHistory{
		ID:              uuid.New(),
		UserID:          requesterUUID,
		Date:            policy.StartDate,
		NewLocation:     models.WorkLocationWFH,
		ChangedByUserID: &requesterUUID,
		Reason:          policy.Reason,
		Action:          models.WorkLocationHistoryActionGlobalPolicyRemoved,
	}
	if err := s.workLocationRepo.CreateHistory(history); err != nil {
		fmt.Printf("Warning: failed to record global policy removal in history: %v\n", err)
	}

	return nil
}

func (s *workLocationService) resolveEffectiveLocation(userID, date string) (models.WorkLocation, string, error) {
	status, err := s.workLocationRepo.FindStatusByUserAndDate(userID, date)
	if err != nil {
		return "", "", err
	}
	if status != nil {
		return status.Location, "explicit", nil
	}
	
	policy, err := s.workLocationRepo.FindActiveGlobalPolicyByDate(date)
	if err != nil {
		return "", "", err
	}
	if policy != nil {
		return models.WorkLocationWFH, "global_policy", nil
	}

	return models.WorkLocationOffice, "default", nil
}

func validateDate(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	return nil
}
