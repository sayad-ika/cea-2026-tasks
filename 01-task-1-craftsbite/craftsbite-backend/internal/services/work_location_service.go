package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type WorkLocationService interface {
	SetMyLocation(userID, date, location string) error
	GetMyLocation(userID, date string) (*WorkLocationResponse, error)
	SetLocationFor(requesterID, targetUserID, date, location string, reason *string) error
	ListByDate(requesterID, date string) ([]WorkLocationResponse, error)
}

type WorkLocationResponse struct {
    UserID   string  `json:"user_id"`
    Date     string  `json:"date"`
    Location string  `json:"location"`
    SetBy    string  `json:"set_by,omitempty"`
    Reason   *string `json:"reason,omitempty"`
}

type workLocationService struct {
	repo     repository.WorkLocationRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

func NewWorkLocationService(
	repo repository.WorkLocationRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
) WorkLocationService {
	return &workLocationService{repo: repo, userRepo: userRepo, teamRepo: teamRepo}
}

func validateLocation(location string) error {
	if !models.WorkLocationType(location).IsValid() {
		return fmt.Errorf("location must be 'office' or 'wfh'")
	}
	return nil
}

func (s *workLocationService) SetMyLocation(userID, date, location string) error {
	if err := validateDate(date); err != nil {
		return err
	}
	if err := validateLocation(location); err != nil {
		return err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	wl := &models.WorkLocation{
		UserID:   userUUID,
		Date:     date,
		Location: models.WorkLocationType(location),
		SetBy:    nil,
	}
	return s.repo.Upsert(wl)
}

func (s *workLocationService) GetMyLocation(userID, date string) (*WorkLocationResponse, error) {
	if err := validateDate(date); err != nil {
		return nil, err
	}

	wl, err := s.repo.FindByUserAndDate(userID, date)
	if err != nil {
		return nil, err
	}
	if wl == nil {
		return &WorkLocationResponse{UserID: userID, Date: date, Location: "not_set"}, nil
	}

	resp := &WorkLocationResponse{UserID: userID, Date: date, Location: string(wl.Location)}
	if wl.SetBy != nil {
		resp.SetBy = wl.SetBy.String()
	}
	
	resp.Reason = wl.Reason
	return resp, nil
}

func (s *workLocationService) SetLocationFor(requesterID, targetUserID, date, location string, reason *string) error {
	if err := validateDate(date); err != nil {
		return err
	}
	if err := validateLocation(location); err != nil {
		return err
	}

	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return fmt.Errorf("requester not found")
	}

	if requester.Role == models.RoleTeamLead {
		isMember, err := s.teamRepo.IsUserInAnyTeamLedBy(requesterID, targetUserID)
		if err != nil {
			return fmt.Errorf("failed to verify team membership: %w", err)
		}
		if !isMember {
			return fmt.Errorf("you can only set work location for your own team members")
		}
	}

	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return fmt.Errorf("invalid target user ID")
	}
	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return fmt.Errorf("invalid requester ID")
	}

	wl := &models.WorkLocation{
		UserID:   targetUUID,
		Date:     date,
		Location: models.WorkLocationType(location),
		SetBy:    &requesterUUID,
	}
	return s.repo.Upsert(wl)
}

func (s *workLocationService) ListByDate(requesterID, date string) ([]WorkLocationResponse, error) {
    if err := validateDate(date); err != nil {
        return nil, err
    }

    requester, err := s.userRepo.FindByID(requesterID)
    if err != nil {
        return nil, fmt.Errorf("requester not found")
    }

    var wls []models.WorkLocation

    if requester.Role == models.RoleTeamLead {
        teams, err := s.teamRepo.FindByTeamLeadID(requesterID)
        if err != nil {
            return nil, fmt.Errorf("failed to load teams: %w", err)
        }
        var memberIDs []string
        for _, team := range teams {
            for _, member := range team.Members {
                memberIDs = append(memberIDs, member.ID.String())
            }
        }
        wls, err = s.repo.FindByDateAndUserIDs(date, memberIDs)
        if err != nil {
            return nil, err
        }
    } else {
        wls, err = s.repo.FindByDate(date)
        if err != nil {
            return nil, err
        }
    }

    var result []WorkLocationResponse
    for _, wl := range wls {
		item := WorkLocationResponse{
		    UserID:   wl.UserID.String(),
		    Date:     wl.Date,
		    Location: string(wl.Location),
		    Reason:   wl.Reason,
		}
		if wl.SetBy != nil {
		    item.SetBy = wl.SetBy.String()
		}
        result = append(result, item)
    }
    if result == nil {
        result = []WorkLocationResponse{}
    }
    return result, nil
}
